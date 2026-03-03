package cronjob

import (
	"context"
	"encoding/json"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func SetCronJobSuspend(ctx context.Context, Kube kubernetes.Interface, namespace, name string, suspend bool) error {
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"suspend": suspend,
		},
	}
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return fmt.Errorf("failed to marshal patch data: %v", err)
	}

	_, err = Kube.BatchV1().CronJobs(namespace).Patch(
		ctx,
		name,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("CronJob %q not found in namespace %q", name, namespace)
		}
		return fmt.Errorf("failed to update CronJob suspend status: %v", err)
	}

	action := "suspended"
	if !suspend {
		action = "resumed"
	}
	global.Logger.Infof("CronJob %q in namespace %q %s successfully", name, namespace, action)
	return nil
}
