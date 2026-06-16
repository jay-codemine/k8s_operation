package pvc

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/requests"
)

func PatchPVC(ctx context.Context, Kube kubernetes.Interface, req *requests.KubePVCResizeRequest) (*corev1.PersistentVolumeClaim, error) {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"resources": map[string]interface{}{
				"requests": map[corev1.ResourceName]string{
					corev1.ResourceStorage: req.Storage,
				},
			},
		},
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	updated, err := Kube.CoreV1().
		PersistentVolumeClaims(req.Namespace).
		Patch(ctx, req.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("patch pvc storage failed: %w", err)
	}
	return updated, nil
}
