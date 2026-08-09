package services

import (
	"encoding/json"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/domain/rbac"
	"k8soperation/internal/infra/adapter"
	"k8soperation/internal/infra/persistence"
)

func (s *Services) rbacSvc() *rbac.RbacService {
	svc := rbac.NewRbacService(persistence.NewRbacRepository(global.DB, s.tenantID), global.DB)
	svc.WithClusterLister(adapter.NewClusterLister(s.clusterSvc()))
	svc.WithUserLookup(adapter.NewUserLookup(s.userSvc()))
	svc.SetTenantID(s.tenantID)
	return svc
}

// ==================== 角色管理 ====================

// RoleCreate 创建角色
func (s *Services) RoleCreate(req *requests.RoleCreateRequest) (*models.SysRole, error) {
	return s.rbacSvc().RoleCreate(
		req.Name, req.DisplayName, req.Description, req.RoleType,
		req.ScopePlatform, req.ScopeCluster, req.ScopeCICD, req.Color, req.Icon,
	)
}

// RoleUpdate 更新角色
func (s *Services) RoleUpdate(req *requests.RoleUpdateRequest) error {
	role, err := s.rbacSvc().RoleGetByID(req.ID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		values := map[string]interface{}{
			"display_name": req.DisplayName, "description": req.Description,
			"scope_platform": req.ScopePlatform, "scope_cluster": req.ScopeCluster, "scope_cicd": req.ScopeCICD,
		}
		return s.rbacSvc().RoleUpdate(req.ID, values)
	}

	values := map[string]interface{}{
		"display_name": req.DisplayName, "description": req.Description,
		"scope_platform": req.ScopePlatform, "scope_cluster": req.ScopeCluster, "scope_cicd": req.ScopeCICD,
		"color": req.Color, "icon": req.Icon, "sort_order": req.SortOrder,
	}
	return s.rbacSvc().RoleUpdate(req.ID, values)
}

// RoleDelete 删除角色
func (s *Services) RoleDelete(id int64) error {
	role, err := s.rbacSvc().RoleGetByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return nil
	}
	return s.rbacSvc().RoleDelete(id)
}

// RoleGetByID 获取角色详情
func (s *Services) RoleGetByID(id int64) (*models.RoleWithPermissions, error) {
	return s.rbacSvc().RoleGetWithPermissions(id)
}

// RoleGetByName 根据名称获取角色（用于初始化等场景）
func (s *Services) RoleGetByName(name string) (*models.SysRole, error) {
	return s.rbacSvc().RoleGetByName(name)
}

// RoleList 获取角色列表
func (s *Services) RoleList(req *requests.RoleListRequest) ([]*models.SysRole, int64, error) {
	return s.rbacSvc().RoleList(req.Name, req.RoleType, req.Page, req.Limit)
}

// RoleListAll 获取所有角色
func (s *Services) RoleListAll() ([]*models.SysRole, error) {
	return s.rbacSvc().RoleListAll()
}

// ==================== 权限管理 ====================

// PermissionList 获取权限列表
func (s *Services) PermissionList() ([]*models.SysPermission, error) {
	return s.rbacSvc().PermissionList()
}

// RolePermissionList 获取角色权限列表
func (s *Services) RolePermissionList(roleID int64) ([]*models.SysPermission, error) {
	return s.rbacSvc().PermissionGetByRoleID(roleID)
}

// RolePermissionUpdate 更新角色权限
func (s *Services) RolePermissionUpdate(roleID int64, permissionIDs []int64) error {
	return s.rbacSvc().RolePermissionAssign(roleID, permissionIDs)
}

// RoleUserList 获取角色绑定的用户列表
func (s *Services) RoleUserList(roleID int64) ([]*models.User, error) {
	return s.rbacSvc().RoleUserList(roleID)
}

// ==================== 用户角色管理 ====================

// UserRoleAssign 分配用户角色
func (s *Services) UserRoleAssign(req *requests.UserRoleAssignRequest, operatorID int64) error {
	return s.rbacSvc().UserRoleAssignSafe(req.UserID, req.RoleIDs, operatorID)
}

// UserRoleAssignSimple 直接参数分配用户角色（用于初始化等场景）
func (s *Services) UserRoleAssignSimple(userID, roleID int64) error {
	return s.rbacSvc().UserRoleAssign(userID, []int64{roleID}, 0)
}

// UserRoleList 获取用户角色列表
func (s *Services) UserRoleList(userID int64) ([]*models.SysRole, error) {
	return s.rbacSvc().UserRoleList(userID)
}

// UserRoleRemove 移除用户角色
func (s *Services) UserRoleRemove(userID, roleID int64) error {
	return s.rbacSvc().UserRoleRemove(userID, roleID)
}

// ==================== 集群权限管理 ====================

