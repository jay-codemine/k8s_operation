package k8srbac

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8soperation/global"
	"k8soperation/internal/app/services/k8srbac"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
)

type K8sRBACController struct{}

func NewK8sRBACController() *K8sRBACController {
	return &K8sRBACController{}
}

// ==================== ServiceAccount ====================

// ListServiceAccounts godoc
// @Summary 获取 ServiceAccount 列表
// @Description 获取 K8s ServiceAccount 列表
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param namespace query string false "命名空间（空表示所有）"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/serviceaccounts [get]
func (c *K8sRBACController) ListServiceAccounts(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	namespace := ctx.Query("namespace")

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	items, err := svc.ListServiceAccounts(ctx.Request.Context(), cli.Kube, namespace)
	if err != nil {
		global.Logger.Error("获取 ServiceAccount 列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceListFail)
		return
	}

	resp.SuccessList(items, len(items))
}

// GetServiceAccount godoc
// @Summary 获取 ServiceAccount 详情
// @Description 获取指定 ServiceAccount 详情
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param namespace query string true "命名空间"
// @Param name query string true "名称"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/serviceaccount [get]
func (c *K8sRBACController) GetServiceAccount(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if namespace == "" || name == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	item, err := svc.GetServiceAccount(ctx.Request.Context(), cli.Kube, namespace, name)
	if err != nil {
		global.Logger.Error("获取 ServiceAccount 详情失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceGetFail)
		return
	}

	resp.Success(item)
}

// CreateServiceAccount godoc
// @Summary 创建 ServiceAccount
// @Description 创建 K8s ServiceAccount
// @Tags K8s RBAC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body object true "创建参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/serviceaccount [post]
func (c *K8sRBACController) CreateServiceAccount(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req struct {
		Name           string            `json:"name" binding:"required"`
		Namespace      string            `json:"namespace" binding:"required"`
		Labels         map[string]string `json:"labels"`
		AutoMountToken bool              `json:"auto_mount_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	_, err := svc.CreateServiceAccount(ctx.Request.Context(), cli.Kube, req.Namespace, req.Name, req.Labels, req.AutoMountToken)
	if err != nil {
		global.Logger.Error("创建 ServiceAccount 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceCreateFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "创建成功"})
}

// DeleteServiceAccount godoc
// @Summary 删除 ServiceAccount
// @Description 删除 K8s ServiceAccount
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param namespace query string true "命名空间"
// @Param name query string true "名称"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/serviceaccount [delete]
func (c *K8sRBACController) DeleteServiceAccount(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if namespace == "" || name == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	err := svc.DeleteServiceAccount(ctx.Request.Context(), cli.Kube, namespace, name)
	if err != nil {
		global.Logger.Error("删除 ServiceAccount 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceDeleteFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// ==================== Role / ClusterRole ====================

// ListRoles godoc
// @Summary 获取 Role 列表
// @Description 获取 K8s Role 和 ClusterRole 列表
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param namespace query string false "命名空间（空表示所有）"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/roles [get]
func (c *K8sRBACController) ListRoles(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	namespace := ctx.Query("namespace")

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	items, err := svc.ListRoles(ctx.Request.Context(), cli.Kube, namespace)
	if err != nil {
		global.Logger.Error("获取 Role 列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceListFail)
		return
	}

	resp.SuccessList(items, len(items))
}

// CreateRole godoc
// @Summary 创建 Role
// @Description 创建 K8s Role 或 ClusterRole
// @Tags K8s RBAC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body object true "创建参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/role [post]
func (c *K8sRBACController) CreateRole(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req struct {
		Name      string              `json:"name" binding:"required"`
		Type      string              `json:"type" binding:"required"` // "Role" or "ClusterRole"
		Namespace string              `json:"namespace"`
		Rules     []rbacv1.PolicyRule `json:"rules" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if req.Type == "Role" && req.Namespace == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("Role 需要指定命名空间"))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	err := svc.CreateRole(ctx.Request.Context(), cli.Kube, req.Type, req.Namespace, req.Name, req.Rules)
	if err != nil {
		global.Logger.Error("创建 Role 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceCreateFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "创建成功"})
}

// DeleteRole godoc
// @Summary 删除 Role
// @Description 删除 K8s Role 或 ClusterRole
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param type query string true "类型（Role/ClusterRole）"
// @Param namespace query string false "命名空间（Role 必填）"
// @Param name query string true "名称"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/role [delete]
func (c *K8sRBACController) DeleteRole(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	roleType := ctx.Query("type")
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if name == "" || roleType == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	err := svc.DeleteRole(ctx.Request.Context(), cli.Kube, roleType, namespace, name)
	if err != nil {
		global.Logger.Error("删除 Role 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceDeleteFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// ==================== RoleBinding / ClusterRoleBinding ====================

// ListRoleBindings godoc
// @Summary 获取 RoleBinding 列表
// @Description 获取 K8s RoleBinding 和 ClusterRoleBinding 列表
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param namespace query string false "命名空间（空表示所有）"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/rolebindings [get]
func (c *K8sRBACController) ListRoleBindings(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	namespace := ctx.Query("namespace")

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	items, err := svc.ListRoleBindings(ctx.Request.Context(), cli.Kube, namespace)
	if err != nil {
		global.Logger.Error("获取 RoleBinding 列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceListFail)
		return
	}

	resp.SuccessList(items, len(items))
}

// CreateRoleBinding godoc
// @Summary 创建 RoleBinding
// @Description 创建 K8s RoleBinding 或 ClusterRoleBinding
// @Tags K8s RBAC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body object true "创建参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/rolebinding [post]
func (c *K8sRBACController) CreateRoleBinding(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req struct {
		Name      string             `json:"name" binding:"required"`
		Type      string             `json:"type" binding:"required"` // "RoleBinding" or "ClusterRoleBinding"
		Namespace string             `json:"namespace"`
		RoleRef   k8srbac.RoleRef    `json:"role_ref" binding:"required"`
		Subjects  []rbacv1.Subject   `json:"subjects" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if req.Type == "RoleBinding" && req.Namespace == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("RoleBinding 需要指定命名空间"))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	err := svc.CreateRoleBinding(ctx.Request.Context(), cli.Kube, req.Type, req.Namespace, req.Name, req.RoleRef, req.Subjects)
	if err != nil {
		global.Logger.Error("创建 RoleBinding 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceCreateFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "创建成功"})
}

// DeleteRoleBinding godoc
// @Summary 删除 RoleBinding
// @Description 删除 K8s RoleBinding 或 ClusterRoleBinding
// @Tags K8s RBAC
// @Produce json
// @Security ApiKeyAuth
// @Param type query string true "类型（RoleBinding/ClusterRoleBinding）"
// @Param namespace query string false "命名空间（RoleBinding 必填）"
// @Param name query string true "名称"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/rolebinding [delete]
func (c *K8sRBACController) DeleteRoleBinding(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	bindingType := ctx.Query("type")
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if name == "" || bindingType == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	err := svc.DeleteRoleBinding(ctx.Request.Context(), cli.Kube, bindingType, namespace, name)
	if err != nil {
		global.Logger.Error("删除 RoleBinding 失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorK8sResourceDeleteFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// ==================== SubjectAccessReview ====================

// CheckSubjectAccess godoc
// @Summary 检查主体权限
// @Description 检查用户或 ServiceAccount 对指定资源的访问权限
// @Tags K8s RBAC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body object true "检查参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/check [post]
func (c *K8sRBACController) CheckSubjectAccess(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req k8srbac.SubjectAccessCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if req.Resource == "" || req.Verb == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("资源类型和操作不能为空"))
		return
	}

	if req.SubjectType == "User" && req.Username == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("用户类型需要指定用户名"))
		return
	}

	if req.SubjectType == "ServiceAccount" && (req.SANamespace == "" || req.SAName == "") {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("ServiceAccount 需要指定命名空间和名称"))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	result, err := svc.CheckSubjectAccess(ctx.Request.Context(), cli.Kube, req)
	if err != nil {
		global.Logger.Error("权限检查失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRBACAccessDenied.WithDetails(err.Error()))
		return
	}

	resp.Success(result)
}

// BatchCheckSubjectAccess godoc
// @Summary 批量检查主体权限
// @Description 批量检查用户或 ServiceAccount 对多个资源的访问权限
// @Tags K8s RBAC
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body object true "检查参数数组"
// @Success 200 {object} string "成功"
// @Router /api/v1/k8s/rbac/check/batch [post]
func (c *K8sRBACController) BatchCheckSubjectAccess(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req struct {
		Checks []k8srbac.SubjectAccessCheckRequest `json:"checks" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := k8srbac.NewK8sRBACService()

	results, err := svc.BatchCheckSubjectAccess(ctx.Request.Context(), cli.Kube, req.Checks)
	if err != nil {
		global.Logger.Error("批量权限检查失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorRBACAccessDenied.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(results, len(results))
}
