package k8sSet

import (
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils"

	"log"
	"os"
)

func init() {
	utils.InitLogger() // Initialize logger
	config.KubeClient = initKubeClient()
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
