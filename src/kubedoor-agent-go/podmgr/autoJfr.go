package podmgr

import (
	"fmt"
	"github.com/google/uuid"
	"kubedoor-agent-go/config"
	"net/http"
	"os"
	"time"
)

// 处理自动 JFR 请求
func autoJFRHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	podName := r.URL.Query().Get("pod_name")
	env := r.URL.Query().Get("env")

	// Start JFR recording
	fileName := fmt.Sprintf("%s-%s-%s.jfr", env, ns, podName)
	command := fmt.Sprintf("jcmd 1 JFR.start duration=5m filename=/%s", fileName)
	status, _ := ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to start JFR", http.StatusInternalServerError)
		return
	}

	// Upload JFR file to OSS
	dlurl := fmt.Sprintf("%s/%s/jfr/%s", os.Getenv("OSS_URL"), env, fileName)
	command = fmt.Sprintf("curl -s -T /%s %s", fileName, dlurl)
	status, _ = ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to upload JFR file", http.StatusInternalServerError)
		return
	}

	// Create background task to upload JFR file
	taskID := uuid.New().String()
	config.TaskManagerObj.UpdateTask(taskID, &config.TaskResult{
		Status: "processing",
	})
	go func() {
		time.Sleep(2 * time.Minute)
		config.TaskManagerObj.UpdateTask(taskID, &config.TaskResult{
			Status: "completed",
			File:   dlurl,
		})

	}()

	sendMd("JFR file uploaded successfully", env, ns, podName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("JFR file available at %s", dlurl)))
}
