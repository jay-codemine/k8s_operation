package namespace

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// PatchLabels 修改 Namespace 标签（添加/删除）
func PatchLabels(ctx context.Context, kube kubernetes.Interface, nsName string, addLabels map[string]string, removeLabels []string) error {
	// 构建 JSON Patch
	patch := make(map[string]interface{})
	metadata := make(map[string]interface{})
	labels := make(map[string]interface{})

	// 添加标签
	for k, v := range addLabels {
		labels[k] = v
	}

	// 删除标签（设置为 null）
	for _, k := range removeLabels {
		labels[k] = nil
	}

	if len(labels) == 0 {
		return fmt.Errorf("no labels to add or remove")
	}

	metadata["labels"] = labels
	patch["metadata"] = metadata

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch failed: %w", err)
	}

	_, err = kube.CoreV1().Namespaces().Patch(ctx, nsName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("patch namespace labels failed: %w", err)
	}

	return nil
}
