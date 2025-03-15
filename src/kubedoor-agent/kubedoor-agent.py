import asyncio
import base64
import json
import sys
from urllib.parse import urlencode
from aiohttp import ClientSession, ClientWebSocketResponse, WSMsgType, web
from kubernetes_asyncio import client, config
from kubernetes_asyncio.client.rest import ApiException
from datetime import datetime
from loguru import logger
import utils

# 配置日志
logger.remove()
logger.add(
    sys.stderr,
    format='<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> [<level>{level}</level>] <level>{message}</level>',
    level='INFO',
)

VERSION = utils.get_version()
# 全局变量
ws_conn = None
v1 = None  # AppsV1Api
batch_v1 = None  # BatchV1Api
core_v1 = None  # CoreV1Api
admission_api = None  # AdmissionregistrationV1Api

# 用于存储 WebSocket 请求的 Future
request_futures = {}


def init_kubernetes():
    """在程序启动时加载 Kubernetes 配置并初始化客户端"""
    global v1, batch_v1, core_v1, admission_api
    try:
        config.load_incluster_config()
        v1 = client.AppsV1Api()
        batch_v1 = client.BatchV1Api()
        core_v1 = client.CoreV1Api()
        admission_api = client.AdmissionregistrationV1Api()
        logger.info("Kubernetes 配置加载成功")
    except Exception as e:
        logger.error(f"加载 Kubernetes 配置失败: {e}")
        raise


async def handle_http_request(ws: ClientWebSocketResponse, request_id: str, method: str, query: dict, body: dict, path: str):
    """异步处理 HTTP 请求并发送响应"""
    try:
        async with ClientSession() as session:
            logger.info(f"转发请求: {method} {path}?{urlencode(query)}【{json.dumps(body)}】")
            if method == "GET":
                async with session.get(path, params=query, ssl=False) as resp:
                    response_data = await resp.json()
            elif method == "POST":
                async with session.post(path, params=query, json=body, ssl=False) as resp:
                    response_data = await resp.json()
            else:
                response_data = {"error": f"Unsupported method: {method}"}
    except Exception as e:
        response_data = {"error": str(e)}

    await ws.send_json({"type": "response", "request_id": request_id, "response": response_data})


async def process_request(ws: ClientWebSocketResponse):
    """处理服务端发送的请求"""
    async for msg in ws:
        if msg.type == WSMsgType.TEXT:
            try:
                data = json.loads(msg.data)
            except json.JSONDecodeError:
                logger.error(f"收到无法解析的消息：{msg.data}")
                continue
            if data.get("type") == "admis":
                request_id = data.get("request_id")
                deploy_res = data.get("deploy_res")
                logger.info(f"收到 admis 消息：{request_id} {deploy_res}")
                if request_id in request_futures:
                    request_futures[request_id].set_result(deploy_res)
                    del request_futures[request_id]
            elif data.get("type") == "request":
                request_id = data["request_id"]
                method = data["method"]
                query = data["query"]
                body = data["body"]
                path = (
                    'http://127.0.0.1:81' + data["path"]
                    if data["path"].startswith('/api/pod/')
                    else 'https://127.0.0.1' + data["path"]
                )
                asyncio.create_task(handle_http_request(ws, request_id, method, query, body, path))
        elif msg.type == WSMsgType.ERROR:
            logger.error(f"WebSocket 错误：{msg.data}")


async def heartbeat(ws: ClientWebSocketResponse):
    """定期发送心跳"""
    while True:
        try:
            await ws.send_json({"type": "heartbeat"})
            logger.debug("成功发送心跳")
            await asyncio.sleep(4)
        except Exception as e:
            logger.error(f"心跳发送失败：{e}")
            break


