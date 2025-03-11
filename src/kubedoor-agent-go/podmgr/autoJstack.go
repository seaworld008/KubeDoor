package podmgr

import (
	"fmt"
	"net/http"
	"os"
)

// 处理自动 JStack 请求
func autoJStackHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	podName := r.URL.Query().Get("pod_name")
	env := r.URL.Query().Get("env")

	// Execute jstack command
	fileName := fmt.Sprintf("%s-%s-%s.jstack", env, ns, podName)
	command := fmt.Sprintf("jstack -l 1 |tee /%s", fileName)
	status, _ := ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to execute jstack", http.StatusInternalServerError)
		return
	}

	// Upload the jstack file to OSS
	dlurl := fmt.Sprintf("%s/%s/jstack/%s", os.Getenv("OSS_URL"), env, fileName)
	command = fmt.Sprintf("curl -s -T /%s %s", fileName, dlurl)
	status, _ = ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to upload jstack file", http.StatusInternalServerError)
		return
	}

	sendMd("Jstack file uploaded successfully", env, ns, podName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Jstack file available at %s", dlurl)))
}
