package models

import dm "k8soperation/internal/domain/rbac"

// ==================== 角色类型常量 ====================

const (
	RoleTypeSuperAdmin    = dm.RoleTypeSuperAdmin
	RoleTypePlatformAdmin = dm.RoleTypePlatformAdmin
	RoleTypeDevOps        = dm.RoleTypeDevOps
	RoleTypeDeveloper     = dm.RoleTypeDeveloper
	RoleTypeTester        = dm.RoleTypeTester
	RoleTypeViewer        = dm.RoleTypeViewer
	RoleTypeClusterAdmin  = dm.RoleTypeClusterAdmin
)

// ==================== 权限域常量 ====================

const (
	ScopePlatform = dm.ScopePlatform
	ScopeCluster  = dm.ScopeCluster
	ScopeCICD     = dm.ScopeCICD
)

// ==================== 权限级别常量 ====================

const (
	AccessLevelNone  = dm.AccessLevelNone
	AccessLevelRead  = dm.AccessLevelRead
	AccessLevelWrite = dm.AccessLevelWrite
	AccessLevelAdmin = dm.AccessLevelAdmin
)

// ==================== 权限操作常量 ====================

const (
	PermissionActionView   = dm.PermissionActionView
	PermissionActionCreate = dm.PermissionActionCreate
	PermissionActionUpdate = dm.PermissionActionUpdate
	PermissionActionDelete = dm.PermissionActionDelete
	PermissionActionExec   = dm.PermissionActionExec
	PermissionActionManage = dm.PermissionActionManage
)

// ==================== 资源类型/模块常量 ====================

const (
	ResourceTypeUser     = dm.ResourceTypeUser
	ResourceTypeRole     = dm.ResourceTypeRole
	ResourceTypeSettings = dm.ResourceTypeSettings
	ResourceTypeAudit    = dm.ResourceTypeAudit

	ResourceTypeCluster  = dm.ResourceTypeCluster
	ResourceTypeWorkload = dm.ResourceTypeWorkload
	ResourceTypeNetwork  = dm.ResourceTypeNetwork
	ResourceTypeConfig   = dm.ResourceTypeConfig
	ResourceTypeStorage  = dm.ResourceTypeStorage
	ResourceTypeNode     = dm.ResourceTypeNode
	ResourceTypeMonitor  = dm.ResourceTypeMonitor

	ResourceTypePipeline = dm.ResourceTypePipeline
	ResourceTypeArtifact = dm.ResourceTypeArtifact

	ResourceTypeNamespace  = dm.ResourceTypeNamespace
	ResourceTypeDeployment = dm.ResourceTypeDeployment
	ResourceTypePod        = dm.ResourceTypePod
	ResourceTypeService    = dm.ResourceTypeService
	ResourceTypeConfigMap  = dm.ResourceTypeConfigMap
	ResourceTypeSecret     = dm.ResourceTypeSecret
	ResourceTypePVC        = dm.ResourceTypePVC
	ResourceTypeIngress    = dm.ResourceTypeIngress
)

// ==================== 类型别名 ====================

type (
	SysRole                 = dm.SysRole
	SysPermission           = dm.SysPermission
	SysRolePermission       = dm.SysRolePermission
	SysUserRole             = dm.SysUserRole
	SysUserCluster          = dm.SysUserCluster
	UserWithRole            = dm.UserWithRole
	RoleWithPermissions     = dm.RoleWithPermissions
	ClusterPermissionDetail = dm.ClusterPermissionDetail
)

// ==================== 函数别名 ====================

var (
	AccessLevelGte            = dm.AccessLevelGte
	GetUserRoles              = dm.GetUserRoles
	GetUserClusterPermissions = dm.GetUserClusterPermissions
	HasClusterPermission      = dm.HasClusterPermission
	HasScopePermission        = dm.HasScopePermission
	IsSuperAdmin              = dm.IsSuperAdmin
	HasUserPermission         = dm.HasUserPermission
	GetUserPermissionNames    = dm.GetUserPermissionNames
)
