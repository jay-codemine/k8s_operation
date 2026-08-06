package services

import (
	"encoding/json"
	"errors"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 角色管理 ====================

// RoleCreate 创建角色
func (s *Services) RoleCreate(req *requests.RoleCreateRequest) (*models.SysRole, error) {
	return s.dao.RoleCreate(
		req.Name,
		req.DisplayName,
		req.Description,
		req.RoleType,
		req.ScopePlatform,
		req.ScopeCluster,
		req.ScopeCICD,
		req.Color,
		req.Icon,
	)
}

// RoleUpdate 更新角色
func (s *Services) RoleUpdate(req *requests.RoleUpdateRequest) error {
	// 检查是否为系统内置角色
	role, err := s.dao.RoleGetByID(req.ID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		// 系统内置角色只能修改显示名称、描述和 scope
		values := map[string]interface{}{
			"display_name":   req.DisplayName,
			"description":    req.Description,
			"scope_platform": req.ScopePlatform,
			"scope_cluster":  req.ScopeCluster,
			"scope_cicd":     req.ScopeCICD,
		}
		return s.dao.RoleUpdate(req.ID, values)
	}

	values := map[string]interface{}{
		"display_name":   req.DisplayName,
		"description":    req.Description,
		"scope_platform": req.ScopePlatform,
		"scope_cluster":  req.ScopeCluster,
		"scope_cicd":     req.ScopeCICD,
		"color":          req.Color,
		"icon":           req.Icon,
		"sort_order":     req.SortOrder,
	}
	return s.dao.RoleUpdate(req.ID, values)
}

// RoleDelete 删除角色
func (s *Services) RoleDelete(id int64) error {
	// 检查是否为系统内置角色
	role, err := s.dao.RoleGetByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return nil // 系统角色不可删除，静默返回
	}
	return s.dao.RoleDelete(id)
}

// RoleGetByID 获取角色详情
func (s *Services) RoleGetByID(id int64) (*models.RoleWithPermissions, error) {
	role, err := s.dao.RoleGetByID(id)
	if err != nil {
		return nil, err
	}
	permissions, _ := s.dao.PermissionGetByRoleID(id)
	return &models.RoleWithPermissions{
		SysRole:     *role,
		Permissions: permissions,
	}, nil
}

// RoleList 获取角色列表
func (s *Services) RoleList(req *requests.RoleListRequest) ([]*models.SysRole, int64, error) {
	return s.dao.RoleList(req.Name, req.RoleType, req.Page, req.Limit)
}

// RoleListAll 获取所有角色
func (s *Services) RoleListAll() ([]*models.SysRole, error) {
	return s.dao.RoleListAll()
}

// ==================== 权限管理 ====================

// PermissionList 获取权限列表
func (s *Services) PermissionList() ([]*models.SysPermission, error) {
	return s.dao.PermissionList()
}

// RolePermissionList 获取角色权限列表
func (s *Services) RolePermissionList(roleID int64) ([]*models.SysPermission, error) {
	return s.dao.PermissionGetByRoleID(roleID)
}

// RolePermissionUpdate 更新角色权限
func (s *Services) RolePermissionUpdate(roleID int64, permissionIDs []int64) error {
	return s.dao.RolePermissionAssign(roleID, permissionIDs)
}

// RoleUserList 获取角色绑定的用户列表
func (s *Services) RoleUserList(roleID int64) ([]*models.User, error) {
	return s.dao.RoleUserList(roleID)
}

// ==================== 用户角色管理 ====================

// UserRoleAssign 分配用户角色
func (s *Services) UserRoleAssign(req *requests.UserRoleAssignRequest, operatorID int64) error {
	// 安全防护：如果操作者在修改自己的角色，确保不会移除最后一个管理员
	if req.UserID == operatorID {
		// 检查新角色列表中是否包含至少一个有 platform:admin 能力的角色
		hasAdmin := false
		for _, rid := range req.RoleIDs {
			var role models.SysRole
			if err := s.dao.DB().Where("id = ? AND is_del = 0", rid).First(&role).Error; err == nil {
				if role.RoleType == models.RoleTypeSuperAdmin || role.RoleType == models.RoleTypePlatformAdmin {
					hasAdmin = true
					break
				}
			}
		}
		if !hasAdmin {
			return errors.New("不能移除自己的管理员权限，请让其他管理员修改你的角色")
		}
	}
	return s.dao.UserRoleAssign(req.UserID, req.RoleIDs, operatorID)
}

