// pkg/k8s/node/yaml.go
package node

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 Node 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, name string) (string, error) {
	node, err := client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanNodeForYaml(node)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(node)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 Node YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, name, yamlContent string) (*corev1.Node, error) {
	// 解析 YAML
	var node corev1.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		return nil, err
	}

	// 确保 name 匹配
	node.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	node.ResourceVersion = existing.ResourceVersion

	// 更新 Node
	updated, err := client.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanNodeForYaml 清理 Node 对象中的运行时字段
func cleanNodeForYaml(node *corev1.Node) {
	// 清理 TypeMeta
	node.APIVersion = "v1"
	node.Kind = "Node"

	// 清理 ObjectMeta 中的运行时字段
	node.ManagedFields = nil
	node.UID = ""
	node.ResourceVersion = ""
	node.Generation = 0
	node.CreationTimestamp = metav1.Time{}
	node.DeletionTimestamp = nil
	node.DeletionGracePeriodSeconds = nil
	node.SelfLink = ""

	// 清理 Status（Node status 包含大量运行时数据）
	node.Status = corev1.NodeStatus{}
}
