package kube_secret

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/secret"
)

// KubeSecretRouter 封装 Secret 路由注册
type KubeSecretRouter struct{}

// 构造函数
func NewKubeSecretRouter() *KubeSecretRouter {
	return &KubeSecretRouter{}
}

// Inject 注册 Secret 相关路由
func (r *KubeSecretRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	secret := v1.NewKubeSecretController()

	// 注册路由
	{
		router.POST("/create", secret.Create)                  // 创建 Secret
		router.POST("/create-from-yaml", secret.CreateFromYaml) // 从 YAML 创建 Secret
		router.GET("/list", secret.List)                        // 获取 Secret 列表
		router.GET("/detail", secret.Detail)                    // 获取 Secret 详情
		router.GET("/yaml", secret.Yaml)                        // 获取 Secret YAML
		router.DELETE("/delete", secret.Delete)                 // 删除 Secret
		router.PATCH("/patch", secret.Patch)                    // Strategic Merge Patch 更新 Secret
		router.POST("/patch_json", secret.PatchJSON)            // JSON Merge Patch 更新 Secret
		router.POST("/decode", secret.Decode)                   // Base64 解码 Secret 内容
		router.PUT("/apply-yaml", secret.ApplyYaml)             // 应用 YAML 更新 Secret
	}
}
