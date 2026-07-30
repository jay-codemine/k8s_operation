package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/middlewares"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// AuditLogController 审计日志控制器
type AuditLogController struct{}

// NewAuditLogController 创建审计日志控制器
func NewAuditLogController() *AuditLogController {
	return &AuditLogController{}
}

// List 获取审计日志列表
// @Summary 获取审计日志列表
// @Description 支持多维度筛选、分页、排序
// @Tags 审计日志
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param username query string false "用户名"
// @Param action query string false "操作类型"
// @Param module query string false "模块"
// @Param status query string false "状态: success/failed"
// @Param start_time query int false "开始时间(unix)"
// @Param end_time query int false "结束时间(unix)"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} models.AuditLogListResponse
// @Router /api/v1/platform/audit/logs [get]
func (ctrl *AuditLogController) List(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	var query models.AuditLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		global.Logger.Error("解析审计日志查询参数失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	result, err := svc.AuditLogList(ctx.Request.Context(), &query)
	if err != nil {
		global.Logger.Error("查询审计日志失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(result)
}

// Detail 获取审计日志详情
// @Summary 获取审计日志详情
// @Description 根据ID获取单条审计日志完整信息
// @Tags 审计日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} models.AuditLog
// @Router /api/v1/platform/audit/logs/{id} [get]
func (ctrl *AuditLogController) Detail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	log, err := svc.AuditLogGetByID(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("获取审计日志详情失败", zap.Error(err), zap.Int64("id", id))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(log)
}

// Statistics 获取审计统计数据
// @Summary 获取审计统计数据
// @Description 返回今日/本周操作量、成功率、用户排行、模块排行等
// @Tags 审计日志
// @Produce json
// @Success 200 {object} models.AuditStatistics
// @Router /api/v1/platform/audit/statistics [get]
func (ctrl *AuditLogController) Statistics(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	stats, err := svc.AuditLogGetStatistics(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("获取审计统计失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(stats)
}

// GetRetention 获取审计日志保留策略
// @Summary 获取审计日志保留策略
// @Description 返回当前保留天数和是否永久保留
// @Tags 审计日志
// @Produce json
// @Success 200 {object} models.AuditRetentionPolicy
// @Router /api/v1/platform/audit/retention [get]
func (ctrl *AuditLogController) GetRetention(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	policy, err := svc.AuditLogGetRetention(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("获取审计保留策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(policy)
}

// UpdateRetention 更新审计日志保留策略
// @Summary 更新审计日志保留策略
// @Description 设置保留天数(0=永久)或切换永久保留
// @Tags 审计日志
// @Accept json
// @Produce json
// @Param body body models.AuditRetentionUpdateReq true "保留策略"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/audit/retention [put]
func (ctrl *AuditLogController) UpdateRetention(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	var req models.AuditRetentionUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.Logger.Error("解析保留策略请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	if err := svc.AuditLogUpdateRetention(ctx.Request.Context(), &req); err != nil {
		global.Logger.Error("更新审计保留策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(gin.H{"message": "保留策略已更新"})
}

// Cleanup 手动触发清理过期审计日志
// @Summary 手动清理过期审计日志
// @Description 根据当前保留策略清理过期数据
// @Tags 审计日志
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/platform/audit/cleanup [post]
func (ctrl *AuditLogController) Cleanup(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	affected, err := svc.AuditLogCleanup(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("清理审计日志失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(gin.H{
		"message":  "清理完成",
		"affected": affected,
	})
}

// Export 导出审计日志
// @Summary 导出审计日志
// @Description 按筛选条件导出审计日志（最多10000条）
// @Tags 审计日志
// @Produce json
// @Param username query string false "用户名"
// @Param action query string false "操作类型"
// @Param module query string false "模块"
// @Param status query string false "状态"
// @Param start_time query int false "开始时间"
// @Param end_time query int false "结束时间"
// @Success 200 {array} models.AuditLog
// @Router /api/v1/platform/audit/export [get]
func (ctrl *AuditLogController) Export(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	var query models.AuditLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	logs, err := svc.AuditLogExport(ctx.Request.Context(), &query)
	if err != nil {
		global.Logger.Error("导出审计日志失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(gin.H{
		"list":  logs,
		"total": len(logs),
	})
}