async def connect_to_server():
    """连接到 WebSocket 服务端，并处理连接断开的情况"""
    uri = f"{utils.KUBEDOOR_MASTER}/ws?env={utils.PROM_K8S_TAG_VALUE}&ver={VERSION}"
    while True:
        try:
            async with ClientSession() as session:
                async with session.ws_connect(uri) as ws:
                    logger.info("成功连接到服务端")
                    global ws_conn
                    ws_conn = ws
                    await asyncio.gather(process_request(ws), heartbeat(ws))
        except Exception as e:
            logger.error(f"连接到服务端失败：{e}")
            logger.info("等待 5 秒后重新连接...")
            await asyncio.sleep(5)


async def delete_cronjob_or_not(cronjob_name, job_type):
    """判断是否是一次性 job，是的话删除"""
    if job_type == "once":
        try:
            await batch_v1.delete_namespaced_cron_job(name=cronjob_name, namespace="kubedoor", body=client.V1DeleteOptions())
            logger.info(f"CronJob '{cronjob_name}' deleted successfully.")
        except ApiException as e:
            logger.exception(f"删除 CronJob '{cronjob_name}' 时出错: {e}")
            utils.send_msg(f"Error when deleting CronJob '{cronjob_name}'!")


async def health_check(request):
    return web.json_response({"ver": VERSION, "status": "healthy"})


async def update_image(request):
    try:
        data = await request.json()
        new_image_tag = data.get('image_tag')
        deployment_name = data.get('deployment')
        namespace = data.get('namespace')

        deployment = await v1.read_namespaced_deployment(deployment_name, namespace)
        current_image = deployment.spec.template.spec.containers[0].image
        image_name = current_image.split(':')[0]
        new_image = f"{image_name}:{new_image_tag}"
        deployment.spec.template.spec.containers[0].image = new_image

        await v1.patch_namespaced_deployment(name=deployment_name, namespace=namespace, body=deployment)
        return web.json_response({"success": True, "message": f"{namespace} {deployment_name} updated with image {new_image}"})
    except Exception as e:
        return web.json_response({"message": str(e)}, status=500)


async def scale(request):
    """批量扩缩容"""
    request_info = await request.json()
    interval = request.query.get("interval")
    error_list = []

    for index, deployment in enumerate(request_info):
        namespace = deployment.get("namespace")
        deployment_name = deployment.get("deployment_name")
        num = deployment.get("num")
        job_name = deployment.get("job_name")
        job_type = deployment.get("job_type")
        logger.info(f"【{namespace}】【{deployment_name}】: {num}")

        try:
            deployment_obj = await v1.read_namespaced_deployment(deployment_name, namespace)
            if not deployment_obj:
                logger.error(f"未找到【{namespace}】【{deployment_name}】")
                error_list.append({'namespace': namespace, 'deployment_name': deployment_name})
                continue

            deployment_obj.spec.replicas = num
            logger.info(f"Deployment【{deployment_name}】副本数更改为 {num}，如已接入准入控制, 实际变更已数据库中数据为准。")
            await v1.patch_namespaced_deployment_scale(deployment_name, namespace, deployment_obj)

            if interval and index != len(request_info) - 1:
                logger.info(f"暂停 {interval}s...")
                await asyncio.sleep(int(interval))

            if job_name:
                utils.send_msg(f"'{deployment_name}' has been scaled!")
                await delete_cronjob_or_not(job_name, job_type)
        except ApiException as e:
            logger.exception(f"调用 AppsV1Api 时出错: {e}")
            error_list.append({'namespace': namespace, 'deployment_name': deployment_name})

    return web.json_response({"message": "ok", "error_list": error_list})


