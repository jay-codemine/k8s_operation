package deployment

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func GetDeploymentDetail(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*appv1.Deployment, error) {
	dp, err := Kube.AppsV1().
		Deployments(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		// 仅当有错误时再细分类型
		if apierrors.IsNotFound(err) { // 别名：apierrors 来自 k8s.io/apimachinery/pkg/api/errors
			global.Logger.Error("deployment not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("deployment %s/%s not found", namespace, name)
		}

		// 其它错误，直接返回并记录
		global.Logger.Error("get deployment failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	// 正常返回
	return dp, nil
}
