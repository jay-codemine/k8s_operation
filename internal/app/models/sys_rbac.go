package models

import (
	"gorm.io/gorm"

	"k8soperation/global"
	"k8soperation/pkg/tenant"
)

// ==================== 角色类型常量 ====================

const (
	RoleTypeSuperAdmin    = "super_admin"    // 超级管理员
	RoleTypePlatformAdmin = "platform_admin" // 平台管理员
	RoleTypeDevOps        = "devops"         // 运维工程师
	RoleTypeDeveloper     = "developer"      // 开发工程师
	RoleTypeTester        = "tester"         // 测试工程师
	RoleTypeViewer        = "viewer"         // 观察者

	// 兼容旧值
	RoleTypeClusterAdmin = "cluster_admin" // deprecated: 映射到 devops
)

// ==================== 权限域常量（三大功能域） ====================

const (
	ScopePlatform = "platform" // 🏛 平台域：用户/角色/系统设置/审计/数据源/告警
	ScopeCluster  = "cluster"  // ☸ 集群域：集群/NS/Workload/Service/Config/Node/监控/日志/终端
	ScopeCICD     = "cicd"     // 🚀 发布域：流水线/构建/部署/制品/镜像/代码扫描/审批
)

// ==================== 权限级别常量（每域通用） ====================

const (
	AccessLevelNone  = "none"  // 不可见
	AccessLevelRead  = "read"  // 只读
	AccessLevelWrite = "write" // 读写（可创建/更新/触发，不可删除系统级资源）
	AccessLevelAdmin = "admin" // 全权（含删除/批量/管理他人资源）
)

// AccessLevelGte 判断 a 级别 >= b 级别
func AccessLevelGte(a, b string) bool {
	return accessLevelOrder(a) >= accessLevelOrder(b)
}

func accessLevelOrder(level string) int {
	switch level {
	case AccessLevelAdmin:
		return 4
	case AccessLevelWrite:
		return 3
	case AccessLevelRead:
		return 2
	case AccessLevelNone:
		return 1
	default:
		return 0
	}
}

// ==================== 权限操作常量（兼容旧代码） ====================

const (
	PermissionActionView   = "view"   // 查看
	PermissionActionCreate = "create" // 创建
	PermissionActionUpdate = "update" // 更新
	PermissionActionDelete = "delete" // 删除
	PermissionActionExec   = "exec"   // 执行（如进入容器）
	PermissionActionManage = "manage" // 管理（完整权限）
)

// ==================== 资源类型/模块常量 ====================

const (
	// 平台域模块
	ResourceTypeUser     = "user"     // 用户管理
	ResourceTypeRole     = "role"     // 角色管理
	ResourceTypeSettings = "settings" // 系统设置
	ResourceTypeAudit    = "audit"    // 审计日志

	// 集群域模块
	ResourceTypeCluster    = "cluster"    // 集群管理
	ResourceTypeWorkload   = "workload"   // 工作负载（Deployment/Pod/DaemonSet...）
	ResourceTypeNetwork    = "network"    // 服务与路由（Service/Ingress）
	ResourceTypeConfig     = "config"     // 配置管理（ConfigMap/Secret）
	ResourceTypeStorage    = "storage"    // 存储管理（PV/PVC/SC）
	ResourceTypeNode       = "node"       // 节点管理
	ResourceTypeMonitor    = "monitor"    // 监控与日志

	// 发布域模块
	ResourceTypePipeline = "pipeline" // 流水线管理
	ResourceTypeArtifact = "artifact" // 制品与镜像

	// deprecated: 保留兼容
	ResourceTypeNamespace  = "namespace"
	ResourceTypeDeployment = "deployment"
	ResourceTypePod        = "pod"
	ResourceTypeService    = "service"
	ResourceTypeConfigMap  = "configmap"
	ResourceTypeSecret     = "secret"
	ResourceTypePVC        = "pvc"
	ResourceTypeIngress    = "ingress"
)

// ==================== 系统角色表 ====================

