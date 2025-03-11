package api

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/k8sSet"
	"kubedoor-agent-go/utils"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScheduleServiceScaleRestart 入口方法
func ScheduleServiceScaleRestart(operatorType string, data config.BodyTimeScaleRestartStruct) (response map[string]interface{}) {
	// 返回结果
	response = map[string]interface{}{
		"message": "ok",
	}
	if data.Cron == "" {
		// 解析时间并转换为 Cron 表达式
		cronTimeExpr, err := getCronExpressionForOneTimeTask(data.Time)
		if err != nil {
			utils.Logger.Error("Failed to parse time", zap.Error(err))
			return
		}
		if err = createK8sCronJobForScheduledRestart(data.Service, cronTimeExpr, operatorType, "once"); err != nil {
			utils.Logger.Error("Failed to create Kubernetes CronJob", zap.Error(err))
			response["errorList"] = []string{err.Error()}
		}
	} else {
		if err := createK8sCronJobForScheduledRestart(data.Service, data.Cron, operatorType, "cron"); err != nil {
			utils.Logger.Error("Failed to create Kubernetes CronJob", zap.Error(err))
			response["errorList"] = []string{err.Error()}
		}
	}
	if response["errorList"] != nil {
		response["message"] = "err"
	}
	return response
}

// getCronExpressionForOneTimeTask 将时间转换为 Cron 表达式
func getCronExpressionForOneTimeTask(timeData interface{}) (string, error) {
	b, err := json.Marshal(timeData)
	if err != nil {
		return "", err
	}

	scheduleTime, err := time.Parse("[2006,1,2,15,4]", string(b))
	if err != nil {
		return "", err
	}

	now := time.Now()
	if now.After(scheduleTime) {
		return "", fmt.Errorf("The scheduled time is in the past")
	}

	return fmt.Sprintf("%d %d %d %d *", scheduleTime.Minute(), scheduleTime.Hour(), scheduleTime.Day(), scheduleTime.Month()), nil
}

// createK8sCronJobForScheduledRestart 创建 Kubernetes CronJob
func createK8sCronJobForScheduledRestart(services []config.BodyScaleRestartStruct, cronExpr string, operatorType, cronType string) (err error) {
	jsonData, err := json.Marshal(services)
	if err != nil {
		utils.Logger.Error("Failed to marshal services", zap.Error(err))
		return
	}
	jobNs := "kubedoor"
	for _, service := range services {
		cronJob := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%s", operatorType, cronType, service.DeploymentName),
				Namespace: jobNs,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: cronExpr,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{
									{
										Name:  "restart-container",
										Image: "registry.cn-shenzhen.aliyuncs.com/starsl/busybox-curl",
										Command: []string{
											"curl",
											"-s",
											"-X",
											"POST",
											"-H",
											"Content-Type: application/json",
											"-d",
											string(jsonData),
											fmt.Sprintf("http://kubedoor-agent.kubedoor/api/%s", operatorType),
										},
									},
								},
							},
						},
					},
				},
			},
		}

		existingCronJob, err := k8sSet.KubeClient.BatchV1().CronJobs(jobNs).Get(context.TODO(), cronJob.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				// 如果 CronJob 不存在，则创建
				_, err = k8sSet.KubeClient.BatchV1().CronJobs(jobNs).Create(context.TODO(), cronJob, metav1.CreateOptions{})
				if err != nil {
					utils.Logger.Error("Failed to create Kubernetes CronJob", zap.Error(err))
					continue
				} else {
					utils.Logger.Info(fmt.Sprintf("Kubernetes CronJob created for %s in namespace %s", cronJob.Name, jobNs))
				}
			} else {
				// 其他错误
				utils.Logger.Error("Failed to check Kubernetes CronJob existence", zap.Error(err))
			}
			continue
		}

		// 如果 CronJob 已存在，则更新
		cronJob.ResourceVersion = existingCronJob.ResourceVersion // 重要：保留 ResourceVersion 以防冲突
		_, err = k8sSet.KubeClient.BatchV1().CronJobs(jobNs).Update(context.TODO(), cronJob, metav1.UpdateOptions{})
		if err != nil {
			utils.Logger.Error("Failed to update Kubernetes CronJob", zap.Error(err))
			continue
		} else {
			utils.Logger.Info(fmt.Sprintf("Kubernetes CronJob updated for %s in namespace %s", cronJob.Name, jobNs))
		}
	}
	return
}

// deleteK8sCronJobForScheduledRestart 执行完定时任务后执行删除操作
func deleteK8sCronJobForScheduledRestart(cronJobName string) {
	jobNs := "kubedoor"
	err := k8sSet.KubeClient.BatchV1().CronJobs(jobNs).Delete(context.TODO(), cronJobName, metav1.DeleteOptions{})
	if err != nil {
		utils.Logger.Error("Failed to delete Kubernetes CronJob", zap.String("cronJobName", cronJobName), zap.Error(err))
		return
	}
	utils.Logger.Info(fmt.Sprintf("Kubernetes CronJob %s deleted in namespace %s", cronJobName, jobNs))
	return
}
