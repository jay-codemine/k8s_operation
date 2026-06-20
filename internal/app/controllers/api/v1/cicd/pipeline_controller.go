package cicd

import (
	"fmt"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/create [post]
func (c *PipelineController) Create(ctx *gin.Context) {
	param := requests.NewPipelineCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineCreateRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, warnings, err := svc.PipelineCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("PipelineCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineCreateFail.WithDetails(err.Error()))
		return
	}

	result := gin.H{"pipeline_id": id}
	if len(warnings) > 0 {
		result["warnings"] = warnings
	}

	// 创建成功后自动触发第一次构建
	runReq := &requests.PipelineRunRequest{ID: id}
	run, runErr := svc.PipelineRun(ctx.Request.Context(), runReq, userID)
	if runErr != nil {
		global.Logger.Warn("[流水线] 创建后自动触发构建失败", zap.Int64("pipeline_id", id), zap.Error(runErr))
		result["auto_run_error"] = runErr.Error()
	} else if run != nil {
		result["run_id"] = run.ID
		result["message"] = "流水线创建成功并已自动触发构建"
	}

	rsp.Success(result)
}

// CheckName godoc
// @Summary 检查应用名称是否可用
// @Description 检查指定应用名称是否已被占用
// @Tags CICD Pipeline
// @Produce json
// @Param name query string true "应用名称"
// @Param exclude_id query int false "排除的流水线 ID（编辑时使用）"
// @Success 200 {object} map[string]any "返回 available"
// @Router /api/v1/k8s/cicd/pipeline/check-name [get]
func (c *PipelineController) CheckName(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	name := ctx.Query("name")
	if name == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("应用名称不能为空"))
		return
	}

	excludeID, _ := strconv.ParseInt(ctx.Query("exclude_id"), 10, 64)

	svc := services.NewServices()
	available, _, err := svc.PipelineCheckName(ctx.Request.Context(), name, excludeID)
	if err != nil {
		global.Logger.Error("PipelineCheckName error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"available": available})
}

// BatchCreate godoc
// @Summary 批量创建流水线
// @Description 一次性导入多个项目的流水线配置，支持跳过已存在项目
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineBatchCreateRequest true "批量创建参数"
// @Success 200 {object} map[string]any "返回批量创建结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/batch-create [post]
func (c *PipelineController) BatchCreate(ctx *gin.Context) {
	param := &requests.PipelineBatchCreateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineBatchCreateRequest); !ok {
		return
	}

	if len(param.Pipelines) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("流水线列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()

	result, err := svc.PipelineBatchCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("PipelineBatchCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorPipelineCreateFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":       fmt.Sprintf("批量创建完成：成功 %d，失败 %d，跳过 %d", result.SuccessCount, result.FailCount, result.SkipCount),
		"success_count": result.SuccessCount,
		"fail_count":    result.FailCount,
		"skip_count":    result.SkipCount,
		"results":       result.Results,
	})
}

// Detail godoc
// @Summary 获取流水线详情
// @Description 获取流水线的完整配置信息
// @Tags CICD Pipeline
// @Produce json
// @Param id query int true "流水线ID"
// @Success 200 {object} map[string]any "返回流水线详情"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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

// BatchRun godoc
// @Summary 批量运行流水线
// @Description 批量触发多个流水线的 Jenkins 构建
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineBatchRunRequest true "批量运行参数"
// @Success 200 {object} map[string]any "返回批量运行结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/batch-run [post]
func (c *PipelineController) BatchRun(ctx *gin.Context) {
	param := &requests.PipelineBatchRunRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineBatchRunRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("流水线ID列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()

	var successCount, failCount int
	var results []map[string]any

	for _, id := range param.IDs {
		runReq := &requests.PipelineRunRequest{ID: id}
		run, err := svc.PipelineRun(ctx.Request.Context(), runReq, userID)
		if err != nil {
			failCount++
			results = append(results, map[string]any{
				"id":      id,
				"success": false,
				"error":   err.Error(),
			})
		} else {
			successCount++
			results = append(results, map[string]any{
				"id":      id,
				"success": true,
				"run_id":  run.ID,
			})
		}
	}

	rsp.Success(gin.H{
		"message":       fmt.Sprintf("批量运行完成：成功 %d，失败 %d", successCount, failCount),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
	})
}

// BatchStop godoc
// @Summary 批量停止流水线
// @Description 批量停止多个正在运行的流水线
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Param body body requests.PipelineBatchStopRequest true "批量停止参数"
// @Success 200 {object} map[string]any "返回批量停止结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/pipeline/batch-stop [post]
func (c *PipelineController) BatchStop(ctx *gin.Context) {
	param := &requests.PipelineBatchStopRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidPipelineBatchStopRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("流水线ID列表不能为空"))
		return
	}

	svc := services.NewServices()

	var successCount, failCount int
	var results []map[string]any

	for _, id := range param.IDs {
		stopReq := &requests.PipelineStopRequest{ID: id}
		err := svc.PipelineStop(ctx.Request.Context(), stopReq)
		if err != nil {
			failCount++
			results = append(results, map[string]any{
				"id":      id,
				"success": false,
				"error":   err.Error(),
			})
		} else {
			successCount++
			results = append(results, map[string]any{
				"id":      id,
				"success": true,
			})
		}
	}

	rsp.Success(gin.H{
		"message":       fmt.Sprintf("批量停止完成：成功 %d，失败 %d", successCount, failCount),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "签名验证失败"
// @Failure 500 {object} map[string]interface{} "内部错误"
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
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
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

// TemplateVerify godoc
// @Summary 验证模板化发布配置
// @Description 返回所有支持的语言模板信息，包括 Jenkins Job 名称、阶段、默认参数等
// @Tags CICD Pipeline
// @Produce json
// @Success 200 {object} map[string]any "返回模板列表"
// @Router /api/v1/k8s/cicd/pipeline/template-verify [get]
func (c *PipelineController) TemplateVerify(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	svc := services.NewServices()

	templates, err := svc.TemplateVerifyAll(ctx.Request.Context())
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"templates": templates,
		"summary": gin.H{
			"total_templates":     len(templates),
			"supported_languages": []string{"go", "java", "frontend", "python"},
			"reuse_model":         "4 个通用 Jenkins Job 服务 100+ 项目",
			"callback_protocol":   "HMAC-SHA256 签名 + 阶段回调 + 最终回调",
		},
	})
}

