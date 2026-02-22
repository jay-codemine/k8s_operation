package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// ListPodsByNode 列出指定 Node 上运行的所有 Pod
func ListPodsByNode(ctx context.Context, kube kubernetes.Interface, nodeName string) ([]corev1.Pod, error) {
	// Pod 是 namespace 级资源，这里需要跨所有命名空间查
	podList, err := kube.CoreV1().
		Pods(""). //空字符串代表“所有命名空间”
		List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		})
	if err != nil {
		global.Logger.Error("list Pods by node failed",
			zap.String("nodeName", nodeName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list Pods on node %q: %v", nodeName, err)
	}

	return podList.Items, nil
}
