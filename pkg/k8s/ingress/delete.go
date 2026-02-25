package ingress

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeleteIngress 删除 Ingress（前台级联 + 轮询确认消失）
func DeleteIngress(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 显式前台级联（虽然 Ingress 基本没有子资源，但语义明确）
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := Kube.NetworkingV1().
		Ingresses(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：已不存在视为成功
		}
		return err
	}

	// 轮询直到真正消失（与上面的 context 同步超时）
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // interval
		30*time.Second, // timeout（通常与 context 超时一致）
		true,           // immediate：先立即检查一次
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.NetworkingV1().
				Ingresses(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)
}
