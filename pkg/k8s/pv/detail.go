package pv

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

// GetPVDetail 获取 PersistentVolume 详情
func GetPVDetail(ctx context.Context, Kube kubernetes.Interface, name string) (*corev1.PersistentVolume, error) {
	// 调用 K8s API 获取 PV（集群级，不需要 namespace）
	pv, err := Kube.CoreV1().
		PersistentVolumes().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// 资源不存在
		if apierrors.IsNotFound(err) {
			global.Logger.Error("PersistentVolume not found",
				zap.String("name", name),
			)
			return nil, fmt.Errorf("PersistentVolume %q not found", name)
		}

		// 其他错误
		global.Logger.Error("get PersistentVolume failed",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	//  成功返回
	return pv, nil
}
