package config

import (
	"encoding/json"
	"sync"
)

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

var (
	TaskManagerObj *TaskManager
)

func init() {
	TaskManagerObj = NewTaskManager()
}

// TaskResult represents the status and details of a task
type TaskResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	File    string `json:"file,omitempty"`
}

// TaskManager manages the task results
type TaskManager struct {
	mu          sync.RWMutex
	taskResults map[string]*TaskResult
}

// NewTaskManager creates a new instance of TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		taskResults: make(map[string]*TaskResult),
	}
}

// UpdateTask updates the task result by task ID
func (tm *TaskManager) UpdateTask(taskID string, result *TaskResult) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.taskResults[taskID] = result
}

// GetTask returns the result of a task by ID
func (tm *TaskManager) GetTask(taskID string) ([]byte, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result, exists := tm.taskResults[taskID]
	if exists {
		resultB, _ := json.Marshal(result)
		return resultB, true
	}
	return []byte{}, false
}
