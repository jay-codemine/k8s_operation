package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1"
)

type PlatformSettingsRouter struct{}

func NewPlatformSettingsRouter() *PlatformSettingsRouter {
	return &PlatformSettingsRouter{}
}

func (r *PlatformSettingsRouter) Inject(router *gin.RouterGroup) {
	c := v1.NewPlatformSettingsController()

	g := router.Group("/platform/settings")
	{
		g.GET("", c.Get)          // GET /api/v1/platform/settings
		g.PUT("", c.Update)       // PUT /api/v1/platform/settings
		g.POST("/reset", c.Reset) // POST /api/v1/platform/settings/reset
	}
}
