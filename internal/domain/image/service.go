package image

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ImageService 镜像领域服务
type ImageService struct {
	db   *gorm.DB
	repo ImageRepository
}

// NewImageService 创建镜像服务（db 兼容旧代码，repo 供新代码使用）
func NewImageService(db *gorm.DB, repo ImageRepository) *ImageService {
	return &ImageService{db: db, repo: repo}
}


// ============================================================
// 镜像仓库 CRUD
// ============================================================

// RegistryCreate 创建镜像仓库
func (s *ImageService) RegistryCreate(registry *ImageRegistry) error {
	return s.repo.RegistrySave(context.Background(), registry)
}

// RegistryUpdate 更新镜像仓库
func (s *ImageService) RegistryUpdate(registry *ImageRegistry) error {
	return s.db.Model(registry).Updates(map[string]interface{}{
		"name":              registry.Name,
		"type":              registry.Type,
		"url":               registry.URL,
		"username":          registry.Username,
		"password":          registry.Password,
		"access_key_id":     registry.AccessKeyID,
		"access_key_secret": registry.AccessKeySecret,
		"region":            registry.Region,
		"insecure":          registry.Insecure,
		"description":       registry.Description,
		"is_default":        registry.IsDefault,
		"modified_at":       registry.ModifiedAt,
	}).Error
}

// RegistryUpdateStatus 更新连接状态
func (s *ImageService) RegistryUpdateStatus(id int64, status, lastError string, checkTime int64) error {
	return s.db.Model(&ImageRegistry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"last_error":    lastError,
		"last_check_at": checkTime,
	}).Error
}

// RegistryDelete 软删除镜像仓库
func (s *ImageService) RegistryDelete(id int64) error {
	return s.repo.RegistryDelete(context.Background(), id)
}

// RegistryGetByID 根据ID获取镜像仓库
func (s *ImageService) RegistryGetByID(id int64) (*ImageRegistry, error) {
	var registry ImageRegistry
	err := s.db.Where("id = ? AND is_del = 0", id).First(&registry).Error
	return &registry, err
}

// RegistryGetByName 根据名称获取镜像仓库
func (s *ImageService) RegistryGetByName(name string) (*ImageRegistry, error) {
	var registry ImageRegistry
	err := s.db.Where("name = ? AND is_del = 0", name).First(&registry).Error
	return &registry, err
}

// RegistryList 获取镜像仓库列表（分页）
func (s *ImageService) RegistryList(keyword string, registryType string, page, pageSize int) ([]ImageRegistry, int64, error) {
	var registries []ImageRegistry
	var total int64

	query := s.db.Model(&ImageRegistry{}).Where("is_del = 0")

	if keyword != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if registryType != "" {
		query = query.Where("type = ?", registryType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&registries).Error; err != nil {
		return nil, 0, err
	}

	return registries, total, nil
}

// RegistryListAll 获取所有镜像仓库（不分页）
func (s *ImageService) RegistryListAll() ([]ImageRegistry, error) {
	var registries []ImageRegistry
	err := s.db.Where("is_del = 0").Order("is_default DESC, id ASC").Find(&registries).Error
	return registries, err
}

// RegistryGetDefault 获取默认仓库
func (s *ImageService) RegistryGetDefault() (*ImageRegistry, error) {
	var registry ImageRegistry
	err := s.db.Where("is_default = 1 AND is_del = 0").First(&registry).Error
	return &registry, err
}

// RegistrySetDefault 设置默认仓库
func (s *ImageService) RegistrySetDefault(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ImageRegistry{}).Where("is_del = 0").Update("is_default", 0).Error; err != nil {
			return err
		}
		return tx.Model(&ImageRegistry{}).Where("id = ?", id).Update("is_default", 1).Error
	})
}

