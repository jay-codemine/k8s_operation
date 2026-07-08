package configmap

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

func GetConfigMapDetail(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*corev1.ConfigMap, error) {
	// 通过 K8s Client 获取 ConfigMap
	cm, err := Kube.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("configmap not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("configmap %s/%s not found", namespace, name)
		}
		global.Logger.Error("get configmap failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return cm, nil
}
