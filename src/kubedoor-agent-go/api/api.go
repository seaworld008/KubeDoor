package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils" // Import the utils module
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClient *kubernetes.Clientset

func init() {
	utils.InitLogger() // Initialize logger
	kubeClient = initKubeClient()
}

func initKubeClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		utils.Logger.Warn("Failed to load incluster config", zap.Error(err))
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			utils.Logger.Fatal("Failed to load kubeconfig", zap.Error(err))
			log.Fatalf("Failed to load kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		utils.Logger.Fatal("Failed to create kubernetes client", zap.Error(err))
		log.Fatalf("Failed to create kubernetes client: %v", err)
	}

	utils.Logger.Info("Successfully initialized Kubernetes client")
	return clientset
}

func StartAPI() {
	router := mux.NewRouter()

	router.HandleFunc("/api/health", getHealthCheck).Methods("GET")
	router.HandleFunc("/api/restart", modifyRestart).Methods("POST")
	router.HandleFunc("/api/scale", modifyScale).Methods("POST")

	http.Handle("/", router)

	port := os.Getenv("PODMGR_PORT")
	if port == "" {
		port = "80" // Default port for podmgr
	}

	utils.Logger.Info("Starting PodMgr server", zap.String("port", port))
	log.Printf("PodMgr server listening on port %s", port)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		utils.Logger.Error("PodMgr server failed to start", zap.Error(err))
		log.Fatalf("PodMgr server failed to start: %v", err)
	}
}

func modifyRestart(w http.ResponseWriter, r *http.Request) {
	var (
		objDatas []config.BodyScaleRestartStruct
	)
	err := json.NewDecoder(r.Body).Decode(&objDatas)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	response := RestartDeployment(objDatas, true)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func modifyScale(w http.ResponseWriter, r *http.Request) {
	var (
		objDatas []config.BodyScaleRestartStruct
	)
	err := json.NewDecoder(r.Body).Decode(&objDatas)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	response := ScaleDeployment(objDatas, true)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func getHealthCheck(w http.ResponseWriter, r *http.Request) {

	version := utils.GetVersion()
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"ver":    version,
		"status": "healthy",
	}
	json.NewEncoder(w).Encode(response)
}

func HealthCheck() map[string]interface{} {
	version := utils.GetVersion()
	return map[string]interface{}{
		"ver":    version,
		"status": "healthy",
	}
}

func UpdateImage(objData config.BodyUpdateImageStruct) (responseData map[string]interface{}) {
	responseData = map[string]interface{}{
		"success": false,
	}
	ctx := context.Background()
	newImageTag := objData.ImageTag
	deploymentName := objData.Deployment
	namespace := objData.Namespace

	// 获取 Deployment 资源
	deployment, err := kubeClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		utils.Logger.Error("Failed to get deployment", zap.Error(err), zap.String("namespace", namespace), zap.String("deployment", deploymentName))
		responseData["error"] = fmt.Sprintf("Failed to get Deployment %s in namespace %s", deploymentName, namespace)
		return
	}

	// 确保 Deployment 至少有一个容器
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		utils.Logger.Error("Deployment has no containers", zap.String("namespace", namespace), zap.String("deployment", deploymentName))
		responseData["error"] = "Deployment has no containers to update"
		return
	}

	// 只修改第一个容器的镜像 tag
	container := &deployment.Spec.Template.Spec.Containers[0]
	imageParts := strings.Split(container.Image, ":")
	imageName := imageParts[0] // 仅取镜像名，去掉旧 tag
	newImage := fmt.Sprintf("%s:%s", imageName, newImageTag)

	utils.Logger.Info("Updating container image",
		zap.String("container", container.Name),
		zap.String("oldImage", container.Image),
		zap.String("newImage", newImage),
	)

	// 直接更新 Deployment 对象
	container.Image = newImage

	_, err = kubeClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		utils.Logger.Error("Failed to update deployment", zap.Error(err), zap.String("namespace", namespace), zap.String("deployment", deploymentName))
		responseData["error"] = fmt.Sprintf("Failed to update Deployment %s in namespace %s", deploymentName, namespace)
		return
	}

	utils.Logger.Info("Deployment image updated successfully", zap.String("namespace", namespace), zap.String("deployment", deploymentName), zap.String("newImage", newImage))
	responseData["success"] = true
	responseData["message"] = fmt.Sprintf("Deployment %s in namespace %s updated with new images", deploymentName, namespace)
	return
}

