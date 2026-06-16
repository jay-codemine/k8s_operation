package crd

import (
	"context"
	"k8soperation/internal/app/dao"

	"k8soperation/internal/app/requests"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
)

func UpdateAppConfig(ctx context.Context, d *dao.KubeAppConfig, req *requests.KubeAppConfigUpdateRequest) (*appv1alpha1.AppConfig, error) {
	app, err := d.Get(ctx, req.Namespace, req.AppName)
	if err != nil {
		return nil, err
	}

	if req.Image != "" {
		app.Spec.Image = req.Image
	}
	if req.Replicas != nil {
		app.Spec.Replicas = req.Replicas
	}

	if err := d.Update(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}
