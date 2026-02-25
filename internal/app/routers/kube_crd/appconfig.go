package kube_crd

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/crd"
)

// KubeAppConfigRouter 封装 AppConfig 路由注册
type KubeAppConfigRouter struct{}

// 构造函数
func NewKubeAppConfigRouter() *KubeAppConfigRouter {
	return &KubeAppConfigRouter{}
}

// Inject 注册 AppConfig 相关路由
func (r *KubeAppConfigRouter) Inject(router *gin.RouterGroup) {
	ac := v1.NewKubeAppConfigController()

	// 基础 CRUD
	router.POST("/create", ac.Create)   // 创建 AppConfig
	router.PUT("/update", ac.Update)    // 更新 AppConfig
	router.GET("/detail", ac.Detail)    // 获取单个 AppConfig 详情
	router.GET("/list", ac.List)        // 列表（支持 ns/name 过滤、分页）
	router.DELETE("/delete", ac.Delete) // 删除 AppConfig
}