// SysRole 系统角色
// 注意 TenantID 不能省略：读路径（IsSuperAdmin/GetUserRoles）强制按 tenant_id 过滤，
// 若结构体没有该字段，INSERT 就不带这一列、由 MySQL 的 DEFAULT 1 兜住，
// 于是非默认租户创建的角色全部落进 1 号租户，该租户自己再也查不到 -> 租户不可用。
// 字段存在后 tenant.RegisterCallbacks 注册的 create 回调会自动填充当前租户。
type SysRole struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID      uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`   // 所属租户
	Name          string `gorm:"column:name;size:50" json:"name"`                      // 角色标识（租户内唯一，见 uk_sys_role_tenant_name）
	DisplayName   string `gorm:"column:display_name;size:100" json:"display_name"`    // 显示名称
	Description   string `gorm:"column:description;size:500" json:"description"`      // 描述
	RoleType      string `gorm:"column:role_type;size:30" json:"role_type"`           // 角色类型
	ScopePlatform string `gorm:"column:scope_platform;size:10;default:'none'" json:"scope_platform"` // 平台域级别
	ScopeCluster  string `gorm:"column:scope_cluster;size:10;default:'none'" json:"scope_cluster"`   // 集群域级别
	ScopeCICD     string `gorm:"column:scope_cicd;size:10;default:'none'" json:"scope_cicd"`         // 发布域级别
	IsSystem      bool   `gorm:"column:is_system;default:false" json:"is_system"`     // 是否系统内置
	Color         string `gorm:"column:color;size:20;default:'#1890ff'" json:"color"` // 角色颜色标识
	Icon          string `gorm:"column:icon;size:50;default:'user'" json:"icon"`      // 图标
	SortOrder     int    `gorm:"column:sort_order;default:0" json:"sort_order"`       // 排序
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del;default:0" json:"is_del"`
}

func (SysRole) TableName() string { return "sys_role" }

// GetEffectiveScope 获取角色在指定域的有效级别
func (r *SysRole) GetEffectiveScope(scope string) string {
	switch scope {
	case ScopePlatform:
		return r.ScopePlatform
	case ScopeCluster:
		return r.ScopeCluster
	case ScopeCICD:
		return r.ScopeCICD
	default:
		return AccessLevelNone
	}
}

// ==================== 系统权限表 ====================

// SysPermission 系统权限定义（v2: 按功能模块聚合，不再逐资源拆分）
type SysPermission struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"column:name;size:100;uniqueIndex" json:"name"`   // 权限标识（唯一）
	DisplayName  string `gorm:"column:display_name;size:100" json:"display_name"` // 显示名称
	Description  string `gorm:"column:description;size:500" json:"description"`   // 描述
	Scope        string `gorm:"column:scope;size:20;default:'cluster'" json:"scope"` // 所属功能域：platform/cluster/cicd
	ResourceType string `gorm:"column:resource_type;size:50" json:"resource_type"` // 模块标识
	Action       string `gorm:"column:action;size:30" json:"action"`             // 操作类型（兼容旧字段）
	Tag          string `gorm:"column:tag;size:50;default:''" json:"tag"`         // 标签分组（用于前端分类展示）
	ParentID     int64  `gorm:"column:parent_id;default:0" json:"parent_id"`     // 父权限ID
	Path         string `gorm:"column:path;size:200" json:"path"`               // 权限路径（用于树形展示）
	SortOrder    int    `gorm:"column:sort_order;default:0" json:"sort_order"` // 排序
	CreatedAt    uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt   uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (SysPermission) TableName() string { return "sys_permission" }

// ==================== 角色权限关联表 ====================

// SysRolePermission 角色权限关联
type SysRolePermission struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"` // 所属租户
	RoleID       int64  `gorm:"column:role_id;index" json:"role_id"`           // 角色ID
	PermissionID int64  `gorm:"column:permission_id;index" json:"permission_id"` // 权限ID
	CreatedAt    uint64 `gorm:"column:created_at" json:"created_at"`
}

func (SysRolePermission) TableName() string { return "sys_role_permission" }

// ==================== 用户角色关联表 ====================

// SysUserRole 用户角色关联
type SysUserRole struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID   uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"` // 所属租户
	UserID     int64  `gorm:"column:user_id;index" json:"user_id"`     // 用户ID
	RoleID     int64  `gorm:"column:role_id;index" json:"role_id"`     // 角色ID
	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	CreatedBy  int64  `gorm:"column:created_by" json:"created_by"`     // 创建人
}

func (SysUserRole) TableName() string { return "sys_user_role" }

// ==================== 用户集群权限表 ====================

