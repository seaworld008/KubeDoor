import os
import sys
import json
import requests
from loguru import logger

logger.remove()
logger.add(
    sys.stderr,
    format='<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> [<level>{level}</level>] <level>{message}</level>',
    level='INFO',
)


# 环境变量

MSG_TOKEN = os.environ.get('MSG_TOKEN')
MSG_TYPE = os.environ.get('MSG_TYPE')
KUBEDOOR_MASTER = os.environ.get('KUBEDOOR_MASTER')
PROM_K8S_TAG_VALUE = os.environ.get('PROM_K8S_TAG_VALUE')
OSS_URL = os.environ.get('OSS_URL')


def get_version():
    try:
        with open('/app/version', 'r') as f:
            return f.read().strip()
    except Exception as e:
        logger.error(f"读取版本号失败：{e}")
        return "unknown"


def send_msg(content):
    response = ""
    if MSG_TYPE == "wecom":
        response = wecom(MSG_TOKEN, content)
    elif MSG_TYPE == "dingding":
        response = dingding(MSG_TOKEN, content)
    elif MSG_TYPE == "feishu":
        response = feishu(MSG_TOKEN, content)
    else:
        logger.warning(f"不支持的消息类型：{MSG_TYPE}")
        return f"不支持的消息类型：{MSG_TYPE}"

    logger.info(f'【{MSG_TYPE}】{response}')
    return f'【{MSG_TYPE}】{response}'


def wecom(webhook, content, at=""):
    webhook = 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {'msgtype': 'markdown', 'markdown': {'content': f"{content}<@{at}>"}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    return response.json()


def dingding(webhook, content, at=""):
    webhook = 'https://oapi.dingtalk.com/robot/send?access_token=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {"msgtype": "markdown", "markdown": {"title": "告警", "text": content}, "at": {"atMobiles": [at]}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
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
    return response.json()
