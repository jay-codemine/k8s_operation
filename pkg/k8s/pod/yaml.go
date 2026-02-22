// pkg/k8s/pod/yaml.go
package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 Pod 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, namespace, name string) (string, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanPodForYaml(pod)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(pod)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 Pod YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, namespace, name, yamlContent string) (*corev1.Pod, error) {
	// 解析 YAML
	var pod corev1.Pod
	if err := yaml.Unmarshal([]byte(yamlContent), &pod); err != nil {
		return nil, err
	}

	// 确保 namespace 和 name 匹配
	pod.Namespace = namespace
	pod.Name = name

	// 获取现有 Pod 以保留必要的字段
	existing, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	pod.ResourceVersion = existing.ResourceVersion

	// 更新 Pod（注意：Pod 的大部分字段是不可变的，只能更新部分字段）
	updated, err := client.CoreV1().Pods(namespace).Update(ctx, &pod, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanPodForYaml 清理 Pod 对象中的运行时字段
func cleanPodForYaml(pod *corev1.Pod) {
	// 清理 TypeMeta
	pod.APIVersion = "v1"
	pod.Kind = "Pod"

	// 清理 ObjectMeta 中的运行时字段
	pod.ManagedFields = nil
	pod.UID = ""
	pod.ResourceVersion = ""
	pod.Generation = 0
	pod.CreationTimestamp = metav1.Time{}
	pod.DeletionTimestamp = nil
	pod.DeletionGracePeriodSeconds = nil
	pod.SelfLink = ""

	// 清理 Status（YAML 编辑时不需要 status）
	pod.Status = corev1.PodStatus{}
}