async def reboot(request):
    """批量重启微服务"""
    request_info = await request.json()
    interval = request.query.get("interval")
    patch = {
        "spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": datetime.now().isoformat()}}}}
    }
    error_list = []

    for index, deployment in enumerate(request_info):
        namespace = deployment.get("namespace")
        deployment_name = deployment.get("deployment_name")
        job_name = deployment.get("job_name")
        job_type = deployment.get("job_type")
        logger.info(f"【{namespace}】【{deployment_name}】")

        try:
            deployment_obj = await v1.read_namespaced_deployment(deployment_name, namespace)
            if not deployment_obj:
                logger.error(f"未找到【{namespace}】【{deployment_name}】")
                error_list.append({'namespace': namespace, 'deployment_name': deployment_name})
                continue

            logger.info(f"重启 Deployment【{deployment_name}】，如已接入准入控制, 实际变更已数据库中数据为准。")
            await v1.patch_namespaced_deployment(deployment_name, namespace, patch)

            if interval and index != len(request_info) - 1:
                logger.info(f"暂停 {interval}s...")
                await asyncio.sleep(int(interval))

            if job_name:
                utils.send_msg(f"'{deployment_name}' has been restarted!")
                await delete_cronjob_or_not(job_name, job_type)
        except ApiException as e:
            logger.exception(f"调用 AppsV1Api 时出错: {e}")
            error_list.append({'namespace': namespace, 'deployment_name': deployment_name})

    return web.json_response({"message": "ok", "error_list": error_list})


async def cron(request):
    """创建定时任务，执行扩缩容或重启"""
    request_info = await request.json()
    cron_expr = request_info.get("cron")
    time_expr = request_info.get("time")
    type_expr = request_info.get("type")
    service = request_info.get("service")

    deployment_name = service[0].get("deployment_name")
    name_pre = f"{type_expr}-{'once' if time_expr else 'cron'}-{deployment_name}"
    job_type = "once" if time_expr else "cron"
    cron_new = f"{time_expr[4]} {time_expr[3]} {time_expr[2]} {time_expr[1]} *" if time_expr else cron_expr
    service[0]["job_name"] = name_pre
    service[0]["job_type"] = job_type

    cronjob = client.V1CronJob(
        metadata=client.V1ObjectMeta(name=name_pre),
        spec=client.V1CronJobSpec(
            schedule=cron_new,
            job_template=client.V1JobTemplateSpec(
                spec=client.V1JobSpec(
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels={"app": name_pre}),
                        spec=client.V1PodSpec(
                            restart_policy="Never",
                            containers=[
                                client.V1Container(
                                    name=name_pre,
                                    image="registry.cn-shenzhen.aliyuncs.com/starsl/busybox-curl",
                                    command=[
                                        "curl",
                                        "-s",
                                        "-k",
                                        "-X",
                                        "POST",
                                        "-H",
                                        "Content-Type: application/json",
                                        "-d",
                                        f'{json.dumps(service)}',
                                        f"https://kubedoor-agent.kubedoor/api/{type_expr}",
                                    ],
                                    env=[client.V1EnvVar(name="CRONJOB_TYPE", value=job_type)],
                                )
                            ],
                        ),
                    )
                )
            ),
        ),
    )

    namespace = "kubedoor"
    try:
        await batch_v1.create_namespaced_cron_job(namespace=namespace, body=cronjob)
        content = f"CronJob '{name_pre}' created successfully in namespace '{namespace}'."
        logger.info(content)
        utils.send_msg(content)
        return web.json_response({"message": "ok"})
    except Exception as e:
        error_message = json.loads(e.body).get("message") if hasattr(e, 'body') and e.body else "执行失败"
        logger.error(error_message)
        return web.json_response({"message": error_message}, status=500)


async def update_namespace_label(ns, action):
    """命名空间修改标签"""
    namespace_name = ns
    label_key = "kubedoor-ignore"
    label_value = 'true' if action == "add" else None
    patch_body = {"metadata": {"labels": {label_key: label_value}}}

    try:
        response = await core_v1.patch_namespace(name=namespace_name, body=patch_body)
        logger.info(
            f"Label '{label_key}: {label_value}' {'added to' if action == 'add' else 'removed from'} namespace '{namespace_name}' successfully."
        )
        logger.info(f"Updated namespace labels: {response.metadata.labels}")
    except ApiException as e:
        logger.error(f"Exception when patching namespace '{namespace_name}': {e}")


