package job

import "k8soperation/pkg/utils"

func stripJobControllerLabels(m map[string]string) map[string]string {
	jobControllerLabels := []string{
		"controller-uid",
		"job-name",
		"batch.kubernetes.io/controller-uid",
		"batch.kubernetes.io/job-name",
	}
	return utils.StripKeys(m, jobControllerLabels)
}
