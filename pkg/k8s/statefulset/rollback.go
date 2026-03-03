package statefulset

import (
	"context"
	"encoding/json"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"time"
)

// RollbackStatefulSet 回滚 StatefulSet 到指定的 ControllerRevision
func RollbackStatefulSet(ctx context.Context, Kube kubernetes.Interface, name, namespace, revisionName string) (*appv1.StatefulSet, error) {

	// 检查 StatefulSet 是否存在
	if _, err := Kube.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("statefulset %s/%s not found: %w", namespace, name, err)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get statefulset %s/%s: %w", namespace, name, err)
		}
		return nil, fmt.Errorf("get statefulset %s/%s failed: %w", namespace, name, err)
	}

	// 获取 StatefulSet 历史版本（ControllerRevision）
	revs, err := GetStatefulSetControllerRevisions(ctx, Kube, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("list statefulset history failed: %w", err)
	}

	var target *appv1.ControllerRevision
	for i := range revs {
		if revs[i].Name == revisionName {
			target = &revs[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("revision %s/%s not found for statefulset %s", namespace, revisionName, name)
	}

	// 解析历史模板
	var stsTemplate appv1.StatefulSet
	if err := json.Unmarshal(target.Data.Raw, &stsTemplate); err != nil {
		return nil, fmt.Errorf("unmarshal statefulset template failed: %w", err)
	}

	var updated *appv1.StatefulSet
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		cur, getErr := Kube.AppsV1().StatefulSets(namespace).Get(cctx, name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("get statefulset %s/%s failed: %w", namespace, name, getErr)
		}

		// 回滚模板
		cur.Spec.Template = stsTemplate.Spec.Template

		// 添加审计注解
		if cur.Annotations == nil {
			cur.Annotations = map[string]string{}
		}
		cur.Annotations["rollback.from.revision"] = revisionName
		cur.Annotations["rollback.at"] = time.Now().Format(time.RFC3339)

		var updateErr error
		updated, updateErr = Kube.AppsV1().StatefulSets(namespace).Update(cctx, cur, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("update statefulset %s/%s failed: %w", namespace, name, updateErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// GetStatefulSetControllerRevisions 获取 StatefulSet 的所有 ControllerRevision
func GetStatefulSetControllerRevisions(ctx context.Context, Kube kubernetes.Interface, namespace, name string) ([]appv1.ControllerRevision, error) {
	list, err := Kube.AppsV1().
		ControllerRevisions(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []appv1.ControllerRevision
	for _, rev := range list.Items {
		if rev.OwnerReferences != nil && len(rev.OwnerReferences) > 0 &&
			rev.OwnerReferences[0].Kind == "StatefulSet" &&
			rev.OwnerReferences[0].Name == name {
			result = append(result, rev)
		}
	}
	return result, nil
}
