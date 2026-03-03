// pkg/k8s/statefulset/yaml.go
package statefulset

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 StatefulSet 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, namespace, name string) (string, error) {
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanStatefulSetForYaml(sts)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(sts)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 StatefulSet YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, namespace, name, yamlContent string) (*appv1.StatefulSet, error) {
	// 解析 YAML
	var sts appv1.StatefulSet
	if err := yaml.Unmarshal([]byte(yamlContent), &sts); err != nil {
		return nil, err
	}

	// 确保 namespace 和 name 匹配
	sts.Namespace = namespace
	sts.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	sts.ResourceVersion = existing.ResourceVersion

	// 更新 StatefulSet
	updated, err := client.AppsV1().StatefulSets(namespace).Update(ctx, &sts, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanStatefulSetForYaml 清理 StatefulSet 对象中的运行时字段
func cleanStatefulSetForYaml(sts *appv1.StatefulSet) {
	// 清理 TypeMeta
	sts.APIVersion = "apps/v1"
	sts.Kind = "StatefulSet"

	// 清理 ObjectMeta 中的运行时字段
	sts.ManagedFields = nil
	sts.UID = ""
	sts.ResourceVersion = ""
	sts.Generation = 0
	sts.CreationTimestamp = metav1.Time{}
	sts.DeletionTimestamp = nil
	sts.DeletionGracePeriodSeconds = nil
	sts.SelfLink = ""

	// 清理 Status
	sts.Status = appv1.StatefulSetStatus{}
}
