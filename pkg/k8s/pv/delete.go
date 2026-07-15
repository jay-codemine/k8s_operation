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
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := Kube.CoreV1().PersistentVolumes().Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete PersistentVolume %q failed: %w", name, err)
	}

	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 30*time.Second, true,
		func(ctx context.Context) (bool, error) {
			_, err := Kube.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}
			return false, nil
		},
	)
}

// GraceDeletePersistentVolume 强制删除 PV（清除 finalizers 后删除，用于卡 Terminating 的 PV）
func GraceDeletePersistentVolume(ctx context.Context, Kube kubernetes.Interface, name string) error {
	pv, err := Kube.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get PV %q failed: %w", name, err)
	}

	if len(pv.Finalizers) > 0 {
		pv.Finalizers = nil
		if _, err := Kube.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("clear PV %q finalizers failed: %w", name, err)
		}
	}

	return DeletePersistentVolume(ctx, Kube, name)
}
