package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func UpdateIngressFromJSON(ctx context.Context, kube kubernetes.Interface, namespace string, content string,
) (*networkingv1.Ingress, error) {

	// Step 1: JSON 解析为 Ingress 对象
	var ing networkingv1.Ingress
	if err := json.Unmarshal([]byte(content), &ing); err != nil {
		return nil, fmt.Errorf("解析 Ingress JSON 失败: %w", err)
	}

	if ing.Name == "" {
		// 兼容：有的人会只写 metadata.name
		if ing.ObjectMeta.Name == "" {
			return nil, fmt.Errorf("metadata.name 不能为空")
		}
		ing.Name = ing.ObjectMeta.Name
	}

	// namespace 以请求入参为准
	ing.Namespace = namespace

	// Step 2: 获取旧对象，继承 ResourceVersion
	old, err := kube.NetworkingV1().
		Ingresses(namespace).
		Get(ctx, ing.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("Ingress 不存在: %s/%s", namespace, ing.Name)
		}
		return nil, fmt.Errorf("获取原 Ingress 失败: %w", err)
	}

	if ing.ResourceVersion == "" {
		ing.ResourceVersion = old.ResourceVersion
	}

	// Step 3: 移除 managedFields，防止 update 冲突
	ing.ManagedFields = nil

	// Step 4: 执行全量覆盖更新（PUT）
	updated, err := kube.NetworkingV1().
		Ingresses(namespace).
		Update(ctx, &ing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	global.Logger.Infof(
		"Ingress [%s] 在命名空间 [%s] 更新成功 (rv=%s)",
		updated.Name, updated.Namespace, updated.ResourceVersion,
	)
	return updated, nil
}
