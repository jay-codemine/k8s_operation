package daemonset

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

func DeleteDaemonSet(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 显式前台级联：先删 Pod，再删 DaemonSet
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除请求
	if err := Kube.AppsV1().
		DaemonSets(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等删除：不存在也算成功
		}
		return err
	}

	// 等待 DaemonSet 真正删除完成
	err := wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 轮询间隔
		30*time.Second, // 最长超时
		true,           // 立即执行一次检查
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.AppsV1().
				DaemonSets(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)

	return err
}
