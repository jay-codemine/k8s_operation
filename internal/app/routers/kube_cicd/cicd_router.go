package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
)

type CicdRouter struct {
	ctrl *cicd.CicdReleaseController
}

func NewCicdRouter() *CicdRouter {
	return &CicdRouter{
		ctrl: cicd.NewCicdReleaseController(),
	}
}

func (r *CicdRouter) Inject(rg *gin.RouterGroup) {
	// 发布单相关
	release := rg.Group("/release")
	{
		release.POST("/create", r.ctrl.Create)           // 创建发布单
		release.GET("/detail", r.ctrl.Detail)            // 发布单详情
		release.GET("/list", r.ctrl.List)                // 发布单列表
		release.POST("/cancel", r.ctrl.Cancel)           // 取消发布（智能判断回滚/取消）
		release.POST("/rollback", r.ctrl.Rollback)       // 回滚发布
		release.POST("/retry", r.ctrl.Retry)             // 重试发布
		release.GET("/tasks", r.ctrl.Tasks)              // 获取发布单下的任务列表
	}

	// 回调接口（可选：给 Jenkins/外部系统回调用）
	callback := rg.Group("/callback")
	{
		callback.POST("/build", r.ctrl.BuildCallback)    // Jenkins 构建回调
	}
}
