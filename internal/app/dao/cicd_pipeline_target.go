package dao

import (
	"context"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/tenant"
	"time"
)

// ==================== PipelineTarget CRUD（流水线多环境部署目标）====================

// PipelineTargetListByPipeline 获取某条流水线的全部环境目标（按 sort_order 升序）
func (d *Dao) PipelineTargetListByPipeline(ctx context.Context, pipelineID int64) ([]*models.CicdPipelineTargetView, error) {
	var list []*models.CicdPipelineTargetView
	tid, ok := tenant.GetTenantID(d.db)
	if !ok {
		return nil, nil
	}

	err := global.DB.WithContext(ctx).
		Table("cicd_pipeline_target AS t").
		Select("t.*, COALESCE(c.cluster_name, '') AS cluster_name").
		Joins("LEFT JOIN kube_cluster c ON c.id = t.cluster_id AND c.tenant_id = ? AND c.is_del = 0", tid).
		Where("t.pipeline_id = ? AND t.is_del = 0 AND t.tenant_id = ?", pipelineID, tid).
		Order("t.sort_order ASC, t.id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PipelineTargetGetByPipelineAndEnv 按流水线+环境获取部署目标
func (d *Dao) PipelineTargetGetByPipelineAndEnv(ctx context.Context, pipelineID int64, env string) (*models.CicdPipelineTarget, error) {
	var t models.CicdPipelineTarget
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND env = ? AND is_del = 0", pipelineID, env).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// PipelineTargetCreate 创建环境目标
func (d *Dao) PipelineTargetCreate(ctx context.Context, t *models.CicdPipelineTarget) error {
	now := uint64(time.Now().Unix())
	t.CreatedAt = now
	t.ModifiedAt = now
	return d.db.WithContext(ctx).Create(t).Error
}

// PipelineTargetUpdate 更新环境目标（按 ID）
func (d *Dao) PipelineTargetUpdate(ctx context.Context, id int64, updates map[string]any) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipelineTarget{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}

// PipelineTargetDelete 软删除环境目标
func (d *Dao) PipelineTargetDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipelineTarget{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]any{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// PipelineTargetUpsert 按 (pipeline_id, env) 插入或更新（幂等保存单个环境目标）
func (d *Dao) PipelineTargetUpsert(ctx context.Context, t *models.CicdPipelineTarget) error {
	exist, err := d.PipelineTargetGetByPipelineAndEnv(ctx, t.PipelineID, t.Env)
	if err == nil && exist != nil {
		return d.PipelineTargetUpdate(ctx, exist.ID, map[string]any{
			"cluster_id":       t.ClusterID,
			"namespace":        t.Namespace,
			"workload_kind":    t.WorkloadKind,
			"workload_name":    t.WorkloadName,
			"container":        t.Container,
			"auto_deploy":      t.AutoDeploy,
			"require_approval": t.RequireApproval,
			"promote_from":     t.PromoteFrom,
			"sort_order":       t.SortOrder,
		})
	}
	return d.PipelineTargetCreate(ctx, t)
}

// ==================== Release 晋级链查询 ====================

// CicdReleaseLatestByPipelineEnv 获取某条流水线在某环境的最新一条发布单
// 用于晋级链可视化：展示各环境当前部署的镜像与状态
func (d *Dao) CicdReleaseLatestByPipelineEnv(ctx context.Context, pipelineID int64, env string) (*models.CicdRelease, error) {
	var rel models.CicdRelease
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND env = ? AND is_del = 0", pipelineID, env).
		Order("id DESC").
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// CicdReleaseLatestBySourceRunID 获取由某次构建(run)产出镜像的最新一条已定环境的发布单
// 用于镜像晋级时反查该镜像当前所处环境，补全晋级链 source_env（避免链路断裂）
func (d *Dao) CicdReleaseLatestBySourceRunID(ctx context.Context, pipelineID, sourceRunID int64) (*models.CicdRelease, error) {
	var rel models.CicdRelease
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND source_run_id = ? AND env <> '' AND is_del = 0", pipelineID, sourceRunID).
		Order("id DESC").
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
