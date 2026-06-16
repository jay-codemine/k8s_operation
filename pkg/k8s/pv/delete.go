package pv

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeletePersistentVolume 删除指定名称的 PV（包含轮询确认）
func DeletePersistentVolume(ctx context.Context, Kube kubernetes.Interface, name string) error {
	//  删除策略：前台级联删除
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 调用 Kubernetes API 发起删除请求
	if err := Kube.CoreV1().
		PersistentVolumes().
		Delete(ctx, name, opts); err != nil {
		// 如果 PV 不存在，视为删除成功（幂等处理）
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete PersistentVolume %q failed: %w", name, err)
	}

	// 轮询检查，直到 PV 确认删除
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 每隔 2 秒检查一次
		30*time.Second, // 最长等待 30 秒
		true,           // 立即执行第一次检查
		func(ctx context.Context) (done bool, err error) {
			_, err = Kube.CoreV1().
				PersistentVolumes().
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				// 已确认删除
				return true, nil
			}
			// 其他错误（如暂时性通信问题）
			if err != nil {
				return false, err
			}
			// 仍存在，继续等待
			return false, nil
		},
	)
}
