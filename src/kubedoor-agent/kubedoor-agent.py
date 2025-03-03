import asyncio
import json
import logging
from urllib.parse import urlencode
from aiohttp import ClientSession, ClientWebSocketResponse, WSMsgType
import utils

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(message)s")


VERSION = utils.get_version()


async def process_request(ws: ClientWebSocketResponse):
    """处理服务端发送的请求"""
    async for msg in ws:
        if msg.type == WSMsgType.TEXT:
            try:
                data = json.loads(msg.data)
            except json.JSONDecodeError:
                logging.error(f"收到无法解析的消息：{msg.data}")
                continue

            if data.get("type") == "request":
                request_id = data["request_id"]
                method = data["method"]
                if data["path"].startswith('/api/pod/'):
                    path = 'http://127.0.0.1:81' + data["path"]
                else:
                    path = 'http://127.0.0.1' + data["path"]
                query = data["query"]
                body = data["body"]
                try:
                    async with ClientSession() as session:
                        logging.info(f"转发请求: {method} {path}?{urlencode(query)}【{json.dumps(body)}】")
                        if method == "GET":
                            async with session.get(path, params=query) as resp:
                                response_data = await resp.json()
                        elif method == "POST":
                            async with session.post(path, params=query, json=body) as resp:
                                response_data = await resp.json()
                        else:
                            response_data = {"error": f"Unsupported method: {method}"}
                except Exception as e:
                    response_data = {"error": str(e)}

                response_message = {
                    "type": "response",
                    "request_id": request_id,
                    "response": response_data,
                }
                await ws.send_json(response_message)  # 使用 JSON 格式发送响应
        elif msg.type == WSMsgType.ERROR:
            logging.error(f"WebSocket 错误：{msg.data}")


async def heartbeat(ws: ClientWebSocketResponse):
    """定期发送心跳"""
    while True:
        try:
            await ws.send_json({"type": "heartbeat"})  # 使用 JSON 格式发送心跳
            logging.debug("成功发送心跳")
            await asyncio.sleep(4)
        except Exception as e:
            logging.error(f"心跳发送失败：{e}")
            break


async def connect_to_server():
    """连接到 WebSocket 服务端，并处理连接断开的情况"""
    uri = f"{utils.KUBEDOOR_MASTER}/ws?env={utils.PROM_K8S_TAG_VALUE}&ver={VERSION}"
    while True:
        try:
            async with ClientSession() as session:
                async with session.ws_connect(uri) as ws:
                    logging.info("成功连接到服务端")
                    # 并发处理心跳和请求
                    await asyncio.gather(process_request(ws), heartbeat(ws))
        except Exception as e:
            logging.error(f"连接到服务端失败：{e}")

        # 等待一段时间后重连
        logging.info("等待 5 秒后重新连接...")
        await asyncio.sleep(5)


if __name__ == "__main__":
    asyncio.run(connect_to_server())