async def get_mutating_webhook():
    """获取 MutatingWebhookConfiguration"""
    webhook_name = "kubedoor-admis-configuration"
    try:
        await admission_api.read_mutating_webhook_configuration(name=webhook_name)
        logger.info(f"MutatingWebhookConfiguration '{webhook_name}' exists.")
        return {"is_on": True}
    except ApiException as e:
        if e.status == 404:
            logger.error(f"MutatingWebhookConfiguration '{webhook_name}' does not exist.")
            return {"is_on": False}
        error_message = json.loads(e.body).get("message") if hasattr(e, 'body') and e.body else "执行失败"
        logger.error(f"Exception when reading MutatingWebhookConfiguration: {e}")
        return {"message": error_message}


async def create_mutating_webhook():
    """创建 MutatingWebhookConfiguration"""
    webhook_name = "kubedoor-admis-configuration"
    namespace = "kubedoor"
    webhook_config = client.V1MutatingWebhookConfiguration(
        metadata=client.V1ObjectMeta(name=webhook_name),
        webhooks=[
            client.V1MutatingWebhook(
                name="kubedoor-admis.mutating.webhook",
                client_config=client.AdmissionregistrationV1WebhookClientConfig(
                    service=client.AdmissionregistrationV1ServiceReference(
                        namespace=namespace, name="kubedoor-agent", path="/api/admis", port=443
                    ),
                    ca_bundle=utils.BASE64CA,
                ),
                rules=[
                    client.V1RuleWithOperations(
                        operations=["CREATE", "UPDATE"],
                        api_groups=["apps"],
                        api_versions=["v1"],
                        resources=["deployments", "deployments/scale"],
                        scope="*",
                    )
                ],
                failure_policy="Fail",
                match_policy="Equivalent",
                namespace_selector=client.V1LabelSelector(
                    match_expressions=[client.V1LabelSelectorRequirement(key="kubedoor-ignore", operator="DoesNotExist")]
                ),
                side_effects="None",
                timeout_seconds=10,
                admission_review_versions=["v1"],
                reinvocation_policy="Never",
            )
        ],
    )

    try:
        response = await admission_api.create_mutating_webhook_configuration(body=webhook_config)
        logger.info(f"MutatingWebhookConfiguration created: {response.metadata.name}")
        await update_namespace_label("kube-system", "add")
        await update_namespace_label("kubedoor", "add")
    except ApiException as e:
        error_message = json.loads(e.body).get("message") if hasattr(e, 'body') and e.body else "执行失败"
        logger.error(f"Exception when creating MutatingWebhookConfiguration: {e}")
        return {"message": error_message}


async def delete_mutating_webhook():
    """删除 MutatingWebhookConfiguration"""
    webhook_name = "kubedoor-admis-configuration"
    try:
        await admission_api.delete_mutating_webhook_configuration(name=webhook_name)
        logger.info(f"MutatingWebhookConfiguration '{webhook_name}' deleted successfully.")
        await update_namespace_label("kube-system", "delete")
        await update_namespace_label("kubedoor", "delete")
    except ApiException as e:
        error_message = json.loads(e.body).get("message") if hasattr(e, 'body') and e.body else "执行失败"
        logger.error(f"Exception when deleting MutatingWebhookConfiguration: {e}")
        return {"message": error_message}


async def admis_switch(request):
    action = request.query.get("action")
    res = await get_mutating_webhook()

    if action == "get":
        return web.json_response(res)
    elif action == "on":
        if res.get("is_on"):
            return web.json_response({"message": "Webhook is already opened!", "success": True})
        create_res = await create_mutating_webhook()
        if create_res:
            return web.json_response(create_res, status=500)
    elif action == "off":
        if not res.get("is_on"):
            return web.json_response({"message": "Webhook is already closed!", "success": True})
        delete_res = await delete_mutating_webhook()
        if delete_res:
            return web.json_response(delete_res, status=500)

    return web.json_response({"message": "执行成功", "success": True})


