package api

import (
	"context"
	"go.uber.org/zap"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubedoor-agent-go/config"
	"kubedoor-agent-go/utils"
)

// Get the MutatingWebhookConfiguration
func getMutatingWebhook() map[string]interface{} {

	admissionAPI := config.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations()

	webhookName := "kubedoor-admis-configuration"
	webhook, err := admissionAPI.Get(context.Background(), webhookName, metav1.GetOptions{})
	if err != nil {
		if err.Error() == "mutatingwebhookconfigurations.admissionregistration.k8s.io \""+webhookName+"\" not found" {
			utils.Logger.Info("MutatingWebhookConfiguration does not exist")
			return map[string]interface{}{"is_on": false}
		}
		utils.Logger.Error("Error reading MutatingWebhookConfiguration", zap.Error(err))
		return map[string]interface{}{"message": "Error reading MutatingWebhookConfiguration", "success": false, "status": 500}
	}

	utils.Logger.Info("MutatingWebhookConfiguration found", zap.String("name", webhook.Name))
	return map[string]interface{}{"is_on": true, "success": true, "status": 200}
}

// Create the MutatingWebhookConfiguration
func createMutatingWebhook() map[string]interface{} {

	admissionAPI := config.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations()

	webhookName := "kubedoor-admis-configuration"
	namespace := "kubedoor"

	webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{

		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				AdmissionReviewVersions: []string{"v1"},
				Name:                    "kubedoor-admis.mutating.webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: namespace,
						Name:      "kubedoor-agent",
						Path: func() *string {
							pt := "/api/admis"
							return &pt
						}(),
						Port: func() *int32 {
							pt := int32(443)
							return &pt
						}(),
					},
					CABundle: []byte("-----BEGIN CERTIFICATE-----\nMIIDITCCAgmgAwIBAgIJAI5Ow/BtqHBiMA0GCSqGSIb3DQEBCwUAMCYxJDAiBgNV\nBAMMG2t1YmVkb29yLWFnZW50Lmt1YmVkb29yLnN2YzAgFw0yNTAzMTAwMzI0Mzla\nGA8yMTI1MDIxNDAzMjQzOVowJjEkMCIGA1UEAwwba3ViZWRvb3ItYWdlbnQua3Vi\nZWRvb3Iuc3ZjMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvcsqgBov\ndZjplW7TS8qiJqhLVn5vxU7kZ7bBE+VgC462PrJnTFN9C8/tmr+HN7QJibplVA0A\nz1FMjQcvO3Gcbybo1z6oKaJmu1Igdlk1csa8I2Qw7/Odt3e/LpohxbikIdKs76gx\nr5ZDiFV1U9eS13fiVdM3/8cv0j+8zhFrFwQiJye4YmaNdPSFP1mRn5bz0o3Ne/SZ\np0xvmMcLU1Ac8sjQmODLh15XF45ue9/4676BZ4PI1Y1fgXvGw4CLPZg9D8+crwWW\n5mhYWe7MUdx1uqnn0KcF77tB7Yr/8G3Oi7JSZf+hb+PYRXx0UjE78E09sWsEecKE\n5T5U8+c29FeJTwIDAQABo1AwTjAdBgNVHQ4EFgQU/oOFa1haaLCt6tsGOAp+Q53T\nFnkwHwYDVR0jBBgwFoAU/oOFa1haaLCt6tsGOAp+Q53TFnkwDAYDVR0TBAUwAwEB\n/zANBgkqhkiG9w0BAQsFAAOCAQEAILkLox0j93R9SngqYRnlEQn74uGLSbBz54/w\nI7IUZxutKYsbCWtTSpk/IqZuYoAf4Y6411YELJ2cre3tU9oZlDmqLZRX+IZQuKjI\nebtC/oQC/bjfgPQE9q7hG0kIch4xE/xWW2M/zF0whNCxkn5TVcOTN8Sm9wfO3XYr\nYjoXOFO2tUf0Qa+Iv0uqbRpfySPSsDX1TzAfP3wxGbrBp+q4P0Y8/HCi9cVQXFbK\nfOGiQF/dbxtgemmdN/rwdllVhJPK3dFDybgNXYK7SWFkVIDur9Zm1jaIsYb4Bvn0\nVNf4ZyS6QE8IRO1NQ3kfXd6Nk3S8w6z2iQL7zbs7VqNJqrT1yQ==\n-----END CERTIFICATE-----"),
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"apps"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments", "deployments/scale"},
						},
					},
				},
				FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
					pt := admissionregistrationv1.Fail
					return &pt
				}(),
				SideEffects: func() *admissionregistrationv1.SideEffectClass {
					se := admissionregistrationv1.SideEffectClassNone
					return &se
				}(),
				MatchPolicy: func() *admissionregistrationv1.MatchPolicyType {
					se := admissionregistrationv1.Equivalent
					return &se
				}(),
			},
		},
	}

	// Create the webhook
	_, err := admissionAPI.Create(context.Background(), webhookConfig, metav1.CreateOptions{})
	if err != nil {
		utils.Logger.Error("Error creating MutatingWebhookConfiguration", zap.Error(err))
		return map[string]interface{}{"message": err.Error(), "success": false, "status": 500}
	}

	utils.Logger.Info("MutatingWebhookConfiguration created successfully")
	return map[string]interface{}{"message": "Created successfully", "success": true, "status": 200}
}

// Delete the MutatingWebhookConfiguration
func deleteMutatingWebhook() map[string]interface{} {

	admissionAPI := config.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations()

	webhookName := "kubedoor-admis-configuration"
	err := admissionAPI.Delete(context.Background(), webhookName, metav1.DeleteOptions{})
	if err != nil {
		utils.Logger.Error("Error deleting MutatingWebhookConfiguration", zap.Error(err))
		return map[string]interface{}{"message": err.Error(), "success": false, "status": 500}
	}

	utils.Logger.Info("MutatingWebhookConfiguration deleted successfully")
	return map[string]interface{}{"message": "Deleted successfully", "success": true, "status": 200}
}

// AdmisSwitch function to handle the webhook switch
func AdmisSwitch(queryData map[string]interface{}) (response map[string]interface{}) {
	action := queryData["action"].(string)
	res := getMutatingWebhook()

	switch action {
	case "get":
		return res
	case "on":
		if res["is_on"].(bool) == true {
			return map[string]interface{}{"message": "Webhook is already opened!", "success": true, "status": 200}
		}
		return createMutatingWebhook()
	case "off":
		if res["is_on"].(bool) == false {
			return map[string]interface{}{"message": "Webhook is already closed!", "success": true, "status": 200}
		}
		return deleteMutatingWebhook()
	}
	return map[string]interface{}{
		"message": "执行成功",
		"success": true,
		"status":  200,
	}
}
