import os
import sys
import time
import json
import requests
from datetime import datetime, timedelta
from collections import defaultdict
from clickhouse_driver import Client
from clickhouse_driver.errors import ServerException
from functools import wraps
from loguru import logger
logger.remove()
logger.add(sys.stderr,format='<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> [<level>{level}</level>] <level>{message}</level>',level='INFO')


# 环境变量
CK_DATABASE = os.environ.get('CK_DATABASE')
CK_HOST = os.environ.get('CK_HOST')
CK_HTTP_PORT = os.environ.get('CK_HTTP_PORT')
CK_PASSWORD = os.environ.get('CK_PASSWORD')
CK_PORT = os.environ.get('CK_PORT')
CK_USER = os.environ.get('CK_USER')
MSG_TOKEN = os.environ.get('MSG_TOKEN')
MSG_TYPE = os.environ.get('MSG_TYPE')
NAMESPACE_LIST = os.environ.get('NAMESPACE_LIST')
PEAK_TIME = os.environ.get('PEAK_TIME')
PROM_K8S_TAG_KEY = os.environ.get('PROM_K8S_TAG_KEY')
PROM_K8S_TAG_VALUE = os.environ.get('PROM_K8S_TAG_VALUE')
PROM_TYPE = os.environ.get('PROM_TYPE')
PROM_URL = os.environ.get('PROM_URL')


ckclient = Client(
    host=CK_HOST,
    port=CK_PORT,
    user=CK_USER,
    password=CK_PASSWORD,
    database=CK_DATABASE,
)


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
    return ckclient.execute(query)


query_list = ["pod_num", "core_usage", "core_usage_percent", "wss_usage_MB", "wss_usage_percent", "limit_core", "limit_mem_MB", "request_core", "request_mem_MB"]
query_dict = {
    #pod数
    "pod_num": 'count(label_replace(kube_pod_container_info{{env}container!="",container!="POD",namespace=~"{namespace}"}, "deployment", "$1", "pod", "^(.*)-[a-z0-9]+-[a-z0-9]+$")) by({env_key}deployment,namespace)',
    #使用核数P95
    "core_usage": 'quantile_over_time(0.95, avg by ({env_key}namespace, deployment) (label_replace(irate(container_cpu_usage_seconds_total{{env}container!="",container!="POD",namespace=~"{namespace}"}[2m]),"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$"))[{duration}:])',
    #CPU使用率P95
    "core_usage_percent": 'quantile_over_time(0.95,(sum by ({env_key}namespace, deployment)(label_replace(irate(container_cpu_usage_seconds_total{{env}container!="",container!="POD",namespace=~"{namespace}"}[2m]),"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$"))/sum by ({env_key}namespace, deployment)(label_replace(container_spec_cpu_quota{{env}container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")/100000)*100)[{duration}:]) != Inf',
    #WSS内存使用MB P95
    "wss_usage_MB": 'quantile_over_time(0.95,avg by ({env_key}namespace, deployment)(label_replace(container_memory_working_set_bytes{{env}container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$"))[{duration}:])/1024/1024',
    #WSS内存使用率P95
    "wss_usage_percent": 'quantile_over_time(0.95,avg by ({env_key}namespace, deployment)(label_replace(container_memory_working_set_bytes{{env}container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$"))[{duration}:])/max(label_replace(kube_pod_container_resource_limits{{env}resource="memory",unit="byte",container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")) by ({env_key}namespace,deployment) *100 != Inf',
    #CPU limit
    "limit_core": 'max(label_replace(kube_pod_container_resource_limits{{env}resource="cpu", unit="core",container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")) by ({env_key}namespace,deployment) *1000',
    #内存limit_MB
    "limit_mem_MB": 'max(label_replace(kube_pod_container_resource_limits{{env}resource="memory",unit="byte",container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")) by ({env_key}namespace,deployment)/1024/1024',
    #CPU request
    "request_core" :'max(label_replace(kube_pod_container_resource_requests{{env}resource="cpu", unit="core",container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")) by ({env_key}namespace,deployment) * 1000',
    #内存request_MB
    "request_mem_MB": 'max(label_replace(kube_pod_container_resource_requests{{env}resource="memory", unit="byte",container!="",container!="POD",namespace=~"{namespace}"},"deployment","$1","pod","^(.*)-[a-z0-9]+-[a-z0-9]+$")) by ({env_key}deployment,namespace)/1024/1024',
}


