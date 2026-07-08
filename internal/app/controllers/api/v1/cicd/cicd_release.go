package cicd

import (
	"fmt"
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

type CicdReleaseController struct {
}

func NewCicdReleaseController() *CicdReleaseController {
	return &CicdReleaseController{}
}

// @Summary 创建 CICD 发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseCreateRequest true "创建参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/cicd/release/create [post]
func (c *CicdReleaseController) Create(ctx *gin.Context) {
	param := requests.NewCicdReleaseCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseCreateRequest); !ok {
		return
	}

	// 获取当前用户 ID
	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.CicdReleaseCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("CicdReleaseCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseCreateFail.WithDetails(err.Error()))
		return
	}

	result := gin.H{"release_id": id}

	// 创建发布单后自动触发流水线构建（如果关联了 pipeline_id）
	if param.PipelineID > 0 {
		runReq := &requests.PipelineRunRequest{ID: param.PipelineID}
		run, runErr := svc.PipelineRun(ctx.Request.Context(), runReq, userID)
		if runErr != nil {
			global.Logger.Warn("[发布] 创建发布单后自动触发构建失败",
				zap.Int64("release_id", id),
				zap.Int64("pipeline_id", param.PipelineID),
				zap.Error(runErr),
			)
			result["auto_build_error"] = runErr.Error()
		} else if run != nil {
			result["run_id"] = run.ID
			result["message"] = "发布单创建成功并已自动触发构建"
			global.Logger.Info("[发布] 创建发布单后自动触发构建成功",
				zap.Int64("release_id", id),
				zap.Int64("pipeline_id", param.PipelineID),
				zap.Int64("run_id", run.ID),
			)
		}
	}

	rsp.Success(result)
}

// Detail godoc
// @Summary 获取发布单详情
// @Tags CICD Release
// @Produce json
// @Param id query int true "发布单ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/detail [get]
func (c *CicdReleaseController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的发布单ID"))
		return
	}

	svc := services.NewServices()
	rel, tasks, err := svc.CicdReleaseDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("CicdReleaseDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"release": rel,
		"tasks":   tasks,
	})
}

