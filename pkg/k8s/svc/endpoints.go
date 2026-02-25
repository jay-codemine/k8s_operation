package svc

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

func GetServiceEndpoints(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*corev1.Endpoints, error) {
	ep, err := Kube.CoreV1().
		Endpoints(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("endpoints not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("endpoints %s/%s not found", namespace, name)
		}
		global.Logger.Error("get endpoints failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}
	return ep, nil
}
