package platform

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1"
)

// AuditLogRouter 审计日志路由
type AuditLogRouter struct{}

func NewAuditLogRouter() *AuditLogRouter {
	return &AuditLogRouter{}
}

func (r *AuditLogRouter) Inject(router *gin.RouterGroup) {
	ctrl := v1.NewAuditLogController()

	g := router.Group("/platform/audit")
	{
		g.GET("/logs", ctrl.List)              // GET /api/v1/platform/audit/logs
		g.GET("/logs/:id", ctrl.Detail)        // GET /api/v1/platform/audit/logs/:id
		g.GET("/statistics", ctrl.Statistics)   // GET /api/v1/platform/audit/statistics
		g.GET("/retention", ctrl.GetRetention)  // GET /api/v1/platform/audit/retention
		g.PUT("/retention", ctrl.UpdateRetention) // PUT /api/v1/platform/audit/retention
		g.POST("/cleanup", ctrl.Cleanup)       // POST /api/v1/platform/audit/cleanup
		g.GET("/export", ctrl.Export)           // GET /api/v1/platform/audit/export
	}
}
