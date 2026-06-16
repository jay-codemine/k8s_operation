package kube_ingress

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/ingress"
)

// KubeIngressRouter 负责注册 Ingress 相关路由
type KubeIngressRouter struct{}

// NewKubeIngressRouter 构造函数
func NewKubeIngressRouter() *KubeIngressRouter {
	return &KubeIngressRouter{}
}

// Inject 注册 Ingress 相关路由
func (r *KubeIngressRouter) Inject(router *gin.RouterGroup) {
	// 创建控制器实例
	ingress := v1.NewKubeIngressController()

	// 注册路由组
	{
		router.POST("/create", ingress.Create)              // 创建 Ingress
		router.POST("/create-from-yaml", ingress.CreateFromYaml) // 从 YAML 创建 Ingress
		router.GET("/list", ingress.List)                   // 获取 Ingress 列表
		router.GET("/detail", ingress.Detail)               // 获取 Ingress 详情
		router.GET("/yaml", ingress.Yaml)                   // 获取 Ingress YAML
		router.PUT("/apply-yaml", ingress.ApplyYaml)        // 应用 YAML 更新
		router.PATCH("/patch", ingress.Patch)               // StrategicMergePatch 更新
		router.POST("/patch_json", ingress.PatchJSON)       // JSON MergePatch 更新
		router.DELETE("/delete", ingress.Delete)            // 删除 Ingress
		//router.GET("/controller", ingress.Controllers) // 获取可用 IngressClass 列表
		//router.GET("/hosts", ingress.Hosts)            // 获取已占用域名列表
		//router.GET("/tls-secrets", ingress.TlsSecrets) // 获取可用 TLS Secret 列表
	}
}
