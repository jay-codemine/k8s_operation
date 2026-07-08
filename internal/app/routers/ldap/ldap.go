package ldap

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/ldap"
)

type LDAPRouter struct{}

func NewLDAPRouter() *LDAPRouter {
	return &LDAPRouter{}
}

// Inject 注册需要认证的 LDAP 管理接口
func (r *LDAPRouter) Inject(router *gin.RouterGroup) {
	ctl := v1.NewLDAPController()

	g := router.Group("/ldap")
	{
		g.GET("/config", ctl.GetConfig)     // 获取 LDAP 配置（需管理员）
		g.POST("/test-connection", ctl.TestConnection) // 测试连接（需管理员）
		g.POST("/sync-users", ctl.SyncUsers) // 全量同步（需管理员）
	}
}

// LDAPPublicRouter 公开接口（无需JWT，用于登录页判断 LDAP 状态）
type LDAPPublicRouter struct{}

func NewLDAPPublicRouter() *LDAPPublicRouter {
	return &LDAPPublicRouter{}
}

func (r *LDAPPublicRouter) Inject(router *gin.RouterGroup) {
	ctl := v1.NewLDAPController()

	g := router.Group("/ldap")
	{
		g.GET("/status", ctl.GetStatus) // 获取 LDAP 状态（公开，用于登录页）
	}
}
