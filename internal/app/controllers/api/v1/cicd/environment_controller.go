package cicd

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type EnvironmentController struct {
}

func NewEnvironmentController() *EnvironmentController {
	return &EnvironmentController{}
}

// List godoc
// @Summary 获取环境列表
// @Description 分页查询部署环境列表
// @Tags CICD Environment
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20"
// @Param keyword query string false "关键字搜索"
// @Success 200 {object} map[string]any "返回环境列表"
// @Router /api/v1/k8s/cicd/environment/list [get]
func (c *EnvironmentController) List(ctx *gin.Context) {
	param := requests.NewEnvironmentListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidEnvironmentListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.EnvironmentList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("EnvironmentList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取环境详情
// @Description 获取单个部署环境的详细信息
// @Tags CICD Environment
// @Produce json
// @Param id query int true "环境ID"
// @Success 200 {object} map[string]any "返回环境详情"
// @Router /api/v1/k8s/cicd/environment/detail [get]
func (c *EnvironmentController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的环境ID"))
		return
	}

	svc := services.NewServices()
	env, err := svc.EnvironmentDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("EnvironmentDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"environment": env})
}

// Create godoc
// @Summary 创建环境
// @Description 创建新的部署环境配置
// @Tags CICD Environment
// @Accept json
// @Produce json
// @Param body body requests.EnvironmentCreateRequest true "创建参数"
// @Success 200 {object} map[string]any "返回环境ID"
// @Router /api/v1/k8s/cicd/environment/create [post]
func (c *EnvironmentController) Create(ctx *gin.Context) {
	param := &requests.EnvironmentCreateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidEnvironmentCreateRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.EnvironmentCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("EnvironmentCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"environment_id": id})
}

// Update godoc
// @Summary 更新环境
// @Description 更新部署环境配置
// @Tags CICD Environment
// @Accept json
// @Produce json
// @Param body body requests.EnvironmentUpdateRequest true "更新参数"
// @Success 200 {object} map[string]any "更新成功"
// @Router /api/v1/k8s/cicd/environment/update [post]
func (c *EnvironmentController) Update(ctx *gin.Context) {
	param := &requests.EnvironmentUpdateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidEnvironmentUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.EnvironmentUpdate(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("EnvironmentUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// Delete godoc
// @Summary 删除环境
// @Description 删除部署环境配置
// @Tags CICD Environment
// @Accept json
// @Produce json
// @Param body body requests.EnvironmentIDRequest true "删除参数"
// @Success 200 {object} map[string]any "删除成功"
// @Router /api/v1/k8s/cicd/environment/delete [post]
func (c *EnvironmentController) Delete(ctx *gin.Context) {
	param := &requests.EnvironmentIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidEnvironmentIDRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.EnvironmentDelete(ctx.Request.Context(), param.ID)
	if err != nil {
		global.Logger.Error("EnvironmentDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}

// ==================== 审批流程 ====================

type ApprovalController struct {
}

func NewApprovalController() *ApprovalController {
	return &ApprovalController{}
}

// List godoc
// @Summary 获取审批列表
// @Description 分页查询审批记录
// @Tags CICD Approval
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param pipeline_id query int false "流水线ID筛选"
// @Param status query string false "状态筛选（pending/approved/rejected/expired）"
// @Success 200 {object} map[string]any "返回审批列表"
// @Router /api/v1/k8s/cicd/approval/list [get]
func (c *ApprovalController) List(ctx *gin.Context) {
	param := requests.NewApprovalListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalListRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()

	// 权限判断：有 cicd:approval:action 权限的是审批人，可看全部；否则只能看自己提交的
	canViewAll := models.HasUserPermission(global.DB, userID, "cicd:approval:action")

	var list []*models.ApprovalListItem
	var total int64
	var err error

	if canViewAll {
		list, total, err = svc.ApprovalList(ctx.Request.Context(), param)
	} else {
		list, total, err = svc.ApprovalListByUser(ctx.Request.Context(), userID, param)
	}
	if err != nil {
		global.Logger.Error("ApprovalList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"list":         list,
		"total":        total,
		"can_approve":  canViewAll,
	})
}

// Detail godoc
// @Summary 获取审批详情
// @Description 获取单个审批记录的详细信息
// @Tags CICD Approval
// @Produce json
// @Param id query int true "审批ID"
// @Success 200 {object} map[string]any "返回审批详情"
// @Router /api/v1/k8s/cicd/approval/detail [get]
func (c *ApprovalController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	svc := services.NewServices()
	approval, err := svc.ApprovalDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("ApprovalDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"approval": approval})
}

// Create godoc
// @Summary 创建审批申请
// @Description 创建部署审批申请
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param body body requests.ApprovalCreateRequest true "创建参数"
// @Success 200 {object} map[string]any "返回审批ID"
// @Router /api/v1/k8s/cicd/approval/create [post]
func (c *ApprovalController) Create(ctx *gin.Context) {
	param := &requests.ApprovalCreateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalCreateRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.ApprovalCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("ApprovalCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"approval_id": id})
}

// Action godoc
// @Summary 审批操作
// @Description 通过或拒绝审批申请
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param body body requests.ApprovalActionRequest true "操作参数"
// @Success 200 {object} map[string]any "操作成功"
// @Router /api/v1/k8s/cicd/approval/action [post]
func (c *ApprovalController) Action(ctx *gin.Context) {
	param := &requests.ApprovalActionRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalActionRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	err := svc.ApprovalAction(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("ApprovalAction error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	message := "审批通过"
	if param.Action == "reject" {
		message = "审批已拒绝"
	}

	rsp.Success(gin.H{"message": message})
}

// BatchAction godoc
// @Summary 批量审批操作
// @Description 批量通过或拒绝审批申请
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param body body requests.ApprovalBatchActionRequest true "批量操作参数"
// @Success 200 {object} map[string]any "操作结果"
// @Router /api/v1/k8s/cicd/approval/batch-action [post]
func (c *ApprovalController) BatchAction(ctx *gin.Context) {
	param := &requests.ApprovalBatchActionRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalBatchActionRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("审批ID列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	success, failures := svc.ApprovalBatchAction(ctx.Request.Context(), param.IDs, param.Action, param.Reason, userID)

	rsp.Success(gin.H{
		"success":  success,
		"failures": failures,
		"total":    len(param.IDs),
	})
}

// Update godoc
// @Summary 更新审批记录
// @Description 更新待审批记录的字段信息
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param body body requests.ApprovalUpdateRequest true "更新参数"
// @Success 200 {object} map[string]any "更新成功"
// @Router /api/v1/k8s/cicd/approval/update [post]
func (c *ApprovalController) Update(ctx *gin.Context) {
	param := &requests.ApprovalUpdateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.ApprovalUpdate(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("ApprovalUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// Delete godoc
// @Summary 删除审批记录
// @Description 删除审批记录（已通过的不允许删除）
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param body body requests.ApprovalDeleteRequest true "删除参数"
// @Success 200 {object} map[string]any "删除成功"
// @Router /api/v1/k8s/cicd/approval/delete [post]
func (c *ApprovalController) Delete(ctx *gin.Context) {
	param := &requests.ApprovalDeleteRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidApprovalDeleteRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.ApprovalDelete(ctx.Request.Context(), param.ID)
	if err != nil {
		global.Logger.Error("ApprovalDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}

// Pending godoc
// @Summary 获取待审批列表
// @Description 获取当前用户待审批的申请列表
// @Tags CICD Approval
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Success 200 {object} map[string]any "返回待审批列表"
// @Router /api/v1/k8s/cicd/approval/pending [get]
func (c *ApprovalController) Pending(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	list, total, err := svc.ApprovalPendingList(ctx.Request.Context(), userID)
	if err != nil {
		global.Logger.Error("ApprovalPendingList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Stats godoc
// @Summary 获取审批统计
// @Description 获取各状态审批数量统计
// @Tags CICD Approval
// @Produce json
// @Success 200 {object} map[string]any "返回统计数据"
// @Router /api/v1/k8s/cicd/approval/stats [get]
func (c *ApprovalController) Stats(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()

	// 权限判断：有审批操作权限的看全局统计，否则只看自己的
	canViewAll := models.HasUserPermission(global.DB, userID, "cicd:approval:action")

	var stats map[string]int64
	var err error
	if canViewAll {
		stats, err = svc.ApprovalStats(ctx.Request.Context())
	} else {
		stats, err = svc.ApprovalStatsByUser(ctx.Request.Context(), userID)
	}
	if err != nil {
		global.Logger.Error("ApprovalStats error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"stats": stats})
}

// FeishuCallback godoc
// @Summary 飞书审批回调
// @Description 飞书交互卡片审批回调接口（公开接口，无需JWT）
// @Tags CICD Approval
// @Accept json
// @Produce json
// @Param token query string true "审批Token"
// @Param action query string true "审批动作: approve/reject"
// @Param reason query string false "审批意见"
// @Success 200 {object} map[string]any "审批处理成功"
// @Router /api/v1/k8s/cicd/approval/feishu-callback [get]
func (c *ApprovalController) FeishuCallback(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	// 支持 GET 参数（飞书卡片按钮跳转链接）和 POST JSON
	req := &services.FeishuApprovalCallbackRequest{
		Token:  ctx.Query("token"),
		Action: ctx.Query("action"),
		Reason: ctx.Query("reason"),
	}

	// 如果 GET 参数为空，尝试从 POST body 解析
	if req.Token == "" {
		_ = ctx.ShouldBindJSON(req)
	}

	if req.Token == "" || req.Action == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("缺少必要参数: token 和 action"))
		return
	}

	svc := services.NewServices()
	err := svc.HandleFeishuApprovalCallback(ctx.Request.Context(), req)
	if err != nil {
		global.Logger.Error("FeishuCallback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	message := "审批通过，部署已触发"
	if req.Action == "reject" {
		message = "审批已拒绝"
	}

	rsp.Success(gin.H{"message": message})
}
