// pkg/k8s/namespace/yaml.go
package namespace

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 Namespace 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, name string) (string, error) {
	ns, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanNamespaceForYaml(ns)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(ns)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 Namespace YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, name, yamlContent string) (*corev1.Namespace, error) {
	// 解析 YAML
	var ns corev1.Namespace
	if err := yaml.Unmarshal([]byte(yamlContent), &ns); err != nil {
		return nil, err
	}

	// 确保 name 匹配
	ns.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	ns.ResourceVersion = existing.ResourceVersion

	// 更新 Namespace
	updated, err := client.CoreV1().Namespaces().Update(ctx, &ns, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanNamespaceForYaml 清理 Namespace 对象中的运行时字段
func cleanNamespaceForYaml(ns *corev1.Namespace) {
	// 清理 TypeMeta
	ns.APIVersion = "v1"
	ns.Kind = "Namespace"

	// 清理 ObjectMeta 中的运行时字段
	ns.ManagedFields = nil
	ns.UID = ""
	ns.ResourceVersion = ""
	ns.Generation = 0
	ns.CreationTimestamp = metav1.Time{}
	ns.DeletionTimestamp = nil
	ns.DeletionGracePeriodSeconds = nil
	ns.SelfLink = ""

	// 清理 Status
	ns.Status = corev1.NamespaceStatus{}
}
