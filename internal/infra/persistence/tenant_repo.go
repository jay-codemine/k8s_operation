package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/tenant"
)

type tenantRepo struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) tenant.TenantRepository {
	return &tenantRepo{db: db}
}

func (r *tenantRepo) FindAll(ctx context.Context, isSuperAdmin bool, tenantID uint32) ([]*tenant.Tenant, error) {
	var tenants []*tenant.Tenant
	query := r.db.WithContext(ctx).Where("is_del = 0")
	if !isSuperAdmin {
		query = query.Where("id = ?", tenantID)
	}
	if err := query.Find(&tenants).Error; err != nil {
		return nil, err
	}
	if tenants == nil {
		tenants = []*tenant.Tenant{}
	}
	return tenants, nil
}

func (r *tenantRepo) Save(ctx context.Context, t *tenant.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tenantRepo) Update(ctx context.Context, id uint32, values map[string]interface{}) error {
	values["modified_at"] = uint32(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&tenant.Tenant{}).Where("id = ? AND is_del = 0", id).Updates(values).Error
}

func (r *tenantRepo) SoftDelete(ctx context.Context, id uint32) error {
	now := uint32(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&tenant.Tenant{}).Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{"is_del": 1, "deleted_at": now, "modified_at": now}).Error
}

func (r *tenantRepo) ExistsByID(ctx context.Context, id uint32) (bool, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&tenant.Tenant{}).Where("id = ? AND is_del = 0", id).Count(&cnt).Error
	return cnt > 0, err
}

func (r *tenantRepo) ExistsByCode(ctx context.Context, code string, excludeID uint32) (bool, error) {
	var cnt int64
	query := r.db.WithContext(ctx).Model(&tenant.Tenant{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&cnt).Error
	return cnt > 0, err
}
