package rbac

// ==================== 角色类型常量 ====================

const (
	RoleTypeSuperAdmin    = "super_admin"
	RoleTypePlatformAdmin = "platform_admin"
	RoleTypeDevOps        = "devops"
	RoleTypeDeveloper     = "developer"
	RoleTypeTester        = "tester"
	RoleTypeViewer        = "viewer"
	RoleTypeClusterAdmin  = "cluster_admin" // deprecated
)

// ==================== 权限域常量 ====================

const (
	ScopePlatform = "platform"
	ScopeCluster  = "cluster"
	ScopeCICD     = "cicd"
)

// ==================== 权限级别常量 ====================

const (
	AccessLevelNone  = "none"
	AccessLevelRead  = "read"
	AccessLevelWrite = "write"
	AccessLevelAdmin = "admin"
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

// ==================== 权限操作常量 ====================

const (
	PermissionActionView   = "view"
	PermissionActionCreate = "create"
	PermissionActionUpdate = "update"
	PermissionActionDelete = "delete"
	PermissionActionExec   = "exec"
	PermissionActionManage = "manage"
)

// ==================== 资源类型/模块常量 ====================

const (
	ResourceTypeUser     = "user"
	ResourceTypeRole     = "role"
	ResourceTypeSettings = "settings"
	ResourceTypeAudit    = "audit"

	ResourceTypeCluster  = "cluster"
	ResourceTypeWorkload = "workload"
	ResourceTypeNetwork  = "network"
	ResourceTypeConfig   = "config"
	ResourceTypeStorage  = "storage"
	ResourceTypeNode     = "node"
	ResourceTypeMonitor  = "monitor"

	ResourceTypePipeline = "pipeline"
	ResourceTypeArtifact = "artifact"

	// deprecated
	ResourceTypeNamespace  = "namespace"
	ResourceTypeDeployment = "deployment"
	ResourceTypePod        = "pod"
	ResourceTypeService    = "service"
	ResourceTypeConfigMap  = "configmap"
	ResourceTypeSecret     = "secret"
	ResourceTypePVC        = "pvc"
	ResourceTypeIngress    = "ingress"
)

// ==================== 模型 ====================

// SysRole 系统角色
type SysRole struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID      uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	Name          string `gorm:"column:name;size:50" json:"name"`
	DisplayName   string `gorm:"column:display_name;size:100" json:"display_name"`
	Description   string `gorm:"column:description;size:500" json:"description"`
	RoleType      string `gorm:"column:role_type;size:30" json:"role_type"`
	ScopePlatform string `gorm:"column:scope_platform;size:10;default:'none'" json:"scope_platform"`
	ScopeCluster  string `gorm:"column:scope_cluster;size:10;default:'none'" json:"scope_cluster"`
	ScopeCICD     string `gorm:"column:scope_cicd;size:10;default:'none'" json:"scope_cicd"`
	IsSystem      bool   `gorm:"column:is_system;default:false" json:"is_system"`
	Color         string `gorm:"column:color;size:20;default:'#1890ff'" json:"color"`
	Icon          string `gorm:"column:icon;size:50;default:'user'" json:"icon"`
	SortOrder     int    `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del;default:0" json:"is_del"`
}

// AggregateID 实现 domain.AggregateRoot 接口
func (r SysRole) AggregateID() int64 { return r.ID }

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

// SysPermission 系统权限定义
type SysPermission struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"column:name;size:100;uniqueIndex" json:"name"`
	DisplayName  string `gorm:"column:display_name;size:100" json:"display_name"`
	Description  string `gorm:"column:description;size:500" json:"description"`
	Scope        string `gorm:"column:scope;size:20;default:'cluster'" json:"scope"`
	ResourceType string `gorm:"column:resource_type;size:50" json:"resource_type"`
	Action       string `gorm:"column:action;size:30" json:"action"`
	Tag          string `gorm:"column:tag;size:50;default:''" json:"tag"`
	ParentID     int64  `gorm:"column:parent_id;default:0" json:"parent_id"`
	Path         string `gorm:"column:path;size:200" json:"path"`
	SortOrder    int    `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt    uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt   uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (SysPermission) TableName() string { return "sys_permission" }

// SysRolePermission 角色权限关联
type SysRolePermission struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	RoleID       int64  `gorm:"column:role_id;index" json:"role_id"`
	PermissionID int64  `gorm:"column:permission_id;index" json:"permission_id"`
	CreatedAt    uint64 `gorm:"column:created_at" json:"created_at"`
}

func (SysRolePermission) TableName() string { return "sys_role_permission" }

// SysUserRole 用户角色关联
type SysUserRole struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID  uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	UserID    int64  `gorm:"column:user_id;index" json:"user_id"`
	RoleID    int64  `gorm:"column:role_id;index" json:"role_id"`
	CreatedAt uint64 `gorm:"column:created_at" json:"created_at"`
	CreatedBy int64  `gorm:"column:created_by" json:"created_by"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

// SysUserCluster 用户集群权限
type SysUserCluster struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      int64  `gorm:"column:user_id;index" json:"user_id"`
	ClusterID   int64  `gorm:"column:cluster_id;index" json:"cluster_id"`
	RoleType    string `gorm:"column:role_type;size:30" json:"role_type"`
	AccessLevel string `gorm:"column:access_level;size:10;default:'read'" json:"access_level"`
	Namespaces  string `gorm:"column:namespaces;type:text" json:"namespaces"`
	CanView     bool   `gorm:"column:can_view;default:true" json:"can_view"`
	CanCreate   bool   `gorm:"column:can_create;default:false" json:"can_create"`
	CanUpdate   bool   `gorm:"column:can_update;default:false" json:"can_update"`
	CanDelete   bool   `gorm:"column:can_delete;default:false" json:"can_delete"`
	CanExec     bool   `gorm:"column:can_exec;default:false" json:"can_exec"`
	ExpireAt    uint64 `gorm:"column:expire_at;default:0" json:"expire_at"`
	CreatedAt   uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt  uint64 `gorm:"column:modified_at" json:"modified_at"`
	CreatedBy   int64  `gorm:"column:created_by" json:"created_by"`
}

func (SysUserCluster) TableName() string { return "sys_user_cluster" }

// GetAccessLevel 获取有效权限级别
func (c *SysUserCluster) GetAccessLevel() string {
	if c.AccessLevel != "" && c.AccessLevel != "none" {
		return c.AccessLevel
	}
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

// ==================== DTOs ====================

// UserWithRole 带角色信息的用户
type UserWithRole struct {
	ID           uint32     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Avatar       string     `json:"avatar"`
	Status       uint8      `json:"status"`
	Roles        []*SysRole `json:"roles" gorm:"-"`
	RoleNames    []string   `json:"role_names" gorm:"-"`
	ClusterIDs   []int64    `json:"cluster_ids" gorm:"-"`
	IsSuperAdmin bool       `json:"is_super_admin" gorm:"-"`
	CreatedAt    uint32     `json:"created_at"`
	ModifiedAt   uint32     `json:"modified_at"`
}

// RoleWithPermissions 带权限的角色
type RoleWithPermissions struct {
	SysRole
	Permissions []*SysPermission `json:"permissions" gorm:"-"`
	UserCount   int64            `json:"user_count" gorm:"-"`
}

// ClusterPermissionDetail 集群权限详情
type ClusterPermissionDetail struct {
	SysUserCluster
	ClusterName string   `json:"cluster_name" gorm:"-"`
	Username    string   `json:"username" gorm:"-"`
	NsList      []string `json:"ns_list" gorm:"-"`
}
