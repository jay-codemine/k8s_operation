package kube_node

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/node"
)

type KubeNodeRouter struct{}

func NewKubeNodeRouter() *KubeNodeRouter {
	return &KubeNodeRouter{}
}

func (r *KubeNodeRouter) Inject(router *gin.RouterGroup) {
	n := v1.NewKubeNodeController()

	// 基础查询
	router.GET("/list", n.List)       // 列表：支持名称模糊、Label 选择器、可调度过滤、分页
	router.GET("/detail", n.Detail)   // 详情：Status/Capacity/Allocatable/Conditions/Summary
	router.GET("/pods", n.ListPods)   // 查看节点上运行的 Pods（支持分页/按命名空间过滤）
	router.GET("/metrics", n.Metrics) // 节点指标（需 metrics-server）
	router.GET("/events", n.Events)   // 节点事件

	// 调度/维护
	router.POST("/cordon", n.Cordon) // 标记不可调度
	router.POST("/drain", n.Drain)   // 驱逐节点上的可驱逐 Pod（维护/下线常用）

	// 标签/污点管理
	router.PATCH("/labels", n.PatchLabels) // 修改节点标签
	router.PATCH("/taints", n.PatchTaints) // 修改节点污点

	// YAML 管理
	router.GET("/yaml", n.Yaml)           // 获取 Node YAML 配置
	router.PUT("/apply_yaml", n.ApplyYaml) // 应用 Node YAML 配置
}
