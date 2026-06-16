package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"net/http"
)

// Logout godoc
// @Summary 退出登录
// @Description 清空会话并让会话 Cookie 立即过期（注意：不会自动使现有 JWT 失效）
// @Tags 认证管理
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)

	// 记录一下当前 SessionID（便于日志追踪；可能为空）
	cookieName := global.CacheSetting.Name
	sessionID, _ := ctx.Cookie(cookieName)

	// 清空会话内的数据
	sess.Clear()

	// 让 Cookie 立刻过期：务必把关键选项一起设上，避免覆盖成零值
	sess.Options(sessions.Options{
		Path:     "/",   // 与发包时一致
		MaxAge:   -1,    // 立刻过期
		HttpOnly: true,  // 保持安全属性
		Secure:   false, // 本地HTTP调试 false；生产HTTPS请设 true
		// SameSite: http.SameSiteLaxMode, // 如生产跨域场景是 None，请按你的发包一致
	})

	if err := sess.Save(); err != nil {
		global.Logger.Error("logout failed", zap.String("sid", sessionID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "退出失败"})
		return
	}

	global.Logger.Info("logout success" /*, zap.String("sid", sessionID)*/)
	ctx.JSON(http.StatusOK, gin.H{"msg": "退出成功"})
}
