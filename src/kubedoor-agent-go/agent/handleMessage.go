package agent

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"kubedoor-agent-go/api"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils"
)

func handleMessage(conn *websocket.Conn, data config.MessageDataStruct) {
	var response map[string]interface{}
	// 返回响应
	responseMessage := config.ResponseMessageStruct{
		ResponseMessageType: "response",
		RequestId:           data.RequestID,
	}
	// 输出请求的日志
	utils.Logger.Info("Forwarding request",
		zap.String("method", data.Method),
		zap.String("path", data.Path),
		zap.Any("query", data.Query),
		zap.Any("body", data.Body),
	)

	// 按照 path 处理不同的 API
	switch data.Path {
	case "/api/health":
		response = api.HealthCheck()
	case "/api/update-image":
		// 更新镜像需要使用 BodyUpdateImageStruct
		var body config.BodyUpdateImageStruct
		if err := json.Unmarshal(data.Body, &body); err != nil {
			utils.Logger.Error("Failed to unmarshal body", zap.Error(err))
			return
		}
		utils.Logger.Info("Handling BodyUpdateImageStruct:", zap.Any("body", body))
		response = api.UpdateImage(body)
	case "/api/scale":
		// 扩缩容需要使用 BodyScaleStruct
		var body []config.BodyScaleRestartStruct
		if err := json.Unmarshal(data.Body, &body); err != nil {
			utils.Logger.Error("Failed to unmarshal body", zap.Error(err))
			return
		}
		utils.Logger.Info("Handling BodyScaleStruct:", zap.Any("body", body))
		response = api.ScaleDeployment(body, false)
	case "/api/restart":
		// 重启服务需要使用 BodyRestartStruct
		var body []config.BodyScaleRestartStruct
		if err := json.Unmarshal(data.Body, &body); err != nil {
			utils.Logger.Error("Failed to unmarshal body", zap.Error(err))
			return
		}
		utils.Logger.Info("Handling BodyRestartStruct:", zap.Any("body", body))
		response = api.RestartDeployment(body, false)
	case "/api/cron":
		// 定时相关操作
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(data.Body, &bodyMap); err != nil {
			utils.Logger.Error("Failed to unmarshal body", zap.Error(err))
			return
		}
		utils.Logger.Info("Handling BodyRestartStruct:", zap.Any("body", bodyMap))

		var body config.BodyTimeScaleRestartStruct
		if err := json.Unmarshal(data.Body, &body); err != nil {
			utils.Logger.Error("Failed to unmarshal body", zap.Error(err))
		}
		response = api.ScheduleServiceScaleRestart(bodyMap["type"].(string), body)

	default:
		utils.Logger.Warn("Unknown path", zap.String("path", data.Path))
	}
	responseMessage.Response = response
	if err := conn.WriteJSON(responseMessage); err != nil {
		utils.Logger.Error("Failed to send response", zap.Error(err))
	}
}
