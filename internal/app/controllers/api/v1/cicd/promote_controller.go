package cicd

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/middlewares"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// PromoteController 镜像晋级 / 环境目标配置控制器
//
// 支撑「一次构建、跨环境晋级」(build once, promote everywhere)：
//   - 环境目标配置（cicd_pipeline_target）：一条流水线预先绑定 dev/test/staging/prod 各环境部署目标
//   - 镜像晋级：复用已构建的不可变镜像发布到目标环境（复用发布单的审批 + 多集群下发能力）
//   - 晋级链可视化：展示各环境当前部署的镜像与发布状态
type PromoteController struct{}

func NewPromoteController() *PromoteController {
	return &PromoteController{}
}

// Targets godoc
// @Summary 获取流水线的环境部署目标列表
// @Tags CICD Promote
// @Produce json
// @Param pipeline_id query int true "流水线ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/promote/targets [get]
func (c *PromoteController) Targets(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	pipelineID, err := strconv.ParseInt(ctx.Query("pipeline_id"), 10, 64)
	if err != nil || pipelineID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	list, err := svc.PipelineTargetList(ctx.Request.Context(), pipelineID)
	if err != nil {
		global.Logger.Error("PipelineTargetList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdPipelineTargetFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"list": list})
}

// SaveTargets godoc
// @Summary 全量保存流水线的环境部署目标
// @Tags CICD Promote
// @Accept json
// @Produce json
// @Param body body requests.PipelineTargetSaveRequest true "环境目标配置"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/promote/targets/save [post]
func (c *PromoteController) SaveTargets(ctx *gin.Context) {
	param := &requests.PipelineTargetSaveRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineTargetSaveRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.PipelineTargetSave(ctx.Request.Context(), param, userID); err != nil {
		global.Logger.Error("PipelineTargetSave error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdPipelineTargetFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"message": "环境目标配置保存成功"})
}

// DeleteTarget godoc
// @Summary 删除单个环境部署目标
// @Tags CICD Promote
// @Accept json
// @Produce json
// @Param body body requests.PipelineTargetDeleteRequest true "删除参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/promote/targets/delete [post]
func (c *PromoteController) DeleteTarget(ctx *gin.Context) {
	param := &requests.PipelineTargetDeleteRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineTargetDeleteRequest); !ok {
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.PipelineTargetDelete(ctx.Request.Context(), param.ID); err != nil {
		global.Logger.Error("PipelineTargetDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdPipelineTargetFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"message": "环境目标删除成功"})
}

// Promote godoc
// @Summary 镜像晋级：将已构建镜像发布到目标环境
// @Tags CICD Promote
// @Accept json
// @Produce json
// @Param body body requests.PipelinePromoteRequest true "晋级参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/promote/run [post]
func (c *PromoteController) Promote(ctx *gin.Context) {
	param := &requests.PipelinePromoteRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelinePromoteRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := middlewares.NewServicesFromContext(ctx)
	releaseID, err := svc.PipelinePromote(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("PipelinePromote error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdPromoteFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{
		"release_id": releaseID,
		"message":    "镜像晋级已提交（如目标环境需审批则进入审批流程）",
	})
}

// Chain godoc
// @Summary 获取流水线的晋级链视图（各环境当前部署镜像/状态）
// @Tags CICD Promote
// @Produce json
// @Param pipeline_id query int true "流水线ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/promote/chain [get]
func (c *PromoteController) Chain(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	pipelineID, err := strconv.ParseInt(ctx.Query("pipeline_id"), 10, 64)
	if err != nil || pipelineID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	nodes, err := svc.PipelinePromotionChain(ctx.Request.Context(), pipelineID)
	if err != nil {
		global.Logger.Error("PipelinePromotionChain error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdPromoteFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"chain": nodes})
}
