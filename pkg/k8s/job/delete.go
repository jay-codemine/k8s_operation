package job

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DeleteJob 删除指定 Job（带前台级联 + 轮询等待删除完成）
func DeleteJob(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 指定级联删除策略（Foreground: 先删 Pods，再删 Job）
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := Kube.BatchV1().
		Jobs(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等处理：不存在也视为成功
		}
		return err
	}

	// 轮询等待 Job 确认删除完成
	err := wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 轮询间隔
		30*time.Second, // 超时时间
		true,           // immediate：立即执行一次
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.BatchV1().
				Jobs(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil // 确认删除完成
			}
			return false, err // 未删除则继续轮询
		},
	)

	return err
}
