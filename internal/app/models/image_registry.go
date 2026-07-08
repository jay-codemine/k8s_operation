package models

import (
	"k8soperation/global"

	"gorm.io/gorm"
)

// ImageRegistry 镜像仓库配置
type ImageRegistry struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"type:varchar(100);not null;uniqueIndex:idx_registry_name" json:"name"`    // 仓库名称
	Type            string `gorm:"type:varchar(50);not null;default:'docker'" json:"type"`                  // 类型: docker, harbor, gcr, ecr, acr, quay
	URL             string `gorm:"type:varchar(500);not null" json:"url"`                                   // 仓库地址
	Username        string `gorm:"type:varchar(100)" json:"username"`                                       // 用户名
	Password        string `gorm:"type:varchar(500)" json:"-"`                                              // 密码（加密存储，不返回给前端）
	AccessKeyID     string `gorm:"type:varchar(100)" json:"access_key_id"`                                  // 阿里云 AccessKey ID（ACR 使用）
	AccessKeySecret string `gorm:"type:varchar(200)" json:"-"`                                              // 阿里云 AccessKey Secret（加密存储）
	Region          string `gorm:"type:varchar(50)" json:"region"`                                          // 区域（如 cn-hangzhou）
	Insecure        bool   `gorm:"type:tinyint(1);default:0" json:"insecure"`                               // 是否跳过 TLS 验证
	Description     string `gorm:"type:varchar(500)" json:"description"`                                    // 描述
	IsDefault       bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`                             // 是否默认仓库
	Status          string `gorm:"type:varchar(50);default:'unknown'" json:"status"`                        // 连接状态: connected, disconnected, unknown
	LastCheckAt     int64  `gorm:"type:bigint" json:"last_check_at"`                                        // 最后检测时间
	LastError       string `gorm:"type:varchar(500)" json:"last_error"`                                     // 最后错误信息
	CreatedBy       int64  `gorm:"type:bigint" json:"created_by"`                                           // 创建人
	CreatedAt       int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt      int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel           int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (ImageRegistry) TableName() string {
	return "image_registry"
}

// ImageRegistryModel 镜像仓库模型操作
type ImageRegistryModel struct {
	db *gorm.DB
}

func NewImageRegistryModel() *ImageRegistryModel {
	return &ImageRegistryModel{db: global.DB}
}

// Create 创建镜像仓库
func (m *ImageRegistryModel) Create(registry *ImageRegistry) error {
	return m.db.Create(registry).Error
}

// Update 更新镜像仓库
func (m *ImageRegistryModel) Update(registry *ImageRegistry) error {
	return m.db.Model(registry).Updates(map[string]interface{}{
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

// UpdateStatus 更新连接状态
func (m *ImageRegistryModel) UpdateStatus(id int64, status, lastError string, checkTime int64) error {
	return m.db.Model(&ImageRegistry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"last_error":    lastError,
		"last_check_at": checkTime,
	}).Error
}

// Delete 软删除镜像仓库
func (m *ImageRegistryModel) Delete(id int64) error {
	return m.db.Model(&ImageRegistry{}).Where("id = ?", id).Update("is_del", 1).Error
}

// GetByID 根据ID获取镜像仓库
func (m *ImageRegistryModel) GetByID(id int64) (*ImageRegistry, error) {
	var registry ImageRegistry
	err := m.db.Where("id = ? AND is_del = 0", id).First(&registry).Error
	return &registry, err
}

// GetByName 根据名称获取镜像仓库
func (m *ImageRegistryModel) GetByName(name string) (*ImageRegistry, error) {
	var registry ImageRegistry
	err := m.db.Where("name = ? AND is_del = 0", name).First(&registry).Error
	return &registry, err
}

// List 获取镜像仓库列表
func (m *ImageRegistryModel) List(keyword string, registryType string, page, pageSize int) ([]ImageRegistry, int64, error) {
	var registries []ImageRegistry
	var total int64

	query := m.db.Model(&ImageRegistry{}).Where("is_del = 0")

	if keyword != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if registryType != "" {
		query = query.Where("type = ?", registryType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&registries).Error; err != nil {
		return nil, 0, err
	}

	return registries, total, nil
}

// ListAll 获取所有镜像仓库（不分页）
func (m *ImageRegistryModel) ListAll() ([]ImageRegistry, error) {
	var registries []ImageRegistry
	err := m.db.Where("is_del = 0").Order("is_default DESC, id ASC").Find(&registries).Error
	return registries, err
}

// GetDefault 获取默认仓库
func (m *ImageRegistryModel) GetDefault() (*ImageRegistry, error) {
	var registry ImageRegistry
	err := m.db.Where("is_default = 1 AND is_del = 0").First(&registry).Error
	return &registry, err
}

// SetDefault 设置默认仓库
func (m *ImageRegistryModel) SetDefault(id int64) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// 先清除所有默认标记
		if err := tx.Model(&ImageRegistry{}).Where("is_del = 0").Update("is_default", 0).Error; err != nil {
			return err
		}
		// 设置新的默认仓库
		return tx.Model(&ImageRegistry{}).Where("id = ?", id).Update("is_default", 1).Error
	})
}

// ExistsByName 检查名称是否存在
func (m *ImageRegistryModel) ExistsByName(name string, excludeID int64) (bool, error) {
	var count int64
	query := m.db.Model(&ImageRegistry{}).Where("name = ? AND is_del = 0", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
