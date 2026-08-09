package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/k8s"
	dtenant "k8soperation/internal/domain/tenant"
	"k8soperation/internal/domain/user"
	"k8soperation/pkg/tenant"
)

// RbacService RBAC 领域服务
type RbacService struct {
	repo          RbacRepository
	db            *gorm.DB      // 仅用于跨域查询（kube_cluster、user 表 JOIN），待各域提供接口后移除
	tenantID      uint32        // 缓存的租户 ID
	clusterLister ClusterLister // 可选：跨域集群查询接口
	userLookup    UserLookup    // 可选：跨域用户查询接口
}

// SetTenantID 注入租户 ID（Services 层调用）
func (s *RbacService) SetTenantID(tid uint32) { s.tenantID = tid }

// scopedDB 返回临时租户隔离 DB（用于独立权限函数，它们依赖 Statement.Context 取 tenant_id）
func (s *RbacService) scopedDB() *gorm.DB {
	if s.tenantID == 0 {
		return s.db
	}
	return tenant.NewScopedDB(s.db, s.tenantID)
}

// NewRbacService 创建 RBAC 服务
func NewRbacService(repo RbacRepository, db *gorm.DB) *RbacService {
	return &RbacService{repo: repo, db: db}
}

// WithClusterLister 注入跨域集群查询器（逐步替代 s.db JOIN）
func (s *RbacService) WithClusterLister(l ClusterLister) *RbacService {
	s.clusterLister = l
	return s
}

// WithUserLookup 注入跨域用户查询器
func (s *RbacService) WithUserLookup(l UserLookup) *RbacService {
	s.userLookup = l
	return s
}

// ==================== 角色管理 ====================

