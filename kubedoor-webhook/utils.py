import os
import sys
import json
import requests
from clickhouse_driver import Client
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

NS_LIST = NAMESPACE_LIST.split(',')

ckclient = Client(
    host=CK_HOST,
    port=CK_PORT,
    user=CK_USER,
    password=CK_PASSWORD,
    database=CK_DATABASE,
)


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
