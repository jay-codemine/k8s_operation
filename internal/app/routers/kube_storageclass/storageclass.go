package kube_storageclass

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/storageclass"
)

// KubeStorageClassRouter 封装 StorageClass 路由注册
type KubeStorageClassRouter struct{}

// 构造函数
func NewKubeStorageClassRouter() *KubeStorageClassRouter {
	return &KubeStorageClassRouter{}
}

// Inject 注册 StorageClass 相关路由
func (r *KubeStorageClassRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	storage := v1.NewKubeStorageClassController()

	// 注册路由
	{
		router.POST("/create", storage.Create)               // 创建 StorageClass
		router.GET("/list", storage.List)                    // 获取 StorageClass 列表
		router.GET("/detail", storage.Detail)                // 获取 StorageClass 详情
		router.DELETE("/delete", storage.Delete)             // 删除 StorageClass
		router.GET("/yaml", storage.GetYaml)                 // 获取 StorageClass YAML
		router.POST("/create-from-yaml", storage.CreateFromYaml) // 从 YAML 创建
		router.POST("/apply-yaml", storage.ApplyYaml)        // 应用 YAML
	}
}