// UserRoleList 获取用户角色列表
func (s *Services) UserRoleList(userID int64) ([]*models.SysRole, error) {
	return s.dao.UserRoleList(userID)
}

// UserRoleRemove 移除用户角色
func (s *Services) UserRoleRemove(userID, roleID int64) error {
	return s.dao.UserRoleRemove(userID, roleID)
}

// ==================== 集群权限管理 ====================

// ClusterPermissionCreate 创建集群权限
func (s *Services) ClusterPermissionCreate(req *requests.ClusterPermissionCreateRequest, operatorID int64) (*models.SysUserCluster, error) {
	return s.dao.ClusterPermissionCreate(
		req.UserID,
		req.ClusterID,
		req.RoleType,
		req.AccessLevel,
		req.Namespaces,
		req.CanView,
		req.CanCreate,
		req.CanUpdate,
		req.CanDelete,
		req.CanExec,
		req.ExpireAt,
		operatorID,
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
		"role_type":    req.RoleType,
		"access_level": req.AccessLevel,
		"namespaces":   nsJSON,
		"can_view":     req.CanView,
		"can_create":   req.CanCreate,
		"can_update":   req.CanUpdate,
		"can_delete":   req.CanDelete,
		"can_exec":     req.CanExec,
		"expire_at":    req.ExpireAt,
	}
	return s.dao.ClusterPermissionUpdate(req.ID, values)
}

// ClusterPermissionDelete 删除集群权限
func (s *Services) ClusterPermissionDelete(id int64) error {
	return s.dao.ClusterPermissionDelete(id)
}

// ClusterPermissionList 获取集群权限列表
func (s *Services) ClusterPermissionList(req *requests.ClusterPermissionListRequest) ([]*models.ClusterPermissionDetail, int64, error) {
	return s.dao.ClusterPermissionList(req.UserID, req.ClusterID, req.Page, req.Limit)
}

// ClusterPermissionListByUser 获取用户的所有集群权限
func (s *Services) ClusterPermissionListByUser(userID int64) ([]*models.ClusterPermissionDetail, error) {
	return s.dao.ClusterPermissionListByUser(userID)
}

// BatchClusterPermissionCreate 批量分配集群权限
func (s *Services) BatchClusterPermissionCreate(req *requests.BatchClusterPermissionRequest, operatorID int64) error {
	return s.dao.BatchClusterPermissionCreate(
		req.UserID,
		req.ClusterIDs,
		req.RoleType,
		req.AccessLevel,
		req.CanView,
		req.CanCreate,
		req.CanUpdate,
		req.CanDelete,
		req.CanExec,
		operatorID,
	)
}

// ==================== 权限检查 ====================

// CheckClusterPermission 检查用户集群权限
func (s *Services) CheckClusterPermission(userID, clusterID int64, action string) bool {
	// 超级管理员拥有所有权限
	if s.dao.IsSuperAdmin(userID) {
		return true
	}
	// 检查用户角色的 scope_cluster 级别
	roles, _ := s.dao.UserRoleList(userID)
	for _, role := range roles {
		switch action {
		case models.PermissionActionView:
			if models.AccessLevelGte(role.ScopeCluster, models.AccessLevelRead) {
				return true
			}
		case models.PermissionActionCreate, models.PermissionActionUpdate, models.PermissionActionExec:
			if models.AccessLevelGte(role.ScopeCluster, models.AccessLevelWrite) {
				return true
			}
		case models.PermissionActionDelete, models.PermissionActionManage:
			if models.AccessLevelGte(role.ScopeCluster, models.AccessLevelAdmin) {
				return true
			}
		}
	}
	// 回退到细粒度集群权限表
	return s.dao.ClusterPermissionCheck(userID, clusterID, action)
}

// CheckScopePermission 检查用户是否满足指定域的最低权限级别
func (s *Services) CheckScopePermission(userID int64, scope, minLevel string) bool {
	if s.dao.IsSuperAdmin(userID) {
		return true
	}
	roles, err := s.dao.UserRoleList(userID)
	if err != nil || len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if models.AccessLevelGte(role.GetEffectiveScope(scope), minLevel) {
			return true
		}
	}
	return false
}

