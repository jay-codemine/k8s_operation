package rbac

import "context"

// RbacRepository RBAC 仓储接口
type RbacRepository interface {
	// 角色
	SaveRole(ctx context.Context, role *SysRole) error
	UpdateRole(ctx context.Context, id int64, values map[string]interface{}) error
	DeleteRole(ctx context.Context, id int64) error
	FindRoleByID(ctx context.Context, id int64) (*SysRole, error)
	FindRoleByName(ctx context.Context, name string) (*SysRole, error)
	QueryRoles(ctx context.Context, name, roleType string, page, limit int) ([]*SysRole, int64, error)
	ListAllRoles(ctx context.Context) ([]*SysRole, error)

	// 权限
	ListPermissions(ctx context.Context) ([]*SysPermission, error)
	FindPermissionsByRoleID(ctx context.Context, roleID int64) ([]*SysPermission, error)
	AssignRolePermissions(ctx context.Context, roleID int64, permissionIDs []int64) error

	// 用户角色
	AssignUserRoles(ctx context.Context, userID int64, roleIDs []int64, createdBy int64) error
	RemoveUserRole(ctx context.Context, userID, roleID int64) error
	FindUserRoles(ctx context.Context, userID int64) ([]*SysRole, error)
	FindRoleUsers(ctx context.Context, roleID int64) ([]*SysRole, int64, error) // users + count
	CountRoleUsers(ctx context.Context, roleID int64) int64

	// 集群权限
	SaveClusterPermission(ctx context.Context, p *SysUserCluster) error
	UpdateClusterPermission(ctx context.Context, id int64, values map[string]interface{}) error
	DeleteClusterPermission(ctx context.Context, id int64) error
	QueryClusterPermissions(ctx context.Context, userID, clusterID int64, page, limit int) ([]*SysUserCluster, int64, error)
	FindClusterPermissionsByUser(ctx context.Context, userID int64) ([]*SysUserCluster, error)

	// 权限检查（沿用 repository.go 中的实现）
	IsSuperAdmin(ctx context.Context, userID int64) bool
	HasUserPermission(ctx context.Context, userID int64, permissionName string) bool
}