// SysUserCluster 用户集群权限（细粒度控制用户对特定集群的访问）
type SysUserCluster struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      int64  `gorm:"column:user_id;index" json:"user_id"`               // 用户ID
	ClusterID   int64  `gorm:"column:cluster_id;index" json:"cluster_id"`         // 集群ID
	RoleType    string `gorm:"column:role_type;size:30" json:"role_type"`          // 在该集群的角色类型
	AccessLevel string `gorm:"column:access_level;size:10;default:'read'" json:"access_level"` // 权限级别：none/read/write/admin
	Namespaces  string `gorm:"column:namespaces;type:text" json:"namespaces"`     // 可访问的命名空间（JSON数组，空表示全部）
	CanView     bool   `gorm:"column:can_view;default:true" json:"can_view"`      // deprecated: 用 access_level 代替
	CanCreate   bool   `gorm:"column:can_create;default:false" json:"can_create"` // deprecated
	CanUpdate   bool   `gorm:"column:can_update;default:false" json:"can_update"` // deprecated
	CanDelete   bool   `gorm:"column:can_delete;default:false" json:"can_delete"` // deprecated
	CanExec     bool   `gorm:"column:can_exec;default:false" json:"can_exec"`     // deprecated
	ExpireAt    uint64 `gorm:"column:expire_at;default:0" json:"expire_at"`       // 过期时间（0表示永不过期）
	CreatedAt   uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt  uint64 `gorm:"column:modified_at" json:"modified_at"`
	CreatedBy   int64  `gorm:"column:created_by" json:"created_by"` // 授权人
}

// GetAccessLevel 获取有效权限级别（兼容旧 bool 字段）
func (c *SysUserCluster) GetAccessLevel() string {
	if c.AccessLevel != "" && c.AccessLevel != "none" {
		return c.AccessLevel
	}
	// 兼容旧数据：从 bool 字段推导
	if c.CanDelete {
		return AccessLevelAdmin
	}
	if c.CanCreate || c.CanUpdate || c.CanExec {
		return AccessLevelWrite
	}
	if c.CanView {
		return AccessLevelRead
	}
	return AccessLevelNone
}

func (SysUserCluster) TableName() string { return "sys_user_cluster" }

// ==================== 扩展用户模型（添加角色相关字段） ====================

// UserWithRole 带角色信息的用户
type UserWithRole struct {
	ID          uint32     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Avatar      string     `json:"avatar"`
	Status      uint8      `json:"status"`
	Roles       []*SysRole `json:"roles" gorm:"-"`           // 用户角色列表
	RoleNames   []string   `json:"role_names" gorm:"-"`      // 角色名称列表
	ClusterIDs  []int64    `json:"cluster_ids" gorm:"-"`     // 可访问的集群ID列表
	IsSuperAdmin bool      `json:"is_super_admin" gorm:"-"`  // 是否超级管理员
	CreatedAt   uint32     `json:"created_at"`
	ModifiedAt  uint32     `json:"modified_at"`
}

// ==================== 角色权限详情 ====================

// RoleWithPermissions 带权限的角色
type RoleWithPermissions struct {
	SysRole
	Permissions []*SysPermission `json:"permissions" gorm:"-"` // 权限列表
	UserCount   int64            `json:"user_count" gorm:"-"`  // 用户数量
}

// ==================== 集群权限详情 ====================

// ClusterPermissionDetail 集群权限详情
type ClusterPermissionDetail struct {
	SysUserCluster
	ClusterName string   `json:"cluster_name" gorm:"-"` // 集群名称
	Username    string   `json:"username" gorm:"-"`     // 用户名
	NsList      []string `json:"ns_list" gorm:"-"`      // 命名空间列表
}

// ==================== DAO 方法 ====================

// GetUserRoles 获取用户角色列表
func GetUserRoles(db *gorm.DB, userID int64) ([]*SysRole, error) {
	tid, ok := tenant.GetTenantID(db)
	if !ok {
		return nil, gorm.ErrMissingWhereClause
	}
	var roles []*SysRole
	err := global.DB.Table("sys_role").
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

// HasClusterPermission 检查用户是否有集群权限（兼容新旧两种模式）
func HasClusterPermission(db *gorm.DB, userID, clusterID int64, action string) bool {
	var record SysUserCluster
	err := db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).First(&record).Error
	if err != nil {
		return false
	}

	// 优先使用新 access_level 字段
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
	// 超管直接通过
	if IsSuperAdmin(db, userID) {
		return true
	}

	// 获取用户所有角色，取最高 scope 级别
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
	global.DB.Table("sys_user_role").
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
	global.DB.Table("sys_role_permission").
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
	err := global.DB.Table("sys_permission").
		Joins("JOIN sys_role_permission ON sys_role_permission.permission_id = sys_permission.id").
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role_permission.role_id").
		Where("sys_user_role.user_id = ? AND sys_user_role.tenant_id = ?", userID, tid).
		Pluck("sys_permission.name", &names).Error
	return names, err
}
