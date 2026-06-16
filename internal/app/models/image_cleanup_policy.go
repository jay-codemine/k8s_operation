package models

import (
	"k8soperation/global"

	"gorm.io/gorm"
)

// ImageCleanupPolicy 镜像清理策略
type ImageCleanupPolicy struct {
	ID                int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	RegistryID        int64  `gorm:"type:bigint;not null;index:idx_registry_id" json:"registry_id"`
	Name              string `gorm:"type:varchar(100);not null" json:"name"`
	Enabled           bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	RepositoryPattern string `gorm:"type:varchar(200);default:'*'" json:"repository_pattern"`
	TagPattern        string `gorm:"type:varchar(200);default:'*'" json:"tag_pattern"`
	KeepLastCount     int    `gorm:"type:int;default:5" json:"keep_last_count"`
	KeepDays          int    `gorm:"type:int;default:30" json:"keep_days"`
	CronExpression    string `gorm:"type:varchar(50);default:'0 2 * * *'" json:"cron_expression"`
	LastRunAt         int64  `gorm:"type:bigint" json:"last_run_at"`
	LastRunResult     string `gorm:"type:varchar(500)" json:"last_run_result"`
	DeletedCount      int64  `gorm:"type:bigint;default:0" json:"deleted_count"`
	Description       string `gorm:"type:varchar(500)" json:"description"`
	CreatedBy         int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt         int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt        int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel             int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (ImageCleanupPolicy) TableName() string {
	return "image_cleanup_policy"
}

// ImageCleanupLog 清理任务日志
type ImageCleanupLog struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PolicyID     int64  `gorm:"type:bigint;not null;index:idx_policy_id" json:"policy_id"`
	RegistryID   int64  `gorm:"type:bigint;not null" json:"registry_id"`
	StartTime    int64  `gorm:"type:bigint;not null;index:idx_start_time" json:"start_time"`
	EndTime      int64  `gorm:"type:bigint" json:"end_time"`
	Status       string `gorm:"type:varchar(20);default:'running'" json:"status"`
	ScannedCount int    `gorm:"type:int;default:0" json:"scanned_count"`
	DeletedCount int    `gorm:"type:int;default:0" json:"deleted_count"`
	FreedSize    int64  `gorm:"type:bigint;default:0" json:"freed_size"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`
	Details      string `gorm:"type:json" json:"details"`
}

func (ImageCleanupLog) TableName() string {
	return "image_cleanup_log"
}

// ImageCleanupPolicyModel 清理策略模型操作
type ImageCleanupPolicyModel struct {
	db *gorm.DB
}

func NewImageCleanupPolicyModel() *ImageCleanupPolicyModel {
	return &ImageCleanupPolicyModel{db: global.DB}
}

// Create 创建清理策略
func (m *ImageCleanupPolicyModel) Create(policy *ImageCleanupPolicy) error {
	return m.db.Create(policy).Error
}

// Update 更新清理策略
func (m *ImageCleanupPolicyModel) Update(policy *ImageCleanupPolicy) error {
	return m.db.Model(policy).Updates(map[string]interface{}{
		"name":               policy.Name,
		"enabled":            policy.Enabled,
		"repository_pattern": policy.RepositoryPattern,
		"tag_pattern":        policy.TagPattern,
		"keep_last_count":    policy.KeepLastCount,
		"keep_days":          policy.KeepDays,
		"cron_expression":    policy.CronExpression,
		"description":        policy.Description,
		"modified_at":        policy.ModifiedAt,
	}).Error
}

// UpdateRunResult 更新执行结果
func (m *ImageCleanupPolicyModel) UpdateRunResult(id int64, result string, deletedCount int64) error {
	return m.db.Model(&ImageCleanupPolicy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_run_at":     gorm.Expr("UNIX_TIMESTAMP()"),
		"last_run_result": result,
		"deleted_count":   gorm.Expr("deleted_count + ?", deletedCount),
	}).Error
}

// Delete 软删除
func (m *ImageCleanupPolicyModel) Delete(id int64) error {
	return m.db.Model(&ImageCleanupPolicy{}).Where("id = ?", id).Update("is_del", 1).Error
}

// GetByID 根据ID获取
func (m *ImageCleanupPolicyModel) GetByID(id int64) (*ImageCleanupPolicy, error) {
	var policy ImageCleanupPolicy
	err := m.db.Where("id = ? AND is_del = 0", id).First(&policy).Error
	return &policy, err
}

// ListByRegistry 根据仓库ID获取策略列表
func (m *ImageCleanupPolicyModel) ListByRegistry(registryID int64) ([]ImageCleanupPolicy, error) {
	var policies []ImageCleanupPolicy
	err := m.db.Where("registry_id = ? AND is_del = 0", registryID).Order("id DESC").Find(&policies).Error
	return policies, err
}

// ListEnabled 获取所有启用的策略
func (m *ImageCleanupPolicyModel) ListEnabled() ([]ImageCleanupPolicy, error) {
	var policies []ImageCleanupPolicy
	err := m.db.Where("enabled = 1 AND is_del = 0").Find(&policies).Error
	return policies, err
}

// List 分页列表
func (m *ImageCleanupPolicyModel) List(registryID int64, keyword string, page, pageSize int) ([]ImageCleanupPolicy, int64, error) {
	var policies []ImageCleanupPolicy
	var total int64

	query := m.db.Model(&ImageCleanupPolicy{}).Where("is_del = 0")
	if registryID > 0 {
		query = query.Where("registry_id = ?", registryID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&policies).Error; err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}

// ImageCleanupLogModel 清理日志模型操作
type ImageCleanupLogModel struct {
	db *gorm.DB
}

func NewImageCleanupLogModel() *ImageCleanupLogModel {
	return &ImageCleanupLogModel{db: global.DB}
}

// Create 创建日志
func (m *ImageCleanupLogModel) Create(log *ImageCleanupLog) error {
	return m.db.Create(log).Error
}

// Update 更新日志
func (m *ImageCleanupLogModel) Update(log *ImageCleanupLog) error {
	return m.db.Save(log).Error
}

// ListByPolicy 根据策略ID获取日志
func (m *ImageCleanupLogModel) ListByPolicy(policyID int64, limit int) ([]ImageCleanupLog, error) {
	var logs []ImageCleanupLog
	err := m.db.Where("policy_id = ?", policyID).Order("start_time DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// ListRecent 获取最近的日志
func (m *ImageCleanupLogModel) ListRecent(limit int) ([]ImageCleanupLog, error) {
	var logs []ImageCleanupLog
	err := m.db.Order("start_time DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
