// pkg/k8s/daemonset/yaml.go
package daemonset

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 DaemonSet 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, namespace, name string) (string, error) {
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanDaemonSetForYaml(ds)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(ds)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 DaemonSet YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, namespace, name, yamlContent string) (*appv1.DaemonSet, error) {
	// 解析 YAML
	var ds appv1.DaemonSet
	if err := yaml.Unmarshal([]byte(yamlContent), &ds); err != nil {
		return nil, err
	}

	// 确保 namespace 和 name 匹配
	ds.Namespace = namespace
	ds.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	ds.ResourceVersion = existing.ResourceVersion

	// 更新 DaemonSet
	updated, err := client.AppsV1().DaemonSets(namespace).Update(ctx, &ds, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanDaemonSetForYaml 清理 DaemonSet 对象中的运行时字段
func cleanDaemonSetForYaml(ds *appv1.DaemonSet) {
	// 清理 TypeMeta
	ds.APIVersion = "apps/v1"
	ds.Kind = "DaemonSet"

	// 清理 ObjectMeta 中的运行时字段
	ds.ManagedFields = nil
	ds.UID = ""
	ds.ResourceVersion = ""
	ds.Generation = 0
	ds.CreationTimestamp = metav1.Time{}
	ds.DeletionTimestamp = nil
	ds.DeletionGracePeriodSeconds = nil
	ds.SelfLink = ""

	// 清理 Status
	ds.Status = appv1.DaemonSetStatus{}
}
