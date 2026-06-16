package kube_namespace

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/namespace"
)

type KubeNamespaceRouter struct{}

func NewKubeNamespaceRouter() *KubeNamespaceRouter {
	return &KubeNamespaceRouter{}
}

func (r *KubeNamespaceRouter) Inject(router *gin.RouterGroup) {
	n := v1.NewKubeNamespaceController()
	router.POST("/create", n.Create)     // 创建 Namespace
	router.GET("/list", n.List)          // 获取 Namespace 列表
	router.GET("/detail", n.Detail)      // 获取 Namespace 详情
	router.DELETE("/delete", n.Delete)   // 删除 Namespace（危险操作）
	router.PATCH("/labels", n.PatchLabels) // 修改 Namespace 标签

	// 如果你希望支持 label/annotation patch：
	router.PATCH("/patch", n.Patch) // 更新 Namespace 的 labels/annotations

	// YAML 管理
	router.GET("/yaml", n.Yaml)           // 获取 Namespace YAML 配置
	router.PUT("/apply_yaml", n.ApplyYaml) // 应用 Namespace YAML 配置
}
