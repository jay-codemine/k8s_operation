package cicd

import (
	"io"

	"github.com/gin-gonic/gin"

	"k8soperation/global"
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
