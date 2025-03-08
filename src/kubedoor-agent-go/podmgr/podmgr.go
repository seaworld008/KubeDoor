package podmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"kubedoor-agent-go/utils" // Use module import path

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		utils.Logger.Warn("Failed to load incluster config, trying kubeconfig", zap.Error(err))
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

func StartPodMgr() {
	router := mux.NewRouter()

	router.HandleFunc("/api/pod/modify_pod", modifyPod).Methods("GET")
	router.HandleFunc("/api/pod/delete_pod", deletePod).Methods("GET")
	router.HandleFunc("/api/pod/auto_dump", autoDump).Methods("GET")
	router.HandleFunc("/api/pod/auto_jstack", autoJstack).Methods("GET")
	router.HandleFunc("/api/pod/auto_jfr", autoJfr).Methods("GET")
	router.HandleFunc("/api/pod/auto_jvm_mem", autoJvmMem).Methods("GET")
	router.HandleFunc("/api/pod/task_status/{task_id}", getTaskStatus).Methods("GET")

	// http.Handle("/", router)

	port := os.Getenv("PODMGR_PORT")
	if port == "" {
		port = "81" // Default port for podmgr
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

func modifyPod(w http.ResponseWriter, r *http.Request) {
	env := r.URL.Query().Get("env")
	ns := r.URL.Query().Get("ns")
	podName := r.URL.Query().Get("pod_name")

	if env == "" || ns == "" || podName == "" {
		utils.Logger.Error("Missing required parameters for modifyPod")
		http.Error(w, "Missing required parameters: env, ns, pod_name", http.StatusBadRequest)
		return
	}

	success, status := modifyPodLabel(r.Context(), ns, podName)
	if !success {
		utils.Logger.Error("Failed to modify pod label", zap.String("namespace", ns), zap.String("podName", podName), zap.String("status", status))
		http.Error(w, status, http.StatusInternalServerError)
		return
	}

	sendMd("app label modified successfully", env, ns, podName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("【%s】【%s】app label modified successfully", ns, podName),
		"success": true,
	})
}

func modifyPodLabel(ctx context.Context, ns string, podName string) (bool, string) {
	utils.Logger.Info("===开始修改标签", zap.String("namespace", ns), zap.String("podName", podName))

	podData, err := kubeClient.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
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
	newLabelValue := labelsApp + "-ALERT"
	currentLabels[isolateLabel] = newLabelValue

	podData.ObjectMeta.Labels = currentLabels
	_, err = kubeClient.CoreV1().Pods(ns).Patch(ctx, podName, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"%s":"%s"}}}`, isolateLabel, newLabelValue)), metav1.PatchOptions{})
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

func deletePod(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement deletePod handler
	utils.Logger.Info("deletePod handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "deletePod handler not implemented yet"})
}

func autoDump(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement autoDump handler
	utils.Logger.Info("autoDump handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "autoDump handler not implemented yet"})
}

func autoJstack(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement autoJstack handler
	utils.Logger.Info("autoJstack handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "autoJstack handler not implemented yet"})
}

func autoJfr(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement autoJfr handler
	utils.Logger.Info("autoJfr handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "autoJfr handler not implemented yet"})
}

func autoJvmMem(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement autoJvmMem handler
	utils.Logger.Info("autoJvmMem handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "autoJvmMem handler not implemented yet"})
}

func getTaskStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement getTaskStatus handler
	utils.Logger.Info("getTaskStatus handler not implemented yet")
	json.NewEncoder(w).Encode(map[string]string{"message": "getTaskStatus handler not implemented yet"})
}
