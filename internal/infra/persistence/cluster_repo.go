package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/k8s"
	"k8soperation/pkg/db"
	"k8soperation/pkg/utils"
)

type clusterRepo struct {
	db *gorm.DB
}

func NewClusterRepository(database *gorm.DB) k8s.ClusterRepository {
	return &clusterRepo{db: database}
}

func (r *clusterRepo) Save(ctx context.Context, c *k8s.Cluster) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *clusterRepo) Update(ctx context.Context, id uint32, values map[string]interface{}) error {
	var exist k8s.Cluster
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&exist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("集群不存在或已删除: id=%d", id)
		}
		return err
	}
	return r.db.WithContext(ctx).Model(&k8s.Cluster{}).
		Where("id = ? AND is_del = 0", id).
		Updates(values).Error
}

func (r *clusterRepo) SoftDelete(ctx context.Context, id uint32) error {
	now := uint32(time.Now().Unix())
	if err := r.db.WithContext(ctx).Model(&k8s.Cluster{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del": 1, "deleted_at": now, "modified_at": now,
		}).Error; err != nil {
		return err
	}
	if r.db.WithContext(ctx).RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", id)
	}
	return nil
}

func (r *clusterRepo) BatchDelete(ctx context.Context, ids []uint32) (int64, error) {
	now := utils.NowUnix()
	tx := r.db.WithContext(ctx).Model(&k8s.Cluster{}).
		Where("id IN ? AND is_del = 0", ids).
		Updates(map[string]interface{}{
			"is_del": 1, "deleted_at": now, "modified_at": now,
		})
	return tx.RowsAffected, tx.Error
}

func (r *clusterRepo) FindByID(ctx context.Context, id uint32) (*k8s.Cluster, error) {
	var out k8s.Cluster
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *clusterRepo) FindByName(ctx context.Context, name string) (*k8s.Cluster, error) {
	var out k8s.Cluster
	if err := r.db.WithContext(ctx).Where("cluster_name = ? AND is_del = 0", name).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *clusterRepo) Query(ctx context.Context, name string, page, limit int) ([]*k8s.Cluster, int64, error) {
	base := r.db.WithContext(ctx).Model(&k8s.Cluster{})

	var total int64
	if err := base.Scopes(
		db.ScopeNotDeleted(),
		db.ScopeLikeName("cluster_name", name),
	).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*k8s.Cluster
	err := base.Scopes(
		db.ScopeNotDeleted(),
		db.ScopeLikeName("cluster_name", name),
		db.ScopeOrderBy("last_check_at", true),
		db.Paginate(page, limit),
	).Find(&list).Error

	return list, total, err
}

func (r *clusterRepo) UpdateHealth(ctx context.Context, id uint32, status uint8, lastErr string) error {
	now := uint32(time.Now().Unix())
	if err := r.db.WithContext(ctx).Model(&k8s.Cluster{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]any{
			"status": status, "last_error": lastErr,
			"last_check_at": now, "modified_at": now,
		}).Error; err != nil {
		return err
	}
	if r.db.WithContext(ctx).RowsAffected == 0 {
		return fmt.Errorf("集群不存在或已删除: id=%d", id)
	}
	return nil
}
