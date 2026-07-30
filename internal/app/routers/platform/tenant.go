package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/platform"
)

type TenantRouter struct{}

func NewTenantRouter() *TenantRouter { return &TenantRouter{} }

func (r *TenantRouter) Inject(router *gin.RouterGroup) {
	c := v1.NewTenantController()
	router.GET("/tenants", c.List)
	router.POST("/tenants", c.Create)
	router.PUT("/tenants/:id", c.Update)
	router.DELETE("/tenants/:id", c.Delete)
}