func RestartDeployment(objDatas []config.BodyScaleRestartStruct, isAPI bool) (responseData map[string]interface{}) {
	errorList := []map[string]string{}
	// 返回结果
	response := map[string]interface{}{
		"message": "ok",
	}
	ctx := context.Background()
	for _, deploymentInfo := range objDatas {
		namespace := deploymentInfo.Namespace
		deploymentName := deploymentInfo.DeploymentName
		// 获取 Deployment
		deployment, err := kubeClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			utils.Logger.Error("Failed to get deployment", zap.Error(err),
				zap.String("namespace", namespace),
				zap.String("deployment", deploymentName),
			)
			errorList = append(errorList, map[string]string{
				"namespace":       namespace,
				"deployment_name": deploymentName,
				"error":           err.Error(),
			})
			continue
		}

		// 修改 Deployment 的标签，触发 Pod 重启
		newLabelValue := fmt.Sprintf("%d", time.Now().UnixNano()) // 用当前时间戳作为新标签的值
		deployment.Spec.Template.ObjectMeta.Labels["restartTimestamp"] = newLabelValue

		// 更新 Deployment
		_, err = kubeClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			utils.Logger.Error("Failed to update deployment for restart", zap.Error(err),
				zap.String("namespace", namespace),
				zap.String("deployment", deploymentName),
			)
			errorList = append(errorList, map[string]string{
				"namespace":       namespace,
				"deployment_name": deploymentName,
				"error":           err.Error(),
			})
			continue
		}

		utils.Logger.Info("Successfully restarted deployment", zap.String("namespace", namespace),
			zap.String("deployment", deploymentName), zap.String("newLabelValue", newLabelValue),
		)

		if isAPI {
			cronJobName := fmt.Sprintf("%s-%s-%s", "restart", "once", deploymentName)
			deleteK8sCronJobForScheduledRestart(cronJobName)
		}
	}

	if len(errorList) > 0 {
		response["message"] = "err"
		response["errorList"] = errorList
	}
	return response
}

// ScaleDeployment 处理多个 Deployment 的缩放请求
func ScaleDeployment(objDatas []config.BodyScaleRestartStruct, isAPI bool) (responseData map[string]interface{}) {
	errorList := []map[string]string{}
	// 返回结果
	response := map[string]interface{}{
		"message": "ok",
	}
	ctx := context.Background()
	for _, deploymentInfo := range objDatas {
		namespace := deploymentInfo.Namespace
		deploymentName := deploymentInfo.DeploymentName
		replaceNum := int32(deploymentInfo.Num)
		// 获取 Deployment
		deployment, err := kubeClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			utils.Logger.Error("Failed to get deployment", zap.Error(err),
				zap.String("namespace", namespace),
				zap.String("deployment", deploymentName),
			)
			errorList = append(errorList, map[string]string{
				"namespace":       namespace,
				"deployment_name": deploymentName,
				"error":           err.Error(),
			})
			continue
		}

		deployment.Spec.Replicas = &replaceNum

		// 更新 Deployment
		_, err = kubeClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			utils.Logger.Error("Failed to scale deployment", zap.Error(err), zap.String("namespace", namespace), zap.String("deployment", deploymentName))
			errorDetail := map[string]string{
				"namespace":       namespace,
				"deployment_name": deploymentName,
				"error":           err.Error(),
			}
			errorList = append(errorList, errorDetail)
		} else {
			utils.Logger.Info("Successfully scaled deployment", zap.String("namespace", namespace), zap.String("deployment", deploymentName), zap.Int32("num", replaceNum))
		}
		if isAPI {
			cronJobName := fmt.Sprintf("%s-%s-%s", "scale", "once", deploymentName)
			deleteK8sCronJobForScheduledRestart(cronJobName)
		}
	}

	if len(errorList) > 0 {
		response["message"] = "err"
		response["errorList"] = errorList
	}
	return response
}
