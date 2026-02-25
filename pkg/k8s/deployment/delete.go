package deployment

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func DeleteDeployment(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 显式前台级联，先删 RS/Pods 再删 Deployment
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := Kube.AppsV1().
		Deployments(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：已不存在视为成功
		}
		return err
	}

	// 等待资源真正消失（与上面同一个 context）
	// 使用 wait.PollUntilContextTimeout 进行轮询操作，直到满足条件或超时
	err := wait.PollUntilContextTimeout(
		ctx,            // 上下文（带 30s 超时）
		2*time.Second,  // interval
		30*time.Second, // timeout
		true,           // immediate (是否立即执行一次)
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.AppsV1().
				Deployments(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)

	return err
}

func DeleteService(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	err := Kube.CoreV1().
		Services(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：已不存在视为成功
		}
	}

	// 等待删除完成（可选）
	waitErr := wait.PollUntilContextTimeout(
		ctx,
		1*time.Second,
		10*time.Second,
		true,
		func(ctx context.Context) (bool, error) {
			_, err := Kube.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)
	return waitErr
}
