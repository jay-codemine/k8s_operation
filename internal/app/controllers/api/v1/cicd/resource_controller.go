package cicd

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// ResourceController 资源配置控制器
type ResourceController struct{}

// NewResourceController 创建控制器
func NewResourceController() *ResourceController {
	return &ResourceController{}
}

// ==================== 资源模板 ====================

// TemplateList godoc
// @Summary 获取资源模板列表
// @Description 根据环境和服务类型获取资源模板列表
// @Tags CICD Resource
// @Accept json
// @Produce json
// @Param env query string false "环境：dev/test/staging/prod"
// @Param service_type query string false "服务类型：java/go/node/python"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/templates [get]
func (c *ResourceController) TemplateList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	env := ctx.Query("env")
	serviceType := ctx.Query("service_type")

	svc := services.NewServices()
	list, err := svc.ResourceTemplateList(ctx.Request.Context(), env, serviceType)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"list":  list,
		"total": len(list),
	})
}

// TemplateDetail godoc
// @Summary 获取资源模板详情
// @Tags CICD Resource
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/template/{id} [get]
func (c *ResourceController) TemplateDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	tpl, err := svc.ResourceTemplateGetByID(ctx.Request.Context(), id)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(tpl)
}

// TemplateDefault godoc
// @Summary 获取默认资源模板
// @Tags CICD Resource
// @Param env query string true "环境"
// @Param service_type query string true "服务类型"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/template/default [get]
func (c *ResourceController) TemplateDefault(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	env := ctx.Query("env")
	serviceType := ctx.Query("service_type")

	if env == "" || serviceType == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("env和service_type必填"))
		return
	}

	svc := services.NewServices()
	tpl, err := svc.ResourceTemplateGetDefault(ctx.Request.Context(), env, serviceType)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(tpl)
}

// TemplateCreate godoc
// @Summary 创建资源模板
// @Tags CICD Resource
// @Accept json
// @Produce json
// @Param body body models.CicdResourceTemplate true "模板信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/template [post]
func (c *ResourceController) TemplateCreate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	var tpl models.CicdResourceTemplate
	if err := ctx.ShouldBindJSON(&tpl); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.ResourceTemplateCreate(ctx.Request.Context(), &tpl); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"id": tpl.ID})
}

// TemplateUpdate godoc
// @Summary 更新资源模板
// @Tags CICD Resource
// @Accept json
// @Param id path int true "模板ID"
// @Param body body models.CicdResourceTemplate true "模板信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/template/{id} [put]
func (c *ResourceController) TemplateUpdate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	var tpl models.CicdResourceTemplate
	if err := ctx.ShouldBindJSON(&tpl); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	tpl.ID = id

	svc := services.NewServices()
	if err := svc.ResourceTemplateUpdate(ctx.Request.Context(), &tpl); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(nil)
}

// TemplateDelete godoc
// @Summary 删除资源模板
// @Tags CICD Resource
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/template/{id} [delete]
func (c *ResourceController) TemplateDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	if err := svc.ResourceTemplateDelete(ctx.Request.Context(), id); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(nil)
}

// ==================== 环境规则 ====================

// RuleList godoc
// @Summary 获取环境规则列表
// @Tags CICD Resource
// @Param env query string false "环境"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/rules [get]
func (c *ResourceController) RuleList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	env := ctx.Query("env")

	svc := services.NewServices()
	list, err := svc.EnvResourceRuleList(ctx.Request.Context(), env)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"list":  list,
		"total": len(list),
	})
}

// RuleUpdate godoc
// @Summary 更新环境规则
// @Tags CICD Resource
// @Accept json
// @Param id path int true "规则ID"
// @Param body body models.CicdEnvResourceRule true "规则信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/rule/{id} [put]
func (c *ResourceController) RuleUpdate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	var rule models.CicdEnvResourceRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	rule.ID = id

	svc := services.NewServices()
	if err := svc.EnvResourceRuleUpdate(ctx.Request.Context(), &rule); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(nil)
}

// ==================== 资源校验 ====================

// Validate godoc
// @Summary 校验资源配置
// @Description 校验资源配置是否符合环境规则，返回警告和错误
// @Tags CICD Resource
// @Accept json
// @Produce json
// @Param body body requests.ResourceValidateRequest true "校验参数"
// @Success 200 {object} models.ResourceValidationResult
// @Router /api/v1/k8s/cicd/resource/validate [post]
func (c *ResourceController) Validate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	var req requests.ResourceValidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	result := svc.ValidateResourceConfig(ctx.Request.Context(), req.Env, req.ServiceType, req.Config)

	resp.Success(result)
}

// ==================== 发布审批 ====================

// ApprovalList godoc
// @Summary 获取审批列表
// @Tags CICD Resource
// @Param status query string false "状态：pending/approved/rejected"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/approvals [get]
func (c *ResourceController) ApprovalList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	svc := services.NewServices()
	list, total, err := svc.DeployApprovalList(ctx.Request.Context(), status, page, pageSize)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"list":  list,
		"total": total,
		"page":  page,
	})
}

// ApprovalDetail godoc
// @Summary 获取审批详情
// @Tags CICD Resource
// @Param id path int true "审批ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/approval/{id} [get]
func (c *ResourceController) ApprovalDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	svc := services.NewServices()
	approval, err := svc.DeployApprovalGetByID(ctx.Request.Context(), id)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(approval)
}

// ApprovalApprove godoc
// @Summary 通过审批
// @Tags CICD Resource
// @Accept json
// @Param id path int true "审批ID"
// @Param body body map[string]string true "审批意见"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/approval/{id}/approve [put]
func (c *ResourceController) ApprovalApprove(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	_ = ctx.ShouldBindJSON(&req)

	userID := ctx.GetInt64("user_id")
	userName := ctx.GetString("user_name")

	svc := services.NewServices()
	if err := svc.DeployApprovalApprove(ctx.Request.Context(), id, uint64(userID), userName, req.Comment); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(nil)
}

// ApprovalReject godoc
// @Summary 拒绝审批
// @Tags CICD Resource
// @Accept json
// @Param id path int true "审批ID"
// @Param body body map[string]string true "拒绝原因"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/cicd/resource/approval/{id}/reject [put]
func (c *ResourceController) ApprovalReject(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if id == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	_ = ctx.ShouldBindJSON(&req)

	userID := ctx.GetInt64("user_id")
	userName := ctx.GetString("user_name")

	svc := services.NewServices()
	if err := svc.DeployApprovalReject(ctx.Request.Context(), id, uint64(userID), userName, req.Comment); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(nil)
}
