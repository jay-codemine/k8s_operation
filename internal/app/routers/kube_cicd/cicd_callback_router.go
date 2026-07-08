package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
)

// CicdCallbackRouter 公开的回调路由（不需要 JWT 认证）
// Jenkins 回调使用 HMAC 签名验证
// 飞书审批回调使用 Token 验证
type CicdCallbackRouter struct {
	pipelineCtrl *cicd.PipelineController
	stageCtrl    *cicd.StageController
	releaseCtrl  *cicd.CicdReleaseController
	artifactCtrl *cicd.ArtifactController
	approvalCtrl *cicd.ApprovalController
	gitOpsCtrl   *cicd.GitOpsController
}

func NewCicdCallbackRouter() *CicdCallbackRouter {
	return &CicdCallbackRouter{
		pipelineCtrl: cicd.NewPipelineController(),
		stageCtrl:    cicd.NewStageController(),
		releaseCtrl:  cicd.NewCicdReleaseController(),
		artifactCtrl: cicd.NewArtifactController(),
		approvalCtrl: cicd.NewApprovalController(),
		gitOpsCtrl:   cicd.NewGitOpsController(),
	}
}

func (r *CicdCallbackRouter) Inject(rg *gin.RouterGroup) {
	// 流水线回调
	// POST /api/v1/k8s/cicd/pipeline/callback
	// POST /api/v1/k8s/cicd/pipeline/sonar-callback （SonarQube 扫描结果回调）
	pipeline := rg.Group("/pipeline")
	{
		pipeline.POST("/callback", r.pipelineCtrl.Callback)
		pipeline.POST("/sonar-callback", r.pipelineCtrl.SonarCallback)
	}

	// 阶段回调
	// POST /api/v1/k8s/cicd/stage/callback
	stage := rg.Group("/stage")
	{
		stage.POST("/callback", r.stageCtrl.StageCallback)
	}

	// 构建回调（发布单用）
	// POST /api/v1/k8s/cicd/callback/build
	callback := rg.Group("/callback")
	{
		callback.POST("/build", r.releaseCtrl.BuildCallback)
	}

	// 制品上传回调（Jenkins 构建完成后自动上传制品）
	// POST /api/v1/k8s/cicd/artifact/upload
	artifact := rg.Group("/artifact")
	{
		artifact.POST("/upload", r.artifactCtrl.Upload)
	}

	// 飞书审批回调（公开接口，使用 Token 验证）
	// GET/POST /api/v1/k8s/cicd/approval/feishu-callback
	approval := rg.Group("/approval")
	{
		approval.GET("/feishu-callback", r.approvalCtrl.FeishuCallback)
		approval.POST("/feishu-callback", r.approvalCtrl.FeishuCallback)
	}

	// GitOps 回调（公开接口，HMAC 签名验证）
	gitops := rg.Group("/gitops")
	{
		gitops.POST("/sync-callback", r.gitOpsCtrl.SyncCallback)
		gitops.POST("/webhook", r.gitOpsCtrl.Webhook)
	}
}
