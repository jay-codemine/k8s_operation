package cicd

import (
	"io"

	"github.com/gin-gonic/gin"

	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
)

// GitOpsController GitOps 控制器
// 不持有 Services：启动期构造的 Services 绑的是 global.DB，绕过租户过滤。
// 每个 handler 按请求从 context 取租户隔离实例。
type GitOpsController struct {
}

func NewGitOpsController() *GitOpsController {
	return &GitOpsController{}
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
	svc := middlewares.NewServicesFromContext(ctx)
	sig := ctx.GetHeader("X-Signature")
	if !svc.GitOpsVerifyHMAC(sig, body) {
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

	if err := svc.GitOpsSyncCallback(ctx, param); err != nil {
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

	status, err := middlewares.NewServicesFromContext(ctx).GitOpsGetAppStatus(ctx, appName)
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

	if err := middlewares.NewServicesFromContext(ctx).GitOpsTriggerSyncByPipeline(ctx, req.PipelineID); err != nil {
		global.Logger.Errorf("[GitOps] 手动触发同步失败: %v", err)
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "同步已触发"})
}
