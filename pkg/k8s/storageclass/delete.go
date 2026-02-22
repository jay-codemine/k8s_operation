package storageclass

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeleteStorageClass 删除 StorageClass
func DeleteStorageClass(ctx context.Context, Kube kubernetes.Interface, name string) error {
	// 删除选项（Foreground 表示前台删除，直到资源清理完）
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := Kube.StorageV1().
		StorageClasses().
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：不存在也算成功
		}
		return err
	}

	// 轮询等待确认删除完成
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 检查间隔
		30*time.Second, // 最大等待时间
		true,           // 立即执行第一次检查
		func(ctx context.Context) (bool, error) {
			_, err := Kube.StorageV1().
				StorageClasses().
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil // 删除确认
			}
			return false, err
		},
	)
}
