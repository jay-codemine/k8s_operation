package namespace

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

func GetNamespaceDetail(ctx context.Context, Kube kubernetes.Interface, name string) (*corev1.Namespace, error) {
	ns, err := Kube.CoreV1().
		Namespaces().
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("Namespace not found", zap.String("name", name))
			return nil, fmt.Errorf("Namespace %q not found", name)
		}

		global.Logger.Error("get Namespace failed",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return ns, nil
}