// TemplateSimulate godoc
// @Summary 模拟模板化发布流程
// @Description 模拟完整发布流程，验证参数和 Jenkins 配置（不实际触发构建）
// @Tags CICD Pipeline
// @Produce json
// @Param language_type query string true "语言类型: go/java/frontend/python"
// @Param git_repo query string true "Git 仓库地址"
// @Param git_branch query string false "Git 分支，默认 main"
// @Param image_repo query string true "镜像仓库地址"
// @Success 200 {object} map[string]any "返回模拟结果"
// @Router /api/v1/k8s/cicd/pipeline/template-simulate [get]
func (c *PipelineController) TemplateSimulate(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	languageType := ctx.Query("language_type")
	gitRepo := ctx.Query("git_repo")
	gitBranch := ctx.DefaultQuery("git_branch", global.DefaultBranch())
	imageRepo := ctx.Query("image_repo")

	if languageType == "" || gitRepo == "" || imageRepo == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("缺少必填参数: language_type, git_repo, image_repo"))
		return
	}

	svc := services.NewServices()
	result, err := svc.TemplateSimulateRun(ctx.Request.Context(), languageType, gitRepo, gitBranch, imageRepo)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(result)
}

// SonarReport godoc
// @Summary 获取流水线 SonarQube 代码质量报告
// @Description 返回指定流水线的代码质量扫描结果，包括 Bug、漏洞、覆盖率等指标
// @Tags CICD Pipeline
// @Produce json
// @Param pipeline_id query int true "流水线ID"
// @Param run_id query int false "运行记录ID（空则获取最新一次）"
// @Success 200 {object} map[string]any "返回代码质量报告"
// @Router /api/v1/k8s/cicd/pipeline/sonar-report [get]
func (c *PipelineController) SonarReport(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	pipelineID, err := strconv.ParseInt(ctx.Query("pipeline_id"), 10, 64)
	if err != nil || pipelineID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的 pipeline_id"))
		return
	}

	var runID int64
	if rid := ctx.Query("run_id"); rid != "" {
		runID, _ = strconv.ParseInt(rid, 10, 64)
	}

	svc := services.NewServices()
	report, err := svc.GetSonarReport(ctx.Request.Context(), pipelineID, runID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(report)
}

// SonarCallback godoc
// @Summary SonarQube 扫描结果回调
// @Description 接收 Jenkins 回传的 SonarQube 扫描结果，存储代码质量报告
// @Tags CICD Pipeline
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any "成功"
// @Router /api/v1/k8s/cicd/pipeline/sonar-callback [post]
func (c *PipelineController) SonarCallback(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		PipelineID       int64   `json:"pipeline_id"`
		RunID            int64   `json:"run_id"`
		ProjectKey       string  `json:"project_key"`
		ProjectName      string  `json:"project_name"`
		QualityGate      string  `json:"quality_gate"`
		DashboardURL     string  `json:"dashboard_url"`
		Bugs             int     `json:"bugs"`
		Vulnerabilities  int     `json:"vulnerabilities"`
		CodeSmells       int     `json:"code_smells"`
		Coverage         float64 `json:"coverage"`
		Duplications     float64 `json:"duplications"`
		LinesOfCode      int     `json:"lines_of_code"`
		SecurityHotspots int     `json:"security_hotspots"`
		ReliabilityRating string `json:"reliability_rating"`
		SecurityRating    string `json:"security_rating"`
		Maintainability   string `json:"maintainability_rating"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	err := svc.SaveSonarReport(ctx.Request.Context(), req.PipelineID, req.RunID, &models.StageSonarInfo{
		ProjectKey:        req.ProjectKey,
		ProjectName:       req.ProjectName,
		QualityGate:       req.QualityGate,
		DashboardURL:      req.DashboardURL,
		Bugs:              req.Bugs,
		Vulnerabilities:   req.Vulnerabilities,
		CodeSmells:        req.CodeSmells,
		Coverage:          req.Coverage,
		Duplications:      req.Duplications,
		LinesOfCode:       req.LinesOfCode,
		SecurityHotspots:  req.SecurityHotspots,
		ReliabilityRating: req.ReliabilityRating,
		SecurityRating:    req.SecurityRating,
		Maintainability:   req.Maintainability,
	})
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "SonarQube 报告已保存"})
}

// DeploySilenceStatus godoc
// @Summary 获取流水线发布静默状态
// @Description 查看当前流水线是否有活跃的发布静默规则
// @Tags CICD Pipeline
// @Produce json
// @Param pipeline_id query int true "流水线ID"
// @Success 200 {object} map[string]any "返回静默状态"
// @Router /api/v1/k8s/cicd/pipeline/deploy-silence-status [get]
func (c *PipelineController) DeploySilenceStatus(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("pipeline_id")
	pipelineID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || pipelineID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的流水线ID"))
		return
	}

	svc := services.NewServices()
	rules, err := svc.GetActiveDeploySilences(ctx.Request.Context(), pipelineID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"active":     len(rules) > 0,
		"count":      len(rules),
		"rules":      rules,
	})
}