def calculate_peak_duration_and_end_time(peak_time):
    # 提取开始和结束时间
    start_str, end_str = peak_time.split('-')
    start_time = datetime.strptime(start_str, '%H:%M:%S')
    end_time = datetime.strptime(end_str, '%H:%M:%S')
    # 计算持续时间
    duration = end_time - start_time
    duration_hours = duration.seconds // 3600
    duration_minutes = (duration.seconds % 3600) // 60
    # 生成持续时间的字符串
    duration_str = f"{duration_hours}h{duration_minutes}m"

    end_time_part = end_time.time()
    return duration_str, end_time_part


def check_and_delete_day_data(date):
    """检查是否有当天的数据，有则删除"""
    query_sql = f"""select * from kubedoor.k8s_resources where date = '{date}'"""
    delete_sql = f"""delete from kubedoor.k8s_resources where date = '{date}'"""
    logger.info(f"query_sql==={query_sql}")
    result = ckclient.execute(query_sql)
    ckclient.disconnect()
    if result:
        logger.info(f"从表k8s_resources删除{date}的数据")
        logger.info(f"delete_sql==={delete_sql}")
        ckclient.execute("SET allow_experimental_lightweight_delete = 1")
        ckclient.execute(delete_sql)
    return result


def get_prom_url():
    """按类型选择查询指标的方式"""
    url = f"{PROM_URL}/api/v1/query_range"
    # if PROM_TYPE == "Prometheus":
    #     url = f"{PROM_URL}/api/v1/query_range"
    # if PROM_TYPE == "Victoria-Metrics-Single":
    #     url = f"{PROM_URL}/api/v1/query_range"
    # if PROM_TYPE == "Victoria-Metrics-Cluster":
    #     url = f"{PROM_URL}/select/0/prometheus/api/v1/query_range"
    return url


def get_prom_data(promql, env, namespace_str, end_time_full, duration):
    """获取指标源数据"""
    url = get_prom_url()
    if PROM_K8S_TAG_KEY:
        k8s_filter = f'{PROM_K8S_TAG_KEY}=~"{PROM_K8S_TAG_VALUE}",'
        query = query_dict.get(promql).replace("{env}", k8s_filter).replace("{env_key}", f"{PROM_K8S_TAG_KEY},")
    else:
        query = query_dict.get(promql).replace("{env}", '').replace("{env_key}", '')
    query = query.replace("{namespace}", namespace_str).replace("{duration}", duration)
    querystring = {"query":query,"start":end_time_full.timestamp(),"end":end_time_full.timestamp(),"step":"15"}
    logger.info(querystring)
    response = requests.request("GET", url, params=querystring).json()
    if response.get("status") == "success":
        result = response["data"]["result"]
        metrics_dict = {}
        for x in result:
            for tv in x['values']:
                if PROM_K8S_TAG_KEY:
                    k8s = x['metric'][PROM_K8S_TAG_KEY]
                else:
                    k8s = "k8s"
                ns = x['metric'].get('namespace',x['metric'].get('k8s_ns')) or x['metric'].get('namespace',x['metric'].get('destination_workload_namespace'))
                ms = x['metric'].get('deployment')
                if promql == "pod_num":
                    metrics_dict[f'{tv[0]}@{k8s}@{ns}@{ms}'] = int(tv[1])
                else:
                    metrics_dict[f'{tv[0]}@{k8s}@{ns}@{ms}'] = float(tv[1])
        logger.info('单个指标数量 {}: {}', promql,len(metrics_dict))
        return metrics_dict
    else:
        logger.error('ERROR {} {}', promql, env)


