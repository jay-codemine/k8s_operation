package dao

import (
	"encoding/json"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/tenant"
	"time"

	"gorm.io/gorm"
)

// ==================== 租户 RBAC 初始化 ====================

// TenantSeedRBAC 为指定租户克隆一份默认角色及其权限授权。
//
// 新建租户必须做这一步，否则该租户完全不可用：读路径（models.IsSuperAdmin /
// GetUserRoles）按 tenant_id 过滤 sys_role，租户下没有 super_admin 行时，
// 它的所有用户都判不出任何权限，连租户管理页都进不去。
//
// 从默认租户 SELECT-INSERT 克隆而非硬编码字段值，这样 scope_platform/scope_cluster/
// scope_cicd 等平台默认值日后调整时新租户自动跟随。sys_permission 是全局共享目录，
// 只克隆 sys_role_permission 的授权关系，按 role_type 对齐新旧角色。
// db 传事务句柄即可与租户记录本身同生共死。
func TenantSeedRBAC(db *gorm.DB, tenantID uint32) error {
	if tenantID == 0 || tenantID == models.DefaultTenantID {
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
		tenantID, now, now, models.DefaultTenantID,
		models.RoleTypeSuperAdmin, models.RoleTypePlatformAdmin, models.RoleTypeDevOps,
	).Error; err != nil {
		return err
	}

	return db.Exec(`
		INSERT INTO sys_role_permission (tenant_id, role_id, permission_id, created_at)
		SELECT ?, dst.id, rp.permission_id, ?
		FROM sys_role_permission rp
		JOIN sys_role src ON src.id = rp.role_id AND src.tenant_id = ? AND src.is_del = 0
		JOIN sys_role dst ON dst.tenant_id = ? AND dst.role_type = src.role_type AND dst.is_del = 0`,
		tenantID, now, models.DefaultTenantID, tenantID,
	).Error
}

// TenantsMissingSuperAdmin 列出启用中但没有 super_admin 角色的租户 ID（即不可用的租户）
func TenantsMissingSuperAdmin(db *gorm.DB) ([]uint32, error) {
	var ids []uint32
	err := db.Raw(`
		SELECT t.id FROM tenant t
		WHERE t.is_del = 0 AND t.status = 1
		  AND NOT EXISTS (
			SELECT 1 FROM sys_role r
			WHERE r.tenant_id = t.id AND r.role_type = ? AND r.is_del = 0
		  )
		ORDER BY t.id`, models.RoleTypeSuperAdmin).Scan(&ids).Error
	return ids, err
}

// ==================== 角色管理 ====================

