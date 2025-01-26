#!/usr/bin/python3
import json, datetime, requests, os
from flask import Flask, Response, request

MSG_TOKEN = os.environ.get('MSG_TOKEN')
MSG_TYPE = os.environ.get('MSG_TYPE')
DEFAULT_AT = os.environ.get('DEFAULT_AT')
ALERTMANAGER_EXTURL = os.environ.get('ALERTMANAGER_EXTURL')


def wecom(webhook, content, at):
    webhook = 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {'msgtype': 'markdown', 'markdown': {'content': f"{content}<@{at}>"}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    print(f'【wecom】{response.json()}', flush=True)


def dingding(webhook, content, at):
    webhook = 'https://oapi.dingtalk.com/robot/send?access_token=' + webhook
    headers = {'Content-Type': 'application/json'}
    params = {"msgtype": "markdown", "markdown": {"title": "告警", "text": content}, "at": {"atMobiles": [at]}}
    data = bytes(json.dumps(params), 'utf-8')
    response = requests.post(webhook, headers=headers, data=data)
    print(f'【dingding】{response.json()}', flush=True)


def feishu(webhook, content, at):
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
    print(f'【feishu】{response.json()}')


app = Flask(__name__)


@app.route("/msg/<token>", methods=['POST'])
def alertnode(token):
    req = request.get_json()
    print('↓↓↓↓↓↓↓↓↓↓↓↓↓↓node↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓', flush=True)
    print(json.dumps(req, indent=2, ensure_ascii=False), flush=True)
    print('↑↑↑↑↑↑↑↑↑↑↑↑↑↑node↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑', flush=True)
    now_utc = datetime.datetime.now(datetime.UTC).replace(tzinfo=None)
    now_cn = datetime.datetime.now()

    # time1830 = datetime.datetime.strptime(str(now_cn.date()) + '18:30', '%Y-%m-%d%H:%M')
    # time0830 = datetime.datetime.strptime(str(now_cn.date()) + '08:30', '%Y-%m-%d%H:%M')

    # if (now_cn > time1830 or now_cn < time0830):
    #    return Response(status=204)
    allmd = ''
    for i in req["alerts"]:
        status = "故障" if i['status'] == "firing" else "恢复"
        try:
            firstime = datetime.datetime.strptime(i['startsAt'], '%Y-%m-%dT%H:%M:%SZ')
            durn_s = (now_utc - firstime).total_seconds()
        except:
            try:
                firstime = datetime.datetime.strptime(i['startsAt'], '%Y-%m-%dT%H:%M:%S.%fZ')
                durn_s = (now_utc - firstime).total_seconds()
            except:
                firstime = datetime.datetime.strptime(i['startsAt'].split(".")[0], '%Y-%m-%dT%H:%M:%S+08:00')
                durn_s = (now_cn - firstime).total_seconds()
        if durn_s < 60:
            durninfo = '小于1分钟'
        elif durn_s < 3600:
            durn = round(durn_s / 60, 1)
            durninfo = f"已持续{durn}分钟"
        else:
            durn = round(durn_s / 3600, 1)
            durninfo = f"已持续{durn}小时"

        summary = f"{i['labels']['alertname']},{durninfo}"
        message = i['annotations']['description']
        at = i['annotations'].get('at', DEFAULT_AT)

        url = f"{ALERTMANAGER_EXTURL}/#/alerts?silenced=false&inhibited=false&active=true&filter=%7Balertname%3D%22{i['labels']['alertname']}%22%7D"

        if status == '恢复':
            info = f"### {status}<font color=\"#6aa84f\">{summary}</font>\n- {message}\n\n"
        else:
            info = f"### {status}<font color=\"#ff0000\">{summary}</font>\n- {message}[【屏蔽】]({url})\n\n"
        allmd = allmd + info

    im, key = token.split('=', 1)
    if im == 'wecom':
        wecom(key, allmd, at)
    elif im == 'dingding':
        dingding(key, allmd, at)
    elif im == 'feishu':
        feishu(key, allmd, at)
    return Response(status=200)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=80)