func (s *RbacService) RoleCreate(name, displayName, description, roleType, scopePlatform, scopeCluster, scopeCICD, color, icon string) (*SysRole, error) {
	now := uint64(time.Now().Unix())
	role := &SysRole{
		Name: name, DisplayName: displayName, Description: description,
		RoleType: roleType, ScopePlatform: scopePlatform, ScopeCluster: scopeCluster, ScopeCICD: scopeCICD,
		Color: color, Icon: icon, IsSystem: false, SortOrder: 0,
		CreatedAt: now, ModifiedAt: now,
	}
	if err := s.repo.SaveRole(context.Background(), role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RbacService) RoleUpdate(id int64, values map[string]interface{}) error {
	return s.repo.UpdateRole(context.Background(), id, values)
}

func (s *RbacService) RoleDelete(id int64) error {
	return s.repo.DeleteRole(context.Background(), id)
}

func (s *RbacService) RoleGetByID(id int64) (*SysRole, error) {
	return s.repo.FindRoleByID(context.Background(), id)
}

func (s *RbacService) RoleGetByName(name string) (*SysRole, error) {
	return s.repo.FindRoleByName(context.Background(), name)
}

func (s *RbacService) RoleList(name, roleType string, page, limit int) ([]*SysRole, int64, error) {
	return s.repo.QueryRoles(context.Background(), name, roleType, page, limit)
}

func (s *RbacService) RoleListAll() ([]*SysRole, error) {
	return s.repo.ListAllRoles(context.Background())
}

func (s *RbacService) RoleGetWithPermissions(id int64) (*RoleWithPermissions, error) {
	role, err := s.RoleGetByID(id)
	if err != nil {
		return nil, err
	}
	permissions, _ := s.PermissionGetByRoleID(id)
	return &RoleWithPermissions{SysRole: *role, Permissions: permissions}, nil
}

func (s *RbacService) RoleListWithCount(page, limit int) ([]*RoleWithPermissions, int64, error) {
	roles, total, err := s.repo.QueryRoles(context.Background(), "", "", page, limit)
	if err != nil {
		return nil, 0, err
	}
	var result []*RoleWithPermissions
	for _, role := range roles {
		count := s.repo.CountRoleUsers(context.Background(), role.ID)
		result = append(result, &RoleWithPermissions{SysRole: *role, UserCount: count})
	}
	return result, total, nil
}

// ==================== 权限管理 ====================

func (s *RbacService) PermissionList() ([]*SysPermission, error) {
	return s.repo.ListPermissions(context.Background())
}

func (s *RbacService) PermissionGetByRoleID(roleID int64) ([]*SysPermission, error) {
	return s.repo.FindPermissionsByRoleID(context.Background(), roleID)
}

func (s *RbacService) RolePermissionAssign(roleID int64, permissionIDs []int64) error {
	return s.repo.AssignRolePermissions(context.Background(), roleID, permissionIDs)
}

// ==================== 用户角色管理 ====================

func (s *RbacService) UserRoleAssign(userID int64, roleIDs []int64, createdBy int64) error {
	return s.repo.AssignUserRoles(context.Background(), userID, roleIDs, createdBy)
}

func (s *RbacService) UserRoleRemove(userID, roleID int64) error {
	return s.repo.RemoveUserRole(context.Background(), userID, roleID)
}

func (s *RbacService) UserRoleList(userID int64) ([]*SysRole, error) {
	return s.repo.FindUserRoles(context.Background(), userID)
}

func (s *RbacService) RoleUserList(roleID int64) ([]*user.User, error) {
	tid, ok := tenant.GetTenantID(s.db)
	if !ok {
		return nil, nil
	}
	var users []*user.User
	err := s.db.Table("user").
		Joins("JOIN sys_user_role ON user.id = sys_user_role.user_id").
		Where("sys_user_role.role_id = ? AND user.is_del = 0 AND user.tenant_id = ? AND sys_user_role.tenant_id = ?", roleID, tid, tid).
		Find(&users).Error
	return users, err
}

// ==================== 集群权限管理 ====================

func (s *RbacService) ClusterPermissionCreate(userID, clusterID int64, roleType, accessLevel string, namespaces []string,
	canView, canCreate, canUpdate, canDelete, canExec bool, expireAt uint64, createdBy int64) (*SysUserCluster, error) {
	now := uint64(time.Now().Unix())
	nsJSON := ""
	if len(namespaces) > 0 {
		if data, err := json.Marshal(namespaces); err == nil {
			nsJSON = string(data)
		}
	}

	// upsert: check existing
	perms, _ := s.repo.FindClusterPermissionsByUser(context.Background(), userID)
	for _, existing := range perms {
		if existing.ClusterID == clusterID {
			updates := map[string]interface{}{
				"role_type": roleType, "access_level": accessLevel, "namespaces": nsJSON,
				"can_view": canView, "can_create": canCreate, "can_update": canUpdate,
				"can_delete": canDelete, "can_exec": canExec, "expire_at": expireAt, "modified_at": now,
			}
			if err := s.repo.UpdateClusterPermission(context.Background(), existing.ID, updates); err != nil {
				return nil, err
			}
			return existing, nil
		}
	}

	perm := &SysUserCluster{
		UserID: userID, ClusterID: clusterID, RoleType: roleType, AccessLevel: accessLevel,
		Namespaces: nsJSON, CanView: canView, CanCreate: canCreate, CanUpdate: canUpdate,
		CanDelete: canDelete, CanExec: canExec, ExpireAt: expireAt,
		CreatedAt: now, ModifiedAt: now, CreatedBy: createdBy,
	}
	if err := s.repo.SaveClusterPermission(context.Background(), perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *RbacService) ClusterPermissionUpdate(id int64, values map[string]interface{}) error {
	return s.repo.UpdateClusterPermission(context.Background(), id, values)
}

func (s *RbacService) ClusterPermissionDelete(id int64) error {
	return s.repo.DeleteClusterPermission(context.Background(), id)
}

func (s *RbacService) ClusterPermissionList(userID, clusterID int64, page, limit int) ([]*ClusterPermissionDetail, int64, error) {
	perms, total, err := s.repo.QueryClusterPermissions(context.Background(), userID, clusterID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	cleanDB := s.db.Session(&gorm.Session{NewDB: true})
	result := make([]*ClusterPermissionDetail, 0, len(perms))
	for _, perm := range perms {
		detail := &ClusterPermissionDetail{SysUserCluster: *perm}
			cleanDB.Table("kube_cluster").Select("cluster_name").Where("id = ?", perm.ClusterID).Scan(&detail.ClusterName)
			cleanDB.Table("user").Select("username").Where("id = ?", perm.UserID).Scan(&detail.Username)
		if perm.Namespaces != "" {
			json.Unmarshal([]byte(perm.Namespaces), &detail.NsList)
		}
		result = append(result, detail)
	}
	return result, total, nil
}

func (s *RbacService) ClusterPermissionListByUser(userID int64) ([]*ClusterPermissionDetail, error) {
	perms, err := s.repo.FindClusterPermissionsByUser(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	result := make([]*ClusterPermissionDetail, 0, len(perms))
	for _, perm := range perms {
		detail := &ClusterPermissionDetail{SysUserCluster: *perm}
		s.db.Table("kube_cluster").Select("cluster_name").Where("id = ? AND is_del = 0", perm.ClusterID).Scan(&detail.ClusterName)
		if perm.Namespaces != "" {
			json.Unmarshal([]byte(perm.Namespaces), &detail.NsList)
		}
		result = append(result, detail)
	}
	return result, nil
}

func (s *RbacService) BatchClusterPermissionCreate(userID int64, clusterIDs []int64, roleType, accessLevel string,
	canView, canCreate, canUpdate, canDelete, canExec bool, createdBy int64) error {
	now := uint64(time.Now().Unix())
	if err := s.db.Where("user_id = ? AND cluster_id IN ?", userID, clusterIDs).Delete(&SysUserCluster{}).Error; err != nil {
		return err
	}
	for _, cid := range clusterIDs {
		perm := &SysUserCluster{
			UserID: userID, ClusterID: cid, RoleType: roleType, AccessLevel: accessLevel,
			CanView: canView, CanCreate: canCreate, CanUpdate: canUpdate,
			CanDelete: canDelete, CanExec: canExec,
			CreatedAt: now, ModifiedAt: now, CreatedBy: createdBy,
		}
		if err := s.repo.SaveClusterPermission(context.Background(), perm); err != nil {
			return err
		}
	}
	return nil
}

// ==================== 权限检查（委托到 repository.go standalone 函数） ====================

func (s *RbacService) CheckClusterPermission(userID, clusterID int64, action string) bool {
	if s.repo.IsSuperAdmin(context.Background(), userID) {
		return true
	}
	roles, _ := s.repo.FindUserRoles(context.Background(), userID)
	for _, role := range roles {
		switch action {
		case PermissionActionView:
			if AccessLevelGte(role.ScopeCluster, AccessLevelRead) {
				return true
			}
		case PermissionActionCreate, PermissionActionUpdate, PermissionActionExec:
			if AccessLevelGte(role.ScopeCluster, AccessLevelWrite) {
				return true
			}
		case PermissionActionDelete, PermissionActionManage:
			if AccessLevelGte(role.ScopeCluster, AccessLevelAdmin) {
				return true
			}
		}
	}
	// 显式传 tenantID，避免依赖 ScopedDB 上下文
	return HasClusterPermission(s.scopedDB(), userID, clusterID, action)
}

func (s *RbacService) CheckScopePermission(userID int64, scope, minLevel string) bool {
	return HasScopePermission(s.scopedDB(), userID, scope, minLevel)
}

func (s *RbacService) IsSuperAdmin(userID int64) bool {
	return s.repo.IsSuperAdmin(context.Background(), userID)
}

func (s *RbacService) HasUserPermission(userID int64, permissionName string) bool {
	return s.repo.HasUserPermission(context.Background(), userID, permissionName)
}

func (s *RbacService) GetUserAccessibleClusters(userID int64) ([]*k8s.Cluster, error) {
	if s.repo.IsSuperAdmin(context.Background(), userID) {
		var clusters []*k8s.Cluster
		if err := s.db.Table("kube_cluster").Where("is_del = 0").Find(&clusters).Error; err != nil {
			return nil, err
		}
		return clusters, nil
	}
	roles, _ := s.repo.FindUserRoles(context.Background(), userID)
	for _, role := range roles {
		if AccessLevelGte(role.ScopeCluster, AccessLevelRead) {
			var clusters []*k8s.Cluster
			if err := s.db.Table("kube_cluster").Where("is_del = 0").Find(&clusters).Error; err != nil {
				return nil, err
			}
			return clusters, nil
		}
	}
	tid, ok := tenant.GetTenantID(s.db)
	if !ok {
		return nil, nil
	}
	var clusters []*k8s.Cluster
	err := s.db.Table("kube_cluster").
		Joins("JOIN sys_user_cluster ON kube_cluster.id = sys_user_cluster.cluster_id AND sys_user_cluster.tenant_id = ?", tid).
		Where("sys_user_cluster.user_id = ? AND (sys_user_cluster.access_level IN ('read','write','admin') OR sys_user_cluster.can_view = 1) AND kube_cluster.is_del = 0 AND kube_cluster.tenant_id = ?", userID, tid).
		Find(&clusters).Error
	return clusters, err
}

func (s *RbacService) GetUserAccessibleNamespaces(userID, clusterID int64) ([]string, error) {
	if s.repo.IsSuperAdmin(context.Background(), userID) {
		return []string{}, nil
	}
	perms, _ := s.repo.FindClusterPermissionsByUser(context.Background(), userID)
	for _, perm := range perms {
		if perm.ClusterID == clusterID {
			if perm.Namespaces == "" {
				return []string{}, nil
			}
			var namespaces []string
			if err := json.Unmarshal([]byte(perm.Namespaces), &namespaces); err != nil {
				return []string{}, nil
			}
			return namespaces, nil
		}
	}
	return []string{"__none__"}, nil
}

// ==================== 用户完整 RBAC 信息 ====================

type UserWithRBACInfo struct {
	UserID             int64                    `json:"user_id"`
	Username           string                   `json:"username"`
	IsSuperAdmin       bool                     `json:"is_super_admin"`
	Roles              []*SysRole               `json:"roles"`
	ClusterPermissions []*ClusterPermissionDetail `json:"cluster_permissions"`
	Scopes             *UserScopes              `json:"scopes"`
	PermissionNames    []string                 `json:"permission_names"`
}

type UserScopes struct {
	Platform string `json:"platform"`
	Cluster  string `json:"cluster"`
	CICD     string `json:"cicd"`
}

func (s *RbacService) GetUserWithRBACInfo(userID int64) (*UserWithRBACInfo, error) {
	roles, err := s.repo.FindUserRoles(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	clusterPerms, err := s.ClusterPermissionListByUser(userID)
	if err != nil {
		return nil, err
	}
	// 从角色列表直接判断（避免单独 DB 查询因租户上下文丢失而失败）
	isSuperAdmin := false
	for _, role := range roles {
		if role.RoleType == RoleTypeSuperAdmin {
			isSuperAdmin = true
			break
		}
	}

	username := ""
	if s.userLookup != nil {
		if name, err := s.userLookup.FindUsername(context.Background(), userID); err == nil {
			username = name
		}
	}
	if username == "" {
		var userRow struct{ Username string }
		s.db.WithContext(context.Background()).Table("user").Select("username").Where("id = ?", userID).Scan(&userRow)
		username = userRow.Username
	}

	scopes := &UserScopes{Platform: AccessLevelNone, Cluster: AccessLevelNone, CICD: AccessLevelNone}
	if isSuperAdmin {
		scopes.Platform = AccessLevelAdmin
		scopes.Cluster = AccessLevelAdmin
		scopes.CICD = AccessLevelAdmin
	} else {
		for _, role := range roles {
			if AccessLevelGte(role.ScopePlatform, scopes.Platform) {
				scopes.Platform = role.ScopePlatform
			}
			if AccessLevelGte(role.ScopeCluster, scopes.Cluster) {
				scopes.Cluster = role.ScopeCluster
			}
			if AccessLevelGte(role.ScopeCICD, scopes.CICD) {
				scopes.CICD = role.ScopeCICD
			}
		}
	}

	permissionNames := s.getUserPermissionNames(userID, isSuperAdmin)

	return &UserWithRBACInfo{
		UserID: userID, Username: username, IsSuperAdmin: isSuperAdmin,
		Roles: roles, ClusterPermissions: clusterPerms,
		Scopes: scopes, PermissionNames: permissionNames,
	}, nil
}

func (s *RbacService) getUserPermissionNames(userID int64, isSuperAdmin bool) []string {
	if isSuperAdmin {
		perms, _ := s.repo.ListPermissions(context.Background())
		names := make([]string, 0, len(perms))
		for _, p := range perms {
			names = append(names, p.Name)
		}
		return names
	}
	names, _ := GetUserPermissionNames(s.scopedDB(), userID)
	return names
}

// ==================== 安全防护 ====================

func (s *RbacService) UserRoleAssignSafe(userID int64, roleIDs []int64, operatorID int64) error {
	if userID == operatorID {
		hasAdmin := false
		for _, rid := range roleIDs {
			role, err := s.repo.FindRoleByID(context.Background(), rid)
			if err == nil && (role.RoleType == RoleTypeSuperAdmin || role.RoleType == RoleTypePlatformAdmin) {
				hasAdmin = true
				break
			}
		}
		if !hasAdmin {
			return errors.New("不能移除自己的管理员权限，请让其他管理员修改你的角色")
		}
	}
	return s.UserRoleAssign(userID, roleIDs, operatorID)
}

// ==================== 租户 RBAC 初始化 ====================

func SeedTenantRBAC(db *gorm.DB, tenantID uint32) error {
	if tenantID == 0 || tenantID == dtenant.DefaultTenantID {
		return nil
	}
	now := uint64(time.Now().Unix())
	if err := db.Exec(`
		INSERT INTO sys_role
			(tenant_id, name, display_name, description, role_type,
			 scope_platform, scope_cluster, scope_cicd,
			 is_system, color, icon, sort_order, created_at, modified_at, is_del)
		SELECT ?, name, display_name, description, role_type,
			   scope_platform, scope_cluster, scope_cicd,
			   is_system, color, icon, sort_order, ?, ?, 0
		FROM sys_role
		WHERE tenant_id = ? AND is_del = 0 AND role_type IN (?, ?, ?)`,
		tenantID, now, now, dtenant.DefaultTenantID,
		RoleTypeSuperAdmin, RoleTypePlatformAdmin, RoleTypeDevOps,
	).Error; err != nil {
		return err
	}
	return db.Exec(`
		INSERT INTO sys_role_permission (tenant_id, role_id, permission_id, created_at)
		SELECT ?, dst.id, rp.permission_id, ?
		FROM sys_role_permission rp
		JOIN sys_role src ON src.id = rp.role_id AND src.tenant_id = ? AND src.is_del = 0
		JOIN sys_role dst ON dst.tenant_id = ? AND dst.role_type = src.role_type AND dst.is_del = 0`,
		tenantID, now, dtenant.DefaultTenantID, tenantID,
	).Error
}

func ListTenantsMissingSuperAdmin(db *gorm.DB) ([]uint32, error) {
	var ids []uint32
	err := db.Raw(`
		SELECT t.id FROM tenant t
		WHERE t.is_del = 0 AND t.status = 1
		  AND NOT EXISTS (
			SELECT 1 FROM sys_role r
			WHERE r.tenant_id = t.id AND r.role_type = ? AND r.is_del = 0
		  )
		ORDER BY t.id`, RoleTypeSuperAdmin).Scan(&ids).Error
	return ids, err
}
