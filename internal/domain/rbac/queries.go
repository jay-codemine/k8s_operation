package rbac

import (
	"gorm.io/gorm"

	"k8soperation/pkg/tenant"
)

// GetUserRoles 获取用户角色列表
func GetUserRoles(db *gorm.DB, userID int64) ([]*SysRole, error) {
	tid, ok := tenant.GetTenantID(db)
	if !ok {
		return nil, gorm.ErrMissingWhereClause
	}
	var roles []*SysRole
	err := db.Table("sys_role").
		Joins("JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ? AND sys_role.is_del = 0 AND sys_role.tenant_id = ? AND sys_user_role.tenant_id = ?", userID, tid, tid).
		Find(&roles).Error
	return roles, err
}

// GetUserClusterPermissions 获取用户集群权限
func GetUserClusterPermissions(db *gorm.DB, userID int64) ([]*SysUserCluster, error) {
	var permissions []*SysUserCluster
	err := db.Where("user_id = ?", userID).Find(&permissions).Error
	return permissions, err
}

// HasClusterPermission 检查用户是否有集群权限
func HasClusterPermission(db *gorm.DB, userID, clusterID int64, action string) bool {
	var record SysUserCluster
	err := db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).First(&record).Error
	if err != nil {
		return false
	}

	level := record.GetAccessLevel()
	switch action {
	case PermissionActionView:
		return AccessLevelGte(level, AccessLevelRead)
	case PermissionActionCreate, PermissionActionUpdate, PermissionActionExec:
		return AccessLevelGte(level, AccessLevelWrite)
	case PermissionActionDelete, PermissionActionManage:
		return AccessLevelGte(level, AccessLevelAdmin)
	default:
		return AccessLevelGte(level, AccessLevelRead)
	}
}

// HasScopePermission 检查用户是否满足指定域的最低权限级别
func HasScopePermission(db *gorm.DB, userID int64, scope, minLevel string) bool {
	if IsSuperAdmin(db, userID) {
		return true
	}

	roles, err := GetUserRoles(db, userID)
	if err != nil || len(roles) == 0 {
		return false
	}

	for _, role := range roles {
		if AccessLevelGte(role.GetEffectiveScope(scope), minLevel) {
			return true
		}
	}
	return false
}

// IsSuperAdmin 检查用户是否为超级管理员
func IsSuperAdmin(db *gorm.DB, userID int64) bool {
	tid, ok := tenant.GetTenantID(db)
	if !ok {
		return false
	}
	var count int64
	db.Table("sys_user_role").
		Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ? AND sys_role.role_type = ? AND sys_role.is_del = 0 AND sys_role.tenant_id = ? AND sys_user_role.tenant_id = ?", userID, RoleTypeSuperAdmin, tid, tid).
		Count(&count)
	return count > 0
}

// HasUserPermission 检查用户是否拥有指定的具体权限
func HasUserPermission(db *gorm.DB, userID int64, permissionName string) bool {
	if IsSuperAdmin(db, userID) {
		return true
	}
	tid, ok := tenant.GetTenantID(db)
	if !ok {
		return false
	}
	var count int64
	db.Table("sys_role_permission").
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role_permission.role_id").
		Joins("JOIN sys_permission ON sys_permission.id = sys_role_permission.permission_id").
		Where("sys_user_role.user_id = ? AND sys_permission.name = ? AND sys_user_role.tenant_id = ?", userID, permissionName, tid).
		Count(&count)
	return count > 0
}

// GetUserPermissionNames 获取用户拥有的所有权限名称列表
func GetUserPermissionNames(db *gorm.DB, userID int64) ([]string, error) {
	tid, ok := tenant.GetTenantID(db)
	if !ok {
		return nil, gorm.ErrMissingWhereClause
	}
	var names []string
	err := db.Table("sys_permission").
		Joins("JOIN sys_role_permission ON sys_role_permission.permission_id = sys_permission.id").
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role_permission.role_id").
		Where("sys_user_role.user_id = ? AND sys_user_role.tenant_id = ?", userID, tid).
		Pluck("sys_permission.name", &names).Error
	return names, err
}
