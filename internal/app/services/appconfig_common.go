package services

import (
	"context"
	"fmt"

	"k8soperation/global"

	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/crd"

	"k8s.io/client-go/rest"
)

// 你之前的 GetRestConfig 放在 services 包里就行：
func getRestConfig(ctx context.Context, clusterID uint32) (*rest.Config, error) {
	if clusterID == 0 {
		clusterID = global.AppSetting.DefaultClusterID
	}

	svc := NewServices()
	cli, err := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: clusterID})
	if err != nil {
		return nil, fmt.Errorf("K8sClusterInit failed: %w", err)
	}
	return cli.Config, nil
}

// 根据 clusterID 构建 AppConfig DAO（services 内部使用）
func (s *Services) buildAppConfigDAO(ctx context.Context, clusterID uint32) (*dao.KubeAppConfig, error) {
	cfg, err := getRestConfig(ctx, clusterID) // 这个 getRestConfig 就放在 services 包里
	if err != nil {
		return nil, err
	}

	cli, err := crd.NewAppConfigRuntimeClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewAppConfigClient failed: %w", err)
	}

	return dao.NewKubeAppConfig(cli), nil
}