// ClusterPermissionCreate 创建集群权限
func (s *Services) ClusterPermissionCreate(req *requests.ClusterPermissionCreateRequest, operatorID int64) (*models.SysUserCluster, error) {
	return s.rbacSvc().ClusterPermissionCreate(
		req.UserID, req.ClusterID, req.RoleType, req.AccessLevel, req.Namespaces,
		req.CanView, req.CanCreate, req.CanUpdate, req.CanDelete, req.CanExec,
		req.ExpireAt, operatorID,
	)
}

// ClusterPermissionUpdate 更新集群权限
func (s *Services) ClusterPermissionUpdate(req *requests.ClusterPermissionUpdateRequest) error {
	nsJSON := ""
	if len(req.Namespaces) > 0 {
		if data, err := json.Marshal(req.Namespaces); err == nil {
			nsJSON = string(data)
		}
	}

	values := map[string]interface{}{
		"role_type": req.RoleType, "access_level": req.AccessLevel, "namespaces": nsJSON,
		"can_view": req.CanView, "can_create": req.CanCreate, "can_update": req.CanUpdate,
		"can_delete": req.CanDelete, "can_exec": req.CanExec, "expire_at": req.ExpireAt,
	}
	return s.rbacSvc().ClusterPermissionUpdate(req.ID, values)
}

// ClusterPermissionDelete 删除集群权限
func (s *Services) ClusterPermissionDelete(id int64) error {
	return s.rbacSvc().ClusterPermissionDelete(id)
}

// ClusterPermissionList 获取集群权限列表
func (s *Services) ClusterPermissionList(req *requests.ClusterPermissionListRequest) ([]*models.ClusterPermissionDetail, int64, error) {
	return s.rbacSvc().ClusterPermissionList(req.UserID, req.ClusterID, req.Page, req.Limit)
}

// ClusterPermissionListByUser 获取用户的所有集群权限
func (s *Services) ClusterPermissionListByUser(userID int64) ([]*models.ClusterPermissionDetail, error) {
	return s.rbacSvc().ClusterPermissionListByUser(userID)
}

// BatchClusterPermissionCreate 批量分配集群权限
func (s *Services) BatchClusterPermissionCreate(req *requests.BatchClusterPermissionRequest, operatorID int64) error {
	return s.rbacSvc().BatchClusterPermissionCreate(
		req.UserID, req.ClusterIDs, req.RoleType, req.AccessLevel,
		req.CanView, req.CanCreate, req.CanUpdate, req.CanDelete, req.CanExec,
		operatorID,
	)
}

// ==================== 权限检查 ====================

// CheckClusterPermission 检查用户集群权限
func (s *Services) CheckClusterPermission(userID, clusterID int64, action string) bool {
	return s.rbacSvc().CheckClusterPermission(userID, clusterID, action)
}

// CheckScopePermission 检查用户是否满足指定域的最低权限级别
func (s *Services) CheckScopePermission(userID int64, scope, minLevel string) bool {
	return s.rbacSvc().CheckScopePermission(userID, scope, minLevel)
}

// IsSuperAdmin 检查用户是否为超级管理员
func (s *Services) IsSuperAdmin(userID int64) bool {
	return s.rbacSvc().IsSuperAdmin(userID)
}

// HasUserPermission 检查用户是否拥有指定权限
func (s *Services) HasUserPermission(userID int64, permissionName string) bool {
	return s.rbacSvc().HasUserPermission(userID, permissionName)
}

// GetUserAccessibleClusters 获取用户可访问的集群
func (s *Services) GetUserAccessibleClusters(userID int64) ([]*models.K8sCluster, error) {
	return s.rbacSvc().GetUserAccessibleClusters(userID)
}

// ==================== 用户完整信息 ====================

// GetUserWithRBACInfo 获取用户完整RBAC信息
func (s *Services) GetUserWithRBACInfo(userID int64) (*rbac.UserWithRBACInfo, error) {
	return s.rbacSvc().GetUserWithRBACInfo(userID)
}

// GetUserAccessibleNamespaces 获取用户在指定集群可访问的命名空间
func (s *Services) GetUserAccessibleNamespaces(userID, clusterID int64) ([]string, error) {
	return s.rbacSvc().GetUserAccessibleNamespaces(userID, clusterID)
}

// ==================== 租户 RBAC 初始化 ====================

// TenantListMissingSuperAdmin 列出缺少 super_admin 角色的租户
func (s *Services) TenantListMissingSuperAdmin() ([]uint32, error) {
	return rbac.ListTenantsMissingSuperAdmin(s.db)
}

// TenantSeedRBAC 为指定租户克隆默认角色及权限
func (s *Services) TenantSeedRBAC(tenantID uint32) error {
	return rbac.SeedTenantRBAC(s.db, tenantID)
}
