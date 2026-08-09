package crd

import (
	"context"
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8soperation/internal/app/requests"
)

func ListAppConfig(ctx context.Context, d *KubeAppConfig, req *requests.KubeAppConfigListRequest) ([]appv1alpha1.AppConfig, error) {
	return d.List(ctx, req.Namespace)
}
