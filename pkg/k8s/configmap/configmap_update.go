package configmap

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"

	"k8soperation/global"
)

// UpdateConfigMapData：直接更新 ConfigMap 的 data 字段
func UpdateConfigMapData(ctx context.Context, kube kubernetes.Interface, namespace, name string, data map[string]string) (*corev1.ConfigMap, error) {
	// 获取旧对象
	old, err := kube.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("ConfigMap 不存在: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("获取原 ConfigMap 失败: %w", err)
	}

	// 更新 data
	old.Data = data

	// 全量 Update（PUT）
	updated, err := kube.CoreV1().
		ConfigMaps(namespace).
		Update(ctx, old, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	global.Logger.Infof("ConfigMap [%s] 在命名空间 [%s] data 更新成功",
		updated.Name, updated.Namespace)

	return updated, nil
}

// UpdateConfigMapJson：JSON -> ConfigMap 全量 Update（PUT）
// 关键点：
// - timeout（10s）
// - namespace 强制一致
// - 继承 ResourceVersion（避免冲突）
// - 可选：继承 Labels/Annotations（如果新对象未提供）
func UpdateConfigMapJson(ctx context.Context, kube kubernetes.Interface, namespace string, content string,
) (*corev1.ConfigMap, error) {
	// 参数校验
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("namespace 不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content 不能为空")
	}

	// JSON 反序列化
	var cm corev1.ConfigMap
	if err := json.Unmarshal([]byte(content), &cm); err != nil {
		return nil, fmt.Errorf("解析 ConfigMap JSON 失败: %w", err)
	}

	// namespace 以入参为准
	cm.Namespace = namespace

	// 兼容：metadata.name 取法
	name := cm.Name
	if name == "" {
		name = cm.ObjectMeta.Name
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	cm.Name = name

	// 获取旧对象（继承 RV / 可选继承 labels/annotations）
	old, err := kube.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("ConfigMap 不存在: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("获取原 ConfigMap 失败: %w", err)
	}

	// 继承 ResourceVersion（Update 必需）
	if cm.ResourceVersion == "" {
		cm.ResourceVersion = old.ResourceVersion
	}

	// 可选：如果未提供 labels/annotations，继承旧值
	if cm.Labels == nil {
		cm.Labels = old.Labels
	}
	if cm.Annotations == nil {
		cm.Annotations = old.Annotations
	}

	// 全量 Update（PUT）
	updated, err := kube.CoreV1().
		ConfigMaps(namespace).
		Update(ctx, &cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	global.Logger.Infof("ConfigMap [%s] 在命名空间 [%s] 更新成功 (rv=%s)",
		updated.Name, updated.Namespace, updated.ResourceVersion)

	return updated, nil
}
