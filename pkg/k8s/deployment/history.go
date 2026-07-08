package deployment

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
)

func GetDeploymentReplicaSet(ctx context.Context, Kube kubernetes.Interface, namespace, name string) ([]appv1.ReplicaSet, error) {
	// 获取 指定 deployment 的详细信息
	d, err := Kube.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("deployment %s not found: %w", name, err)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get deployment %s: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get deployment %s: %w", name, err)
	}

	selector := ""
	if d.Spec.Selector != nil {
		selector = metav1.FormatLabelSelector(d.Spec.Selector)
	}

	rsList, err := Kube.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("list replicasets failed (ns=%s, selector=%q): %w", namespace, selector, err)
	}

	ownend := make([]appv1.ReplicaSet, 0, len(rsList.Items))
	for i := range rsList.Items {
		rs := rsList.Items[i]
		if isControlledBy(&rs, d) {
			ownend = append(ownend, rs)
		}
	}

	sort.SliceStable(ownend, func(i, j int) bool {
		// 获取 ownend 切片中第 i 和第 j 个元素的 CreationTimestamp 字段的时间部分
		ti, tj := ownend[i].CreationTimestamp.Time, ownend[j].CreationTimestamp.Time
		// 比较两个时间是否相等
		if !ti.Equal(tj) {
			// 如果时间不相等，则返回 ti 是否在 tj 之后
			return ti.After(tj)
		}

		// 如果时间相等，则返回 ownend 切片中第 i 和第 j 个元素的 Name 字段是否相等
		return ownend[i].Name < ownend[j].Name //
	})

	return ownend, nil
}

// 判断 replicas 是否由 deployment 控制
func isControlledBy(rs *appv1.ReplicaSet, d *appv1.Deployment) bool {
	for _, o := range rs.ObjectMeta.OwnerReferences {
		if o.Controller != nil && *o.Controller &&
			o.UID == d.UID &&
			o.Kind == "Deployment" &&
			o.APIVersion == "apps/v1" {
			return true
		}
	}
	return false
}
