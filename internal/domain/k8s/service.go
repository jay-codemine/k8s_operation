package k8s

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"k8soperation/internal/domain/events"
	"k8soperation/pkg/logger"
	"k8soperation/pkg/utils"
)

// ClusterService K8s 集群领域服务
type ClusterService struct {
	repo      ClusterRepository
	logger    *logger.Logger
	publisher events.EventPublisher
}

// NewClusterService 创建集群服务
func NewClusterService(repo ClusterRepository, logger *logger.Logger, publisher events.EventPublisher) *ClusterService {
	return &ClusterService{repo: repo, logger: logger, publisher: publisher}
}

// ClusterWithPlain kubeconfig 解密后的集群信息
type ClusterWithPlain struct {
	ID          uint32
	ClusterName string
	KubeConfig  string
	ClusterVer  string
	Status      uint8
	ModifiedAt  uint64
}

// Create 创建集群（kubeconfig 明文入，加密落库 —— 领域逻辑 + 值对象验证）
func (s *ClusterService) Create(ctx context.Context, clusterName, clusterVersion, kubeConfigPlain string) error {
	name, err := NewClusterName(clusterName)
	if err != nil {
		return err
	}
	ver, _ := NewClusterVersion(clusterVersion)
	kcCfg, err := NewKubeConfigFromPlain(kubeConfigPlain)
	if err != nil {
		return err
	}
	now := utils.NowUnix()
	kc := &Cluster{
		ClusterName:    name.String(),
		ClusterVersion: ver.String(),
		KubeConfig:     kcCfg.Encrypted(),
		Status:         0,
		CreatedAt:      now,
		ModifiedAt:     now,
	}
	if err := s.repo.Save(ctx, kc); err != nil {
		s.logger.Error("创建集群失败", zap.String("cluster_name", clusterName), zap.Error(err))
		return err
	}
	s.publish(NewClusterCreated(uint32(kc.ID), clusterName))
	return nil
}

// GetByName 根据名称获取集群
func (s *ClusterService) GetByName(ctx context.Context, name string) (*Cluster, error) {
	return s.repo.FindByName(ctx, name)
}

// GetByID 根据 ID 获取集群（解密 kubeconfig —— 值对象解密）
func (s *ClusterService) GetByID(ctx context.Context, id uint32) (*ClusterWithPlain, error) {
	cluster, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cluster not found: %d", id)
		}
		return nil, err
	}
	kc := NewKubeConfigFromEncrypted(cluster.KubeConfig)
	plain, err := kc.Decrypt()
	if err != nil {
		return nil, err
	}
	return &ClusterWithPlain{
		ID: cluster.ID, ClusterName: cluster.ClusterName,
		KubeConfig: plain, ClusterVer: cluster.ClusterVersion,
		Status: cluster.Status, ModifiedAt: cluster.ModifiedAt,
	}, nil
}

// GetByIDEncrypted 根据 ID 获取集群（不解密）
func (s *ClusterService) GetByIDEncrypted(ctx context.Context, id uint32) (*Cluster, error) {
	return s.repo.FindByID(ctx, id)
}

// MarkCheckResult 标记集群健康检查结果
func (s *ClusterService) MarkCheckResult(ctx context.Context, id uint32, status uint8, checkAt uint64, lastErr string) error {
	return s.repo.Update(ctx, id, map[string]interface{}{
		"status": status, "last_check_at": checkAt,
		"last_error": lastErr, "modified_at": utils.NowUnix(),
	})
}

// Update 更新集群（值对象验证 + kubeconfig 变更时重置健康状态）
func (s *ClusterService) Update(ctx context.Context, id uint32, clusterName, clusterVersion, kubeConfigPlain string, hasKC bool) error {
	name, err := NewClusterName(clusterName)
	if err != nil {
		return err
	}
	ver, _ := NewClusterVersion(clusterVersion)
	values := map[string]interface{}{
		"cluster_name": name.String(), "cluster_version": ver.String(),
		"modified_at": utils.NowUnix(),
	}
	if hasKC {
		kcCfg, err := NewKubeConfigFromPlain(kubeConfigPlain)
		if err != nil {
			return err
		}
		values["kube_config"] = kcCfg.Encrypted()
		values["status"] = uint8(0)
		values["last_check_at"] = uint32(0)
		values["last_error"] = ""
	}
	if err := s.repo.Update(ctx, id, values); err != nil {
		s.logger.Error("更新集群失败", zap.Uint32("cluster_id", id), zap.Error(err))
		return err
	}
	return nil
}

// List 集群列表
func (s *ClusterService) List(ctx context.Context, clusterName string, page, limit int) ([]*Cluster, int64, error) {
	return s.repo.Query(ctx, clusterName, page, limit)
}

// Delete 软删除集群
func (s *ClusterService) Delete(ctx context.Context, id uint32) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		s.logger.Error("软删除集群失败", zap.Uint32("cluster_id", id), zap.Error(err))
		return err
	}
	s.logger.Info("软删除集群成功", zap.Uint32("cluster_id", id))
	s.publish(NewClusterDeleted(id, ""))
	return nil
}

// BatchDelete 批量软删除
func (s *ClusterService) BatchDelete(ctx context.Context, ids []uint32) (int64, error) {
	return s.repo.BatchDelete(ctx, ids)
}

// UpdateHealth 更新集群健康状态
func (s *ClusterService) UpdateHealth(ctx context.Context, clusterID uint32, status uint8, lastErr string) error {
	if err := s.repo.UpdateHealth(ctx, clusterID, status, lastErr); err != nil {
		s.logger.Error("更新集群健康状态失败", zap.Uint32("cluster_id", clusterID), zap.Error(err))
		return err
	}
	return nil
}

// publish 发布领域事件（静默失败，不影响主流程）
func (s *ClusterService) publish(event events.DomainEvent) {
	if s.publisher == nil {
		return
	}
	defer func() { _ = recover() }()
	s.publisher.Publish(event)
}
