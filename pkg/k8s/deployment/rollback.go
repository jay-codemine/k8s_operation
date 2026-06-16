package deployment

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"time"
)

func RollbackDeployment(ctx context.Context, Kube kubernetes.Interface, name, namespace, rsName string) (*appv1.Deployment, error) {
	if _, err := Kube.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("deployment %s/%s not found: %w", namespace, name, err)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get deployment %s/%s: %w", namespace, name, err)
		}
		return nil, fmt.Errorf("get deployment %s/%s failed: %w", namespace, name, err)
	}

	replicaSets, err := GetDeploymentReplicaSet(ctx, Kube, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("list owned replicasets failed: %w", err)
	}
	var target *appv1.ReplicaSet
	for i := range replicaSets {
		if replicaSets[i].Name == rsName {
			target = &replicaSets[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("replicaset %s/%s not found for deployment %s", namespace, rsName, name)
	}

	var updated *appv1.Deployment
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cur, getErr := Kube.AppsV1().Deployments(namespace).Get(cctx, name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("get deployment %s/%s failed: %w", namespace, name, getErr)
		}
		// 回滚模板
		cur.Spec.Template = target.Spec.Template

		// 审计注解（不触发滚动）
		if cur.Annotations == nil {
			cur.Annotations = map[string]string{}
		}
		cur.Annotations["rollback.from.replicaset"] = rsName
		cur.Annotations["rollback.at"] = time.Now().Format(time.RFC3339)

		var updateErr error
		updated, updateErr = Kube.AppsV1().Deployments(namespace).Update(cctx, cur, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("update deployment %s/%s failed: %w", namespace, name, updateErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
