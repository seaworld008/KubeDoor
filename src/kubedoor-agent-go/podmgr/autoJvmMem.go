package podmgr

import (
	"fmt"
	"os"

	"net/http"
)

// 处理自动 JVM 内存请求
func autoJvmMemHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	podName := r.URL.Query().Get("pod_name")
	env := r.URL.Query().Get("env")

	// Execute jcmd memory command
	fileName := fmt.Sprintf("%s-%s-%s.mem", env, ns, podName)
	command := fmt.Sprintf("jcmd 1 VM.native_memory summary output=/%s", fileName)
	status, _ := ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to collect JVM memory statistics", http.StatusInternalServerError)
		return
	}
	// Upload the memory file to OSS
	dlurl := fmt.Sprintf("%s/%s/jvm_mem/%s", os.Getenv("OSS_URL"), env, fileName)
	command = fmt.Sprintf("curl -s -T /%s %s", fileName, dlurl)
	status, _ = ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to upload JVM memory file", http.StatusInternalServerError)
		return
	}

	sendMd("JVM memory statistics file uploaded successfully", env, ns, podName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("JVM memory file available at %s", dlurl)))
}
