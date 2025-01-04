import sys, requests, json, base64, time
from flask import Flask, request, jsonify
from clickhouse_driver import Client
from loguru import logger
from functools import wraps
import utils
logger.remove()
logger.add(sys.stderr,format='<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> [<level>{level}</level>] <level>{message}</level>',level='INFO')



def retry_on_exception(retries=3, delay=1, backoff=2):
    def decorator_retry(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            attempt = 0
            while attempt < retries:
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    attempt += 1
                    logger.warning(f"尝试第 {attempt} 次遇到错误: {e}")
                    if attempt < retries:
                        time.sleep(delay * (backoff ** (attempt - 1)))
            raise Exception("达到最大重试次数，无数据可用")
        return wrapper
    return decorator_retry

@retry_on_exception()
def execute_query(query):
    return utils.ckclient.execute(query)


app = Flask(__name__)


@app.route('/health', methods=['GET'])
def health():
    return jsonify({"message": "Service is running"}), 200


def pass_webhook(uid):
    # deploytment不做任何操作，跳过webhook
    return jsonify(
        {
            "apiVersion": "admission.k8s.io/v1", 
            "kind": "AdmissionReview", 
            "response": {
                "uid": uid, 
                "allowed": True
            }
        }
    )


def connect_fail(uid):
    return jsonify(
        {
            "apiVersion": "admission.k8s.io/v1",
            "kind": "AdmissionReview",
            "response": {
                "uid": uid,
                "allowed": False,
                "status": {"code": 500, "message": f"数据库查询失败"},
            },
        }
    )

def scale_only(uid, replicas):
    # 仅修改副本数，不重启
    patch_replicas = {"op": "replace", "path": "/spec/replicas", "value": replicas}  # 设置副本数
    code = base64.b64encode(json.dumps([patch_replicas]).encode()).decode()
    return jsonify(
        {
            "apiVersion": "admission.k8s.io/v1",
            "kind": "AdmissionReview",
            "response": {"uid": uid, "allowed": True, "patchType": "JSONPatch", "patch": code},
        }
    )


def update_all(
    replicas, namespace, deployment_name, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb, resources_dict, uid
):
    # 按照数据库修改所有参数
    patch_replicas = {"op": "replace", "path": "/spec/replicas", "value": replicas}  # 设置副本数
    # request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb，有则改，无则不动
    logger.info(f"改前：{resources_dict}")
    if request_cpu_m > 0:
        resources_dict["requests"]["cpu"] = f"{request_cpu_m}m"
    else:
        utils.send_wecom(f"webhook:【prod】【{namespace}】服务【{deployment_name}】未配置request_cpu_m")
    if request_mem_mb > 0:
        resources_dict["requests"]["memory"] = f"{request_mem_mb}Mi"
    else:
        utils.send_wecom(f"webhook:【prod】【{namespace}】服务【{deployment_name}】未配置request_mem_mb")
    if limit_cpu_m > 0:
        resources_dict["limits"]["cpu"] = f"{limit_cpu_m}m"
    else:
        utils.send_wecom(f"webhook:【prod】【{namespace}】服务【{deployment_name}】未配置limit_cpu_m")
    if limit_mem_mb > 0:
        resources_dict["limits"]["memory"] = f"{limit_mem_mb}Mi"
    else:
        utils.send_wecom(f"webhook:【prod】【{namespace}】服务【{deployment_name}】未配置limit_mem_mb")
    logger.info(f"改后：{resources_dict}")
    resources = {"op": "add", "path": "/spec/template/spec/containers/0/resources", "value": resources_dict}
    code = base64.b64encode(json.dumps([patch_replicas, resources]).encode()).decode()
    return jsonify(
        {
            "apiVersion": "admission.k8s.io/v1",
            "kind": "AdmissionReview",
            "response": {"uid": uid, "allowed": True, "patchType": "JSONPatch", "patch": code},
        }
    )


@app.route('/mutate', methods=['POST'])
def mutate():
    request_info = request.get_json()
    object = request_info['request']['object']
    old_object = request_info['request']['oldObject']
    kind = request_info['request']['kind']['kind']
    operation = request_info['request']['operation']
    uid = request_info['request']['uid']
    namespace = request_info['request']['object']['metadata']['namespace']
    deployment_name = request_info['request']['object']['metadata']['name']

    # 按命名空间区分是否需要更新
    if namespace not in utils.NS_LIST:
        content = f"【{namespace}】【{deployment_name}】非业务命名空间，跳过"
        logger.info(content, flush=True)
        return pass_webhook(uid)

    # 查询 ClickHouse 获取 CPU 和内存限制
    query = (
        f"SELECT pod_count, pod_count_ai, pod_count_manual, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb "
        f"FROM kubedoor.k8s_res_control "
        f"WHERE namespace='{namespace}' "
        f"AND deployment='{deployment_name}'"
    )
    logger.info(query, flush=True)
    try:
        result = execute_query(query)
    except Exception as e:
        content = f"webhook:【prod】【{namespace}】服务【{deployment_name}】查询数据库失败：{e}"
        logger.error(content)
        utils.send_wecom(content)
        return connect_fail(uid)

    pod_count, pod_count_ai, pod_count_manual, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb = (
        result[0] if result else (-1, -1, -1, -1, -1, -1, -1)
    )
    # 副本数取值优先级：专家建议→ai建议→原本的副本数
    if pod_count_manual >= 0:
        replicas = pod_count_manual
    elif pod_count_ai >= 0:
        replicas = pod_count_ai
    else:
        replicas = pod_count

    # 如果数据库中request_cpu_m为0，设置为30；如果request_mem_mb为0，设置为1
    if 0 <= request_cpu_m < 30:
        logger.info("request_cpu_m为0，设置为30", flush=True)
        request_cpu_m = 30
    if request_mem_mb == 0:
        logger.info("request_mem_mb为0，设置为1", flush=True)
        request_mem_mb = 1

    logger.info(
        f"副本数:{replicas}, 请求CPU:{request_cpu_m}m, 请求内存:{request_mem_mb}mb, 限制CPU:{limit_cpu_m}m, 限制内存:{limit_mem_mb}mb",
        flush=True,
    )

    try:
        if not result:
            # 服务未在表中的，部署失败
            content = f"webhook:【prod】【{namespace}】服务【{deployment_name}】部署失败:服务不在表里"
            logger.info(content, flush=True)
            utils.send_wecom(content)
            return jsonify(
                {
                    "apiVersion": "admission.k8s.io/v1",
                    "kind": "AdmissionReview",
                    "response": {
                        "uid": uid,
                        "allowed": False,
                        "status": {"code": 400, "message": "部署失败:未从数据库中查到该服务"},
                    },
                }
            )
        elif kind == 'Scale' and operation == 'UPDATE':
            # 把spec.replicas修改成数据库的pod数一致
            logger.info(f"webhook:【prod】【{namespace}】服务【{deployment_name}】收到scale请求，仅修改replicas", flush=True)
            return scale_only(uid, replicas)
        elif kind == 'Deployment' and operation == 'CREATE':
            # 按照数据库修改所有参数
            logger.info(f"webhook:【prod】【{namespace}】服务【{deployment_name}】收到create请求，修改所有参数", flush=True)
            resources_dict = request_info['request']['object']['spec']['template']['spec']['containers'][0]['resources']
            if not resources_dict:
                resources_dict = {"requests": {}, "limits": {}}
            return update_all(
                replicas, namespace, deployment_name, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb, resources_dict, uid
            )
        elif kind == 'Deployment' and operation == 'UPDATE':
            template = object['spec']['template']
            old_template = old_object['spec']['template']
            if object == old_object:
                logger.info(f"webhook:【prod】【{namespace}】服务【{deployment_name}】object和oldObject相等，跳过", flush=True)
                return pass_webhook(uid)
            elif template != old_template:
                # spec.template 变了,触发重启逻辑,按照数据库修改所有参数
                logger.info(
                    f"webhook:【prod】【{namespace}】服务【{deployment_name}】收到update请求，修改所有参数",
                    flush=True,
                )
                resources_dict = request_info['request']['object']['spec']['template']['spec']['containers'][0]['resources']
                if not resources_dict:
                    resources_dict = {"requests": {}, "limits": {}}
                return update_all(
                    replicas, namespace, deployment_name, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb, resources_dict, uid
                )
            elif template == old_template and replicas != object['spec']['replicas']:
                # spec.template 没变,spec.replicas 变了,只把修改spec.replicas和数据库的pod数一致
                logger.info(f"webhook:【prod】【{namespace}】服务【{deployment_name}】收到update请求，但仅修改replicas", flush=True)
                return scale_only(uid, replicas)
            elif template == old_template and replicas == object['spec']['replicas']:
                # spec.template 没变,spec.replicas 没变,什么也不做
                logger.info(f"webhook:【prod】【{namespace}】服务【{deployment_name}】template和replicas没变，不处理", flush=True)
                return pass_webhook(uid)
        else:
            content = f"webhook:【prod】【{namespace}】服务【{deployment_name}】不符合预设情况，未做任何处理"
            logger.info(content, flush=True)
            utils.send_wecom(content)
            return pass_webhook(uid)
    except Exception as e:
        logger.error(f"【{namespace}】服务【{deployment_name}】Webhook 处理错误：")
        logger.exception(e)
        content = f"webhook:【prod】【{namespace}】服务【{deployment_name}】处理错误：{e}"
        utils.send_wecom(content)


if __name__ == '__main__':
    context = ('/serving-certs/tls.crt', '/serving-certs/tls.key')
    app.run(host='0.0.0.0', ssl_context=context, port=443, debug=True)

