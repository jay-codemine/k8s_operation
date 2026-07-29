package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1"
)

// LicenseRouter License 授权路由（公开：激活页需在未登录/未授权时可用）
type LicenseRouter struct{}

func NewLicenseRouter() *LicenseRouter {
	return &LicenseRouter{}
}

func (r *LicenseRouter) Inject(router *gin.RouterGroup) {
	c := v1.NewLicenseController()

	g := router.Group("/platform/license")
	{
		g.GET("/status", c.Status)     // GET  /api/v1/platform/license/status
		g.POST("/activate", c.Activate) // POST /api/v1/platform/license/activate
	}
}
