package cicd

import (
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// GitOpsController GitOps 控制器
type GitOpsController struct {
	svc *services.Services
}

func NewGitOpsController() *GitOpsController {
	return &GitOpsController{
		svc: services.NewServices(),
	}
}

// SyncCallback ArgoCD 同步状态 Webhook 回调
// POST /api/v1/k8s/cicd/gitops/sync-callback
func (c *GitOpsController) SyncCallback(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	// 读取请求体用于 HMAC 验证
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("读取请求体失败"))
		return
	}

	// HMAC 签名验证
	sig := ctx.GetHeader("X-Signature")
	if !c.svc.GitOpsVerifyHMAC(sig, body) {
		global.Logger.Warn("[GitOps] HMAC 签名验证失败")
		rsp.ToErrorResponse(errorcode.UnauthorizedTokenError.WithDetails("HMAC 签名验证失败"))
		return
	}

	// 解析请求参数（支持 form 和 query）
	param := &requests.GitOpsSyncCallbackRequest{
		AppName:      ctx.PostForm("app_name"),
		SyncStatus:   ctx.PostForm("sync_status"),
		SyncRevision: ctx.PostForm("sync_revision"),
		HealthStatus: ctx.PostForm("health_status"),
	}
	if param.AppName == "" {
		param.AppName = ctx.Query("app_name")
	}
	if param.SyncStatus == "" {
		param.SyncStatus = ctx.Query("sync_status")
	}
	if param.SyncRevision == "" {
		param.SyncRevision = ctx.Query("sync_revision")
	}
	if param.HealthStatus == "" {
		param.HealthStatus = ctx.Query("health_status")
	}

	if param.AppName == "" || param.SyncStatus == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("缺少必要参数: app_name, sync_status"))
		return
	}

	if err := c.svc.GitOpsSyncCallback(ctx, param); err != nil {
		global.Logger.Errorf("[GitOps] 同步回调处理失败: %v", err)
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "ok"})
}

// Webhook Argo Workflows 状态回调
// POST /api/v1/k8s/cicd/gitops/webhook
func (c *GitOpsController) Webhook(ctx *gin.Context) {
	// GitOps workflow webhook 委托给 PipelineController.Callback 处理
	pipelineCtrl := NewPipelineController()
	pipelineCtrl.Callback(ctx)
}

// GetAppStatus 获取 ArgoCD Application 同步状态
// GET /api/v1/k8s/cicd/gitops/app-status
func (c *GitOpsController) GetAppStatus(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	appName := ctx.Query("app_name")
	if appName == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("缺少必要参数: app_name"))
		return
	}

	status, err := c.svc.GitOpsGetAppStatus(ctx, appName)
	if err != nil {
		global.Logger.Errorf("[GitOps] 查询 Application 状态失败: %v", err)
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"app_name":      appName,
		"sync_status":   status.SyncStatus,
		"sync_revision": status.SyncRevision,
		"health_status": status.HealthStatus,
		"phase":         status.Phase,
	})
}

// GetSyncHistory 获取流水线的 GitOps 同步历史
// GET /api/v1/k8s/cicd/gitops/sync-history
func (c *GitOpsController) GetSyncHistory(ctx *gin.Context) {
	// 复用 PipelineController.History 逻辑（内部自行处理响应）
	pipelineCtrl := NewPipelineController()
	pipelineCtrl.History(ctx)
}

// TriggerSync 手动触发 ArgoCD Application 同步
// POST /api/v1/k8s/cicd/gitops/sync
func (c *GitOpsController) TriggerSync(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	type triggerRequest struct {
		PipelineID int64 `json:"pipeline_id"`
	}

	var req triggerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("参数解析失败"))
		return
	}

	if req.PipelineID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("缺少必要参数: pipeline_id"))
		return
	}

	if err := c.svc.GitOpsTriggerSyncByPipeline(ctx, req.PipelineID); err != nil {
		global.Logger.Errorf("[GitOps] 手动触发同步失败: %v", err)
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "同步已触发"})
}

// ReleaseStats 获取 GitOps 发布统计数据
// GET /api/v1/k8s/cicd/gitops/release-stats
func (c *GitOpsController) ReleaseStats(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	svc := services.NewServices()
	stats, err := svc.GitOpsReleaseStats(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("GitOpsReleaseStats error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"stats": stats})
}

// ReleaseSearch 增强 GitOps 发布搜索
// GET /api/v1/k8s/cicd/gitops/release-search
func (c *GitOpsController) ReleaseSearch(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	var req models.GitOpsReleaseSearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	svc := services.NewServices()
	list, total, err := svc.GitOpsReleaseSearch(ctx.Request.Context(), &req)
	if err != nil {
		global.Logger.Error("GitOpsReleaseSearch error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"list": list, "total": total})
}
