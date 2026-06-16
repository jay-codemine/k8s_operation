package kube_statefulset

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/statefulset"
)

type KubeStatefulSetmentRouter struct{}

func NewKubeStatefulSetmentRouter() *KubeStatefulSetmentRouter {
	return &KubeStatefulSetmentRouter{}
}

// Inject 注册 StatefulSet 相关路由
func (r *KubeStatefulSetmentRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	statefulset := v1.NewKubeStatefulSetController()

	// 注册路由
	{
		router.GET("/list", statefulset.List)              // 获取 StatefulSet 列表
		router.GET("/detail", statefulset.Detail)          // 获取 StatefulSet 详情
		router.POST("/create", statefulset.Create)         // 创建 StatefulSet
		router.POST("/create-from-yaml", statefulset.CreateFromYaml) // 从 YAML 创建 StatefulSet
		router.DELETE("/delete", statefulset.Delete)       // 删除 StatefulSet
		router.DELETE("/delete_svc", statefulset.DeleteService) // 删除 Service
		router.PUT("/scale", statefulset.Scale)            // 扩缩容
		router.PUT("/update_image", statefulset.UpdateImage)    // 更新镜像
		router.POST("/restart", statefulset.Restart)       // 滚动重启
		router.GET("/sts_pods", statefulset.PodList)       // 获取关联 Pod 列表
		router.POST("/events", statefulset.EventList)      // 获取事件
		router.GET("/history", statefulset.History)        // 获取历史版本
		router.POST("/rollback", statefulset.Rollback)     // 回滚到指定版本
		router.GET("/yaml", statefulset.Yaml)              // 获取 YAML 配置
		router.PUT("/apply_yaml", statefulset.ApplyYaml)   // 应用 YAML 配置
	}
}
