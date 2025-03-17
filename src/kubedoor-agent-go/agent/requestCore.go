package agent

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"kubedoor-agent-go/utils"
	"net/http"
	"time"
)

var baseURL = "http://127.0.0.1:81"

func sendRequest(endpoint string, queryData map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.Logger.Error("requst py response", zap.Error(err))
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range queryData {
		q.Add(key, fmt.Sprintf("%v", value))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		utils.Logger.Error("requst py response", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		utils.Logger.Error("requst py response", zap.Error(err))
		return nil, err
	}
	s, err := json.Marshal(response)
	if err != nil {
		utils.Logger.Error("requst py response", zap.Error(err))
		return nil, err
	}
	utils.Logger.Info("requst py response", zap.String("response", string(s)))
	return response, nil
}

//
//func AutoDump(queryData map[string]interface{}) map[string]interface{} {
//	response, err := sendRequest("/auto_dump", queryData)
//	if err != nil {
//		zap.L().Error("Failed to perform auto dump", zap.Error(err))
//		return map[string]interface{}{"message": "Failed to perform auto dump", "success": false}
//	}
//	return response
//}
//
//func AutoJstack(queryData map[string]interface{}) map[string]interface{} {
//	response, err := sendRequest("/auto_jstack", queryData)
//	if err != nil {
//		zap.L().Error("Failed to perform auto jstack", zap.Error(err))
//		return map[string]interface{}{"message": "Failed to perform auto jstack", "success": false}
//	}
//	return response
//}
//
//func AutoJfr(queryData map[string]interface{}) map[string]interface{} {
//	response, err := sendRequest("/auto_jfr", queryData)
//	if err != nil {
//		zap.L().Error("Failed to perform auto jfr", zap.Error(err))
//		return map[string]interface{}{"message": "Failed to perform auto jfr", "success": false}
//	}
//	return response
//}
//
//func AutoJvmMem(queryData map[string]interface{}) map[string]interface{} {
//	response, err := sendRequest("/auto_jvm_mem", queryData)
//	if err != nil {
//		zap.L().Error("Failed to perform auto JVM memory check", zap.Error(err))
//		return map[string]interface{}{"message": "Failed to perform auto JVM memory check", "success": false}
//	}
//	return response
//}
//
//func GetTaskStatus(taskID string, env string) map[string]interface{} {
//	queryData := map[string]interface{}{"env": env}
//	endpoint := fmt.Sprintf("/task_status/%s", taskID)
//	response, err := sendRequest(endpoint, queryData)
//	if err != nil {
//		zap.L().Error("Failed to get task status", zap.String("taskID", taskID), zap.Error(err))
//		return map[string]interface{}{"message": "Failed to get task status", "success": false}
//	}
//	return response
//}
