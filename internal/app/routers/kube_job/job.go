package kube_job

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/job" // ← 新增 job 控制器包
)

type KubeJobRouter struct{}

func NewKubeJobRouter() *KubeJobRouter { // ← 名称修正
	return &KubeJobRouter{}
}

// Inject 注册 Job 相关路由
func (r *KubeJobRouter) Inject(router *gin.RouterGroup) {
	job := v1.NewKubeJobController() // ← 创建 Job 控制器

	{
		// 基础 CRUD
		router.POST("/create", job.Create)   // 创建 Job
		router.POST("/create-from-yaml", job.CreateFromYaml) // 从 YAML 创建 Job
		router.GET("/list", job.List)        // 获取 Job 列表（支持 name/namespace 过滤 + 分页）
		router.GET("/detail", job.Detail)    // 获取 Job 详情
		router.DELETE("/delete", job.Delete) // 删除 Job（支持级联删除 Pod）

		// 运行管理
		router.PUT("/suspend", job.Suspend)  // 暂停 Job（spec.suspend=true） 恢复 Job（spec.suspend=false）
		router.POST("/restart", job.Restart) // 重新运行 Job（克隆/重建，见说明）
		
		// 镜像更新
		router.PUT("/update-image", job.UpdateImage) // 更新 Job 镜像

		// YAML 操作
		router.GET("/yaml", job.GetYaml)       // 获取 Job YAML
		router.PUT("/apply-yaml", job.ApplyYaml) // 应用 YAML 修改
	}
}
