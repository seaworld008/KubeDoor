package api

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	admisv1 "k8s.io/api/admission/v1"
	"k8s.io/api/apps/v1" // 导入 Deployment
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils"
	"net/http"
	"time"
)

func updateAll(replicas int, namespace, deploymentName string, requestCPU, requestMem, limitCPU, limitMem int, resourcesObj corev1.ResourceRequirements, uid string) *admissionv1.AdmissionReview {
	// 按照数据库修改所有参数
	patchReplicas := map[string]interface{}{
		"op":    "replace",
		"path":  "/spec/replicas",
		"value": replicas,
	}
	if resourcesObj.Requests == nil {
		resourcesObj.Requests = corev1.ResourceList{}
	}
	if resourcesObj.Limits == nil {
		resourcesObj.Limits = corev1.ResourceList{}
	}

	b, err := json.Marshal(resourcesObj)
	if err != nil {
		utils.Logger.Error("Error marshalling resourcesObj: ", zap.Error(err))
	}
	// request_cpu_m, request_mem_mb, limit_cpu_m, limit_mem_mb，有则改，无则不动
	utils.Logger.Info(fmt.Sprintf("改前：%s", string(b)))
	if requestCPU > 0 {
		resourcesObj.Requests[corev1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%dm", requestCPU))
	} else {
		utils.Logger.Info(fmt.Sprintf("admis:【%s】【%s】【%s】未配置request_cpu_m", namespace, deploymentName, "PROM_K8S_TAG_VALUE"))
	}

	if requestMem > 0 {
		resourcesObj.Requests[corev1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%dMi", requestMem))
	} else {
		// utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deploymentName}】未配置request_mem_mb")
		utils.Logger.Info(fmt.Sprintf("admis:【%s】【%s】【%s】未配置request_mem_mb", namespace, deploymentName, "PROM_K8S_TAG_VALUE"))
	}

	if limitCPU > 0 {
		resourcesObj.Limits[corev1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%dm", limitCPU))

	} else {
		// utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deploymentName}】未配置limit_cpu_m")
		utils.Logger.Info(fmt.Sprintf("admis:【%s】【%s】【%s】未配置limit_cpu_m", namespace, deploymentName, "PROM_K8S_TAG_VALUE"))
	}

	if limitMem > 0 {
		resourcesObj.Limits[corev1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%dMi", limitMem))
	} else {
		// utils.send_msg(f"admis:【{utils.PROM_K8S_TAG_VALUE}】【{namespace}】【{deploymentName}】未配置limit_mem_mb")
		utils.Logger.Info(fmt.Sprintf("admis:【%s】【%s】【%s】未配置limit_mem_mb", namespace, deploymentName, "PROM_K8S_TAG_VALUE"))
	}
	b, err = json.Marshal(resourcesObj)
	if err != nil {
		utils.Logger.Error("Error marshalling resourcesObj: ", zap.Error(err))
	}
	utils.Logger.Info(fmt.Sprintf("改后：%s", string(b)))

	// 构造资源更新的 patch
	resources := map[string]interface{}{
		"op":   "replace",
		"path": "/spec/template/spec/containers/0/resources",
	}
	resources["value"] = resourcesObj

	// 将更新操作编码为 JSON Patch 格式
	patchOps := []interface{}{patchReplicas, resources}
	patchJSONB, err := json.Marshal(patchOps)
	if err != nil {
		utils.Logger.Error("Error marshalling patch ops: ", zap.Error(err))
	}
	utils.Logger.Info(string(patchJSONB))
	// 创建 AdmissionReview 响应
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       types.UID(uid),
			Allowed:   true,
			PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
			Patch:     patchJSONB,
		},
	}
}

// admisMutateHandler 请求处理
func admisMutateHandler(w http.ResponseWriter, r *http.Request) {
	// 读取 r.Body 内容
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Logger.Error("Error reading body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	// 解析 AdmissionReview 对象
	var admissionReview admisv1.AdmissionReview
	err = json.Unmarshal(body, &admissionReview)
	if err != nil {
		utils.Logger.Error("Error unmarshalling body", zap.Error(err))
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// 获取 AdmissionRequest 体
	request := admissionReview.Request
	if request == nil {
		utils.Logger.Error("Invalid AdmissionRequest")
		http.Error(w, "Invalid AdmissionRequest", http.StatusBadRequest)
		return
	}
	// 正确示例，解码之后才能访问
	var deployObject v1.Deployment
	err = json.Unmarshal(request.Object.Raw, &deployObject)
	if err != nil {
		utils.Logger.Error("Error unmarshalling object:", zap.Error(err))
		http.Error(w, "Failed to parse object", http.StatusBadRequest)
		return
	}
	if request.OldObject.Raw == nil {
		utils.Logger.Info("create deployment skip admis")
		json.NewEncoder(w).Encode(admisPass(string(request.UID)))
		return
	}
	// 正确示例，解码之后才能访问
	var deployOldObject v1.Deployment
	err = json.Unmarshal(request.OldObject.Raw, &deployOldObject)
	if err != nil {
		utils.Logger.Error("Error unmarshalling Old object:", zap.Error(err))
		utils.Logger.Warn("Error unmarshalling Old object admissionReview", zap.String("body", string(body)))
		http.Error(w, "Failed to parse object", http.StatusBadRequest)
		return
	}
	object := deployObject
	oldObject := deployOldObject
	kind := request.Kind.Kind
	operation := request.Operation
	uid := string(request.UID)
	namespace := request.Namespace
	deploymentName := request.Name

	var (
		resultChan = make(chan *json.RawMessage, 10)
	)

	config.RequestFutures.Set(uid, resultChan)

	// Log收到请求
	utils.Logger.Info("Received admis request", zap.String("namespace", namespace), zap.String("deploymentName", deploymentName))

	// 如果 WebSocket 连接不可用，返回 503 错误
	if config.WebSocketConcent == nil {
		utils.Logger.Error("Failed to connect to kubedoor-master")
		http.Error(w, "Failed to connect to kubedoor-master", http.StatusServiceUnavailable)
		return
	}

	// 发送消息给 wsConn
	if err = config.WebSocketConcent.WriteJSON(map[string]interface{}{
		"type":       "admis",
		"request_id": uid,
		"namespace":  namespace,
		"deployment": deploymentName,
	}); err != nil {
		utils.Logger.Error("Failed to send message to kubedoor-master", zap.Error(err))
		json.NewEncoder(w).Encode(admisFail(uid, http.StatusInternalServerError, "Failed to send message to kubedoor-master"))
		return
	}
	// 等待响应
	var (
		podCount, podCountAI, podCountManual, requestCPUm, requestMemMB, limitCPUm, limitMemMB int
		cloudReplicas                                                                          int
	)

	select {
	case result := <-resultChan:
		utils.Logger.Info("resultChan->", zap.String("result", string(*result)))
		abcd := []interface{}{}
		if err := json.Unmarshal(*result, &abcd); err != nil {
			utils.Logger.Error("Failed to unmarshal requestRes", zap.Error(err))
			json.NewEncoder(w).Encode(admisFail(uid, http.StatusInternalServerError, "Failed to unmarshal requestRes"))
			return
		}
		if len(abcd) == 2 {
			if abcd[0].(float64) == 200 {
				utils.Logger.Info("Request handled successfully", zap.String("uid", uid))
				json.NewEncoder(w).Encode(admisPass(uid))
				return
			}
			// 失败
			utils.Logger.Error("Request failed", zap.Float64("status", abcd[0].(float64)), zap.String("message", abcd[1].(string)))
			json.NewEncoder(w).Encode(admisFail(uid, int32(abcd[0].(float64)), abcd[1].(string)))
			return
		}

		if len(abcd) == 7 || len(abcd) == 8 {
			var resIntSlice []int
			err := json.Unmarshal(*result, &resIntSlice)
			if err != nil {
				utils.Logger.Error(err.Error())
			}
			podCount = resIntSlice[0]
			podCountAI = resIntSlice[1]
			podCountManual = resIntSlice[2]
			requestCPUm = resIntSlice[3]
			requestMemMB = resIntSlice[4]
			limitCPUm = resIntSlice[5]
			limitMemMB = resIntSlice[6]
		}

	case <-time.After(60 * time.Second):
		// 超时
		utils.Logger.Error("Request timeout", zap.String("uid", uid))
		//http.Error(w, "Timeout waiting for response from kubedoor-master", http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(admisFail(uid, http.StatusGatewayTimeout, "Request timeout"))
		return
	}
	// 副本数取值优先级：专家建议→AI建议→原本的副本数
	if podCountManual >= 0 {
		cloudReplicas = podCountManual
	} else if podCountAI >= 0 {
		cloudReplicas = podCountAI
	} else {
		cloudReplicas = podCount
	}
	// 如果数据库中request_cpu_m为0，设置为30；如果request_mem_mb为0，设置为1
	if 0 <= requestCPUm && requestCPUm < 30 {
		utils.Logger.Info("request_cpu_m为0，设置为30")
		requestCPUm = 30
	}
	if requestMemMB == 0 {
		utils.Logger.Info("request_mem_mb为0，设置为1")
		requestMemMB = 1
	}
	utils.Logger.Info(fmt.Sprintf("副本数: %d, 请求CPU: %dm, 请求内存: %dMB, 限制CPU: %dm, 限制内存: %dMB\n",
		cloudReplicas, requestCPUm, requestMemMB, limitCPUm, limitMemMB))

	// 根据 kind 和 operation 执行具体逻辑
	switch {
	case kind == "Scale" && operation == "UPDATE":
		// 处理仅修改副本数的逻辑
		utils.Logger.Info("Received scale request, updating replicas", zap.String("namespace", namespace), zap.String("deploymentName", deploymentName))

		err = json.NewEncoder(w).Encode(scaleOnly(uid, int32(cloudReplicas)))
		if err != nil {
			utils.Logger.Error("err", zap.Error(err))
		}
		return

	case kind == "Deployment" && operation == "CREATE":
		// 处理创建请求并更新所有参数
		utils.Logger.Info("Received create request, updating all parameters", zap.String("namespace", namespace), zap.String("deploymentName", deploymentName))

		// 获取资源请求和限制
		err = json.NewEncoder(w).Encode(updateAll(cloudReplicas,
			namespace,
			deploymentName,
			requestCPUm,
			requestMemMB,
			limitCPUm,
			limitMemMB,
			object.Spec.Template.Spec.Containers[0].Resources,
			uid))
		if err != nil {
			utils.Logger.Error("err", zap.Error(err))
		}
		return

	case kind == "Deployment" && operation == "UPDATE":
		// 处理更新请求
		template := object.Spec.Template
		oldTemplate := oldObject.Spec.Template
		replicas := object.Spec.Replicas
		oldReplicas := oldObject.Spec.Replicas

		// 日志打印函数，减少代码重复
		logUpdateInfo := func(message, namespace, deploymentName string) {
			utils.Logger.Info(message, zap.String("namespace", namespace), zap.String("deploymentName", deploymentName))
		}

		// 比较 template 是否变化
		if fmt.Sprintf("%v", template) == fmt.Sprintf("%v", oldTemplate) {
			// spec.template 没变，检查副本数变化
			if replicas != oldReplicas {
				logUpdateInfo("Replicas updated, scaling only", namespace, deploymentName)
				json.NewEncoder(w).Encode(scaleOnly(uid, *replicas))
				return
			} else {
				logUpdateInfo("No changes in template or replicas, skipping", namespace, deploymentName)
			}
		} else {
			// spec.template 变了，触发重启逻辑，更新所有参数
			logUpdateInfo("Template changed, updating all parameters", namespace, deploymentName)

			resourcesDict := object.Spec.Template.Spec.Containers[0].Resources

			// 返回更新请求
			err := json.NewEncoder(w).Encode(updateAll(
				cloudReplicas,
				namespace,
				deploymentName,
				requestCPUm,
				requestMemMB,
				limitCPUm,
				limitMemMB,
				resourcesDict,
				uid,
			))

			if err != nil {
				utils.Logger.Error("Failed to encode update request", zap.Error(err))
			}
			return
		}
	default:
		utils.Logger.Info("Unknown operation or kind, skipping", zap.String("namespace", namespace), zap.String("deploymentName", deploymentName))
	}

	// 成功响应时
	json.NewEncoder(w).Encode(admisPass(uid))
	return
}

// admisPass 处理允许的请求（类似于 Python 中的 admis_pass 函数）
func admisPass(uid string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     types.UID(uid),
			Allowed: true,
		},
	}
}

// admisFail 处理拒绝的请求（类似于 Python 中的 admis_fail 函数）
func admisFail(uid string, code int32, message string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     types.UID(uid),
			Allowed: false,
			Result: &metav1.Status{
				Code:    code,
				Message: message,
			},
		},
	}
}

func scaleOnly(uid string, replicas int32) *admissionv1.AdmissionReview {
	// 仅修改副本数，不重启
	patchReplicas := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/replicas",
			"value": replicas,
		},
	}

	// 将 patch 转换为 JSON
	patchJSON, err := json.Marshal(patchReplicas)
	if err != nil {
		return &admissionv1.AdmissionReview{}
	}

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     types.UID(uid),
			Allowed: true,
			PatchType: func() *admissionv1.PatchType {
				pt := admissionv1.PatchTypeJSONPatch
				return &pt
			}(),
			Patch: patchJSON,
		},
	}
}
