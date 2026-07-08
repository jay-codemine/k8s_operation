package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1"
	"k8soperation/internal/app/services"
)

type PlatformHealthRouter struct {
	factory *services.ClusterClientFactory
}

func NewPlatformHealthRouter() *PlatformHealthRouter {
	return &PlatformHealthRouter{}
}

func NewPlatformHealthRouterWithFactory(factory *services.ClusterClientFactory) *PlatformHealthRouter {
	return &PlatformHealthRouter{factory: factory}
}

func (r *PlatformHealthRouter) Inject(router *gin.RouterGroup) {
	c := v1.NewPlatformHealthControllerWithFactory(r.factory)

	g := router.Group("/platform/health")
	{
		g.GET("", c.GetFullHealth)                                     // GET /api/v1/platform/health
		g.GET("/ping", c.Ping)                                         // GET /api/v1/platform/health/ping
		g.GET("/component/:component", c.CheckComponent)               // GET /api/v1/platform/health/component/:component
		g.GET("/cluster/:cluster_id/connectivity", c.CheckClusterConnectivity) // GET /api/v1/platform/health/cluster/:cluster_id/connectivity
	}
}
