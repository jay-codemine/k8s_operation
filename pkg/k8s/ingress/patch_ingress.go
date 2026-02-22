package ingress

import (
	"context"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// 通用 Patch：传入 StrategicMergePatch 的 bytes，返回最新的 Ingress
func PatchIngress(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*networkingv1.Ingress, error) {
	ing, err := Kube.
		NetworkingV1().
		Ingresses(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return ing, nil
}
