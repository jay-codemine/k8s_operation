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

type PipelineController struct {
}

func NewPipelineController() *PipelineController {
	return &PipelineController{}
}

// Create godoc
// @Summary 创建流水线
// @Description 创建新的 CI/CD 流水线配置，关联 Git 仓库和 Jenkins Job
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineCreateRequest true "创建参数"
// @Success 200 {object} map[string]any "返回 pipeline_id"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/create [post]
func (c *PipelineController) Create(ctx *gin.Context) {
	param := requests.NewPipelineCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineCreateRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.PipelineCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("PipelineCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineCreateFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"pipeline_id": id})
}

// Detail godoc
// @Summary 获取流水线详情
// @Description 获取流水线的完整配置信息
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Success 200 {object} map[string]any "返回流水线详情"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/detail [get]
func (c *PipelineController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	svc := services.NewServices()
	pipeline, err := svc.PipelineDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("PipelineDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"pipeline": pipeline})
}

// List godoc
// @Summary 获取流水线列表
// @Description 分页查询流水线列表，支持关键字和状态筛选
// @Tags CICD Pipeline
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param keyword query string false "关键字（名称/描述/Git仓库）"
// @Param status query string false "状态筛选（idle/running/disabled）"
// @Success 200 {object} map[string]any "返回列表和总数"
// @Router /api/v1/k8s/cicd/pipeline/list [get]
func (c *PipelineController) List(ctx *gin.Context) {
	param := requests.NewPipelineListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.PipelineList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("PipelineList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Update godoc
// @Summary 更新流水线
// @Description 更新流水线配置信息
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineUpdateRequest true "更新参数"
// @Success 200 {object} map[string]any "更新成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/update [post]
func (c *PipelineController) Update(ctx *gin.Context) {
	param := requests.NewPipelineUpdateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.PipelineUpdate(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("PipelineUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineUpdateFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// Delete godoc
// @Summary 删除流水线
// @Description 软删除流水线（运行中的流水线无法删除）
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineIDRequest true "删除参数"
// @Success 200 {object} map[string]any "删除成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/delete [post]
func (c *PipelineController) Delete(ctx *gin.Context) {
	param := &requests.PipelineIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineIDRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.PipelineDelete(ctx.Request.Context(), param.ID)
	if err != nil {
		global.Logger.Error("PipelineDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineDeleteFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}

// Run godoc
// @Summary 运行流水线
// @Description 触发 Jenkins 构建，开始执行流水线
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineRunRequest true "运行参数"
// @Success 200 {object} map[string]any "返回运行记录ID"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/run [post]
func (c *PipelineController) Run(ctx *gin.Context) {
	param := &requests.PipelineRunRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineRunRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	run, err := svc.PipelineRun(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("PipelineRun error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineRunFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message": "流水线启动成功",
		"run_id":  run.ID,
	})
}

// Stop godoc
// @Summary 停止流水线
// @Description 停止正在运行的 Jenkins 构建
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineStopRequest true "停止参数"
// @Success 200 {object} map[string]any "停止成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/stop [post]
func (c *PipelineController) Stop(ctx *gin.Context) {
	param := &requests.PipelineStopRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineStopRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.PipelineStop(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("PipelineStop error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineStopFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "流水线已停止"})
}

// Logs godoc
// @Summary 获取流水线日志
// @Description 获取 Jenkins 构建的控制台日志，支持增量获取
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Param build_number query int false "构建号（不传则获取最新的）"
// @Param start_line query int false "起始行号（用于增量获取）"
// @Success 200 {object} map[string]any "返回日志内容"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/logs [get]
func (c *PipelineController) Logs(ctx *gin.Context) {
	param := &requests.PipelineLogsRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineLogsRequest); !ok {
		return
	}

	svc := services.NewServices()
	log, err := svc.PipelineLogs(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("PipelineLogs error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineLogsFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"logs": log})
}

// Status godoc
// @Summary 获取流水线实时状态
// @Description 获取流水线和当前构建的实时状态
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Success 200 {object} map[string]any "返回状态信息"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/status [get]
func (c *PipelineController) Status(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	svc := services.NewServices()
	pipeline, buildInfo, latestRun, err := svc.PipelineStatusWithRun(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("PipelineStatus error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	result := gin.H{
		"pipeline": pipeline,
	}
	if buildInfo != nil {
		result["build_info"] = buildInfo
	}
	// 返回最新运行记录（包含错误信息）
	if latestRun != nil {
		result["latest_run"] = latestRun
	}

	rsp.Success(result)
}

// History godoc
// @Summary 获取流水线运行历史
// @Description 分页获取流水线的历史运行记录
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Success 200 {object} map[string]any "返回历史记录列表"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/history [get]
func (c *PipelineController) History(ctx *gin.Context) {
	param := requests.NewPipelineHistoryRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineHistoryRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.PipelineHistory(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("PipelineHistory error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Callback godoc
// @Summary Jenkins 构建状态回调（生产级）
// @Description Jenkins 构建完成后调用此接口通知平台更新状态
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param X-Signature header string false "HMAC-SHA256 签名（用于验证请求真实性）"
// @Param body body requests.PipelineCallbackRequest true "回调参数"
// @Success 200 {object} map[string]any "回调处理成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 401 {object} errorcode.Error "签名验证失败"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/callback [post]
func (c *PipelineController) Callback(ctx *gin.Context) {
	param := &requests.PipelineCallbackRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineCallbackRequest); !ok {
		return
	}

	// HMAC 签名验证
	// 签名格式: HMAC-SHA256(secret, "job_name:build_number:status")
	svc := services.NewServices()
	signature := ctx.GetHeader("X-Signature")
	if !svc.VerifyHMACSignature(signature, param.JobName, param.BuildNumber, param.Status) {
		global.Logger.Warn("[回调] HMAC 签名验证失败",
			zap.String("job_name", param.JobName),
			zap.Int("build_number", param.BuildNumber),
			zap.String("status", param.Status),
			zap.String("signature", signature),
		)
		rsp.ToErrorResponse(errorcode.UnauthorizedTokenError.WithDetails("回调签名验证失败"))
		return
	}

	result, err := svc.PipelineCallback(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("Callback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineCallbackFail.WithDetails(err.Error()))
		return
	}

	// 返回部署结果给 Jenkins，让用户在 Jenkins 看到最终状态
	rsp.Success(gin.H{
		"message":        result.Message,
		"deploy_enabled": result.DeployEnabled,
		"deploy_success": result.DeploySuccess,
		"deploy_message": result.DeployMessage,
		"namespace":      result.Namespace,
		"deployment":     result.Deployment,
		"image":          result.Image,
	})
}

// Stages godoc
// @Summary 获取流水线阶段数据
// @Description 获取 Jenkins Pipeline 的阶段执行数据（来自 Jenkins Workflow API）
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Param build_number query int false "构建号（不传则获取最新的）"
// @Success 200 {object} map[string]any "返回阶段数据"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/stages [get]
func (c *PipelineController) Stages(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	buildNumber := 0
	if bn := ctx.Query("build_number"); bn != "" {
		buildNumber, _ = strconv.Atoi(bn)
	}

	svc := services.NewServices()
	stages, err := svc.PipelineStages(ctx.Request.Context(), id, buildNumber)
	if err != nil {
		global.Logger.Error("PipelineStages error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"stages": stages})
}
