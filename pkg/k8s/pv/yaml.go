// pkg/k8s/pv/yaml.go
package pv

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 PersistentVolume 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, name string) (string, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanPVForYaml(pv)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(pv)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 PersistentVolume YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, name, yamlContent string) (*corev1.PersistentVolume, error) {
	// 解析 YAML
	var pv corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(yamlContent), &pv); err != nil {
		return nil, err
	}

	// 确保 name 匹配
	pv.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	pv.ResourceVersion = existing.ResourceVersion

	// 更新 PV
	updated, err := client.CoreV1().PersistentVolumes().Update(ctx, &pv, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// CreateFromYaml 从 YAML 创建 PersistentVolume
func CreateFromYaml(ctx context.Context, client kubernetes.Interface, yamlContent string) (*corev1.PersistentVolume, error) {
	// 解析 YAML
	var pv corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(yamlContent), &pv); err != nil {
		return nil, err
	}

	// 创建 PV
	created, err := client.CoreV1().PersistentVolumes().Create(ctx, &pv, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return created, nil
}

// cleanPVForYaml 清理 PersistentVolume 对象中的运行时字段
func cleanPVForYaml(pv *corev1.PersistentVolume) {
	// 清理 TypeMeta
	pv.APIVersion = "v1"
	pv.Kind = "PersistentVolume"

	// 清理 ObjectMeta 中的运行时字段
	pv.ManagedFields = nil
	pv.UID = ""
	pv.ResourceVersion = ""
	pv.Generation = 0
	pv.CreationTimestamp = metav1.Time{}
	pv.DeletionTimestamp = nil
	pv.DeletionGracePeriodSeconds = nil
	pv.SelfLink = ""

	// 清理 Status
	pv.Status = corev1.PersistentVolumeStatus{}
}
