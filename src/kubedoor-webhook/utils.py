import os
import json
import requests
from clickhouse_driver import Client

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

def send_wecom(content):
    webhook = f'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key={MSG_TOKEN}'
    headers = {'Content-Type': 'application/json'}
    params = {'msgtype': 'markdown', 'markdown': {'content': content}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    return f'【wecom】{response.json()}'
