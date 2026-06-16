package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1"
	"k8soperation/internal/app/services"
)

type AppStoreRouter struct {
	factory *services.ClusterClientFactory
}

func NewAppStoreRouter() *AppStoreRouter {
	return &AppStoreRouter{}
}

func NewAppStoreRouterWithFactory(factory *services.ClusterClientFactory) *AppStoreRouter {
	return &AppStoreRouter{factory: factory}
}

func (r *AppStoreRouter) Inject(router *gin.RouterGroup) {
	c := v1.NewAppStoreControllerWithFactory(r.factory)

	g := router.Group("/platform/appstore")
	{
		// 应用管理
		g.GET("/list", c.List)             // GET /api/v1/platform/appstore/list
		g.GET("/detail/:id", c.Detail)     // GET /api/v1/platform/appstore/detail/:id
		g.GET("/categories", c.Categories) // GET /api/v1/platform/appstore/categories
		g.POST("/create", c.Create)        // POST /api/v1/platform/appstore/create
		g.PUT("/update", c.Update)         // PUT /api/v1/platform/appstore/update
		g.DELETE("/delete/:id", c.Delete)  // DELETE /api/v1/platform/appstore/delete/:id

		// 安装管理
		g.POST("/install", c.Install)                  // POST /api/v1/platform/appstore/install
		g.POST("/uninstall/:id", c.Uninstall)          // POST /api/v1/platform/appstore/uninstall/:id
		g.GET("/installs", c.InstallList)               // GET /api/v1/platform/appstore/installs
		g.GET("/installs/:id", c.InstallDetail)         // GET /api/v1/platform/appstore/installs/:id
		g.GET("/installs/:id/status", c.InstallStatus)  // GET /api/v1/platform/appstore/installs/:id/status
		g.PUT("/installs/:id/update", c.InstallUpdate)   // PUT /api/v1/platform/appstore/installs/:id/update

		// 组件管理
		g.GET("/components/:id", c.ComponentList)              // GET /api/v1/platform/appstore/components/:id (按应用ID)
		g.POST("/components/create", c.ComponentCreate)        // POST /api/v1/platform/appstore/components/create
		g.PUT("/components/update", c.ComponentUpdate)          // PUT /api/v1/platform/appstore/components/update
		g.DELETE("/components/delete/:comp_id", c.ComponentDelete) // DELETE /api/v1/platform/appstore/components/delete/:comp_id
		g.POST("/components/batch-delete", c.ComponentBatchDelete) // POST /api/v1/platform/appstore/components/batch-delete
		g.PUT("/components/sort", c.ComponentSort)                  // PUT /api/v1/platform/appstore/components/sort
	}
}
