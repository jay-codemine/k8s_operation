package kube_cronjob

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/cronjob"
)

type KubeCronJobRouter struct{}

func NewKubeCronJobRouter() *KubeCronJobRouter {
	return &KubeCronJobRouter{}
}

// Inject 注册 CronJob 相关路由
func (r *KubeCronJobRouter) Inject(router *gin.RouterGroup) {
	cj := v1.NewKubeCronJobController()

	// 基础 CRUD
	router.POST("/create", cj.Create)               // 创建 CronJob
	router.POST("/create-from-yaml", cj.CreateFromYaml)  // 从 YAML 创建 CronJob
	router.PUT("/update-from-yaml", cj.UpdateFromYaml)   // 从 YAML 更新 CronJob
	router.GET("/list", cj.List)                    // 列表（支持 ns/name 过滤、分页、排序）
	router.GET("/detail", cj.Detail)                // 详情
	router.DELETE("/delete", cj.Delete)             // 删除 CronJob（可选同时清理历史 Job）
	// 调度与运行控制
	router.PUT("/suspend", cj.Suspend)   // 暂停/恢复：spec.suspend = true/false
	router.POST("/trigger", cj.Trigger)  // 手动触发：立即创建一个 Job
}
