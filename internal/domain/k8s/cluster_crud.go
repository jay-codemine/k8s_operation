package k8s

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/pkg/db"
	"k8soperation/pkg/utils"
)

// ========== KubeConfig 加解密方法 ==========

// SetKubeConfig 加密并设置 KubeConfig
func (k *Cluster) SetKubeConfig(plaintext string) error {
	if plaintext == "" {
		k.KubeConfig = ""
		return nil
	}

	encrypted, err := utils.GlobalEncryptKubeConfig(plaintext)
	if err != nil {
		return err
	}
	k.KubeConfig = encrypted
	return nil
}

// GetKubeConfig 解密并获取 KubeConfig
func (k *Cluster) GetKubeConfig() (string, error) {
	if k.KubeConfig == "" {
		return "", nil
	}

	decrypted, err := utils.GlobalDecryptKubeConfig(k.KubeConfig)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

// IsKubeConfigEncrypted 检查 KubeConfig 是否已加密
func (k *Cluster) IsKubeConfigEncrypted() bool {
	return utils.IsEncrypted(k.KubeConfig)
}

// EncryptKubeConfigIfNeeded 如果未加密则加密（用于迁移旧数据）
func (k *Cluster) EncryptKubeConfigIfNeeded(tx *gorm.DB) error {
	if k.KubeConfig == "" || k.IsKubeConfigEncrypted() {
		return nil
	}

	plaintext := k.KubeConfig
	if err := k.SetKubeConfig(plaintext); err != nil {
		return err
	}

	now := uint64(time.Now().Unix())
	return tx.Model(&Cluster{}).Where("id = ?", k.ID).Updates(map[string]interface{}{
		"kube_config": k.KubeConfig,
		"modified_at": now,
	}).Error
}

// Create 新增
func (k *Cluster) Create(tx *gorm.DB) error {
	return tx.Create(k).Error
}

// GetByName 根据名称获取（未删除）
func (k *Cluster) GetByName(tx *gorm.DB) (*Cluster, error) {
	var out Cluster
	if err := tx.Where("cluster_name = ? AND is_del = 0", k.ClusterName).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// GetByID 根据ID获取（未删除）
func (Cluster) GetByID(tx *gorm.DB, id uint32) (*Cluster, error) {
	var out Cluster
	if err := tx.Where("id = ? AND is_del = 0", id).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// Update 通用更新（未删除）
func (k *Cluster) Update(tx *gorm.DB, values map[string]interface{}) error {
	var exist Cluster
	if err := tx.Where("id = ? AND is_del = 0", k.ID).First(&exist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
		}
		return err
	}

	return tx.Model(&Cluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(values).Error
}

// List 列表（支持按名称模糊查询）
func (k *Cluster) List(tx *gorm.DB, page, limit int) ([]*Cluster, int64, error) {
	base := tx.Model(&Cluster{})

	var total int64
	if err := base.
		Scopes(
			db.ScopeNotDeleted(),
			db.ScopeLikeName("cluster_name", k.ClusterName),
		).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*Cluster
	err := base.
		Scopes(
			db.ScopeNotDeleted(),
			db.ScopeLikeName("cluster_name", k.ClusterName),
			db.ScopeOrderBy("last_check_at", true),
			db.Paginate(page, limit),
		).
		Find(&list).Error

	return list, total, err
}

// Delete 软删除
func (k *Cluster) Delete(tx *gorm.DB) error {
	now := uint32(time.Now().Unix())
	if err := tx.Model(&Cluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
	}
	return nil
}

// UpdateHealth 更新健康状态
func (k *Cluster) UpdateHealth(tx *gorm.DB, status uint8, lastErr string) error {
	now := uint32(time.Now().Unix())
	if err := tx.Model(&Cluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(map[string]any{
			"status":        status,
			"last_error":    lastErr,
			"last_check_at": now,
			"modified_at":   now,
		}).Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
	}
	return nil
}
