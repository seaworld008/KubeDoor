package podmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils" // Use module import path
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//var kubeClient *kubernetes.Clientset
//
//func init() {
//	utils.InitLogger() // Initialize logger
//	kubeClient = initKubeClient()
//}
//
//func initKubeClient() *kubernetes.Clientset {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		utils.Logger.Warn("Failed to load incluster config, trying kubeconfig", zap.Error(err))
//		kubeconfig := os.Getenv("KUBECONFIG")
//		if kubeconfig == "" {
//			kubeconfig = os.Getenv("HOME") + "/.kube/config"
//		}
//		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
//		if err != nil {
//			utils.Logger.Fatal("Failed to load kubeconfig", zap.Error(err))
//			log.Fatalf("Failed to load kubeconfig: %v", err)
//		}
//	}
//
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		utils.Logger.Fatal("Failed to create kubernetes client", zap.Error(err))
//		log.Fatalf("Failed to create kubernetes client: %v", err)
//	}
//
//	utils.Logger.Info("Successfully initialized Kubernetes client")
//	return clientset
//}

func ModifyPod(queryData map[string]interface{}) (response map[string]interface{}) {
	env := queryData["env"].(string)
	ns := queryData["ns"].(string)
	podName := queryData["pod_name"].(string)
	ctx := context.Background()
	if env == "" || ns == "" || podName == "" {
		utils.Logger.Error("Missing required parameters for modifyPod")
		response = map[string]interface{}{
			"message": fmt.Sprintf("【%s】【%s】 Missing required parameters for modifyPod", ns, podName),
			"success": false,
		}
		return
	}

	success, status := modifyPodLabel(ctx, ns, podName)
	if !success {
		utils.Logger.Error("Failed to modify pod label", zap.String("namespace", ns), zap.String("podName", podName), zap.String("status", status))
		response = map[string]interface{}{
			"message": fmt.Sprintf("【%s】【%s】app label modified error", ns, podName),
			"success": false,
		}
		return
	}
	if status == "Isolated" {
		response = map[string]interface{}{
			"message": fmt.Sprintf("【%s】【%s】Isolated", ns, podName),
			"success": true,
		}
		return
	}
	go sendMd("app label modified successfully", env, ns, podName)

	response = map[string]interface{}{
		"message": fmt.Sprintf("【%s】【%s】app label modified successfully", ns, podName),
		"success": true,
	}
	return
}

func modifyPodLabel(ctx context.Context, ns string, podName string) (bool, string) {
	utils.Logger.Info("===开始修改标签", zap.String("namespace", ns), zap.String("podName", podName))

	podData, err := config.KubeClient.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		utils.Logger.Error("Failed to get pod", zap.Error(err), zap.String("namespace", ns), zap.String("podName", podName))
		return false, "===pod_not_found"
	}
	currentLabels := podData.ObjectMeta.Labels
	if currentLabels == nil {
		currentLabels = make(map[string]string)
	}

	isolateLabel := getPodIsolateLabel(podName)
	labelsApp := currentLabels[isolateLabel]
	if labelsApp == "" {
		return false, "===app_label_not_found"
	}
	if strings.HasSuffix(labelsApp, "-ALERT") {
		return true, "Isolated"
	}
	newLabelValue := labelsApp + "-ALERT"
	currentLabels[isolateLabel] = newLabelValue

	podData.ObjectMeta.Labels = currentLabels
	_, err = config.KubeClient.CoreV1().Pods(ns).Patch(ctx, podName, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"%s":"%s"}}}`, isolateLabel, newLabelValue)), metav1.PatchOptions{})
	if err != nil {
		utils.Logger.Error("Failed to patch pod label", zap.Error(err), zap.String("namespace", ns), zap.String("podName", podName))
		return false, "===modify_label_failed"
	}

	return true, ""
}

func getPodIsolateLabel(podName string) string {
	return "app" // Fixed label key
}

func sendMd(msg string, env string, ns string, podName string) {
	text := fmt.Sprintf("# 【<font color=\"#5bcc85\">%s</font>】%s\n## %s\n", env, ns, podName)
	text += fmt.Sprintf("%s\n", msg)
	utils.SendMsg(os.Getenv("MSG_TYPE"), os.Getenv("MSG_TOKEN"), text)
}

func DeletePodHandler(queryData map[string]interface{}) (response map[string]interface{}) {
	//errorList := []map[string]string{}
	env := queryData["env"].(string)
	ns := queryData["ns"].(string)
	podName := queryData["pod_name"].(string)
	success := deletePod(ns, podName)
	if !success {
		utils.Logger.Error("pod deleted error ", zap.String("namespace", ns), zap.String("podName", podName))
		response = map[string]interface{}{
			"message": fmt.Sprintf("【%s】【%s】pod deleted error", ns, podName),
			"success": false,
		}
		return
	}
	sendMd("pod deleted successfully", env, ns, podName)
	response = map[string]interface{}{
		"message": fmt.Sprintf("【%s】【%s】pod deleted successfully", ns, podName),
		"success": true,
	}
	return
}

func deletePod(ns string, podName string) bool {

	err := config.KubeClient.CoreV1().Pods(ns).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting pod: %v", err)
		return false
	}

	return true
}

func GetTaskStatus(taskId string) (response map[string]interface{}) {
	if result, exists := config.TaskManagerObj.GetTask(taskId); exists {
		json.Unmarshal(result, &response)
		return
	}
	return map[string]interface{}{
		"status": "not found",
	}
}
