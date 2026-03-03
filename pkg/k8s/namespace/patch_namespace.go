package namespace

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func PatchNamespace(ctx context.Context, Kube kubernetes.Interface, name string, patch []byte) (*corev1.Namespace, error) {
	return Kube.CoreV1().
		Namespaces().
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
}
