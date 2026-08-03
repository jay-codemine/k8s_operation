package ldap

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	ldapclient "k8soperation/pkg/ldap"
)

// LDAPController LDAP 管理控制器
type LDAPController struct{}

func NewLDAPController() *LDAPController {
	return &LDAPController{}
}

// GetConfig 获取 LDAP 配置
// @Summary 获取 LDAP 配置
// @Description 获取当前 LDAP 认证配置（脱敏）
// @Tags LDAP管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.LDAPConfigResponse
// @Router /api/v1/ldap/config [get]
func (c *LDAPController) GetConfig(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)
	config := svc.LDAPGetConfig()
	resp.Success(config)
}

// TestConnection 测试 LDAP 连接
// @Summary 测试 LDAP 连接
// @Description 测试 LDAP 服务器是否可连接
// @Tags LDAP管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ldap/test [post]
func (c *LDAPController) TestConnection(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if !ldapclient.IsEnabled() {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("LDAP 未启用，请先在 config.yaml 中配置"))
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.LDAPTestConnection(); err != nil {
		global.Logger.Error("LDAP 连接测试失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"msg":    "LDAP 连接测试成功",
		"status": "connected",
	})
}

// SyncUsers 全量同步 LDAP 用户
// @Summary 全量同步 LDAP 用户
// @Description 从 LDAP 同步所有用户到平台
// @Tags LDAP管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.LDAPSyncResult
// @Router /api/v1/ldap/sync [post]
func (c *LDAPController) SyncUsers(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if !ldapclient.IsEnabled() {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("LDAP 未启用"))
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	result, err := svc.LDAPSyncAllUsers()
	if err != nil {
		global.Logger.Error("LDAP 用户同步失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(result)
}

// GetStatus 获取 LDAP 状态
// @Summary 获取 LDAP 状态
// @Description 获取 LDAP 启用状态和连接状态
// @Tags LDAP管理
// @Produce json
// @Router /api/v1/ldap/status [get]
func (c *LDAPController) GetStatus(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	status := gin.H{
		"enabled": ldapclient.IsEnabled(),
	}

	if ldapclient.IsEnabled() {
		client := ldapclient.NewClient()
		if err := client.TestConnection(); err != nil {
			status["connected"] = false
			status["error"] = err.Error()
		} else {
			status["connected"] = true
		}
		status["host"] = global.LDAPSetting.Host
		status["port"] = global.LDAPSetting.Port
		status["auto_create"] = global.LDAPSetting.AutoCreate
		status["sync_on_login"] = global.LDAPSetting.SyncOnLogin
	}

	resp.Success(status)
}
