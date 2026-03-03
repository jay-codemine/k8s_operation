package pod

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// PatchPodImage：对 Pod 的指定容器做 strategic merge patch 更新镜像
func PatchPodImage(ctx context.Context, kube kubernetes.Interface, namespace, podName, container, newImage string) error {
	patchObj := map[string]any{
		"spec": map[string]any{
			"containers": []map[string]string{
				{"name": container, "image": newImage},
			},
		},
	}

	b, err := json.Marshal(patchObj)
	if err != nil {
		return err
	}

	_, err = kube.CoreV1().
		Pods(namespace).
		Patch(ctx, podName, types.StrategicMergePatchType, b, metav1.PatchOptions{})
	return err
}
