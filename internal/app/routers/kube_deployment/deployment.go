package kube_deployment

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/deployment"
)

type KubeDeploymentRouter struct{}

func NewKubeDeploymentRouter() *KubeDeploymentRouter {
	return &KubeDeploymentRouter{}
}

// Inject 注册 Deployment 相关路由
func (r *KubeDeploymentRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	deployment := v1.NewKubeDeploymentController()

	// 注册路由
	{
		router.GET("/list", deployment.List)                       // 获取 Deployment 列表
		router.GET("/detail", deployment.Detail)                   // 获取 Deployment 详情
		router.POST("/create", deployment.Create)                  // 创建 Deployment（可选创建 Service）
		router.POST("/create-from-yaml", deployment.CreateFromYaml) // 从 YAML 创建 Deployment
		router.DELETE("/delete", deployment.Delete)                // 删除 Deployment
		router.DELETE("/delete_service", deployment.DeleteService) // 删除 Deployment Services
		router.POST("/scale", deployment.Scale)                    // 扩缩容（修改副本数）
		router.POST("/update-image", deployment.UpdateImage)       // 更新镜像
		router.POST("/patch_template", deployment.PatchTemplate)   // Patch 模板（JSONPatch / Merge）
		router.POST("/rollback", deployment.Rollback)              // 回滚到指定 ReplicaSet
		router.POST("/restart", deployment.Restart)                // 滚动重启 Deployment
		router.GET("/deploy_pods", deployment.DeploymentPodList)   // 获取 Deployment 对应的 Pod 列表
		router.POST("/events", deployment.EventList)               // 获取 Deployment 事件
		router.GET("/history", deployment.History)                 // 获取 Deployment 历史版本
		router.GET("/yaml", deployment.Yaml)                        // 获取 Deployment YAML 配置
		router.PUT("/apply_yaml", deployment.ApplyYaml)             // 应用 Deployment YAML 配置
	}
}
