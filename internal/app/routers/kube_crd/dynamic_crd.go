package kube_crd

import (
	"github.com/gin-gonic/gin"
	dynamiccrd "k8soperation/internal/app/controllers/api/v1/dynamiccrd"
)

// KubeDynamicCRDRouter CRD/CR 动态资源管理路由
type KubeDynamicCRDRouter struct{}

func NewKubeDynamicCRDRouter() *KubeDynamicCRDRouter {
	return &KubeDynamicCRDRouter{}
}

// Inject 注册 CRD/CR 动态资源路由
func (r *KubeDynamicCRDRouter) Inject(router *gin.RouterGroup) {
	ctrl := dynamiccrd.NewDynamicCRDController()

	// === CRD 管理 ===
	// GET /crd/list   - 列出所有 CRD
	// GET /crd/detail - 获取 CRD 详情
	// DELETE /crd/delete - 删除 CRD
	crdGroup := router.Group("/crd")
	{
		crdGroup.GET("/list", ctrl.ListCRDs)
		crdGroup.GET("/detail", ctrl.GetCRD)
		crdGroup.DELETE("/delete", ctrl.DeleteCRD)
	}

	// === CR 实例管理 ===
	// GET    /cr/list     - 列出 CR 实例
	// GET    /cr/detail   - 获取 CR 详情
	// POST   /cr/create   - 创建 CR
	// PUT    /cr/update   - 更新 CR
	// DELETE /cr/delete   - 删除 CR
	// GET    /cr/yaml     - 获取 CR YAML
	// POST   /cr/dry-run  - DryRun 校验
	crGroup := router.Group("/cr")
	{
		crGroup.GET("/list", ctrl.ListCRs)
		crGroup.GET("/detail", ctrl.GetCR)
		crGroup.POST("/create", ctrl.CreateCR)
		crGroup.PUT("/update", ctrl.UpdateCR)
		crGroup.DELETE("/delete", ctrl.DeleteCR)
		crGroup.GET("/yaml", ctrl.GetCRYaml)
		crGroup.POST("/dry-run", ctrl.DryRunCR)
	}
}
