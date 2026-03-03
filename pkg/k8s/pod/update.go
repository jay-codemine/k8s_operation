package pod

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// kube_pod/update.go
func UpdatePod(ctx context.Context, kube kubernetes.Interface, namespace, name string, content json.RawMessage) error {
	pod, err := kube.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &pod.Spec); err != nil {
		return fmt.Errorf("反序列化 Pod.Spec 失败: %w", err)
	}

	_, err = kube.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}
