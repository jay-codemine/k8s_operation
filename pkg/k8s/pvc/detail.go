package pvc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// GetPVCDetail 获取 PersistentVolumeClaim 详情
func GetPVCDetail(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	// 调用 K8s API 获取 PVC（需要 namespace）
	pvc, err := Kube.CoreV1().
		PersistentVolumeClaims(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// 资源不存在
		if apierrors.IsNotFound(err) {
			global.Logger.Error("PersistentVolumeClaim not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", name, namespace)
		}

		// 其他错误
		global.Logger.Error("get PersistentVolumeClaim failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	// 成功返回
	return pvc, nil
}
