package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
)

// ControllerRevisionItem 历史版本信息（前端友好）
type ControllerRevisionItem struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Revision        int64             `json:"revision"`
	CreationTime    metav1.Time       `json:"creation_time"`
	Labels          map[string]string `json:"labels,omitempty"`
	OwnerReferences []string          `json:"owner_references,omitempty"`
}

// GetStatefulSetHistory 获取 StatefulSet 的历史版本（ControllerRevision 列表）
func GetStatefulSetHistory(ctx context.Context, Kube kubernetes.Interface, namespace, name string) ([]ControllerRevisionItem, error) {
	// 获取 StatefulSet
	sts, err := Kube.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("statefulset %s not found: %w", name, err)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get statefulset %s: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get statefulset %s: %w", name, err)
	}

	// 构造 selector
	selector := ""
	if sts.Spec.Selector != nil {
		selector = metav1.FormatLabelSelector(sts.Spec.Selector)
	}

	// 获取 ControllerRevision 列表
	revList, err := Kube.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("list controllerrevisions failed (ns=%s, selector=%q): %w", namespace, selector, err)
	}

	// 过滤出属于该 StatefulSet 的版本
	owned := make([]appv1.ControllerRevision, 0, len(revList.Items))
	for i := range revList.Items {
		rev := revList.Items[i]
		if isControlledByStatefulSet(&rev, sts) {
			owned = append(owned, rev)
		}
	}

	// 按 Revision 倒序排序
	sort.SliceStable(owned, func(i, j int) bool {
		return owned[i].Revision > owned[j].Revision
	})

	// 转换为前端友好格式
	result := make([]ControllerRevisionItem, 0, len(owned))
	for _, rev := range owned {
		ownerNames := make([]string, 0)
		for _, o := range rev.OwnerReferences {
			ownerNames = append(ownerNames, fmt.Sprintf("%s/%s", o.Kind, o.Name))
		}
		result = append(result, ControllerRevisionItem{
			Name:            rev.Name,
			Namespace:       rev.Namespace,
			Revision:        rev.Revision,
			CreationTime:    rev.CreationTimestamp,
			Labels:          rev.Labels,
			OwnerReferences: ownerNames,
		})
	}

	return result, nil
}

// isControlledByStatefulSet 判断 ControllerRevision 是否由 StatefulSet 控制
func isControlledByStatefulSet(rev *appv1.ControllerRevision, sts *appv1.StatefulSet) bool {
	for _, o := range rev.OwnerReferences {
		if o.Controller != nil && *o.Controller &&
			o.UID == sts.UID &&
			o.Kind == "StatefulSet" &&
			o.APIVersion == "apps/v1" {
			return true
		}
	}
	return false
}
