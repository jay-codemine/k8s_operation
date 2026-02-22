package namespace

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeleteNamespace 删除指定 Namespace（包含轮询确认删干净）
func DeleteNamespace(ctx context.Context, Kube kubernetes.Interface, name string) error {
	// 前台删除：确保 namespace 内所有资源删除完成后才返回
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := Kube.CoreV1().
		Namespaces().
		Delete(ctx, name, opts); err != nil {

		// 幂等处理：不存在就视为成功
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("delete Namespace %q failed: %w", name, err)
	}

	// 轮询检查 Namespace 是否真的删除
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 每 2 秒检查一次
		60*time.Second, // Namespace 删除可能慢一些，给 60 秒
		true,           // 第一次立即执行
		func(ctx context.Context) (done bool, err error) {
			ns, err := Kube.CoreV1().
				Namespaces().
				Get(ctx, name, metav1.GetOptions{})

			// Namespace 已经不存在 → 删除成功
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			// 其他异常
			if err != nil {
				return false, err
			}

			// 如果 Namespace 仍在 Terminating，继续等待
			if ns.Status.Phase == corev1.NamespaceTerminating {
				return false, nil
			}

			// 正常情况不会出现 Active，但以防万一
			return false, nil
		},
	)
}
