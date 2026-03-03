package middlewares

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	jwt2 "k8soperation/pkg/jwt"
)

// Auth认证中间件
// 建议放到 internal/app/routers 或 middlewares 里
// AuthJWT 是一个JWT认证的中间件函数，用于验证请求中的JWT令牌
// 它返回一个gin.HandlerFunc，可以在Gin路由中使用
func AuthJWT() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		// ================== public 路由跳过 ==================
		if skip, ok := ctx.Get("skip_jwt"); ok {
			if b, ok := skip.(bool); ok && b {
				ctx.Next()
				return
			}
		}
		// =====================================================

		rsp := response.NewResponse(ctx) // 创建响应对象，用于返回错误信息

		// 1) 从 Header 取 Bearer token
		tokenStr, err := jwt2.GetTokenFromHeader(ctx)
		if err != nil {
			// 头部缺失 / 格式不对
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		// 2) 解析/验签（用新的 Manager）
		m := jwt2.NewManager() // 想复用可提到包级：var defaultJWT = jwt.NewManager()
		claims, err := m.ParseToken(tokenStr)
		if err != nil {
			// 区分错误（可选：用 UnauthorizedTokenTimeout 等更细错误码）
			switch err {
			case errorcode.ErrTokenExpired:
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError) // 或 UnauthorizedTokenTimeout
			default:
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			}
			ctx.Abort()
			return
		}

		// 3) 按用户ID查库，确保用户存在/可用
		u := models.NewUser().GetUserByID(claims.UserID)
		if u.ID == 0 {
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		// 4) 将 claims 和用户写入上下文，供后续 handler 使用
		// 设置当前用户ID到上下文中
		ctx.Set("current_user_id", u.GetStringID())
		// 设置当前用户名到上下文中
		ctx.Set("current_user_name", u.Username)
		// 设置当前用户对象到上下文中
		ctx.Set("current_user", u)

		ctx.Next()

	}
}

// AuthJWTSkip 用于 public 路由，标记该请求跳过 JWT 校验
func AuthJWTSkip() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1给当前请求打一个标记
		c.Set("skip_jwt", true)

		// 2.放行，继续执行后面的中间件 / handler
		c.Next()
	}
}
