package kube_configmap

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/configmap"
)

// KubeConfigMapRouter 封装 ConfigMap 路由注册
type KubeConfigMapRouter struct{}

// 构造函数
func NewKubeConfigMapRouter() *KubeConfigMapRouter {
	return &KubeConfigMapRouter{}
}

// Inject 注册 ConfigMap 相关路由
func (r *KubeConfigMapRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	configmap := v1.NewKubeConfigMapController()

	// 注册路由
	{
		router.POST("/create", configmap.Create)           // 创建 ConfigMap
		router.GET("/list", configmap.List)                // 获取 ConfigMap 列表
		router.GET("/detail", configmap.Detail)            // 获取 ConfigMap 详情
		router.GET("/yaml", configmap.Yaml)                // 获取 ConfigMap YAML
		router.DELETE("/delete", configmap.Delete)         // 删除 ConfigMap
		router.PATCH("/patch", configmap.Patch)            // Strategic Merge Patch 更新
		router.PUT("/update-data", configmap.UpdateData)   // 更新 ConfigMap data 字段
		router.POST("/apply-yaml", configmap.ApplyYaml)    // 从 YAML 更新 ConfigMap
		router.POST("/patch-json", configmap.PatchJSON)    // JSON Merge Patch 更新
	}
}
