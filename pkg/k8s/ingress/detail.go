package ingress

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// GetIngressDetail 获取指定命名空间下的 Ingress 详情
func GetIngressDetail(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*networkingv1.Ingress, error) {
	// 直接通过 K8s Client 调用
	ing, err := Kube.NetworkingV1().
		Ingresses(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			global.Logger.Error("ingress not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("ingress %s/%s not found", namespace, name)
		}

		global.Logger.Error("get ingress failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return ing, nil
}
