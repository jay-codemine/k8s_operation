package svc

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"time"
)

// 通用 Patch：传入 patch bytes，返回最新的 Service
func PatchService(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*corev1.Service, error) {
	svc, err := Kube.CoreV1().
		Services(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func PatchJsonService(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*corev1.Service, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	svc, err := Kube.CoreV1().
		Services(namespace).
		Patch(c, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return svc, nil
}
