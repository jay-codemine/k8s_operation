package crd

import (
	"context"
	"k8soperation/internal/app/requests"
)

func DeleteAppConfig(ctx context.Context, d *KubeAppConfig, req *requests.KubeAppConfigDeleteRequest) error {
	return d.Delete(ctx, req.Namespace, req.AppName)
}
