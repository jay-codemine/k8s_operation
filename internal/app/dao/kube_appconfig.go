package dao

import (
	"context"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeAppConfig struct {
	cli client.Client
}

func NewKubeAppConfig(cli client.Client) *KubeAppConfig {
	return &KubeAppConfig{cli: cli}
}

func (d *KubeAppConfig) Get(ctx context.Context, namespace, name string) (*appv1alpha1.AppConfig, error) {
	var app appv1alpha1.AppConfig
	if err := d.cli.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (d *KubeAppConfig) List(ctx context.Context, namespace string) ([]appv1alpha1.AppConfig, error) {
	var list appv1alpha1.AppConfigList
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := d.cli.List(ctx, &list, opts...); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (d *KubeAppConfig) Create(ctx context.Context, app *appv1alpha1.AppConfig) error {
	return d.cli.Create(ctx, app)
}

func (d *KubeAppConfig) Update(ctx context.Context, app *appv1alpha1.AppConfig) error {
	return d.cli.Update(ctx, app)
}

func (d *KubeAppConfig) Delete(ctx context.Context, namespace, name string) error {
	app := &appv1alpha1.AppConfig{}
	app.Namespace = namespace
	app.Name = name
	return d.cli.Delete(ctx, app)
}
