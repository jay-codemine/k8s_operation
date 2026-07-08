// internal/app/services/appconfig.go
package services

import (
	"context"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/crd"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
)

func (s *Services) KubeAppConfigGet(ctx context.Context, req *requests.KubeAppConfigDetailRequest) (*appv1alpha1.AppConfig, error) {
	d, err := s.buildAppConfigDAO(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return crd.GetAppConfig(ctx, d, req)
}

func (s *Services) KubeAppConfigUpdate(ctx context.Context, req *requests.KubeAppConfigUpdateRequest) (*appv1alpha1.AppConfig, error) {
	d, err := s.buildAppConfigDAO(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return crd.UpdateAppConfig(ctx, d, req)
}

func (s *Services) KubeAppConfigCreate(ctx context.Context, req *requests.KubeAppConfigCreateRequest) (*appv1alpha1.AppConfig, error) {
	d, err := s.buildAppConfigDAO(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return crd.CreateAppConfig(ctx, d, req)
}

func (s *Services) KubeAppConfigDelete(ctx context.Context, req *requests.KubeAppConfigDeleteRequest) error {
	d, err := s.buildAppConfigDAO(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	return crd.DeleteAppConfig(ctx, d, req)
}

func (s *Services) KubeAppConfigList(ctx context.Context, req *requests.KubeAppConfigListRequest) ([]appv1alpha1.AppConfig, error) {
	d, err := s.buildAppConfigDAO(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return crd.ListAppConfig(ctx, d, req)
}
