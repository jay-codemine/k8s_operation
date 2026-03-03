package models

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8soperation/global"
)

// ClusterStatus 集群状态定义
type ClusterStatus uint8

const (
	ClusterStatusOK      uint8 = 0 // 正常
	ClusterStatusBad     uint8 = 1 // 异常
	ClusterStatusPending uint8 = 2 // 未检测
)

type K8sCluster struct {
	ID             uint32 `gorm:"primaryKey;column:id" json:"id"`
	ClusterName    string `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterVersion string `gorm:"column:cluster_version" json:"cluster_version"`

	// DB 中存 base64，严禁直接对外返回
	KubeConfig string `gorm:"column:kube_config" json:"-"`

	// 状态：0=OK，1=BAD，2=UNKNOWN（新建/未检查）
	Status uint8 `gorm:"column:status" json:"status"`

	// 时间字段
	CreatedAt   uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt  uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt   uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	LastCheckAt uint64 `gorm:"column:last_check_at" json:"last_check_at"`

	// 最近一次异常原因
	LastError string `gorm:"column:last_error" json:"last_error"`

	// 软删除
	IsDel uint8 `gorm:"column:is_del" json:"-"`
}

func (k *K8sCluster) TableName() string { return "kube_cluster" }

// Create 新增
func (k *K8sCluster) Create(db *gorm.DB) error {
	tx := db.Create(k)
	if tx.Error != nil {
		global.Logger.Error("创建集群失败",
			zap.String("cluster_name", k.ClusterName),
			zap.Error(tx.Error),
		)
		return tx.Error
	}
	return nil
}

// GetByName 根据名称获取（未删除）
func (k *K8sCluster) GetByName(db *gorm.DB) (*K8sCluster, error) {
	var out K8sCluster
	tx := db.Where("cluster_name = ? AND is_del = 0", k.ClusterName).First(&out)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &out, nil
}

// GetByID 根据ID获取（未删除）
func (k *K8sCluster) GetByID(db *gorm.DB, id uint32) (*K8sCluster, error) {
	var out K8sCluster
	tx := db.Where("id = ? AND is_del = 0", id).First(&out)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &out, nil
}

// Update 通用更新（未删除）
func (k *K8sCluster) Update(db *gorm.DB, values map[string]interface{}) error {
	tx := db.Model(&K8sCluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(values)

	if tx.Error != nil {
		global.Logger.Error("更新集群失败",
			zap.Uint32("cluster_id", k.ID),
			zap.Error(tx.Error),
		)
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
	}
	return nil
}

// List 列表（支持按名称模糊查询）
func (k *K8sCluster) List(db *gorm.DB, page, limit int) ([]*K8sCluster, int64, error) {
	base := db.Model(&K8sCluster{})

	// total（不分页）
	var total int64
	if err := base.
		Scopes(
			ScopeNotDeleted(),
			ScopeLikeName("cluster_name", k.ClusterName),
		).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// list（分页 + 排序）
	var list []*K8sCluster
	err := base.
		Scopes(
			ScopeNotDeleted(),
			ScopeLikeName("cluster_name", k.ClusterName),
			ScopeOrderBy("last_check_at", true),
			Paginate(page, limit),
		).
		Find(&list).Error

	return list, total, err
}

// Delete 软删除
func (k *K8sCluster) Delete(db *gorm.DB) error {
	now := uint32(time.Now().Unix())
	tx := db.Model(&K8sCluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		})

	if tx.Error != nil {
		global.Logger.Error("软删除集群失败",
			zap.Uint32("cluster_id", k.ID),
			zap.Error(tx.Error),
		)
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		global.Logger.Warn("软删除集群未生效",
			zap.Uint32("cluster_id", k.ID),
		)
		return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
	}
	global.Logger.Info("软删除集群成功",
		zap.Uint32("cluster_id", k.ID),
	)
	return nil
}

// UpdateHealth 更新健康状态（你加 last_error/last_check_at 后用）
func (k *K8sCluster) UpdateHealth(db *gorm.DB, status uint8, lastErr string) error {
	now := uint32(time.Now().Unix())
	tx := db.Model(&K8sCluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(map[string]any{
			"status":        status,
			"last_error":    lastErr,
			"last_check_at": now,
			"modified_at":   now,
		})

	if tx.Error != nil {
		global.Logger.Error("更新集群健康状态失败",
			zap.Uint32("cluster_id", k.ID),
			zap.Error(tx.Error),
		)
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
	}
	return nil
}
