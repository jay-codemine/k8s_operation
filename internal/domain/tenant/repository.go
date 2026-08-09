package tenant

import "context"

// TenantRepository 租户仓储接口
type TenantRepository interface {
	FindAll(ctx context.Context, isSuperAdmin bool, tenantID uint32) ([]*Tenant, error)
	Save(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, id uint32, values map[string]interface{}) error
	SoftDelete(ctx context.Context, id uint32) error
	ExistsByID(ctx context.Context, id uint32) (bool, error)
	ExistsByCode(ctx context.Context, code string, excludeID uint32) (bool, error)
}
