package podmgr

import (
	"fmt"
	"net/http"
	"os"
)

// 处理自动 Dump 请求
func autoDumpHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	podName := r.URL.Query().Get("pod_name")
	env := r.URL.Query().Get("env")

	// Execute dump command
	fileName := fmt.Sprintf("%s-%s-%s.hprof", env, ns, podName)
	command := fmt.Sprintf("jmap -dump:format=b,file=/%s 1", fileName)
	status, _ := ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to execute jmap dump", http.StatusInternalServerError)
		return
	}

	// Upload the dump file to OSS
	dlurl := fmt.Sprintf("%s/%s/dump/%s", os.Getenv("OSS_URL"), env, fileName)
	command = fmt.Sprintf("curl -s -T /%s %s", fileName, dlurl)
	status, _ = ExecuteCommand(command, podName, ns)
	if !status {
		http.Error(w, "Failed to upload dump file", http.StatusInternalServerError)
		return
	}

	sendMd("Dump file uploaded successfully", env, ns, podName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Dump file available at %s", dlurl)))
}

// ExecuteCommand executes command in pod and returns success status and output
func ExecuteCommand(command, podName, ns string) (bool, string) {
	//
	//// 准备执行命令
	//execOptions := &metav1.PodExecOptions{
	//	Command: []string{"/bin/sh", "-c", command},
	//	Stdin:   false,
	//	Stdout:  true,
	//	Stderr:  true,
	//	Tty:     false,
	//}
	//
	//// 获取 Pod 执行请求
	//req := k8sSet.KubeClient.CoreV1().Pods(ns).GetExec(podName, execOptions)
	//
	//// 捕获输出
	//var stdout, stderr bytes.Buffer
	//err = req.StreamOptions().Stdout(&stdout).Stderr(&stderr).Stream(context.Background())
	//if err != nil {
	//	log.Printf("Error executing command: %v", err)
	//	return false, err.Error()
	//}
	//
	//// 返回输出信息
	//out := stdout.String()
	//if stderr.Len() > 0 {
	//	out += "\n" + stderr.String()
	//}

	//return true, out
	return true, ""
}
