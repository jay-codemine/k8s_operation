package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ==================== 角色管理 ====================

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Name          string `json:"name" form:"name" valid:"name"`
	DisplayName   string `json:"display_name" form:"display_name" valid:"display_name"`
	Description   string `json:"description" form:"description" valid:"description"`
	RoleType      string `json:"role_type" form:"role_type" valid:"role_type"`
	ScopePlatform string `json:"scope_platform" form:"scope_platform"` // none/read/write/admin
	ScopeCluster  string `json:"scope_cluster" form:"scope_cluster"`   // none/read/write/admin
	ScopeCICD     string `json:"scope_cicd" form:"scope_cicd"`         // none/read/write/admin
	Color         string `json:"color" form:"color" valid:"color"`
	Icon          string `json:"icon" form:"icon" valid:"icon"`
}

func NewRoleCreateRequest() *RoleCreateRequest {
	return &RoleCreateRequest{
		Color:         "#1890ff",
		Icon:          "user",
		ScopePlatform: "none",
		ScopeCluster:  "none",
		ScopeCICD:     "none",
	}
}

func ValidRoleCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":         []string{"required", "min:2", "max:50"},
		"display_name": []string{"required", "min:2", "max:50"},
		"role_type":    []string{"required", "in:super_admin,platform_admin,devops,developer,tester,viewer,custom"},
	}
	messages := govalidator.MapData{
		"name":         []string{"required:角色标识为必填项", "min:角色标识至少2个字符", "max:角色标识最多50个字符"},
		"display_name": []string{"required:角色名称为必填项", "min:角色名称至少2个字符", "max:角色名称最多50个字符"},
		"role_type":    []string{"required:角色类型为必填项", "in:角色类型无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	ID            int64  `json:"id" form:"id" valid:"id"`
	DisplayName   string `json:"display_name" form:"display_name" valid:"display_name"`
	Description   string `json:"description" form:"description" valid:"description"`
	ScopePlatform string `json:"scope_platform" form:"scope_platform"` // none/read/write/admin
	ScopeCluster  string `json:"scope_cluster" form:"scope_cluster"`
	ScopeCICD     string `json:"scope_cicd" form:"scope_cicd"`
	Color         string `json:"color" form:"color" valid:"color"`
	Icon          string `json:"icon" form:"icon" valid:"icon"`
	SortOrder     int    `json:"sort_order" form:"sort_order" valid:"sort_order"`
}

func NewRoleUpdateRequest() *RoleUpdateRequest {
	return &RoleUpdateRequest{}
}

func ValidRoleUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":           []string{"required"},
		"display_name": []string{"required", "min:2", "max:50"},
	}
	messages := govalidator.MapData{
		"id":           []string{"required:角色ID为必填项"},
		"display_name": []string{"required:角色名称为必填项", "min:角色名称至少2个字符", "max:角色名称最多50个字符"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// RoleListRequest 角色列表请求
type RoleListRequest struct {
	Name     string `json:"name,omitempty" form:"name"`
	RoleType string `json:"role_type,omitempty" form:"role_type"`
	Page     int    `json:"page,omitempty" form:"page" valid:"page"`
	Limit    int    `json:"limit,omitempty" form:"limit" valid:"limit"`
}

func NewRoleListRequest() *RoleListRequest {
	return &RoleListRequest{Page: 1, Limit: 20}
}

func ValidRoleListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required"},
		"limit": []string{"required"},
	}
	messages := govalidator.MapData{
		"page":  []string{"required:页码为必填项"},
		"limit": []string{"required:每页数量为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// RolePermissionsUpdateRequest 更新角色权限请求
type RolePermissionsUpdateRequest struct {
	RoleID        int64   `json:"role_id" form:"role_id" valid:"role_id"`
	PermissionIDs []int64 `json:"permission_ids" form:"permission_ids" valid:"permission_ids"`
}

func NewRolePermissionsUpdateRequest() *RolePermissionsUpdateRequest {
	return &RolePermissionsUpdateRequest{}
}

func ValidRolePermissionsUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"role_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"role_id": []string{"required:角色ID为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 用户角色分配 ====================

// UserRoleAssignRequest 分配用户角色请求
type UserRoleAssignRequest struct {
	UserID  int64   `json:"user_id" form:"user_id" valid:"user_id"`
	RoleIDs []int64 `json:"role_ids" form:"role_ids" valid:"role_ids"`
}

func NewUserRoleAssignRequest() *UserRoleAssignRequest {
	return &UserRoleAssignRequest{}
}

func ValidUserRoleAssignRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"user_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"user_id": []string{"required:用户ID为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// UserRoleListRequest 获取用户角色请求
type UserRoleListRequest struct {
	UserID int64 `json:"user_id,omitempty" form:"user_id"`
	Page   int   `json:"page,omitempty" form:"page" valid:"page"`
	Limit  int   `json:"limit,omitempty" form:"limit" valid:"limit"`
}

func NewUserRoleListRequest() *UserRoleListRequest {
	return &UserRoleListRequest{Page: 1, Limit: 20}
}

func ValidUserRoleListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required"},
		"limit": []string{"required"},
	}
	messages := govalidator.MapData{
		"page":  []string{"required:页码为必填项"},
		"limit": []string{"required:每页数量为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 集群权限管理 ====================

// ClusterPermissionCreateRequest 创建集群权限请求
type ClusterPermissionCreateRequest struct {
	UserID      int64    `json:"user_id" form:"user_id" valid:"user_id"`
	ClusterID   int64    `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	RoleType    string   `json:"role_type" form:"role_type" valid:"role_type"`
	AccessLevel string   `json:"access_level" form:"access_level"` // none/read/write/admin
	Namespaces  []string `json:"namespaces" form:"namespaces"`     // 可访问的命名空间（空=全部）
	CanView     bool     `json:"can_view" form:"can_view"`         // deprecated
	CanCreate   bool     `json:"can_create" form:"can_create"`     // deprecated
	CanUpdate   bool     `json:"can_update" form:"can_update"`     // deprecated
	CanDelete   bool     `json:"can_delete" form:"can_delete"`     // deprecated
	CanExec     bool     `json:"can_exec" form:"can_exec"`         // deprecated
	ExpireAt    uint64   `json:"expire_at" form:"expire_at"`       // 过期时间，0=永不过期
}

func NewClusterPermissionCreateRequest() *ClusterPermissionCreateRequest {
	return &ClusterPermissionCreateRequest{
		CanView:     true,
		AccessLevel: "read",
	}
}

func ValidClusterPermissionCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"user_id":    []string{"required"},
		"cluster_id": []string{"required"},
		"role_type":  []string{"required", "in:super_admin,platform_admin,devops,developer,tester,viewer,custom"},
	}
	messages := govalidator.MapData{
		"user_id":    []string{"required:用户ID为必填项"},
		"cluster_id": []string{"required:集群ID为必填项"},
		"role_type":  []string{"required:角色类型为必填项", "in:角色类型无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ClusterPermissionUpdateRequest 更新集群权限请求
type ClusterPermissionUpdateRequest struct {
	ID          int64    `json:"id" form:"id" valid:"id"`
	RoleType    string   `json:"role_type" form:"role_type" valid:"role_type"`
	AccessLevel string   `json:"access_level" form:"access_level"` // none/read/write/admin
	Namespaces  []string `json:"namespaces" form:"namespaces"`
	CanView     bool     `json:"can_view" form:"can_view"`     // deprecated
	CanCreate   bool     `json:"can_create" form:"can_create"` // deprecated
	CanUpdate   bool     `json:"can_update" form:"can_update"` // deprecated
	CanDelete   bool     `json:"can_delete" form:"can_delete"` // deprecated
	CanExec     bool     `json:"can_exec" form:"can_exec"`     // deprecated
	ExpireAt    uint64   `json:"expire_at" form:"expire_at"`
}

func NewClusterPermissionUpdateRequest() *ClusterPermissionUpdateRequest {
	return &ClusterPermissionUpdateRequest{}
}

func ValidClusterPermissionUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":        []string{"required"},
		"role_type": []string{"required", "in:super_admin,platform_admin,devops,developer,tester,viewer,custom"},
	}
	messages := govalidator.MapData{
		"id":        []string{"required:权限ID为必填项"},
		"role_type": []string{"required:角色类型为必填项", "in:角色类型无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ClusterPermissionListRequest 集群权限列表请求
type ClusterPermissionListRequest struct {
	UserID    int64 `json:"user_id,omitempty" form:"user_id"`
	ClusterID int64 `json:"cluster_id,omitempty" form:"cluster_id"`
	Page      int   `json:"page,omitempty" form:"page" valid:"page"`
	Limit     int   `json:"limit,omitempty" form:"limit" valid:"limit"`
}

func NewClusterPermissionListRequest() *ClusterPermissionListRequest {
	return &ClusterPermissionListRequest{Page: 1, Limit: 20}
}

func ValidClusterPermissionListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required"},
		"limit": []string{"required"},
	}
	messages := govalidator.MapData{
		"page":  []string{"required:页码为必填项"},
		"limit": []string{"required:每页数量为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 权限检查 ====================

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID    int64  `json:"user_id" form:"user_id"`
	ClusterID int64  `json:"cluster_id" form:"cluster_id"`
	Action    string `json:"action" form:"action"` // view/create/update/delete/exec
	Namespace string `json:"namespace" form:"namespace"`
}

func NewPermissionCheckRequest() *PermissionCheckRequest {
	return &PermissionCheckRequest{}
}

// ==================== 用户完整信息请求 ====================

// UserDetailRequest 获取用户详情请求（包含角色和集群权限）
type UserDetailRequest struct {
	UserID int64 `json:"user_id" form:"user_id" valid:"user_id"`
}

func NewUserDetailRequest() *UserDetailRequest {
	return &UserDetailRequest{}
}

func ValidUserDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"user_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"user_id": []string{"required:用户ID为必填项"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 批量集群权限分配 ====================

// BatchClusterPermissionRequest 批量分配集群权限请求
type BatchClusterPermissionRequest struct {
	UserID      int64   `json:"user_id" form:"user_id" valid:"user_id"`
	ClusterIDs  []int64 `json:"cluster_ids" form:"cluster_ids"`
	RoleType    string  `json:"role_type" form:"role_type" valid:"role_type"`
	AccessLevel string  `json:"access_level" form:"access_level"` // none/read/write/admin
	CanView     bool    `json:"can_view" form:"can_view"`         // deprecated
	CanCreate   bool    `json:"can_create" form:"can_create"`     // deprecated
	CanUpdate   bool    `json:"can_update" form:"can_update"`     // deprecated
	CanDelete   bool    `json:"can_delete" form:"can_delete"`     // deprecated
	CanExec     bool    `json:"can_exec" form:"can_exec"`         // deprecated
}

func NewBatchClusterPermissionRequest() *BatchClusterPermissionRequest {
	return &BatchClusterPermissionRequest{CanView: true, AccessLevel: "read"}
}

func ValidBatchClusterPermissionRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"user_id":   []string{"required"},
		"role_type": []string{"required", "in:super_admin,platform_admin,devops,developer,tester,viewer,custom"},
	}
	messages := govalidator.MapData{
		"user_id":   []string{"required:用户ID为必填项"},
		"role_type": []string{"required:角色类型为必填项", "in:角色类型无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}
