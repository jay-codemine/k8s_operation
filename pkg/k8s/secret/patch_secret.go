package secret

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// 通用 StrategicMergePatch（适合结构化对象）
func PatchSecret(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*corev1.Secret, error) {
	sec, err := Kube.CoreV1().Secrets(namespace).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return sec, nil
}
