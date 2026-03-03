package kube_pod

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/pod"
)

type kubePodRouter struct{}

func NewkubePodRouter() *kubePodRouter {
	return &kubePodRouter{}
}

func (r *kubePodRouter) Inject(router *gin.RouterGroup) {
	// 创建一个新的Pod控制器实例
	pod := v1.NewPodController()

	// 设置HTTP路由和处理函数
	router.POST("/create", pod.Create)                              // 创建Pod
	router.POST("/create-from-yaml", pod.CreateFromYaml)           // 从 YAML 创建 Pod
	router.GET("/list", pod.List)                                   // 获取Pod列表
	router.GET("/detail", pod.Detail)                               // 获取Pod详情
	router.PUT("/update", pod.Update)                               // 更新Pod
	router.DELETE("/grace_delete_pod", pod.DeletePod)               // 删除Pod
	router.GET("/container_name", pod.GetContainerName)             // 获取容器名称
	router.GET("/init_container_name", pod.GetInitContainerName)    // 获取初始化容器名称
	router.GET("/container_image", pod.GetContainerImages)          // 获取容器镜像
	router.GET("/init_container_image", pod.GetInitContainerImages) // 获取初始化容器镜像
	router.GET("/container_logs", pod.GetContainerLog)              // 获取容器日志（一次性）
	router.GET("/container_log", pod.GetContainerLogs)              // 获取容器日志（支持实时流式）
	router.PUT("/patch_image", pod.PatchImage)                      // 更新Pod容器镜像
	router.PATCH("/labels", pod.PatchLabels)                        // 修改Pod标签
	router.POST("/evict", pod.Evict)                                // 驱逐指定 Pod（可指定原因）
	router.GET("/metrics", pod.Metrics)                             // 获取单个 Pod 的资源使用情况
	router.GET("/metrics/list", pod.MetricsList)                    // 批量获取命名空间下所有 Pod 的资源使用情况
	router.GET("/yaml", pod.Yaml)                                   // 获取 Pod 的 YAML 配置
	router.PUT("/apply_yaml", pod.ApplyYaml)                        // 应用 Pod YAML 配置

}
