package middlewares

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/tenant"
)

// TenantScope 多租户中间件：将租户隔离的 DB 注入 context
// 后续 handler/DAO 通过 GetTenantDB(c) 获取自动过滤 tenant_id 的 DB
func TenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.DB == nil {
			c.Next()
			return
		}

		tidVal, exists := c.Get("tenant_id")
		if !exists {
			c.Next()
			return
		}

		tid, ok := tidVal.(uint32)
		if !ok || tid == 0 {
			c.Next()
			return
		}

		// 注入租户隔离 DB：所有查询自动 WHERE tenant_id = tid
		scopedDB := tenant.NewScopedDB(global.DB, tid)
		c.Set("db", scopedDB)
		c.Next()
	}
}

// GetTenantDB 从 gin context 获取租户隔离的 DB（兜底 global.DB）
func GetTenantDB(c *gin.Context) *gorm.DB {
	if db, exists := c.Get("db"); exists {
		if gdb, ok := db.(*gorm.DB); ok {
			return gdb
		}
	}
	return global.DB
}

// NewServicesFromContext 从 gin context 获取租户隔离 DB 并创建 Services
// Controller 层替换 services.NewServices() → middlewares.NewServicesFromContext(c)
func NewServicesFromContext(c *gin.Context) *services.Services {
	return services.NewServicesWithDB(GetTenantDB(c))
}
