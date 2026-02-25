package kube_pvc

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/pvc"
)

type KubePersistentVolumeClaimRouter struct{}

func NewKubePersistentVolumeClaimRouter() *KubePersistentVolumeClaimRouter {
	return &KubePersistentVolumeClaimRouter{}
}

func (r *KubePersistentVolumeClaimRouter) Inject(router *gin.RouterGroup) {
	pvc := v1.NewKubePVCController()

	{
		// 基础 CRUD
		router.POST("/create", pvc.Create)   // 创建 PVC
		router.POST("/create-from-yaml", pvc.CreateFromYaml) // 从 YAML 创建 PVC
		router.GET("/list", pvc.List)        // 获取 PVC 列表（支持 namespace、name 模糊、分页等）
		router.GET("/detail", pvc.Detail)    // 获取 PVC 详情
		router.DELETE("/delete", pvc.Delete) // 删除 PVC
		//
		//// 常用改动
		router.PATCH("/resize", pvc.Resize) // 扩容（更新 spec.resources.requests.storage）
		router.PUT("/apply-yaml", pvc.ApplyYaml) // 应用 YAML 更改
	}
}