def admis_pass(uid):
    return {"apiVersion": "admission.k8s.io/v1", "kind": "AdmissionReview", "response": {"uid": uid, "allowed": True}}


def admis_fail(uid, code, message):
    return {
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {"uid": uid, "allowed": False, "status": {"code": code, "message": message}},
    }


def scale_only(uid, replicas):
    patch_replicas = {"op": "replace", "path": "/spec/replicas", "value": replicas}
    code = base64.b64encode(json.dumps([patch_replicas]).encode()).decode()
    return {
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {"uid": uid, "allowed": True, "patchType": "JSONPatch", "patch": code},
    }


def update_all(
    replicas, namespace, deployment_name, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb, resources_dict, uid
):
    # 按照数据库修改所有参数
    patch_replicas = {"op": "replace", "path": "/spec/replicas", "value": replicas}
    # request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb，有则改，无则不动
    logger.info(f"改前：{resources_dict}")
    if request_cpu_m > 0:
        resources_dict["requests"]["cpu"] = f"{request_cpu_m}m"
    else:
        utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】未配置 request_cpu_m")
    if request_mem_mb > 0:
        resources_dict["requests"]["memory"] = f"{request_mem_mb}Mi"
    else:
        utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】未配置 request_mem_mb")
    if limit_cpu_m > 0:
        resources_dict["limits"]["cpu"] = f"{limit_cpu_m}m"
    else:
        utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】未配置 limit_cpu_m")
    if limit_mem_mb > 0:
        resources_dict["limits"]["memory"] = f"{limit_mem_mb}Mi"
    else:
        utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】未配置 limit_mem_mb")
    logger.info(f"改后：{resources_dict}")
    resources = {"op": "add", "path": "/spec/template/spec/containers/0/resources", "value": resources_dict}
    code = base64.b64encode(json.dumps([patch_replicas, resources]).encode()).decode()
    return {
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {"uid": uid, "allowed": True, "patchType": "JSONPatch", "patch": code},
    }


