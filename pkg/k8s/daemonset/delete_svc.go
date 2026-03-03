package daemonset

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"time"
)

// DeleteDaemonSetService 删除与 DaemonSet 关联的 Service（如果存在）
// 支持幂等：即使 Service 不存在也不会报错。
func DeleteDaemonSetService(ctx context.Context, Kube kubernetes.Interface, name, namespace string) error {
	// 尝试删除 Service
	err := Kube.CoreV1().
		Services(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Infof("service %s/%s not found, skip delete", namespace, name)
			return nil //  幂等：Service 已不存在
		}
		global.Logger.Errorf("delete daemonset service %s/%s failed: %v", namespace, name, err)
		return err
	}

	// 等待删除完成（确认 Service 被真正移除）
	waitErr := wait.PollUntilContextTimeout(
		ctx,
		1*time.Second,  // 每秒检查一次
		10*time.Second, // 最多等 10 秒
		true,           // 是否立即执行第一次检查
		func(ctx context.Context) (bool, error) {
			_, err := Kube.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				global.Logger.Infof("service %s/%s deleted successfully", namespace, name)
				return true, nil //  删除完成
			}
			return false, err
		},
	)

	return waitErr
}
