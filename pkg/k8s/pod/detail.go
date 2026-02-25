package pod

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// GetPodDetail 获取Pod详情
// pkg/k8s/kube_pod/kube_pod.go
func GetPodDetail(ctx context.Context, kube kubernetes.Interface, namespace, name string) (*corev1.Pod, error) {
	pod, err := kube.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		global.Logger.Errorf("Failed to get kube_pod detail: %v", err)
		return nil, err
	}
	global.Logger.Infof("Pod detail: %v", pod)
	return pod, nil
}
