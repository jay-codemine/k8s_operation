package rbac

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type RBACController struct{}

func NewRBACController() *RBACController {
	return &RBACController{}
}

// ==================== 角色管理 ====================

// RoleList godoc
// @Summary 获取角色列表
// @Description 获取角色列表（分页）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param name query string false "角色名称"
// @Param role_type query string false "角色类型"
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/list [get]
func (c *RBACController) RoleList(ctx *gin.Context) {
	param := requests.NewRoleListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidRoleListRequest); !ok {
		return
	}

	svc := services.NewServices()
	roles, total, err := svc.RoleList(param)
	if err != nil {
		global.Logger.Error("获取角色列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleListFail)
		return
	}

	resp.SuccessList(roles, int(total))
}

// RoleListAll godoc
// @Summary 获取所有角色
// @Description 获取所有角色（不分页，用于下拉选择）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/all [get]
func (c *RBACController) RoleListAll(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := services.NewServices()

	roles, err := svc.RoleListAll()
	if err != nil {
		global.Logger.Error("获取角色列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleListFail)
		return
	}

	resp.SuccessList(roles, len(roles))
}

// RoleDetail godoc
// @Summary 获取角色详情
// @Description 获取角色详情（包含权限列表）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "角色ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/detail [get]
func (c *RBACController) RoleDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	role, err := svc.RoleGetByID(id)
	if err != nil {
		global.Logger.Error("获取角色详情失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleNotFound)
		return
	}

	resp.Success(role)
}

// RoleCreate godoc
// @Summary 创建角色
// @Description 创建新角色
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.RoleCreateRequest true "角色信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/create [post]
func (c *RBACController) RoleCreate(ctx *gin.Context) {
	param := requests.NewRoleCreateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidRoleCreateRequest); !ok {
		return
	}

	svc := services.NewServices()
	role, err := svc.RoleCreate(param)
	if err != nil {
		global.Logger.Error("创建角色失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleCreateFail)
		return
	}

	resp.Success(role)
}

// RoleUpdate godoc
// @Summary 更新角色
// @Description 更新角色信息
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.RoleUpdateRequest true "角色信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/update [post]
func (c *RBACController) RoleUpdate(ctx *gin.Context) {
	param := requests.NewRoleUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidRoleUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.RoleUpdate(param); err != nil {
		global.Logger.Error("更新角色失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleUpdateFail)
		return
	}

	resp.Success("更新成功")
}

// RoleDelete godoc
// @Summary 删除角色
// @Description 删除角色（系统内置角色不可删除）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "角色ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/delete [post]
func (c *RBACController) RoleDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	if err := svc.RoleDelete(id); err != nil {
		global.Logger.Error("删除角色失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleDeleteFail)
		return
	}

	resp.Success("删除成功")
}

// RolePermissions godoc
// @Summary 获取角色权限
// @Description 获取角色关联的权限列表
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param role_id query int true "角色ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/permissions [get]
func (c *RBACController) RolePermissions(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	roleIDStr := ctx.Query("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil || roleID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	permissions, err := svc.RolePermissionList(roleID)
	if err != nil {
		global.Logger.Error("获取角色权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorPermissionListFail)
		return
	}

	resp.SuccessList(permissions, len(permissions))
}

// RolePermissionsUpdate godoc
// @Summary 更新角色权限
// @Description 更新角色关联的权限（覆盖原有权限）
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.RolePermissionsUpdateRequest true "角色权限信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/permissions/update [post]
func (c *RBACController) RolePermissionsUpdate(ctx *gin.Context) {
	param := requests.NewRolePermissionsUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidRolePermissionsUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.RolePermissionUpdate(param.RoleID, param.PermissionIDs); err != nil {
		global.Logger.Error("更新角色权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRoleUpdateFail)
		return
	}

	resp.Success("更新成功")
}

// RoleUsers godoc
// @Summary 获取角色绑定用户
// @Description 获取角色绑定的用户列表
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param role_id query int true "角色ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/role/users [get]
func (c *RBACController) RoleUsers(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	roleIDStr := ctx.Query("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil || roleID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	users, err := svc.RoleUserList(roleID)
	if err != nil {
		global.Logger.Error("获取角色用户失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUserListFail)
		return
	}

	resp.SuccessList(users, len(users))
}

// ==================== 权限管理 ====================

// PermissionList godoc
// @Summary 获取权限列表
// @Description 获取所有权限定义
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/permission/list [get]
func (c *RBACController) PermissionList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := services.NewServices()

	permissions, err := svc.PermissionList()
	if err != nil {
		global.Logger.Error("获取权限列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorPermissionListFail)
		return
	}

	resp.SuccessList(permissions, len(permissions))
}

// ==================== 用户角色管理 ====================

// UserRoleAssign godoc
// @Summary 分配用户角色
// @Description 为用户分配角色（覆盖原有角色）
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.UserRoleAssignRequest true "分配信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user-role/assign [post]
func (c *RBACController) UserRoleAssign(ctx *gin.Context) {
	param := requests.NewUserRoleAssignRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidUserRoleAssignRequest); !ok {
		return
	}

	// 获取当前操作者ID（从JWT中获取）
	operatorID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		operatorID = uid.(int64)
	}

	svc := services.NewServices()
	if err := svc.UserRoleAssign(param, operatorID); err != nil {
		global.Logger.Error("分配用户角色失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUserRoleAssignFail)
		return
	}

	resp.Success("分配成功")
}

// UserRoleList godoc
// @Summary 获取用户角色
// @Description 获取用户的角色列表
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query int true "用户ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user-role/list [get]
func (c *RBACController) UserRoleList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	userIDStr := ctx.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	roles, err := svc.UserRoleList(userID)
	if err != nil {
		global.Logger.Error("获取用户角色失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUserRoleListFail)
		return
	}

	resp.SuccessList(roles, len(roles))
}

// ==================== 集群权限管理 ====================

// ClusterPermissionCreate godoc
// @Summary 创建集群权限
// @Description 为用户创建集群访问权限
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.ClusterPermissionCreateRequest true "权限信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/cluster-permission/create [post]
func (c *RBACController) ClusterPermissionCreate(ctx *gin.Context) {
	param := requests.NewClusterPermissionCreateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidClusterPermissionCreateRequest); !ok {
		return
	}

	operatorID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		operatorID = uid.(int64)
	}

	svc := services.NewServices()
	perm, err := svc.ClusterPermissionCreate(param, operatorID)
	if err != nil {
		global.Logger.Error("创建集群权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionCreateFail)
		return
	}

	resp.Success(perm)
}

// ClusterPermissionUpdate godoc
// @Summary 更新集群权限
// @Description 更新用户的集群访问权限
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.ClusterPermissionUpdateRequest true "权限信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/cluster-permission/update [post]
func (c *RBACController) ClusterPermissionUpdate(ctx *gin.Context) {
	param := requests.NewClusterPermissionUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidClusterPermissionUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.ClusterPermissionUpdate(param); err != nil {
		global.Logger.Error("更新集群权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionUpdateFail)
		return
	}

	resp.Success("更新成功")
}

// ClusterPermissionDelete godoc
// @Summary 删除集群权限
// @Description 删除用户的集群访问权限
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "权限ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/cluster-permission/delete [post]
func (c *RBACController) ClusterPermissionDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	if err := svc.ClusterPermissionDelete(id); err != nil {
		global.Logger.Error("删除集群权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionDeleteFail)
		return
	}

	resp.Success("删除成功")
}

// ClusterPermissionList godoc
// @Summary 获取集群权限列表
// @Description 获取集群权限列表（支持按用户或集群筛选）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query int false "用户ID"
// @Param cluster_id query int false "集群ID"
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/cluster-permission/list [get]
func (c *RBACController) ClusterPermissionList(ctx *gin.Context) {
	param := requests.NewClusterPermissionListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidClusterPermissionListRequest); !ok {
		return
	}

	svc := services.NewServices()
	permissions, total, err := svc.ClusterPermissionList(param)
	if err != nil {
		global.Logger.Error("获取集群权限列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionListFail)
		return
	}

	resp.SuccessList(permissions, int(total))
}

// BatchClusterPermissionCreate godoc
// @Summary 批量分配集群权限
// @Description 为用户批量分配多个集群的权限
// @Tags RBAC权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.BatchClusterPermissionRequest true "权限信息"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/cluster-permission/batch [post]
func (c *RBACController) BatchClusterPermissionCreate(ctx *gin.Context) {
	param := requests.NewBatchClusterPermissionRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidBatchClusterPermissionRequest); !ok {
		return
	}

	operatorID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		operatorID = uid.(int64)
	}

	svc := services.NewServices()
	if err := svc.BatchClusterPermissionCreate(param, operatorID); err != nil {
		global.Logger.Error("批量分配集群权限失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionCreateFail)
		return
	}

	resp.Success("批量分配成功")
}

// ==================== 用户完整信息 ====================

// UserRBACInfo godoc
// @Summary 获取用户RBAC信息
// @Description 获取用户完整的RBAC信息（角色+集群权限）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query int true "用户ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user/info [get]
func (c *RBACController) UserRBACInfo(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	userIDStr := ctx.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	info, err := svc.GetUserWithRBACInfo(userID)
	if err != nil {
		global.Logger.Error("获取用户RBAC信息失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUserRoleListFail)
		return
	}

	resp.Success(info)
}

// UserAccessibleClusters godoc
// @Summary 获取用户可访问的集群
// @Description 获取当前用户可访问的集群列表
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user/clusters [get]
func (c *RBACController) UserAccessibleClusters(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	// 从JWT获取当前用户ID
	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}
	if userID <= 0 {
		resp.ToErrorResponse(errorcode.ErrorRBACAccessDenied)
		return
	}

	svc := services.NewServices()
	clusters, err := svc.GetUserAccessibleClusters(userID)
	if err != nil {
		global.Logger.Error("获取用户可访问集群失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionListFail)
		return
	}

	resp.SuccessList(clusters, len(clusters))
}

// CheckPermission godoc
// @Summary 检查权限
// @Description 检查用户是否有指定集群的操作权限
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param cluster_id query int true "集群ID"
// @Param action query string true "操作类型(view/create/update/delete/exec)"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/check [get]
func (c *RBACController) CheckPermission(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	clusterIDStr := ctx.Query("cluster_id")
	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil || clusterID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	action := ctx.Query("action")
	if action == "" {
		action = "view"
	}

	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}

	svc := services.NewServices()
	hasPermission := svc.CheckClusterPermission(userID, clusterID, action)

	resp.Success(gin.H{
		"has_permission": hasPermission,
		"cluster_id":     clusterID,
		"action":         action,
	})
}

// UserPermissions godoc
// @Summary 获取当前用户完整权限
// @Description 获取当前用户的完整权限信息（权限隔离用）
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user/permissions [get]
func (c *RBACController) UserPermissions(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	// 从jwt获取当前用户ID
	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}
	if userID <= 0 {
		resp.ToErrorResponse(errorcode.ErrorRBACAccessDenied)
		return
	}

	svc := services.NewServices()
	
	// 获取用户完整RBAC信息
	userInfo, err := svc.GetUserWithRBACInfo(userID)
	if err != nil {
		global.Logger.Error("获取用户权限信息失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUserRoleListFail)
		return
	}

	// 返回扁平结构（前端期望格式，含 v2 三域 scopes）
	resp.Success(gin.H{
		"user_id":             userInfo.UserID,
		"username":            userInfo.Username,
		"is_super_admin":      userInfo.IsSuperAdmin,
		"roles":               userInfo.Roles,
		"cluster_permissions": userInfo.ClusterPermissions,
		"scopes":              userInfo.Scopes,
	})
}

// UserAccessibleNamespaces godoc
// @Summary 获取用户可访问的命名空间
// @Description 获取当前用户在指定集群可访问的命名空间列表
// @Tags RBAC权限管理
// @Produce json
// @Security ApiKeyAuth
// @Param cluster_id query int true "集群ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/rbac/user/namespaces [get]
func (c *RBACController) UserAccessibleNamespaces(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	// 从jwt获取当前用户ID
	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}
	if userID <= 0 {
		resp.ToErrorResponse(errorcode.ErrorRBACAccessDenied)
		return
	}

	// 获取集群ID
	clusterIDStr := ctx.Query("cluster_id")
	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil || clusterID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	namespaces, err := svc.GetUserAccessibleNamespaces(userID, clusterID)
	if err != nil {
		global.Logger.Error("获取用户可访问命名空间失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorClusterPermissionListFail)
		return
	}

	resp.Success(gin.H{
		"namespaces": namespaces,
		"cluster_id": clusterID,
	})
}
