package tenant

import (
	"context"

	"gorm.io/gorm"
)

// TenantService 租户领域服务
type TenantService struct {
	repo TenantRepository
}

// NewTenantService 创建租户服务
func NewTenantService(repo TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

// List 获取租户列表
func (s *TenantService) List(ctx context.Context, isSuperAdmin bool, tenantID uint32) ([]*Tenant, error) {
	return s.repo.FindAll(ctx, isSuperAdmin, tenantID)
}

// Create 创建租户（纯 CRUD，不含 RBAC 初始化）
func (s *TenantService) Create(ctx context.Context, name, code string) (*Tenant, error) {
	tenant := &Tenant{Name: name, Code: code, Status: 1}
	if err := s.repo.Save(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

// CreateWithTx 在事务中创建租户（跨域事务用，保留 *gorm.DB 参数）
func (s *TenantService) CreateWithTx(tx *gorm.DB, name, code string) (*Tenant, error) {
	tenant := &Tenant{Name: name, Code: code, Status: 1}
	if err := tx.Create(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

// Update 更新租户
func (s *TenantService) Update(ctx context.Context, id uint32, values map[string]interface{}) error {
	return s.repo.Update(ctx, id, values)
}

// SoftDelete 软删除租户
func (s *TenantService) SoftDelete(ctx context.Context, id uint32) error {
	return s.repo.SoftDelete(ctx, id)
}

// ExistsByID 检查 ID 是否存在
func (s *TenantService) ExistsByID(ctx context.Context, id uint32) (bool, error) {
	return s.repo.ExistsByID(ctx, id)
}

// ExistsByCode 检查 code 是否已存在（可选排除指定 ID）
func (s *TenantService) ExistsByCode(ctx context.Context, code string, excludeID uint32) (bool, error) {
	return s.repo.ExistsByCode(ctx, code, excludeID)
}
