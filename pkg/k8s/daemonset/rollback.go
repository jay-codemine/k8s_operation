package daemonset

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

func RollbackDaemonSet(ctx context.Context, Kube kubernetes.Interface, name, namespace, revisionName string) (*appv1.DaemonSet, error) {

	// 检查 DS 是否存在
	if _, err := Kube.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("daemonset %s/%s not found: %w", namespace, name, err)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get daemonset %s/%s: %w", namespace, name, err)
		}
		return nil, fmt.Errorf("get daemonset %s/%s failed: %w", namespace, name, err)
	}

	// 获取 DS 历史版本（ControllerRevision）
	revs, err := GetDaemonSetControllerRevisions(ctx, Kube, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("list daemonset history failed: %w", err)
	}

	var target *appv1.ControllerRevision
	for i := range revs {
		if revs[i].Name == revisionName {
			target = &revs[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("revision %s/%s not found for daemonset %s", namespace, revisionName, name)
	}

	// 解析历史模板
	var dsTemplate appv1.DaemonSet
	if err := json.Unmarshal(target.Data.Raw, &dsTemplate); err != nil {
		return nil, fmt.Errorf("unmarshal daemonset template failed: %w", err)
	}

	var updated *appv1.DaemonSet
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		cur, getErr := Kube.AppsV1().DaemonSets(namespace).Get(cctx, name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("get daemonset %s/%s failed: %w", namespace, name, getErr)
		}

		// 回滚模板
		cur.Spec.Template = dsTemplate.Spec.Template

		// 添加审计注解
		if cur.Annotations == nil {
			cur.Annotations = map[string]string{}
		}
		cur.Annotations["rollback.from.revision"] = revisionName
		cur.Annotations["rollback.at"] = time.Now().Format(time.RFC3339)

		var updateErr error
		updated, updateErr = Kube.AppsV1().DaemonSets(namespace).Update(cctx, cur, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("update daemonset %s/%s failed: %w", namespace, name, updateErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func GetDaemonSetControllerRevisions(ctx context.Context, Kube kubernetes.Interface, namespace, name string) ([]appv1.ControllerRevision, error) {
	list, err := Kube.AppsV1().
		ControllerRevisions(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []appv1.ControllerRevision
	for _, rev := range list.Items {
		if rev.OwnerReferences != nil && len(rev.OwnerReferences) > 0 &&
			rev.OwnerReferences[0].Kind == "DaemonSet" &&
			rev.OwnerReferences[0].Name == name {
			result = append(result, rev)
		}
	}
	return result, nil
}