async def admis_mutate(request):
    request_info = await request.json()
    object = request_info['request']['object']
    old_object = request_info['request']['oldObject']
    kind = request_info['request']['kind']['kind']
    operation = request_info['request']['operation']
    uid = request_info['request']['uid']
    namespace = object['metadata']['namespace']
    deployment_name = object['metadata']['name']
    logger.info(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】收到请求")

    if ws_conn is None or ws_conn.closed:
        return web.json_response(admis_fail(uid, 503, "连接 kubedoor-master 失败"))

    response_future = asyncio.get_event_loop().create_future()
    request_futures[uid] = response_future
    await ws_conn.send_json({"type": "admis", "request_id": uid, "namespace": namespace, "deployment": deployment_name})
    try:
        result = await asyncio.wait_for(response_future, timeout=10)
        logger.info(f"response_future 收到 admis 响应：{uid} {result}")
    except asyncio.TimeoutError:
        del request_futures[uid]
        return web.json_response(admis_fail(uid, 504, "等待 kubedoor-master 响应超时"))

    if len(result) == 2:
        if result[0] == 200:
            return web.json_response(admis_pass(uid))
        return web.json_response(admis_fail(uid, result[0], result[1]))

    pod_count, pod_count_ai, pod_count_manual, request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb = result
    # 副本数取值优先级：专家建议→ai建议→原本的副本数
    replicas = pod_count_manual if pod_count_manual >= 0 else (pod_count_ai if pod_count_ai >= 0 else pod_count)
    # 如果数据库中request_cpu_m为0，设置为30；如果request_mem_mb为0，设置为1
    request_cpu_m = 30 if 0 <= request_cpu_m < 30 else request_cpu_m
    request_mem_mb = 1 if request_mem_mb == 0 else request_mem_mb

    logger.info(
        f"副本数:{replicas}, 请求CPU:{request_cpu_m}m, 请求内存:{request_mem_mb}MB, 限制CPU:{limit_cpu_m}m, 限制内存:{limit_mem_mb}MB"
    )

    try:
        if kind == 'Scale' and operation == 'UPDATE':
            # 把spec.replicas修改成数据库的pod数一致
            logger.info(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】收到scale请求，仅修改replicas")
            return web.json_response(scale_only(uid, replicas))
        elif kind == 'Deployment' and operation == 'CREATE':
            # 按照数据库修改所有参数
            logger.info(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】收到 create 请求，修改所有参数")
            resources_dict = object['spec']['template']['spec']['containers'][0].get('resources', {"requests": {}, "limits": {}})
            return web.json_response(
                update_all(
                    replicas,
                    namespace,
                    deployment_name,
                    request_cpu_m,
                    request_mem_mb,
                    limit_cpu_m,
                    limit_mem_mb,
                    resources_dict,
                    uid,
                )
            )
        elif kind == 'Deployment' and operation == 'UPDATE':
            template = object['spec']['template']
            old_template = old_object['spec']['template']
            if object == old_object:
                logger.info(
                    f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】object 和 oldObject 相等，跳过"
                )
                return web.json_response(admis_pass(uid))
            elif template != old_template:
                # spec.template 变了,触发重启逻辑,按照数据库修改所有参数
                logger.info(
                    f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】收到 update 请求，修改所有参数"
                )
                resources_dict = object['spec']['template']['spec']['containers'][0].get(
                    'resources', {"requests": {}, "limits": {}}
                )
                return web.json_response(
                    update_all(
                        replicas,
                        namespace,
                        deployment_name,
                        request_cpu_m,
                        request_mem_mb,
                        limit_cpu_m,
                        limit_mem_mb,
                        resources_dict,
                        uid,
                    )
                )
            elif template == old_template and replicas != object['spec']['replicas']:
                # spec.template 没变,spec.replicas 变了,只把修改spec.replicas和数据库的pod数一致
                logger.info(
                    f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】收到 update 请求，仅修改 replicas"
                )
                return web.json_response(scale_only(uid, replicas))
            elif template == old_template and replicas == object['spec']['replicas']:
                # spec.template 没变,spec.replicas 没变,什么也不做
                logger.info(
                    f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】template 和 replicas 没变，不处理"
                )
                return web.json_response(admis_pass(uid))
        else:
            content = f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】不符合预设情况，未做任何处理"
            logger.info(content)
            utils.send_msg(content)
            return web.json_response(admis_pass(uid))
    except Exception as e:
        logger.error(f"【{namespace}】【{deployment_name}】Webhook 处理错误：{e}")
        utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deployment_name}】处理错误：{e}")
        return web.json_response({"error": str(e)}, status=500)


async def setup_routes(app):
    app.router.add_get('/api/health', health_check)
    app.router.add_post('/api/update-image', update_image)
    app.router.add_post('/api/scale', scale)
    app.router.add_post('/api/restart', reboot)
    app.router.add_post('/api/cron', cron)
    app.router.add_get('/api/admis_switch', admis_switch)
    app.router.add_post('/api/admis', admis_mutate)


async def start_https_server():
    """启动 HTTPS 服务器"""
    app = web.Application()
    await setup_routes(app)
    import ssl

    ssl_context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    ssl_context.load_cert_chain('/serving-certs/tls.crt', '/serving-certs/tls.key')

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 443, ssl_context=ssl_context)
    await site.start()
    logger.info("HTTPS 服务器已启动，监听端口 443")
    while True:
        await asyncio.sleep(3600)


async def main():
    """主函数"""
    init_kubernetes()  # 初始化 Kubernetes 配置
    await asyncio.gather(connect_to_server(), start_https_server())


if __name__ == "__main__":
    asyncio.run(main())
