package services

import (
	"context"
	"errors"

	"gorm.io/gorm"

	dtenant "k8soperation/internal/domain/tenant"
	"k8soperation/internal/domain/rbac"
	"k8soperation/internal/infra/persistence"
)

var (
	ErrTenantCodeExists         = errors.New("租户编码已存在")
	ErrTenantNotFound           = errors.New("租户不存在")
	ErrTenantDefaultCannotDelete = errors.New("不能删除默认租户")
)

func (s *Services) tenantSvc() *dtenant.TenantService {
	return dtenant.NewTenantService(persistence.NewTenantRepository(s.db))
}

// TenantList 获取租户列表
func (s *Services) TenantList(ctx context.Context, isSuperAdmin bool, tenantID uint32) ([]*dtenant.Tenant, error) {
	return s.tenantSvc().List(ctx, isSuperAdmin, tenantID)
}

// TenantCreate 创建租户（含 RBAC 种子，事务保护）
func (s *Services) TenantCreate(ctx context.Context, name, code string) (*dtenant.Tenant, error) {
	svc := s.tenantSvc()

	exists, err := svc.ExistsByCode(ctx, code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTenantCodeExists
	}

	var tenant *dtenant.Tenant
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		t, e := svc.CreateWithTx(tx, name, code)
		if e != nil {
			return e
		}
		tenant = t
		return rbac.SeedTenantRBAC(tx, tenant.ID)
	})
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// TenantUpdate 更新租户
func (s *Services) TenantUpdate(ctx context.Context, id uint32, name, code string, status *int8) error {
	svc := s.tenantSvc()

	exists, err := svc.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTenantNotFound
	}

	if code != "" {
		dup, err := svc.ExistsByCode(ctx, code, id)
		if err != nil {
			return err
		}
		if dup {
			return ErrTenantCodeExists
		}
	}

	values := make(map[string]interface{})
	if name != "" {
		values["name"] = name
	}
	if code != "" {
		values["code"] = code
	}
	if status != nil {
		values["status"] = *status
	}
	if len(values) == 0 {
		return nil
	}
	return svc.Update(ctx, id, values)
}

// TenantDelete 软删除租户
func (s *Services) TenantDelete(ctx context.Context, id uint32) error {
	if id == 1 {
		return ErrTenantDefaultCannotDelete
	}
	exists, err := s.tenantSvc().ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTenantNotFound
	}
	return s.tenantSvc().SoftDelete(ctx, id)
}
