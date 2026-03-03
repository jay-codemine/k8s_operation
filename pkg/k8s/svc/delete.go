package svc

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

func DeleteService(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 删除选项（Service 没有级联删除，直接默认即可）
	opts := metav1.DeleteOptions{}

	// 发起删除请求
	if err := Kube.CoreV1().
		Services(namespace).
		Delete(ctx, name, opts); err != nil {

		// 如果 Service 不存在则视为成功（幂等性处理）
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// 等待 Service 确认删除成功
	err := wait.PollUntilContextTimeout(
		ctx,            // 上下文（带 30s 超时）
		2*time.Second,  // 轮询间隔
		30*time.Second, // 超时时间
		true,           // 是否立即执行一次
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.CoreV1().
				Services(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)

	return err
}