// RegistryExistsByName 检查名称是否存在
func (s *ImageService) RegistryExistsByName(name string, excludeID int64) (bool, error) {
	var count int64
	query := s.db.Model(&ImageRegistry{}).Where("name = ? AND is_del = 0", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// RegistryCreateWithValidation 创建镜像仓库（含业务校验）
func (s *ImageService) RegistryCreateWithValidation(name, regType, url, username, password, accessKeyID, accessKeySecret, region, description string, insecure, isDefault bool, userID int64) (*ImageRegistry, error) {
	exists, err := s.RegistryExistsByName(name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("仓库名称 '%s' 已存在", name)
	}

	registry := &ImageRegistry{
		Name:            name,
		Type:            regType,
		URL:             strings.TrimSuffix(url, "/"),
		Username:        username,
		Password:        password,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Region:          region,
		Insecure:        insecure,
		Description:     description,
		IsDefault:       isDefault,
		Status:          "unknown",
		CreatedBy:       userID,
	}

	if err := s.RegistryCreate(registry); err != nil {
		return nil, err
	}

	if isDefault {
		_ = s.RegistrySetDefault(registry.ID)
	}

	return registry, nil
}

// RegistryUpdateWithValidation 更新镜像仓库（含业务校验）
func (s *ImageService) RegistryUpdateWithValidation(id int64, name, regType, url, username, password, accessKeyID, accessKeySecret, region, description string, insecure, isDefault bool) (*ImageRegistry, error) {
	existing, err := s.RegistryGetByID(id)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	exists, err := s.RegistryExistsByName(name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("仓库名称 '%s' 已存在", name)
	}

	existing.Name = name
	existing.Type = regType
	existing.URL = strings.TrimSuffix(url, "/")
	existing.Username = username
	if password != "" {
		existing.Password = password
	}
	existing.AccessKeyID = accessKeyID
	if accessKeySecret != "" {
		existing.AccessKeySecret = accessKeySecret
	}
	existing.Region = region
	existing.Insecure = insecure
	existing.Description = description
	existing.IsDefault = isDefault

	if err := s.RegistryUpdate(existing); err != nil {
		return nil, err
	}

	if isDefault {
		_ = s.RegistrySetDefault(existing.ID)
	}

	return existing, nil
}

// RegistryDeleteWithValidation 删除镜像仓库（含存在性校验）
func (s *ImageService) RegistryDeleteWithValidation(id int64) error {
	_, err := s.RegistryGetByID(id)
	if err != nil {
		return fmt.Errorf("仓库不存在")
	}
	return s.RegistryDelete(id)
}

// RegistrySetDefaultWithValidation 设置默认仓库（含存在性校验）
func (s *ImageService) RegistrySetDefaultWithValidation(id int64) error {
	_, err := s.RegistryGetByID(id)
	if err != nil {
		return fmt.Errorf("仓库不存在")
	}
	return s.RegistrySetDefault(id)
}

// ============================================================
// 清理策略 CRUD
// ============================================================

// PolicyCreate 创建清理策略
func (s *ImageService) PolicyCreate(policy *ImageCleanupPolicy) error {
	return s.repo.PolicySave(context.Background(), policy)
}

// PolicyUpdate 更新清理策略
func (s *ImageService) PolicyUpdate(policy *ImageCleanupPolicy) error {
	return s.db.Model(policy).Updates(map[string]interface{}{
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

// PolicyUpdateRunResult 更新执行结果
func (s *ImageService) PolicyUpdateRunResult(id int64, result string, deletedCount int64) error {
	return s.db.Model(&ImageCleanupPolicy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_run_at":     gorm.Expr("UNIX_TIMESTAMP()"),
		"last_run_result": result,
		"deleted_count":   gorm.Expr("deleted_count + ?", deletedCount),
	}).Error
}

// PolicyDelete 软删除清理策略
func (s *ImageService) PolicyDelete(id int64) error {
	return s.repo.PolicyDelete(context.Background(), id)
}

// PolicyGetByID 根据ID获取清理策略
func (s *ImageService) PolicyGetByID(id int64) (*ImageCleanupPolicy, error) {
	var policy ImageCleanupPolicy
	err := s.db.Where("id = ? AND is_del = 0", id).First(&policy).Error
	return &policy, err
}

// PolicyListByRegistry 根据仓库ID获取策略列表
func (s *ImageService) PolicyListByRegistry(registryID int64) ([]ImageCleanupPolicy, error) {
	var policies []ImageCleanupPolicy
	err := s.db.Where("registry_id = ? AND is_del = 0", registryID).Order("id DESC").Find(&policies).Error
	return policies, err
}

// PolicyListEnabled 获取所有启用的策略
func (s *ImageService) PolicyListEnabled() ([]ImageCleanupPolicy, error) {
	var policies []ImageCleanupPolicy
	err := s.db.Where("enabled = 1 AND is_del = 0").Find(&policies).Error
	return policies, err
}

// PolicyList 分页查询清理策略
func (s *ImageService) PolicyList(registryID int64, keyword string, page, pageSize int) ([]ImageCleanupPolicy, int64, error) {
	var policies []ImageCleanupPolicy
	var total int64

	query := s.db.Model(&ImageCleanupPolicy{}).Where("is_del = 0")
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

// PolicyDeleteWithValidation 删除清理策略（含存在性校验）
func (s *ImageService) PolicyDeleteWithValidation(id int64) error {
	_, err := s.PolicyGetByID(id)
	if err != nil {
		return fmt.Errorf("策略不存在")
	}
	return s.PolicyDelete(id)
}

// PolicyToggle 启用/禁用清理策略
func (s *ImageService) PolicyToggle(id int64, enabled bool) error {
	policy, err := s.PolicyGetByID(id)
	if err != nil {
		return fmt.Errorf("策略不存在")
	}
	policy.Enabled = enabled
	return s.PolicyUpdate(policy)
}

// ============================================================
// 清理日志 CRUD
// ============================================================

// LogCreate 创建清理日志
func (s *ImageService) LogCreate(log *ImageCleanupLog) error {
	return s.db.Create(log).Error
}

// LogUpdate 更新清理日志
func (s *ImageService) LogUpdate(log *ImageCleanupLog) error {
	return s.db.Save(log).Error
}

// LogListByPolicy 根据策略ID获取日志
func (s *ImageService) LogListByPolicy(policyID int64, limit int) ([]ImageCleanupLog, error) {
	var logs []ImageCleanupLog
	err := s.db.Where("policy_id = ?", policyID).Order("start_time DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// LogListRecent 获取最近的日志
func (s *ImageService) LogListRecent(limit int) ([]ImageCleanupLog, error) {
	var logs []ImageCleanupLog
	err := s.db.Order("start_time DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
