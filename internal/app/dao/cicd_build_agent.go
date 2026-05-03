package dao

import (
	"context"
	"k8soperation/internal/app/models"
	"time"

	"gorm.io/gorm"
)

// BuildAgentCreate 创建构建探针
func (d *Dao) BuildAgentCreate(ctx context.Context, agent *models.CicdBuildAgent) error {
	now := uint64(time.Now().Unix())
	agent.CreatedAt = now
	agent.ModifiedAt = now
	return d.db.WithContext(ctx).Create(agent).Error
}

// BuildAgentGetByID 根据 ID 获取探针
func (d *Dao) BuildAgentGetByID(ctx context.Context, id int64) (*models.CicdBuildAgent, error) {
	var agent models.CicdBuildAgent
	err := d.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&agent).Error
	return &agent, err
}

// BuildAgentGetByName 根据名称获取探针
func (d *Dao) BuildAgentGetByName(ctx context.Context, name string) (*models.CicdBuildAgent, error) {
	var agent models.CicdBuildAgent
	err := d.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&agent).Error
	return &agent, err
}

// BuildAgentList 探针列表（分页 + 筛选）
func (d *Dao) BuildAgentList(ctx context.Context, category, scope, status, keyword string, page, pageSize int) ([]*models.CicdBuildAgent, int64, error) {
	db := d.db.WithContext(ctx).Model(&models.CicdBuildAgent{}).Where("is_del = 0")

	if category != "" {
		db = db.Where("category = ?", category)
	}
	if scope != "" {
		db = db.Where("scope = ? OR scope = ?", scope, models.AgentScopeAll)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*models.CicdBuildAgent
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// BuildAgentListByScope 获取指定语言的已启用探针（流水线构建时使用）
func (d *Dao) BuildAgentListByScope(ctx context.Context, scope string) ([]*models.CicdBuildAgent, error) {
	var list []*models.CicdBuildAgent
	err := d.db.WithContext(ctx).
		Where("is_del = 0 AND status = ? AND (scope = ? OR scope = ?)", models.AgentStatusActive, scope, models.AgentScopeAll).
		Order("category, name").
		Find(&list).Error
	return list, err
}

// BuildAgentUpdate 更新探针信息
func (d *Dao) BuildAgentUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdBuildAgent{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}

// BuildAgentDelete 软删除探针
func (d *Dao) BuildAgentDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdBuildAgent{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// BuildAgentIncrDownload 增加下载计数
func (d *Dao) BuildAgentIncrDownload(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.CicdBuildAgent{}).
		Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

// BuildAgentIncrUsed 增加引用计数
func (d *Dao) BuildAgentIncrUsed(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.CicdBuildAgent{}).
		Where("id = ?", id).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}
