package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/rbac"
)

type rbacRepoImpl struct {
	db       *gorm.DB
	tenantID uint32
}

func NewRbacRepository(db *gorm.DB, tenantID uint32) rbac.RbacRepository {
	return &rbacRepoImpl{db: db, tenantID: tenantID}
}

// ——— 角色 ———

func (r *rbacRepoImpl) SaveRole(ctx context.Context, role *rbac.SysRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *rbacRepoImpl) UpdateRole(ctx context.Context, id int64, values map[string]interface{}) error {
	values["modified_at"] = uint64(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&rbac.SysRole{}).Where("id = ? AND is_del = 0", id).Updates(values).Error
}

func (r *rbacRepoImpl) DeleteRole(ctx context.Context, id int64) error {
	now := uint64(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&rbac.SysRole{}).Where("id = ? AND is_del = 0 AND is_system = 0", id).
		Updates(map[string]interface{}{"is_del": 1, "deleted_at": now, "modified_at": now}).Error
}

func (r *rbacRepoImpl) FindRoleByID(ctx context.Context, id int64) (*rbac.SysRole, error) {
	var role rbac.SysRole
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rbacRepoImpl) FindRoleByName(ctx context.Context, name string) (*rbac.SysRole, error) {
	var role rbac.SysRole
	if err := r.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rbacRepoImpl) QueryRoles(ctx context.Context, name, roleType string, page, limit int) ([]*rbac.SysRole, int64, error) {
	var roles []*rbac.SysRole
	var total int64
	query := r.db.WithContext(ctx).Model(&rbac.SysRole{}).Where("is_del = 0")
	if name != "" {
		query = query.Where("name LIKE ? OR display_name LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	if roleType != "" {
		query = query.Where("role_type = ?", roleType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *rbacRepoImpl) ListAllRoles(ctx context.Context) ([]*rbac.SysRole, error) {
	var roles []*rbac.SysRole
	if err := r.db.WithContext(ctx).Where("is_del = 0").Order("sort_order ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ——— 权限 ———

func (r *rbacRepoImpl) ListPermissions(ctx context.Context) ([]*rbac.SysPermission, error) {
	var permissions []*rbac.SysPermission
	if err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *rbacRepoImpl) FindPermissionsByRoleID(ctx context.Context, roleID int64) ([]*rbac.SysPermission, error) {
	var permissions []*rbac.SysPermission
	err := r.db.WithContext(ctx).Table("sys_permission").
		Joins("JOIN sys_role_permission ON sys_permission.id = sys_role_permission.permission_id").
		Where("sys_role_permission.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *rbacRepoImpl) AssignRolePermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&rbac.SysRolePermission{}).Error; err != nil {
		return err
	}
	now := uint64(time.Now().Unix())
	for _, pid := range permissionIDs {
		rp := &rbac.SysRolePermission{RoleID: roleID, PermissionID: pid, CreatedAt: now}
		if err := r.db.WithContext(ctx).Create(rp).Error; err != nil {
			return err
		}
	}
	return nil
}

// ——— 用户角色 ———

func (r *rbacRepoImpl) AssignUserRoles(ctx context.Context, userID int64, roleIDs []int64, createdBy int64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&rbac.SysUserRole{}).Error; err != nil {
		return err
	}
	now := uint64(time.Now().Unix())
	for _, rid := range roleIDs {
		ur := &rbac.SysUserRole{UserID: userID, RoleID: rid, CreatedAt: now, CreatedBy: createdBy}
		if err := r.db.WithContext(ctx).Create(ur).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *rbacRepoImpl) RemoveUserRole(ctx context.Context, userID, roleID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&rbac.SysUserRole{}).Error
}

func (r *rbacRepoImpl) FindUserRoles(ctx context.Context, userID int64) ([]*rbac.SysRole, error) {
	tid, ok := r.tenantID, r.tenantID != 0
	if !ok {
		return nil, gorm.ErrMissingWhereClause
	}
	var roles []*rbac.SysRole
	err := r.db.WithContext(ctx).Table("sys_role").
		Joins("JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ? AND sys_role.is_del = 0 AND sys_role.tenant_id = ? AND sys_user_role.tenant_id = ?", userID, tid, tid).
		Find(&roles).Error
	return roles, err
}

func (r *rbacRepoImpl) CountRoleUsers(ctx context.Context, roleID int64) int64 {
	var count int64
	r.db.WithContext(ctx).Model(&rbac.SysUserRole{}).Where("role_id = ?", roleID).Count(&count)
	return count
}

func (r *rbacRepoImpl) FindRoleUsers(ctx context.Context, roleID int64) ([]*rbac.SysRole, int64, error) {
	var roles []*rbac.SysRole
	var total int64
	query := r.db.WithContext(ctx).Model(&rbac.SysRole{}).Where("is_del = 0")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("sort_order ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

// ——— 集群权限 ———

func (r *rbacRepoImpl) SaveClusterPermission(ctx context.Context, p *rbac.SysUserCluster) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *rbacRepoImpl) UpdateClusterPermission(ctx context.Context, id int64, values map[string]interface{}) error {
	values["modified_at"] = uint64(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&rbac.SysUserCluster{}).Where("id = ?", id).Updates(values).Error
}

func (r *rbacRepoImpl) DeleteClusterPermission(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&rbac.SysUserCluster{}).Error
}

func (r *rbacRepoImpl) QueryClusterPermissions(ctx context.Context, userID, clusterID int64, page, limit int) ([]*rbac.SysUserCluster, int64, error) {
	var perms []*rbac.SysUserCluster
	var total int64
	query := r.db.WithContext(ctx).Model(&rbac.SysUserCluster{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&perms).Error; err != nil {
		return nil, 0, err
	}
	return perms, total, nil
}

func (r *rbacRepoImpl) FindClusterPermissionsByUser(ctx context.Context, userID int64) ([]*rbac.SysUserCluster, error) {
	var perms []*rbac.SysUserCluster
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// ——— 权限检查（委托到 repository.go 中的 standalone 函数）———

func (r *rbacRepoImpl) IsSuperAdmin(ctx context.Context, userID int64) bool {
	tid, ok := r.tenantID, r.tenantID != 0
	if !ok {
		return false
	}
	var count int64
	r.db.Table("sys_user_role").
		Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ? AND sys_role.role_type = ? AND sys_role.is_del = 0 AND sys_role.tenant_id = ? AND sys_user_role.tenant_id = ?", userID, rbac.RoleTypeSuperAdmin, tid, tid).
		Count(&count)
	return count > 0
}

func (r *rbacRepoImpl) HasUserPermission(ctx context.Context, userID int64, permissionName string) bool {
	if r.IsSuperAdmin(ctx, userID) {
		return true
	}
	tid := r.tenantID
	if tid == 0 {
		return false
	}
	var count int64
	r.db.Table("sys_role_permission").
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role_permission.role_id").
		Joins("JOIN sys_permission ON sys_permission.id = sys_role_permission.permission_id").
		Where("sys_user_role.user_id = ? AND sys_permission.name = ? AND sys_user_role.tenant_id = ? AND sys_role_permission.tenant_id = ?", userID, permissionName, tid, tid).
		Count(&count)
	return count > 0
}
