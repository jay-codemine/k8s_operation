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
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := Kube.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete PersistentVolumeClaim %q failed: %w", name, err)
	}

	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 30*time.Second, true,
		func(ctx context.Context) (bool, error) {
			_, err := Kube.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
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

// GraceDeletePersistentVolumeClaim 强制删除 PVC（清除 finalizers，用于卡 Terminating 的 PVC）
func GraceDeletePersistentVolumeClaim(ctx context.Context, Kube kubernetes.Interface, namespace, name string) error {
	pvc, err := Kube.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get PVC %q failed: %w", name, err)
	}

	if len(pvc.Finalizers) > 0 {
		pvc.Finalizers = nil
		if _, err := Kube.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("clear PVC %q finalizers failed: %w", name, err)
		}
	}

	return DeletePersistentVolumeClaim(ctx, Kube, namespace, name)
}
