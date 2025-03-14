package podmgr

import (
	"bytes"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"kubedoor-agent-go/config"
	"os"
)

func executeCom() {
	inConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get in-cluster config: ", err)
		return
	}
	// 执行命令并获取结果
	_, err = executeCommandInPod(config.KubeClient, inConfig, "default", "your-pod", "your-container", "ls -l > /a.log")
	if err != nil {
		fmt.Println("Error executing command: ", err)
		return
	}

	// 下载文件
	err = downloadFileFromPod(config.KubeClient, inConfig, "default", "your-pod", "your-container", "/a.log", "./a.log")
	if err != nil {
		fmt.Println("Error downloading file: ", err)
		return
	}

	// 输出文件下载成功信息
	fmt.Println("File downloaded successfully to ./a.log")
}

func executeCommandInPod(clientset *kubernetes.Clientset, config *rest.Config, namespace, podName, containerName, command string) (string, error) {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"/bin/sh", "-c", command},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func downloadFileFromPod(clientset *kubernetes.Clientset, config *rest.Config, namespace, podName, containerName, filePath, localPath string) error {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"cat", filePath},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return fmt.Errorf("file read failed: %v, stderr: %s", err, stderr.String())
	}

	if err := os.WriteFile(localPath, stdout.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write local file: %v", err)
	}

	return nil
}
