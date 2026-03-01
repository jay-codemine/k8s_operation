package cicd

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// StageController 流水线阶段控制器
type StageController struct {
}

// NewStageController 创建阶段控制器
func NewStageController() *StageController {
	return &StageController{}
}

// GetStages 获取运行记录的阶段列表
// @Summary 获取流水线运行阶段列表
// @Tags CICD-Stage
// @Accept json
// @Produce json
// @Param run_id query int true "运行记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/k8s/cicd/stage/list [get]
func (c *StageController) GetStages(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	runIDStr := ctx.Query("run_id")
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil || runID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的运行记录ID"))
		return
	}

	svc := services.NewServices()
	stages, err := svc.GetRunStages(ctx.Request.Context(), runID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"stages": stages,
	})
}

// GetStageLogs 获取阶段日志
// @Summary 获取阶段执行日志
// @Tags CICD-Stage
// @Accept json
// @Produce json
// @Param id query int true "阶段ID"
// @Success 200 {object} response.Response
// @Router /api/v1/k8s/cicd/stage/logs [get]
func (c *StageController) GetStageLogs(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的阶段ID"))
		return
	}

	svc := services.NewServices()
	logs, err := svc.GetStageLogs(ctx.Request.Context(), id)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"logs": logs,
	})
}

// ApproveStage 审批阶段
// @Summary 审批通过/拒绝阶段
// @Tags CICD-Stage
// @Accept json
// @Produce json
// @Param body body requests.StageApproveRequest true "审批请求"
// @Success 200 {object} response.Response
// @Router /api/v1/k8s/cicd/stage/approve [post]
func (c *StageController) ApproveStage(ctx *gin.Context) {
	param := &requests.StageApproveRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidStageApproveRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	var err error
	if param.Action == "approve" {
		err = svc.ApproveStage(ctx.Request.Context(), param.StageID, userID, param.Comment)
	} else {
		err = svc.RejectStage(ctx.Request.Context(), param.StageID, userID, param.Comment)
	}

	if err != nil {
		rsp.ToErrorResponse(errorcode.ErrorPipelineRunFail.WithDetails(err.Error()))
		return
	}

	action := "通过"
	if param.Action == "reject" {
		action = "拒绝"
	}
	rsp.Success(gin.H{
		"message": "审批" + action + "成功",
	})
}

// DeployStage 执行部署阶段
// @Summary 执行部署阶段
// @Tags CICD-Stage
// @Accept json
// @Produce json
// @Param body body requests.StageDeployRequest true "部署请求"
// @Success 200 {object} response.Response
// @Router /api/v1/k8s/cicd/stage/deploy [post]
func (c *StageController) DeployStage(ctx *gin.Context) {
	param := &requests.StageDeployRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidStageDeployRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	err := svc.ExecuteDeployStage(ctx.Request.Context(), param, userID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ErrorPipelineRunFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message": "部署已启动",
	})
}

// StageCallback Jenkins 阶段回调（实时更新阶段状态）
// @Summary Jenkins 阶段回调
// @Tags CICD-Stage
// @Accept json
// @Produce json
// @Param X-Signature header string false "HMAC-SHA256 签名"
// @Param body body requests.StageCallbackRequest true "阶段回调请求"
// @Success 200 {object} response.Response
// @Router /api/v1/k8s/cicd/stage/callback [post]
func (c *StageController) StageCallback(ctx *gin.Context) {
	param := &requests.StageCallbackRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidStageCallbackRequest); !ok {
		return
	}

	// HMAC 签名验证（可选，与最终回调保持一致）
	svc := services.NewServices()
	signature := ctx.GetHeader("X-Signature")
	if signature != "" {
		// 如果提供了签名，则验证
		if !svc.VerifyHMACSignatureSimple(signature, param.JobName, param.BuildNumber, param.StageType) {
			// 签名验证失败不报错，仅记录日志，避影响 Jenkins 构建
			rsp.Success(gin.H{"message": "ok"})
			return
		}
	}

	err := svc.StageCallback(ctx.Request.Context(), param)
	if err != nil {
		// 阶段回调失败不返回错误，避影响 Jenkins 构建
		rsp.Success(gin.H{"message": "ok"})
		return
	}

	rsp.Success(gin.H{"message": "ok"})
}
