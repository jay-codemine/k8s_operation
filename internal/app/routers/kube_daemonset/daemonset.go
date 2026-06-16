package kube_daemonset

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/daemonset"
)

type KubeDaemonSetRouter struct{}

func NewKubeDaemonSetRouter() *KubeDaemonSetRouter {
	return &KubeDaemonSetRouter{}
}

func (r *KubeDaemonSetRouter) Inject(router *gin.RouterGroup) {
	ds := v1.NewKubeDaemonSetController()
	{
		router.POST("/create", ds.Create)
		router.POST("/create-from-yaml", ds.CreateFromYaml) // 从 YAML 创建 DaemonSet
		router.GET("/list", ds.List)
		router.GET("/detail", ds.Detail)
		router.DELETE("/delete", ds.Delete)
		router.DELETE("/delete_service", ds.DeleteService)
		router.PUT("/update_image", ds.UpdateImage)
		router.POST("/restart", ds.Restart)
		router.POST("/rollback", ds.Rollback)
		router.GET("/ds_pods", ds.Pods)
		router.GET("/history", ds.History)
		router.POST("/events", ds.Events)
		router.GET("/yaml", ds.Yaml)             // 获取 DaemonSet YAML 配置
		router.PUT("/apply_yaml", ds.ApplyYaml)  // 应用 DaemonSet YAML 配置
	}
}
