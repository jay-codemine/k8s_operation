package secret

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

// UpdateSecretFromJSON：JSON -> Secret 全量 Update（PUT）
// 关键点：
// - timeout 10s
// - namespace 强制一致
// - 继承 ResourceVersion（避免冲突）
// - type 为空时继承旧值（避免变成 Opaque 或报错）
func UpdateSecretFromJSON(ctx context.Context, kube kubernetes.Interface, namespace string, content string,
) (*corev1.Secret, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("namespace 不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content 不能为空")
	}

	var sec corev1.Secret
	if err := json.Unmarshal([]byte(content), &sec); err != nil {
		return nil, fmt.Errorf("解析 Secret JSON 失败: %w", err)
	}

	// 兼容 metadata.name 写法
	name := sec.Name
	if name == "" {
		name = sec.ObjectMeta.Name
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	sec.Name = name
	sec.Namespace = namespace

	// 获取旧对象：继承 RV / type
	old, err := kube.CoreV1().
		Secrets(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("Secret 不存在: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("获取原 Secret 失败: %w", err)
	}

	if sec.ResourceVersion == "" {
		sec.ResourceVersion = old.ResourceVersion
	}
	if sec.Type == "" {
		sec.Type = old.Type
	}

	// 全量覆盖 Update
	updated, err := kube.CoreV1().
		Secrets(namespace).
		Update(ctx, &sec, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Secret 失败: %w", err)
	}

	global.Logger.Infof("Secret [%s] 在命名空间 [%s] 更新成功 (rv=%s, type=%s)",
		updated.Name, updated.Namespace, updated.ResourceVersion, updated.Type)

	return updated, nil
}
