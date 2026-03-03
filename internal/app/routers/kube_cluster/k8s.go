package kube_cluster

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/k8s_cluster"
)

type KubeRouter struct{}

func NewKubeRouter() *KubeRouter {
	return &KubeRouter{}
}

func (r *KubeRouter) Inject(router *gin.RouterGroup) {
	kc := v1.NewK8sClusterController()
	router.POST("/create", kc.Create)
	router.POST("/update", kc.Update)
	router.POST("/delete", kc.Delete)
	router.GET("/list", kc.List)
	router.POST("/init", kc.Init)
}
