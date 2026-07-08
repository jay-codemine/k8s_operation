package crd

import (
	"context"
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/requests"
)

func CreateAppConfig(ctx context.Context, d *dao.KubeAppConfig, req *requests.KubeAppConfigCreateRequest) (*appv1alpha1.AppConfig, error) {
	app := &appv1alpha1.AppConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.AppName,
			Namespace: req.Namespace,
		},
		Spec: appv1alpha1.AppConfigSpec{
			AppName:  req.AppName,
			Image:    req.Image,
			Replicas: req.Replicas,
		},
	}

	if err := d.Create(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}
