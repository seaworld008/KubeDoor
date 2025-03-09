package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kubedoor-agent-go/asset"
	"net/http"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func Wecom(webhook, content string) string {
	wecomWebhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + webhook
	headers := map[string]string{"Content-Type": "application/json"}
	params := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}
	payload, err := json.Marshal(params)
	if err != nil {
		Logger.Error("Failed to marshal Wecom payload", zap.Error(err))
		return fmt.Sprintf("Failed to marshal Wecom payload: %v", err)
	}

	req, err := http.NewRequest("POST", wecomWebhook, bytes.NewBuffer(payload))
	if err != nil {
		Logger.Error("Failed to create Wecom request", zap.Error(err))
		return fmt.Sprintf("Failed to create Wecom request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Error("Failed to send Wecom message", zap.Error(err))
		return fmt.Sprintf("Failed to send Wecom message: %v", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		Logger.Error("Failed to decode Wecom response", zap.Error(err))
		return fmt.Sprintf("Failed to decode Wecom response: %v", err)
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON)
}

func InitLogger() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	Logger = logger
}

func GetVersion() string {
	content, err := asset.Asset("version")
	if err != nil {
		Logger.Error("Failed to read version file", zap.Error(err))
		return "unknown"
	}
	return string(content)
}

func SendMsg(msgType, token, content string) string {
	switch msgType {
	case "wecom":
		return Wecom(token, content)
	case "dingding":
		return Dingding(token, content)
	case "feishu":
		return Feishu(token, content)
	default:
		Logger.Warn("Unsupported message type", zap.String("type", msgType))
		return fmt.Sprintf("Unsupported message type: %s", msgType)
	}
}

func Dingding(webhook, content string) string {
	dingdingWebhook := "https://oapi.dingtalk.com/robot/send?access_token=" + webhook
	headers := map[string]string{"Content-Type": "application/json"}
	params := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": "告警",
			"text":  content,
		},
		"at": map[string]interface{}{
			"atMobiles": []string{}, // TODO: Support @ mobiles
		},
	}
	payload, err := json.Marshal(params)
	if err != nil {
		Logger.Error("Failed to marshal Dingding payload", zap.Error(err))
		return fmt.Sprintf("Failed to marshal Dingding payload: %v", err)
	}

	req, err := http.NewRequest("POST", dingdingWebhook, bytes.NewBuffer(payload))
	if err != nil {
		Logger.Error("Failed to create Dingding request", zap.Error(err))
		return fmt.Sprintf("Failed to create Dingding request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Error("Failed to send Dingding message", zap.Error(err))
		return fmt.Sprintf("Failed to send Dingding message: %v", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		Logger.Error("Failed to decode Dingding response", zap.Error(err))
		return fmt.Sprintf("Failed to decode Dingding response: %v", err)
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON)
}

func Feishu(webhook, content string) string {
	feishuWebhook := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", webhook)
	headers := map[string]string{"Content-Type": "application/json"}
	params := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "告警通知",
				},
				"template": "red",
			},
			"elements": []interface{}{
				map[string]interface{}{
					"tag":     "markdown",
					"content": content + "\n<at id=></at>", // TODO: Support @ user
				},
			},
		},
	}
	payload, err := json.Marshal(params)
	if err != nil {
		Logger.Error("Failed to marshal Feishu payload", zap.Error(err))
		return fmt.Sprintf("Failed to marshal Feishu payload: %v", err)
	}

	req, err := http.NewRequest("POST", feishuWebhook, bytes.NewBuffer(payload))
	if err != nil {
		Logger.Error("Failed to create Feishu request", zap.Error(err))
		return fmt.Sprintf("Failed to create Feishu request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Error("Failed to send Feishu message", zap.Error(err))
		return fmt.Sprintf("Failed to send Feishu message: %v", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		Logger.Error("Failed to decode Feishu response", zap.Error(err))
		return fmt.Sprintf("Failed to decode Feishu response: %v", err)
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON)
}
