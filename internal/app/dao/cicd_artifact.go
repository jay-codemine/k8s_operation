package dao

import (
	"context"
	"fmt"
	"k8soperation/internal/app/models"
	"time"

	"gorm.io/gorm"
)

// ArtifactCreate 创建制品记录
func (d *Dao) ArtifactCreate(ctx context.Context, artifact *models.CicdArtifact) error {
	now := uint64(time.Now().Unix())
	artifact.CreatedAt = now
	artifact.ModifiedAt = now
	return d.db.WithContext(ctx).Create(artifact).Error
}

// ArtifactGetByID 根据 ID 获取制品
func (d *Dao) ArtifactGetByID(ctx context.Context, id int64) (*models.CicdArtifact, error) {
	var artifact models.CicdArtifact
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&artifact).Error
	return &artifact, err
}

// ArtifactList 制品列表（分页 + 筛选）
func (d *Dao) ArtifactList(ctx context.Context, pipelineID int64, artifactType, languageType, status string, page, pageSize int) ([]*models.CicdArtifact, int64, error) {
	db := d.db.WithContext(ctx).Model(&models.CicdArtifact{}).Where("is_del = 0")

	if pipelineID > 0 {
		db = db.Where("pipeline_id = ?", pipelineID)
	}
	if artifactType != "" {
		db = db.Where("artifact_type = ?", artifactType)
	}
	if languageType != "" {
		db = db.Where("language_type = ?", languageType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*models.CicdArtifact
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ArtifactListByRunID 获取某次运行产出的所有制品
func (d *Dao) ArtifactListByRunID(ctx context.Context, runID int64) ([]*models.CicdArtifact, error) {
	var list []*models.CicdArtifact
	err := d.db.WithContext(ctx).
		Where("run_id = ? AND is_del = 0", runID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// ArtifactUpdate 更新制品信息
func (d *Dao) ArtifactUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdArtifact{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}

// ArtifactDelete 软删除制品
func (d *Dao) ArtifactDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdArtifact{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
			"status":      models.ArtifactStatusDeleted,
		}).Error
}

// ArtifactIncrDownload 增加下载计数
func (d *Dao) ArtifactIncrDownload(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.CicdArtifact{}).
		Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

// ArtifactBatchDelete 批量软删除制品
func (d *Dao) ArtifactBatchDelete(ctx context.Context, ids []int64) (int64, error) {
	now := time.Now().Unix()
	result := d.db.WithContext(ctx).
		Model(&models.CicdArtifact{}).
		Where("id IN ? AND is_del = 0", ids).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
			"status":      models.ArtifactStatusDeleted,
		})
	return result.RowsAffected, result.Error
}

// ArtifactGetByIDs 批量获取制品（用于批量删除时清理文件）
func (d *Dao) ArtifactGetByIDs(ctx context.Context, ids []int64) ([]*models.CicdArtifact, error) {
	var list []*models.CicdArtifact
	err := d.db.WithContext(ctx).
		Where("id IN ? AND is_del = 0", ids).
		Find(&list).Error
	return list, err
}

// ArtifactStats 制品统计（按类型分组）
func (d *Dao) ArtifactStats(ctx context.Context, pipelineID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	db := d.db.WithContext(ctx).
		Model(&models.CicdArtifact{}).
		Select("artifact_type, COUNT(*) as count, SUM(file_size) as total_size").
		Where("is_del = 0 AND status = ?", models.ArtifactStatusReady)

	if pipelineID > 0 {
		db = db.Where("pipeline_id = ?", pipelineID)
	}

	err := db.Group("artifact_type").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询制品统计失败: %w", err)
	}
	return results, nil
}
