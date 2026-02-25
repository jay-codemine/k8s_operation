package configmap

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8soperation/global"
)

// GetConfigMapYaml 获取 ConfigMap 的 YAML 表示
func GetConfigMapYaml(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (string, error) {
	cm, err := Kube.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 ConfigMap 失败: %w", err)
	}

	// 清理不必要的字段
	cm.ManagedFields = nil
	cm.UID = ""
	cm.ResourceVersion = ""
	cm.Generation = 0
	cm.CreationTimestamp = metav1.Time{}

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(cm)
	if err != nil {
		return "", fmt.Errorf("转换 YAML 失败: %w", err)
	}

	return string(yamlBytes), nil
}

// ApplyConfigMapYaml 从 YAML 字符串创建或更新 ConfigMap
func ApplyConfigMapYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*corev1.ConfigMap, error) {
	// 解析 YAML
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(yamlContent), &cm); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 检查必要字段
	if cm.Name == "" {
		return nil, fmt.Errorf("YAML 中缺少 metadata.name")
	}
	if cm.Namespace == "" {
		cm.Namespace = "default"
	}

	// 尝试获取现有 ConfigMap
	existing, err := Kube.CoreV1().
		ConfigMaps(cm.Namespace).
		Get(ctx, cm.Name, metav1.GetOptions{})
	
	if err != nil {
		// ConfigMap 不存在，创建新的
		created, createErr := Kube.CoreV1().
			ConfigMaps(cm.Namespace).
			Create(ctx, &cm, metav1.CreateOptions{})
		if createErr != nil {
			return nil, fmt.Errorf("创建 ConfigMap 失败: %w", createErr)
		}
		
		global.Logger.Infof("ConfigMap [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		return created, nil
	}

	// ConfigMap 已存在，更新它
	// 继承 ResourceVersion
	cm.ResourceVersion = existing.ResourceVersion

	// 更新 ConfigMap
	updated, err := Kube.CoreV1().
		ConfigMaps(cm.Namespace).
		Update(ctx, &cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	global.Logger.Infof("ConfigMap [%s] 在命名空间 [%s] YAML 更新成功",
		updated.Name, updated.Namespace)

	return updated, nil
}
