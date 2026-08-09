package services

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
	"k8soperation/internal/infra/persistence"
	"k8soperation/pkg/utils"

	dm "k8soperation/internal/domain/k8s"
)

func (s *Services) clusterSvc() *dm.ClusterService {
	return dm.NewClusterService(persistence.NewClusterRepository(s.db), s.logger, s.eventBus)
}

func k8sConnTimeout() int {
	if global.ClusterTTL != nil {
		return global.ClusterTTL.ConnectionTimeout
	}
	return 0
}

func (s *Services) K8sClusterCreate(ctx context.Context, param *requests.K8sClusterCreateRequest) error {
	return s.clusterSvc().Create(ctx, param.ClusterName, param.ClusterVersion, param.KubeConfig)
}

// K8sClusterGetByName 按名称查询集群（用于启动时复用已有记录，避免重复创建）
func (s *Services) K8sClusterGetByName(ctx context.Context, name string) (*models.K8sCluster, error) {
	return s.clusterSvc().GetByName(ctx, name)
}

func (s *Services) K8sClusterUpdate(ctx context.Context, param *requests.K8sClusterUpdateRequest) error {
	kubeConfigPlain := strings.TrimSpace(param.KubeConfig)
	hasKC := kubeConfigPlain != ""
	return s.clusterSvc().Update(ctx,
		param.ID,
		param.ClusterName,
		param.ClusterVersion,
		kubeConfigPlain,
		hasKC,
	)
}

func (s *Services) K8sClusterDelete(ctx context.Context, param *requests.K8sClusterDeleteRequest) error {
	return s.clusterSvc().Delete(ctx, param.ID)
}

// K8sClusterBatchDelete 批量删除集群
func (s *Services) K8sClusterBatchDelete(ctx context.Context, ids []uint32) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("集群ID列表不能为空")
	}
	return s.clusterSvc().BatchDelete(ctx, ids)
}

func (s *Services) K8sClusterList(ctx context.Context, param *requests.K8sClusterListRequest,
) (list []*models.K8sCluster, total int64, err error) {
	return s.clusterSvc().List(ctx, param.ClusterName, param.Page, param.Limit)
}

func (s *Services) K8sClusterInit(ctx context.Context, param *requests.K8sClusterInitRequest) (*K8sClients, error) {
	now := utils.NowUnix()

	cfg, err := s.mustFromDB(ctx, param.ID)
	if err != nil {
		_ = s.clusterSvc().MarkCheckResult(ctx, param.ID, models.ClusterStatusBad, now, err.Error())
		return nil, errorcode.ErrorClusterInitFailed
	}

	clients, err := dm.BuildClients(cfg, k8sConnTimeout())
	if err != nil {
		_ = s.clusterSvc().MarkCheckResult(ctx, param.ID, models.ClusterStatusBad, now, err.Error())
		return nil, errorcode.ErrorClusterInitFailed
	}

	if err := s.clusterSvc().MarkCheckResult(ctx, param.ID, models.ClusterStatusOK, now, ""); err != nil {
		global.Logger.Warn("mark check result failed", zap.Uint32("cluster_id", param.ID), zap.Error(err))
	}

	return clients, nil
}

// GetCluster 实现 domain/k8s.ClusterClientProvider 接口
func (s *Services) GetCluster(ctx context.Context, clusterID uint32) (*dm.Cluster, error) {
	c, err := s.clusterSvc().GetByIDEncrypted(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	return &dm.Cluster{
		ID:             c.ID,
		ClusterName:    c.ClusterName,
		ClusterVersion: c.ClusterVersion,
		KubeConfig:     c.KubeConfig,
		Status:         c.Status,
		ModifiedAt:     c.ModifiedAt,
	}, nil
}

// BuildClientsForCluster 实现 domain/k8s.ClusterClientProvider 接口
func (s *Services) BuildClientsForCluster(ctx context.Context, clusterID uint32) (*dm.K8sClients, error) {
	cfg, err := s.mustFromDB(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	return dm.BuildClients(cfg, k8sConnTimeout())
}

// mustFromDB 从DB获取明文kubeconfig
func (s *Services) mustFromDB(ctx context.Context, clusterID uint32) (*rest.Config, error) {
	kc, err := s.clusterSvc().GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	plain := strings.TrimSpace(kc.KubeConfig)
	if plain == "" {
		return nil, fmt.Errorf("empty kubeconfig in DB, cluster=%d", clusterID)
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(plain))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig failed: %w", err)
	}

	global.Logger.Info("init from DB kubeconfig (plain)", zap.Uint32("cluster_id", clusterID))
	return cfg, nil
}