// RoleCreate 创建角色
func (d *Dao) RoleCreate(name, displayName, description, roleType, scopePlatform, scopeCluster, scopeCICD, color, icon string) (*models.SysRole, error) {
	now := uint64(time.Now().Unix())
	role := &models.SysRole{
		Name:          name,
		DisplayName:   displayName,
		Description:   description,
		RoleType:      roleType,
		ScopePlatform: scopePlatform,
		ScopeCluster:  scopeCluster,
		ScopeCICD:     scopeCICD,
		Color:         color,
		Icon:          icon,
		IsSystem:      false,
		SortOrder:     0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}
	if err := d.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// RoleUpdate 更新角色
func (d *Dao) RoleUpdate(id int64, values map[string]interface{}) error {
	values["modified_at"] = uint64(time.Now().Unix())
	return d.db.Model(&models.SysRole{}).Where("id = ? AND is_del = 0", id).Updates(values).Error
}

// RoleDelete 删除角色（软删除）
func (d *Dao) RoleDelete(id int64) error {
	now := uint64(time.Now().Unix())
	return d.db.Model(&models.SysRole{}).Where("id = ? AND is_del = 0 AND is_system = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// RoleGetByID 根据ID获取角色
func (d *Dao) RoleGetByID(id int64) (*models.SysRole, error) {
	var role models.SysRole
	if err := d.db.Where("id = ? AND is_del = 0", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// RoleGetByName 根据名称获取角色
func (d *Dao) RoleGetByName(name string) (*models.SysRole, error) {
	var role models.SysRole
	if err := d.db.Where("name = ? AND is_del = 0", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// RoleList 获取角色列表
func (d *Dao) RoleList(name, roleType string, page, limit int) ([]*models.SysRole, int64, error) {
	var roles []*models.SysRole
	var total int64

	query := d.db.Model(&models.SysRole{}).Where("is_del = 0")
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

// RoleListAll 获取所有角色（不分页）
func (d *Dao) RoleListAll() ([]*models.SysRole, error) {
	var roles []*models.SysRole
	if err := d.db.Where("is_del = 0").Order("sort_order ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ==================== 权限管理 ====================

// PermissionList 获取权限列表
func (d *Dao) PermissionList() ([]*models.SysPermission, error) {
	var permissions []*models.SysPermission
	if err := d.db.Order("sort_order ASC, id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// PermissionGetByRoleID 获取角色的权限列表
func (d *Dao) PermissionGetByRoleID(roleID int64) ([]*models.SysPermission, error) {
	var permissions []*models.SysPermission
	err := d.db.Table("sys_permission").
		Joins("JOIN sys_role_permission ON sys_permission.id = sys_role_permission.permission_id").
		Where("sys_role_permission.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// RolePermissionAssign 分配角色权限
func (d *Dao) RolePermissionAssign(roleID int64, permissionIDs []int64) error {
	// 先删除原有权限
	if err := d.db.Where("role_id = ?", roleID).Delete(&models.SysRolePermission{}).Error; err != nil {
		return err
	}
	// 重新分配
	now := uint64(time.Now().Unix())
	for _, pid := range permissionIDs {
		rp := &models.SysRolePermission{
			RoleID:       roleID,
			PermissionID: pid,
			CreatedAt:    now,
		}
		if err := d.db.Create(rp).Error; err != nil {
			return err
		}
	}
	return nil
}

// ==================== 用户角色管理 ====================

// UserRoleAssign 分配用户角色
func (d *Dao) UserRoleAssign(userID int64, roleIDs []int64, createdBy int64) error {
	// 先删除原有角色
	if err := d.db.Where("user_id = ?", userID).Delete(&models.SysUserRole{}).Error; err != nil {
		return err
	}
	// 重新分配
	now := uint64(time.Now().Unix())
	for _, rid := range roleIDs {
		ur := &models.SysUserRole{
			UserID:    userID,
			RoleID:    rid,
			CreatedAt: now,
			CreatedBy: createdBy,
		}
		if err := d.db.Create(ur).Error; err != nil {
			return err
		}
	}
	return nil
}

// UserRoleRemove 移除用户角色
func (d *Dao) UserRoleRemove(userID, roleID int64) error {
	return d.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.SysUserRole{}).Error
}

// UserRoleList 获取用户角色列表
func (d *Dao) UserRoleList(userID int64) ([]*models.SysRole, error) {
	return models.GetUserRoles(d.db, userID)
}

// RoleUserList 获取角色绑定的用户列表
func (d *Dao) RoleUserList(roleID int64) ([]*models.User, error) {
	tid, ok := tenant.GetTenantID(d.db)
	if !ok {
		return nil, nil
	}
	var users []*models.User
	err := global.DB.Table("user").
		Joins("JOIN sys_user_role ON user.id = sys_user_role.user_id").
		Where("sys_user_role.role_id = ? AND user.is_del = 0 AND user.tenant_id = ? AND sys_user_role.tenant_id = ?", roleID, tid, tid).
		Find(&users).Error
	return users, err
}

// UserRoleListWithCount 获取用户角色关联列表（带用户数统计）
func (d *Dao) UserRoleListWithCount(page, limit int) ([]*models.RoleWithPermissions, int64, error) {
	var roles []*models.SysRole
	var total int64

	query := d.db.Model(&models.SysRole{}).Where("is_del = 0")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	// 统计每个角色的用户数
	var result []*models.RoleWithPermissions
	for _, role := range roles {
		var count int64
		d.db.Model(&models.SysUserRole{}).Where("role_id = ?", role.ID).Count(&count)
		rwp := &models.RoleWithPermissions{
			SysRole:   *role,
			UserCount: count,
		}
		result = append(result, rwp)
	}

	return result, total, nil
}

// ==================== 集群权限管理 ====================

// ClusterPermissionCreate 创建集群权限
func (d *Dao) ClusterPermissionCreate(userID, clusterID int64, roleType, accessLevel string, namespaces []string,
	canView, canCreate, canUpdate, canDelete, canExec bool, expireAt uint64, createdBy int64) (*models.SysUserCluster, error) {
	now := uint64(time.Now().Unix())

	// 序列化命名空间
	nsJSON := ""
	if len(namespaces) > 0 {
		if data, err := json.Marshal(namespaces); err == nil {
			nsJSON = string(data)
		}
	}

	// 检查是否已存在，存在则更新
	var existing models.SysUserCluster
	err := d.db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).First(&existing).Error
	if err == nil {
		// 已存在，更新
		updates := map[string]interface{}{
			"role_type":    roleType,
			"access_level": accessLevel,
			"namespaces":   nsJSON,
			"can_view":     canView,
			"can_create":   canCreate,
			"can_update":   canUpdate,
			"can_delete":   canDelete,
			"can_exec":     canExec,
			"expire_at":    expireAt,
			"modified_at":  now,
		}
		if err := d.db.Model(&existing).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}

	// 不存在，创建新记录
	perm := &models.SysUserCluster{
		UserID:      userID,
		ClusterID:   clusterID,
		RoleType:    roleType,
		AccessLevel: accessLevel,
		Namespaces:  nsJSON,
		CanView:     canView,
		CanCreate:   canCreate,
		CanUpdate:   canUpdate,
		CanDelete:   canDelete,
		CanExec:     canExec,
		ExpireAt:    expireAt,
		CreatedAt:   now,
		ModifiedAt:  now,
		CreatedBy:   createdBy,
	}
	if err := d.db.Create(perm).Error; err != nil {
		return nil, err
	}
	return perm, nil
}

// ClusterPermissionUpdate 更新集群权限
func (d *Dao) ClusterPermissionUpdate(id int64, values map[string]interface{}) error {
	values["modified_at"] = uint64(time.Now().Unix())
	return d.db.Model(&models.SysUserCluster{}).Where("id = ?", id).Updates(values).Error
}

// ClusterPermissionDelete 删除集群权限
func (d *Dao) ClusterPermissionDelete(id int64) error {
	return d.db.Where("id = ?", id).Delete(&models.SysUserCluster{}).Error
}

// ClusterPermissionGetByID 根据ID获取集群权限
func (d *Dao) ClusterPermissionGetByID(id int64) (*models.SysUserCluster, error) {
	var perm models.SysUserCluster
	if err := d.db.Where("id = ?", id).First(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

// ClusterPermissionList 获取集群权限列表
func (d *Dao) ClusterPermissionList(userID, clusterID int64, page, limit int) ([]*models.ClusterPermissionDetail, int64, error) {
	var perms []*models.SysUserCluster
	var total int64

	query := d.db.Model(&models.SysUserCluster{})
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

	// 填充集群名称和用户名
	result := make([]*models.ClusterPermissionDetail, 0, len(perms))
	for _, perm := range perms {
		detail := &models.ClusterPermissionDetail{
			SysUserCluster: *perm,
		}

		// 获取集群名称
		var cluster models.K8sCluster
		if err := d.db.Select("cluster_name").Where("id = ?", perm.ClusterID).First(&cluster).Error; err == nil {
			detail.ClusterName = cluster.ClusterName
		}

		// 获取用户名
		var user models.User
		if err := d.db.Select("username").Where("id = ?", perm.UserID).First(&user).Error; err == nil {
			detail.Username = user.Username
		}

		// 解析命名空间
		if perm.Namespaces != "" {
			var nsList []string
			if err := json.Unmarshal([]byte(perm.Namespaces), &nsList); err == nil {
				detail.NsList = nsList
			}
		}

		result = append(result, detail)
	}

	return result, total, nil
}

// ClusterPermissionListByUser 获取用户的所有集群权限
func (d *Dao) ClusterPermissionListByUser(userID int64) ([]*models.ClusterPermissionDetail, error) {
	var perms []*models.SysUserCluster
	if err := d.db.Where("user_id = ?", userID).Find(&perms).Error; err != nil {
		return nil, err
	}

	result := make([]*models.ClusterPermissionDetail, 0, len(perms))
	for _, perm := range perms {
		detail := &models.ClusterPermissionDetail{
			SysUserCluster: *perm,
		}

		// 获取集群名称
		var cluster models.K8sCluster
		if err := d.db.Select("cluster_name").Where("id = ? AND is_del = 0", perm.ClusterID).First(&cluster).Error; err == nil {
			detail.ClusterName = cluster.ClusterName
		}

		// 解析命名空间
		if perm.Namespaces != "" {
			var nsList []string
			if err := json.Unmarshal([]byte(perm.Namespaces), &nsList); err == nil {
				detail.NsList = nsList
			}
		}

		result = append(result, detail)
	}

	return result, nil
}

// ClusterPermissionCheck 检查用户集群权限
func (d *Dao) ClusterPermissionCheck(userID, clusterID int64, action string) bool {
	return models.HasClusterPermission(d.db, userID, clusterID, action)
}

// IsSuperAdmin 检查用户是否为超级管理员
func (d *Dao) IsSuperAdmin(userID int64) bool {
	return models.IsSuperAdmin(d.db, userID)
}

// ==================== 批量操作 ====================

// BatchClusterPermissionCreate 批量创建集群权限
func (d *Dao) BatchClusterPermissionCreate(userID int64, clusterIDs []int64, roleType, accessLevel string,
	canView, canCreate, canUpdate, canDelete, canExec bool, createdBy int64) error {
	now := uint64(time.Now().Unix())

	// 先删除用户在这些集群的现有权限
	if err := d.db.Where("user_id = ? AND cluster_id IN ?", userID, clusterIDs).Delete(&models.SysUserCluster{}).Error; err != nil {
		return err
	}

	// 批量创建新权限
	for _, cid := range clusterIDs {
		perm := &models.SysUserCluster{
			UserID:      userID,
			ClusterID:   cid,
			RoleType:    roleType,
			AccessLevel: accessLevel,
			CanView:     canView,
			CanCreate:   canCreate,
			CanUpdate:   canUpdate,
			CanDelete:   canDelete,
			CanExec:     canExec,
			CreatedAt:   now,
			ModifiedAt:  now,
			CreatedBy:   createdBy,
		}
		if err := d.db.Create(perm).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetUserAccessibleClusters 获取用户可访问的集群列表
func (d *Dao) GetUserAccessibleClusters(userID int64) ([]*models.K8sCluster, error) {
	// 先检查是否是超级管理员
	if d.IsSuperAdmin(userID) {
		var clusters []*models.K8sCluster
		if err := d.db.Where("is_del = 0").Find(&clusters).Error; err != nil {
			return nil, err
		}
		return clusters, nil
	}

	// 检查用户角色的 scope_cluster >= read，则可访问所有集群
	roles, _ := d.UserRoleList(userID)
	for _, role := range roles {
		if models.AccessLevelGte(role.ScopeCluster, models.AccessLevelRead) {
			var clusters []*models.K8sCluster
			if err := d.db.Where("is_del = 0").Find(&clusters).Error; err != nil {
				return nil, err
			}
			return clusters, nil
		}
	}

	// 否则只返回有权限的集群（兼容旧 can_view + 新 access_level）
	tid, ok := tenant.GetTenantID(d.db)
	if !ok {
		return nil, nil
	}
	var clusters []*models.K8sCluster
	err := global.DB.Table("kube_cluster").
		Joins("JOIN sys_user_cluster ON kube_cluster.id = sys_user_cluster.cluster_id AND sys_user_cluster.tenant_id = ?", tid).
		Where("sys_user_cluster.user_id = ? AND (sys_user_cluster.access_level IN ('read','write','admin') OR sys_user_cluster.can_view = 1) AND kube_cluster.is_del = 0 AND kube_cluster.tenant_id = ?", userID, tid).
		Find(&clusters).Error
	return clusters, err
}

// GetUserAccessibleNamespaces 获取用户在指定集群可访问的命名空间
func (d *Dao) GetUserAccessibleNamespaces(userID, clusterID int64) ([]string, error) {
	// 超级管理员可访问所有
	if d.IsSuperAdmin(userID) {
		return []string{}, nil
	}

	// 查询用户在该集群的权限
	var perm models.SysUserCluster
	err := d.db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).First(&perm).Error
	if err != nil {
		return []string{"__none__"}, nil // 返回特殊标记表示无权限
	}

	// 解析命名空间 JSON
	if perm.Namespaces == "" {
		return []string{}, nil // 空字符串表示可访问所有
	}

	var namespaces []string
	if err := json.Unmarshal([]byte(perm.Namespaces), &namespaces); err != nil {
		return []string{}, nil
	}

	return namespaces, nil
}
