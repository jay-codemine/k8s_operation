package pvc

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeletePersistentVolumeClaim 删除指定 namespace/name 的 PVC，并轮询确认删除完成
func DeletePersistentVolumeClaim(ctx context.Context, Kube kubernetes.Interface, namespace, name string) error {
	// 一般使用 Foreground，让依赖对象先清理；PVC 带有 pvc-protection finalizer，K8s会保证安全删除
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := Kube.CoreV1().
		PersistentVolumeClaims(namespace).
		Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等
		}
		return fmt.Errorf("delete PersistentVolumeClaim %q in namespace %q failed: %w", name, namespace, err)
	}

	// 轮询直至确认删除；PVC 被 Pod 使用时会等待，直到 Pod 释放并移除 pvc-protection finalizer
	return wait.PollUntilContextTimeout(
		ctx,
		2*time.Second,  // 每 2 秒检查一次
		30*time.Second, // 最长等待 30 秒（可按需调大）
		true,           // 立即执行第一次检查
		func(ctx context.Context) (bool, error) {
			_, err := Kube.CoreV1().
				PersistentVolumeClaims(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil // 确认删除
			}
			if err != nil {
				return false, err // 临时错误，继续轮询
			}
			return false, nil // 仍存在，继续轮询
		},
	)
}
