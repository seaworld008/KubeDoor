package config

import "encoding/json"

type MessageDataStruct struct {
	MessageType string                 `json:"type"` // request
	RequestID   string                 `json:"request_id"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Query       map[string]interface{} `json:"query"`
	Body        json.RawMessage        `json:"body"`
}

// BodyUpdateImageStruct 更新镜像
// {"image_tag": "xxx", "deployment": "testhelloworld", "namespace": "ai" }
type BodyUpdateImageStruct struct {
	ImageTag   string `json:"image_tag"`
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
}

// BodyScaleRestartStruct 扩缩容 重启
// [{"namespace":"ai","deployment_name":"testhelloworld"}]
// [{"namespace":"ai","deployment_name":"testhelloworld","num":1}]
type BodyScaleRestartStruct struct {
	Namespace      string `json:"namespace"`
	DeploymentName string `json:"deployment_name"`
	Num            int    `json:"num,omitempty"`
}

// BodyTimeScaleRestartStruct 周期性扩容 定时扩容 周期性重启 定时重启
// {"type":"restart","service":[{"namespace":"ai","deployment_name":"testhelloworld"}],"time":[2025,3,27,0,0],"cron":""}
// {"type":"restart","service":[{"namespace":"ai","deployment_name":"testhelloworld"}],"time":"","cron":"0 0 3 * *"}
// {"type":"scale","service":[{"namespace":"ai","deployment_name":"testhelloworld","num":1}],"time":"","cron":"0 0 3 * *"}
// {"type":"scale","service":[{"namespace":"ai","deployment_name":"testhelloworld","num":1}],"time":[2025,3,25,0,0],"cron":""}
type BodyTimeScaleRestartStruct struct {
	BodyType string                   `json:"type"`
	Service  []BodyScaleRestartStruct `json:"service"`
	Time     interface{}              `json:"time"`
	Cron     string                   `json:"cron"`
}

type ResponseMessageStruct struct {
	ResponseMessageType string      `json:"type"`
	RequestId           string      `json:"request_id"`
	Response            interface{} `json:"response"`
}
