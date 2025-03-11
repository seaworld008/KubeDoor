package main

import (
	"kubedoor-agent-go/agent"
	"kubedoor-agent-go/api"
	_ "kubedoor-agent-go/config"
	_ "kubedoor-agent-go/k8sSet"
	"kubedoor-agent-go/utils"
)

func main() {

	utils.InitLogger() // Initialize logger
	go func() {
		api.StartAPI()
	}()
	//// Start PodMgr server in a goroutine
	//go func() {
	//	podmgr.StartPodMgr()
	//}()
	agent.StartAgent()
}
