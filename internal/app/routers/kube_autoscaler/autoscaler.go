package kube_autoscaler

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/autoscaler"
)

// ---------------- HPA Router ----------------

type KubeHPARouter struct{}

func NewKubeHPARouter() *KubeHPARouter {
	return &KubeHPARouter{}
}

// Inject 注册 HPA 相关路由
func (r *KubeHPARouter) Inject(router *gin.RouterGroup) {
	c := v1.NewKubeHPAController()
	router.GET("/list", c.List)                       // 列表
	router.GET("/detail", c.Detail)                   // 详情
	router.POST("/create", c.Create)                  // 创建
	router.POST("/create-from-yaml", c.CreateFromYaml) // 通过 YAML 创建
	router.POST("/update", c.Update)                  // 更新
	router.DELETE("/delete", c.Delete)                // 删除
	router.POST("/scale", c.Scale)                    // 单独修改副本数 min/max
	router.POST("/batch-scale", c.BatchScale)         // 批量扩缩容（618）
	router.POST("/batch-status", c.BatchStatus)       // 批量查询当前状态
}

// ---------------- VPA Router ----------------

type KubeVPARouter struct{}

func NewKubeVPARouter() *KubeVPARouter {
	return &KubeVPARouter{}
}

// Inject 注册 VPA 相关路由
func (r *KubeVPARouter) Inject(router *gin.RouterGroup) {
	c := v1.NewKubeVPAController()
	router.GET("/available", c.Available)
	router.GET("/list", c.List)
	router.GET("/detail", c.Detail)
	router.POST("/create", c.Create)
	router.POST("/create-from-yaml", c.CreateFromYaml) // 通过 YAML 创建
	router.POST("/update", c.Update)
	router.DELETE("/delete", c.Delete)
}
