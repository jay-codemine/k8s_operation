package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
)

// CicdCallbackRouter 公开的回调路由（不需要 JWT 认证）
// Jenkins 回调使用 HMAC 签名验证
type CicdCallbackRouter struct {
	pipelineCtrl *cicd.PipelineController
	stageCtrl    *cicd.StageController
	releaseCtrl  *cicd.CicdReleaseController
}

func NewCicdCallbackRouter() *CicdCallbackRouter {
	return &CicdCallbackRouter{
		pipelineCtrl: cicd.NewPipelineController(),
		stageCtrl:    cicd.NewStageController(),
		releaseCtrl:  cicd.NewCicdReleaseController(),
	}
}

func (r *CicdCallbackRouter) Inject(rg *gin.RouterGroup) {
	// 流水线回调
	// POST /api/v1/k8s/cicd/pipeline/callback
	pipeline := rg.Group("/pipeline")
	{
		pipeline.POST("/callback", r.pipelineCtrl.Callback)
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
}
