package agent

import (
	"encoding/json"
	"fmt"
	"kubedoor-agent-go/config"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kubedoor-agent-go/utils" // Import the utils module

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func StartAgent() {
	utils.InitLogger() // Initialize logger

	version := utils.GetVersion()
	kubeDoorMaster := os.Getenv("KUBEDOOR_MASTER")
	promK8sTagValue := os.Getenv("PROM_K8S_TAG_VALUE")

	if kubeDoorMaster == "" {
		utils.Logger.Error("KUBEDOOR_MASTER environment variable is not set")
		log.Fatal("KUBEDOOR_MASTER environment variable is not set")
		return
	}

	wsURL := fmt.Sprintf("%s/ws?env=%s&ver=%s", kubeDoorMaster, promK8sTagValue, version)
	u, err := url.Parse(wsURL)
	if err != nil {
		utils.Logger.Error("Failed to parse websocket URL", zap.Error(err))
		log.Fatalf("Failed to parse websocket URL: %v", err)
		return
	}

	log.Printf("Connecting to %s", u.String())

	// Set up signal handling for Ctrl+C (SIGINT) termination
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the WebSocket connection loop
	for {
		select {
		case <-stopChan:
			// Gracefully handle termination
			log.Println("Received termination signal (Ctrl+C). Shutting down...")
			return // Exit the loop and terminate the program
		default:
			// Proceed with normal WebSocket connection
			conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				// 如果有 HTTP 响应，检查状态码
				if resp != nil {
					if resp.StatusCode != http.StatusConflict { // 不等于409 状态码
						utils.Logger.Error("WebSocket connection failed", zap.Int("status_code", resp.StatusCode), zap.Error(err))
						return
					}
					utils.Logger.Warn("Agent already registered", zap.Int("status_code", resp.StatusCode))
				} else {
					utils.Logger.Error("WebSocket connection failed", zap.Error(err))
				}

				// 继续尝试重连
				log.Printf("Failed to connect: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}
			defer conn.Close()

			log.Println("Connected to server")

			// Start heartbeat goroutine
			go func() {
				heartbeatTicker := time.NewTicker(4 * time.Second)
				defer heartbeatTicker.Stop()
				for range heartbeatTicker.C {
					err := conn.WriteJSON(map[string]string{"type": "heartbeat"})
					if err != nil {
						utils.Logger.Error("Failed to send heartbeat", zap.Error(err))
						log.Println("Heartbeat failed:", err)
						return // Exit goroutine if heartbeat fails
					}
					utils.Logger.Debug("Heartbeat sent successfully")
				}
			}()

			// Process requests
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					utils.Logger.Error("Failed to read message from websocket", zap.Error(err))
					log.Println("Read failed:", err)
					break // Exit request processing loop, reconnect
				}

				utils.Logger.Debug("Received message:", zap.String("message", string(message)))

				var data config.MessageDataStruct
				if err := json.Unmarshal(message, &data); err != nil {
					utils.Logger.Error("Failed to unmarshal message", zap.Error(err), zap.String("message", string(message)))
					log.Println("Failed to unmarshal message:", err)
					continue
				}

				switch data.MessageType {
				case "request":
					go handleMessage(conn, data)
				default:
					utils.Logger.Warn("Unknown message type", zap.String("type", fmt.Sprintf("%s", data.MessageType)))
				}
			}

			log.Println("Reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
}