// List godoc
// @Summary 获取发布单列表
// @Tags CICD Release
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param app_name query string false "应用名称"
// @Param status query string false "状态"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/list [get]
func (c *CicdReleaseController) List(ctx *gin.Context) {
	param := requests.NewCicdReleaseListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.CicdReleaseList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("CicdReleaseList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Stats godoc
// @Summary 获取发布单统计
// @Tags CICD Release
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/stats [get]
func (c *CicdReleaseController) Stats(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	svc := services.NewServices()
	stats, err := svc.CicdReleaseStats(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("CicdReleaseStats error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"stats": stats})
}

// Update godoc
// @Summary 编辑发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseUpdateRequest true "编辑参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/update [post]
func (c *CicdReleaseController) Update(ctx *gin.Context) {
	param := &requests.CicdReleaseUpdateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.CicdReleaseUpdate(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("CicdReleaseUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseUpdateFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"message": "更新成功"})
}

// Delete godoc
// @Summary 删除发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "删除参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/delete [post]
func (c *CicdReleaseController) Delete(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.CicdReleaseDelete(ctx.Request.Context(), param.ID); err != nil {
		global.Logger.Error("CicdReleaseDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseDeleteFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"message": "删除成功"})
}

// Cancel godoc
// @Summary 取消发布单
// @Description 智能取消：已部署成功/运行中的会触发回滚，未部署的直接取消
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "取消参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/cancel [post]
func (c *CicdReleaseController) Cancel(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	result, err := svc.CicdReleaseCancel(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseCancel error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseCancelFail.WithDetails(err.Error()))
		return
	}

	// 根据操作类型返回不同的消息
	if result.Action == "rollback" {
		rsp.Success(gin.H{
			"message":            "发布单已触发回滚",
			"action":             result.Action,
			"rollback_release_id": result.RollbackReleaseID,
		})
	} else {
		rsp.Success(gin.H{
			"message": "取消成功",
			"action":  result.Action,
		})
	}
}

// Rollback godoc
// @Summary 回滚发布单
// @Description 将已部署的工作负载回滚到上一个版本，会创建新的发布单执行回滚操作
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "回滚参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/rollback [post]
func (c *CicdReleaseController) Rollback(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	newID, err := svc.CicdReleaseRollback(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseRollback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseRollbackFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":        "回滚成功",
		"rollback_release_id": newID,
	})
}

// Retry godoc
// @Summary 重试发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "重试参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/retry [post]
func (c *CicdReleaseController) Retry(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	newID, err := svc.CicdReleaseRetry(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseRetry error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseRetryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"release_id": newID})
}

// Tasks godoc
// @Summary 获取发布单下的任务列表
// @Tags CICD Release
// @Produce json
// @Param release_id query int true "发布单ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/tasks [get]
func (c *CicdReleaseController) Tasks(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("release_id")
	releaseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || releaseID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的发布单ID"))
		return
	}

	svc := services.NewServices()
	tasks, err := svc.CicdTasksByRelease(ctx.Request.Context(), releaseID)
	if err != nil {
		global.Logger.Error("CicdTasksByRelease error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"tasks": tasks})
}

// BuildCallback godoc
// @Summary Jenkins 构建回调
// @Tags CICD Callback
// @Accept json
// @Produce json
// @Param body body requests.CicdBuildCallbackRequest true "回调参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/callback/build [post]
func (c *CicdReleaseController) BuildCallback(ctx *gin.Context) {
	param := &requests.CicdBuildCallbackRequest{}
	rsp := response.NewResponse(ctx)

	if err := ctx.ShouldBindJSON(param); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.CicdBuildCallback(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("CicdBuildCallback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdBuildCallbackFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "回调处理成功"})
}

// BatchRetry godoc
// @Summary 批量重新发布
// @Description 批量重新发布（根据最近一次发布记录重新发布）
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseBatchRetryRequest true "批量发布参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/batch-retry [post]
func (c *CicdReleaseController) BatchRetry(ctx *gin.Context) {
	param := requests.NewCicdReleaseBatchRetryRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseBatchRetryRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("发布单ID列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	results, err := svc.CicdReleaseBatchRetry(ctx.Request.Context(), param.IDs, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseBatchRetry error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	// 统计成功/失败
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	rsp.Success(gin.H{
		"message": fmt.Sprintf("批量发布完成：成功 %d / 共 %d", successCount, len(results)),
		"results": results,
		"success": successCount,
		"total":   len(results),
	})
}

// BatchRollback godoc
// @Summary 批量回滚发布单
// @Description 批量回滚（根据最近一次发布记录回滚到上一个版本）
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseBatchRollbackRequest true "批量回滚参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/batch-rollback [post]
func (c *CicdReleaseController) BatchRollback(ctx *gin.Context) {
	param := requests.NewCicdReleaseBatchRollbackRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseBatchRollbackRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("发布单ID列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	results, err := svc.CicdReleaseBatchRollback(ctx.Request.Context(), param.IDs, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseBatchRollback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	// 统计成功/失败
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	rsp.Success(gin.H{
		"message": fmt.Sprintf("批量回滚完成：成功 %d / 共 %d", successCount, len(results)),
		"results": results,
		"success": successCount,
		"total":   len(results),
	})
}

// BatchCancel godoc
// @Summary 批量取消发布单
// @Description 批量取消（智能判断：已部署成功/运行中的会触发回滚，未部署的直接取消）
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseBatchCancelRequest true "批量取消参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/batch-cancel [post]
func (c *CicdReleaseController) BatchCancel(ctx *gin.Context) {
	param := requests.NewCicdReleaseBatchCancelRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseBatchCancelRequest); !ok {
		return
	}

	if len(param.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("发布单ID列表不能为空"))
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	results, err := svc.CicdReleaseBatchCancel(ctx.Request.Context(), param.IDs, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseBatchCancel error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	// 统计成功/失败
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	rsp.Success(gin.H{
		"message": fmt.Sprintf("批量取消完成：成功 %d / 共 %d", successCount, len(results)),
		"results": results,
		"success": successCount,
		"total":   len(results),
	})
}

// SyncFromPipeline godoc
// @Summary 同步流水线运行记录到发布管理
// @Description 将最近的流水线运行记录同步到发布管理页面
// @Tags CICD Release
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/sync-from-pipeline [post]
func (c *CicdReleaseController) SyncFromPipeline(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	svc := services.NewServices()
	synced, err := svc.CicdReleaseSyncFromPipeline(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("CicdReleaseSyncFromPipeline error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message": fmt.Sprintf("同步完成：新增 %d 条发布记录", synced),
		"synced":  synced,
	})
}

// History godoc
// @Summary 应用发布历史查询（增强版）
// @Description 支持时间范围、命名空间、应用名筛选的发布历史查询
// @Tags CICD Release
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param app_name query string false "应用名称"
// @Param namespace query string false "命名空间"
// @Param status query string false "状态"
// @Param start_time query int false "开始时间戳"
// @Param end_time query int false "结束时间戳"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/history [get]
func (c *CicdReleaseController) History(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	appName := ctx.Query("app_name")
	namespace := ctx.Query("namespace")
	status := ctx.Query("status")
	startTime, _ := strconv.ParseInt(ctx.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(ctx.Query("end_time"), 10, 64)

	svc := services.NewServices()
	list, total, err := svc.CicdReleaseHistory(ctx.Request.Context(), appName, namespace, status, startTime, endTime, page, pageSize)
	if err != nil {
		global.Logger.Error("CicdReleaseHistory error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// StatsEnhanced godoc
// @Summary 增强版发布统计
// @Description 返回今日/本周发布数、成功率、状态分布
// @Tags CICD Release
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/stats-enhanced [get]
func (c *CicdReleaseController) StatsEnhanced(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	svc := services.NewServices()
	stats, err := svc.CicdReleaseStatsEnhanced(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("CicdReleaseStatsEnhanced error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(stats)
}