// IsSuperAdmin 检查用户是否为超级管理员
func (s *Services) IsSuperAdmin(userID int64) bool {
	return s.dao.IsSuperAdmin(userID)
}

// GetUserAccessibleClusters 获取用户可访问的集群
func (s *Services) GetUserAccessibleClusters(userID int64) ([]*models.K8sCluster, error) {
	return s.dao.GetUserAccessibleClusters(userID)
}

// ==================== 用户完整信息 ====================

// UserWithRBACInfo 用户完整信息（包含角色和集群权限）
type UserWithRBACInfo struct {
	UserID             int64                            `json:"user_id"`
	Username           string                           `json:"username"`
	IsSuperAdmin       bool                             `json:"is_super_admin"`
	Roles              []*models.SysRole                `json:"roles"`
	ClusterPermissions []*models.ClusterPermissionDetail `json:"cluster_permissions"`
	Scopes             *UserScopes                      `json:"scopes"`              // 三域有效权限级别
	PermissionNames    []string                         `json:"permission_names"`    // 用户拥有的所有权限名称
}

// UserScopes 用户三域有效权限级别（取所有角色的最高值）
type UserScopes struct {
	Platform string `json:"platform"` // none/read/write/admin
	Cluster  string `json:"cluster"`  // none/read/write/admin
	CICD     string `json:"cicd"`     // none/read/write/admin
}

// GetUserWithRBACInfo 获取用户完整RBAC信息
func (s *Services) GetUserWithRBACInfo(userID int64) (*UserWithRBACInfo, error) {
	// 获取用户角色
	roles, err := s.dao.UserRoleList(userID)
	if err != nil {
		return nil, err
	}

	// 获取集群权限
	clusterPerms, err := s.dao.ClusterPermissionListByUser(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否超级管理员
	isSuperAdmin := s.dao.IsSuperAdmin(userID)

	// 获取用户名
	username := ""
	if user, err := s.dao.UserGetByID(userID); err == nil && user != nil {
		username = user.Username
	}

	// 计算三域有效级别（取所有角色中的最高值）
	scopes := &UserScopes{
		Platform: models.AccessLevelNone,
		Cluster:  models.AccessLevelNone,
		CICD:     models.AccessLevelNone,
	}
	if isSuperAdmin {
		scopes.Platform = models.AccessLevelAdmin
		scopes.Cluster = models.AccessLevelAdmin
		scopes.CICD = models.AccessLevelAdmin
	} else {
		for _, role := range roles {
			if models.AccessLevelGte(role.ScopePlatform, scopes.Platform) {
				scopes.Platform = role.ScopePlatform
			}
			if models.AccessLevelGte(role.ScopeCluster, scopes.Cluster) {
				scopes.Cluster = role.ScopeCluster
			}
			if models.AccessLevelGte(role.ScopeCICD, scopes.CICD) {
				scopes.CICD = role.ScopeCICD
			}
		}
	}

	return &UserWithRBACInfo{
		UserID:             userID,
		Username:           username,
		IsSuperAdmin:       isSuperAdmin,
		Roles:              roles,
		ClusterPermissions: clusterPerms,
		Scopes:             scopes,
		PermissionNames:    s.getUserPermissionNames(userID, isSuperAdmin),
	}, nil
}

// GetUserAccessibleNamespaces 获取用户在指定集群可访问的命名空间
func (s *Services) GetUserAccessibleNamespaces(userID, clusterID int64) ([]string, error) {
	// 超级管理员可访问所有命名空间
	if s.dao.IsSuperAdmin(userID) {
		return []string{}, nil // 空数组表示所有
	}

	return s.dao.GetUserAccessibleNamespaces(userID, clusterID)
}

// getUserPermissionNames 获取用户拥有的所有权限名称
func (s *Services) getUserPermissionNames(userID int64, isSuperAdmin bool) []string {
	if isSuperAdmin {
		// 超管返回所有权限
		var names []string
		s.dao.DB().Model(&models.SysPermission{}).Pluck("name", &names)
		return names
	}
	names, _ := models.GetUserPermissionNames(s.dao.DB(), userID)
	return names
}
