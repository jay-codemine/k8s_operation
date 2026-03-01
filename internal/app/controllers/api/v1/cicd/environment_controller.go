package cicd

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

	svc := services.NewServices()
	list, total, err := svc.ApprovalList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("ApprovalList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
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
