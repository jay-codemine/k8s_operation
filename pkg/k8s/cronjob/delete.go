package cronjob

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

func DeleteCronJob(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// Foreground：先删所有受控 Job，再删除 CronJob 本体
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := Kube.BatchV1().
		CronJobs(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：不存在视为成功
		}
		return err
	}

	// 等待 CronJob 被真正删除
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second, 30*time.Second, true,
		func(ctx context.Context) (bool, error) {
			_, err := Kube.BatchV1().
				CronJobs(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)
}