def merged_dict(env, namespace_str, duration_str, end_time_full):
    """解析指标源数据，处理成列表"""
    metrics_list = []
    metrics_keys_list = []
    k8s_metrics_list = []
    for promql in query_list:
        metrics_dict = get_prom_data(promql, env, namespace_str, end_time_full, duration_str)
        metrics_keys_list = metrics_keys_list + list(metrics_dict.keys())
        metrics_list.append(metrics_dict)
    metrics_keys_list = list(set(metrics_keys_list))
    new_metrics_list = []
    for i in metrics_list:
        for j in metrics_keys_list:
            if j not in i:
                i[j] = -1
        logger.info('处理完成后（补全查不到的值）的数据行数: {}',len(i))
        new_metrics_list.append(i)

    merged_dict = defaultdict(list)
    for d in new_metrics_list:
        for k, v in d.items():
            merged_dict[k].append(v)
    for k,v in merged_dict.items():
        key_list = k.split('@')
        key_list[0] = datetime.fromtimestamp(int(key_list[0]))
        if v[0] == -1:  # 如果未从指标查到pod数，则置为0
            v[0] = 0
        logger.info(key_list + v)
        k8s_metrics_list.append(key_list + v + [-1, -1, -1])
    return k8s_metrics_list

def metrics_to_ck(k8s_metrics_list):
    """将指标数据存入ck"""
    batch_size = 10000
    for i in range(0, len(k8s_metrics_list), batch_size):
        begin = time.time()
        batch_data = k8s_metrics_list[i : i + batch_size]
        try:
            ckclient.execute("INSERT INTO k8s_resources VALUES", batch_data)
            logger.info(
                f"== count: insert batch {i//batch_size}",
                "耗时：{:.2f}s".format(time.time() - begin),
            )
        except ServerException as e:
            logger.exception("Failed to insert batch {}:// {}", i//batch_size, e)

    ckclient.disconnect()
    return True


def send_msg(content):
    response = ""
    if MSG_TYPE  == "wecom":
        response = wecom(MSG_TOKEN, content)
    if MSG_TYPE  == "dingding":
        response = wecom(MSG_TOKEN, content)
    if MSG_TYPE  == "feishu":
        response = wecom(MSG_TOKEN, content)
    return f'【{MSG_TYPE}】{response}'


def wecom(webhook, content, at=""):
    webhook = 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {'msgtype': 'markdown', 'markdown': {'content': f"{content}<@{at}>"}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    logger.info(f'【wecom】{response.json()}')
    return response.json()


def dingding(webhook, content, at=""):
    webhook = 'https://oapi.dingtalk.com/robot/send?access_token=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {"msgtype": "markdown", "markdown": {"title": "告警", "text": content}, "at": {"atMobiles": [at]}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    logger.info(f'【dingding】{response.json()}')
    return response.json()


def feishu(webhook, content, at=""):
    title = "告警通知"
    webhook = f'https://open.feishu.cn/open-apis/bot/v2/hook/{webhook}'
    headers = {'Content-Type': 'application/json'}
    params = {
        "msg_type": "interactive",
        "card": {
            "header": {"title": {"tag": "plain_text", "content": title}, "template": "red"},
            "elements": [
                {
                    "tag": "markdown",
                    "content": f"{content}\n<at id={at}></at>",
                }
            ],
        },
    }
    data = json.dumps(params)
    response = requests.post(webhook, headers=headers, data=data)
    logger.info(f'【feishu】{response.json()}')
    return response.json()


def get_list_from_resources():
    """获取资源表信息，取最近10天cpu数据最高的一天的数据"""
    query = f"""
        select
            `date`,
            env,
            namespace,
            deployment,
            pod_count,
            p95_pod_cpu_pct,
            p95_pod_wss_pct,
            request_pod_cpu_m,
            request_pod_mem_mb,
            limit_pod_cpu_m,
            limit_pod_mem_mb,
            p95_pod_load,
            p95_pod_wss_mb
        from kubedoor.k8s_resources
        where date = (
            SELECT `date`
            FROM kubedoor.k8s_resources
            WHERE `date` >= toDate(today() - 10)
            GROUP BY `date`
            order by SUM(pod_count * p95_pod_load) desc
            limit 1
        )
    """
    result = ckclient.execute(query)
    ckclient.disconnect()
    logger.info("提取最近10天cpu最高的一天的数据：")
    for i in result:
        logger.info(i)
    return result


def is_init_or_update():
    """判断管控表是初始化还是更新"""
    query = """select * from kubedoor.k8s_res_control"""
    result = ckclient.execute(query)
    if not result: # 初始化
        return True
    else: # 更新
        return False


def parse_insert_data(srv):
    """将从resource表查到的指标数据，解析为可以存入管控表的数据"""
    # 把request-cpu,request-mem,limit-cpu,limit-mem这四个值转化为整数
    srv = list(srv)
    for j in range(7, 11):
        srv[j] = int(srv[j])
    tmp = [srv[1], srv[2], srv[3], srv[4], srv[4], -1, srv[5], srv[6], int(srv[11]*1000), int(srv[12]), srv[9], srv[10], srv[0], -1, -1,
           -1, -1, -1, -1, -1, datetime(2000, 1, 1, 0, 0, 0)]
    return tmp

def init_control_data(metrics_list_ck):
    '''初始化管控表'''
    metrics_list = list()
    for srv in metrics_list_ck:
        tmp = parse_insert_data(srv)
        logger.info(tmp)
        metrics_list.append(tmp)
    batch_size = 10000
    for i in range(0, len(metrics_list), batch_size):
        begin = time.time()
        batch_data = metrics_list[i : i + batch_size]
        try:
            ckclient.execute("INSERT INTO k8s_res_control VALUES", batch_data, types_check=True)
            logger.info(
                f"== count: insert batch {i//batch_size}",
                "耗时：{:.2f}s".format(time.time() - begin),
            )
        except ServerException as e:
            logger.exception("Failed to insert batch {}: {}", i//batch_size, e)
            return False

    ckclient.disconnect()
    return True


def update_control_data(metrics_list_ck):
    """更新管控表"""
    for i in metrics_list_ck:
        date, env, namespace, deployment, pod_count, p95_pod_cpu_pct, p95_pod_wss_pct, request_pod_cpu_m, request_pod_mem_mb, limit_pod_cpu_m, limit_pod_mem_mb, p95_pod_load, p95_pod_wss_mb= i
        date_str = date.strftime('%Y-%m-%d %H:%M:%S')
        sql = f"select * from kubedoor.k8s_res_control where namespace = '{namespace}' and deployment = '{deployment}'"
        data = ckclient.execute(sql)
        if data:  # 更新
            if data[0][0] == env:  # PROM_K8S_TAG_KEY未改变
                request_cpu_m = p95_pod_load * 1000
                try:
                    update_sql = f"""
                        alter table kubedoor.k8s_res_control
                        update
                            `update` = '{date_str}',
                            pod_count = {pod_count},
                            p95_pod_cpu_pct = {p95_pod_cpu_pct},
                            p95_pod_mem_pct = {p95_pod_wss_pct},
                            request_cpu_m = {int(request_cpu_m)},
                            request_mem_mb = {int(p95_pod_wss_mb)}
                        where
                            namespace = '{namespace}'
                            and deployment = '{deployment}'
                    """
                    update_data = ckclient.execute(update_sql)
                except Exception as e:
                    logger.exception("Failed to execute {}: {}", update_sql, e)
                    ckclient.disconnect()
                    return False
            else:  # PROM_K8S_TAG_KEY改变了
                try:  # 先删除
                    ckclient.execute("SET allow_experimental_lightweight_delete = 1")
                    del_sql = f"DELETE from k8s_res_control WHERE namespace = '{namespace}' and deployment = '{deployment}'"
                    ckclient.execute(del_sql)
                except Exception as e:
                    logger.exception("Failed to execute {}: {}", del_sql, e)
                    ckclient.disconnect()
                    return False
                try:  # 再重新插入数据
                    tmp = parse_insert_data(i)
                    ckclient.execute("INSERT INTO k8s_res_control VALUES", [tuple(tmp)], types_check=True)
                except Exception as e:
                    logger.exception("Failed to insert {}: {}", [tuple(tmp)], e)
                    ckclient.disconnect()
                    return False
        else:  # 添加
            content = f"执行更新管控表脚本，【{env}】命名空间【{namespace}】下服务【{deployment}】未配置，新增数据"
            logger.info(content)
            send_msg(content)
            tmp = ""
            try:
                tmp = parse_insert_data(i)
                ckclient.execute("INSERT INTO k8s_res_control VALUES", [tuple(tmp)], types_check=True)
            except Exception as e:
                logger.exception("Failed to insert {}: {}", [tuple(tmp)], e)
                ckclient.disconnect()
                return False
    ckclient.disconnect()

