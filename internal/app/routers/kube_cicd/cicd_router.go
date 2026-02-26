package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
)

type CicdRouter struct {
	releaseCtrl  *cicd.CicdReleaseController
	pipelineCtrl *cicd.PipelineController
}

func NewCicdRouter() *CicdRouter {
	return &CicdRouter{
		releaseCtrl:  cicd.NewCicdReleaseController(),
		pipelineCtrl: cicd.NewPipelineController(),
	}
}

func (r *CicdRouter) Inject(rg *gin.RouterGroup) {
	// ==================== 流水线管理 ====================
	// /api/v1/k8s/cicd/pipeline/...
	pipeline := rg.Group("/pipeline")
	{
		pipeline.GET("/list", r.pipelineCtrl.List)         // 获取流水线列表
		pipeline.GET("/detail", r.pipelineCtrl.Detail)     // 获取流水线详情
		pipeline.POST("/create", r.pipelineCtrl.Create)    // 创建流水线
		pipeline.POST("/update", r.pipelineCtrl.Update)    // 更新流水线
		pipeline.POST("/delete", r.pipelineCtrl.Delete)    // 删除流水线
		pipeline.POST("/run", r.pipelineCtrl.Run)          // 运行流水线（触发Jenkins构建）
		pipeline.POST("/stop", r.pipelineCtrl.Stop)        // 停止流水线
		pipeline.GET("/logs", r.pipelineCtrl.Logs)         // 获取构建日志
		pipeline.GET("/status", r.pipelineCtrl.Status)     // 获取实时状态
		pipeline.GET("/history", r.pipelineCtrl.History)   // 获取运行历史
		pipeline.POST("/callback", r.pipelineCtrl.JenkinsCallback) // Jenkins状态回调
	}

	// ==================== 发布单管理 ====================
	// /api/v1/k8s/cicd/release/...
	release := rg.Group("/release")
	{
		release.POST("/create", r.releaseCtrl.Create)     // 创建发布单
		release.GET("/detail", r.releaseCtrl.Detail)      // 发布单详情
		release.GET("/list", r.releaseCtrl.List)          // 发布单列表
		release.POST("/cancel", r.releaseCtrl.Cancel)     // 取消发布（智能判断回滚/取消）
		release.POST("/rollback", r.releaseCtrl.Rollback) // 回滚发布
		release.POST("/retry", r.releaseCtrl.Retry)       // 重试发布
		release.GET("/tasks", r.releaseCtrl.Tasks)        // 获取发布单下的任务列表
	}

	// ==================== 回调接口 ====================
	// /api/v1/k8s/cicd/callback/...
	callback := rg.Group("/callback")
	{
		callback.POST("/build", r.releaseCtrl.BuildCallback) // Jenkins 构建回调（发布单用）
	}
}
