package ai

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
)

// AIOpsController AIOps 智能运维控制器
type AIOpsController struct{}

// NewAIOpsController 创建 AIOps 控制器
func NewAIOpsController() *AIOpsController {
	return &AIOpsController{}
}

// AnalyzeAlert AI 分析告警事件
// POST /api/v1/ai/ops/alert/analyze
func (c *AIOpsController) AnalyzeAlert(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.AlertAnalysisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if req.EventID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("event_id 必填"))
		return
	}

	// 获取当前用户（从 session 或 token）
	userID := int64(getUserID(ctx))

	result, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().AnalyzeAlert(ctx.Request.Context(), &req, userID)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(result)
}

// DiagnoseLogs AI 日志诊断
// POST /api/v1/ai/ops/log/diagnose
func (c *AIOpsController) DiagnoseLogs(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.LogDiagnosisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if req.Query == "" && req.Namespace == "" && req.Pod == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请指定查询条件(query/namespace/pod)"))
		return
	}

	userID := int64(getUserID(ctx))

	result, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().DiagnoseLogs(ctx.Request.Context(), &req, userID)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(result)
}

// RunInspection 手动触发巡检
// POST /api/v1/ai/ops/inspection/run
func (c *AIOpsController) RunInspection(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	userID := int64(getUserID(ctx))

	report, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().RunInspection(ctx.Request.Context(), userID)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(report)
}

// GetInspectionReports 巡检报告列表
// GET /api/v1/ai/ops/inspection/list
func (c *AIOpsController) GetInspectionReports(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	reports, total, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().GetInspectionReports(ctx.Request.Context(), page, pageSize)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"list":  reports,
		"total": total,
		"page":  page,
	})
}

// GetInspectionDetail 巡检报告详情
// GET /api/v1/ai/ops/inspection/:id
func (c *AIOpsController) GetInspectionDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	report, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().GetInspectionReport(ctx.Request.Context(), id)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(report)
}

// GetAnalysisRecords 分析记录列表
// GET /api/v1/ai/ops/records
func (c *AIOpsController) GetAnalysisRecords(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	recordType := ctx.Query("type") // alert_analysis/log_diagnosis/inspection
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	records, total, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().GetAnalysisRecords(ctx.Request.Context(), recordType, page, pageSize)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"list":  records,
		"total": total,
		"page":  page,
	})
}

// GetDashboard AIOps 仪表盘数据
// GET /api/v1/ai/ops/dashboard
func (c *AIOpsController) GetDashboard(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	stats, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().GetDashboardStats(ctx.Request.Context())
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(stats)
}

// ExportReport 导出巡检报告为 Markdown
// GET /api/v1/ai/ops/inspection/:id/export
func (c *AIOpsController) ExportReport(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	markdown, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().ExportReportMarkdown(ctx.Request.Context(), id)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	// 如果请求 download=true 则返回文件下载
	if ctx.Query("download") == "true" {
		ctx.Header("Content-Type", "text/markdown; charset=utf-8")
		ctx.Header("Content-Disposition", "attachment; filename=inspection_report_"+strconv.FormatInt(id, 10)+".md")
		ctx.String(200, markdown)
		return
	}

	resp.Success(gin.H{"content": markdown, "format": "markdown"})
}

// NotifyReport 发送巡检报告到通知渠道
// POST /api/v1/ai/ops/inspection/:id/notify
func (c *AIOpsController) NotifyReport(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	var req struct {
		ChannelIDs []int64 `json:"channel_ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("channel_ids 必填"))
		return
	}
	if len(req.ChannelIDs) == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请至少选择一个通知渠道"))
		return
	}

	results, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().NotifyReport(ctx.Request.Context(), &services.NotifyReportRequest{
		ReportID:   id,
		ChannelIDs: req.ChannelIDs,
	})
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(results)
}

// GetNotifyChannels 获取可用通知渠道
// GET /api/v1/ai/ops/channels
func (c *AIOpsController) GetNotifyChannels(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	channels, err := middlewares.NewServicesFromContext(ctx).AIOpsSvc().GetNotifyChannels(ctx.Request.Context())
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(channels)
}
