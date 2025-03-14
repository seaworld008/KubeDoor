package config

import (
	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"os"
)

var (
	PROM_K8S_TAG_VALUE = os.Getenv("PROM_K8S_TAG_VALUE")
	KubeClient         *kubernetes.Clientset
	WebSocketConcent   *websocket.Conn // 你需要定义 WebSocketClient 或者其他的 WebSocket 客户端
)
