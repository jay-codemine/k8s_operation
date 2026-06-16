package secret

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

func GetSecretList(ctx context.Context, Kube kubernetes.Interface, name, namespace string, page, limit int) ([]corev1.Secret, int, error) {
	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取 Secret 列表
	list, err := Kube.CoreV1().
		Secrets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, 0, fmt.Errorf("namespace %s not found", namespace)
		}
		return nil, 0, err
	}

	// 选择器：名称过滤 + 排序 + 分页（与你的 Ingress 一致）
	selector := dataselect.NewSecretSelector(list.Items, name, page, limit)

	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromSecretCells(data.GenericDataList), total, nil
}
