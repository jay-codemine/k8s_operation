package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/metrics"
	jwt2 "k8soperation/pkg/jwt"
)

// recordUserOnline 将用户最近活跃时间写入 Redis ZSET（best-effort，失败静默）
func recordUserOnline(userID int64) {
	if global.RedisCli == nil || userID <= 0 {
		return
	}
	defer func() { _ = recover() }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	global.RedisCli.ZAdd(ctx, global.OnlineUsersKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: fmt.Sprintf("%d", userID),
	})
}

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
			metrics.AuthTokenValidationTotal.WithLabelValues("invalid").Inc()
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
				metrics.AuthTokenValidationTotal.WithLabelValues("expired").Inc()
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError) // 或 UnauthorizedTokenTimeout
			default:
				metrics.AuthTokenValidationTotal.WithLabelValues("invalid").Inc()
				rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			}
			ctx.Abort()
			return
		}

		// Token 校验成功
		metrics.AuthTokenValidationTotal.WithLabelValues("success").Inc()

		// 3) 按用户ID查库，确保用户存在/可用
		u := models.NewUser().GetUserByID(claims.UserID)
		if u.ID == 0 {
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		// 4) 将 claims 和用户写入上下文，供后续 handler 使用
		// 设置当前用户ID到上下文中（int64类型，用于RBAC权限检查）
		ctx.Set("user_id", int64(u.ID))
		// 设置当前用户ID字符串（兼容旧代码）
		ctx.Set("current_user_id", u.GetStringID())
		// 设置当前用户名到上下文中
		ctx.Set("current_user_name", u.Username)
		// 设置当前用户对象到上下文中
		ctx.Set("current_user", u)

		// 多租户：从 JWT 提取 tenant_id，兜底用用户记录
		tid := claims.TenantID
		if tid == 0 {
			tid = u.TenantID
		}
		if tid == 0 {
			tid = models.DefaultTenantID
		}
		ctx.Set("tenant_id", tid)

		// 5) 记录用户活跃时间（异步，用于"在线用户"统计，不阻塞请求）
		go recordUserOnline(int64(u.ID))

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

// RequireCICDPermission CICD 细粒度权限检查中间件
// 使用方式: router.POST("/run", middlewares.RequireCICDPermission("cicd:pipeline:run"), ctrl.Run)
func RequireCICDPermission(permissionName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetInt64("user_id")
		if userID <= 0 {
			rsp := response.NewResponse(ctx)
			rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
			ctx.Abort()
			return
		}

		// 检查细粒度权限
		if !models.HasUserPermission(global.DB, userID, permissionName) {
			rsp := response.NewResponse(ctx)
			rsp.ToErrorResponse(errorcode.ErrorRBACAccessDenied)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// CheckCICDPermission 控制器内使用的权限检查帮助函数
// 返回 true 表示有权限，false 表示无权限（已自动返回错误响应）
func CheckCICDPermission(ctx *gin.Context, permissionName string) bool {
	userID := ctx.GetInt64("user_id")
	if userID <= 0 {
		rsp := response.NewResponse(ctx)
		rsp.ToErrorResponse(errorcode.UnauthorizedTokenError)
		return false
	}
	if !models.HasUserPermission(global.DB, userID, permissionName) {
		rsp := response.NewResponse(ctx)
		rsp.ToErrorResponse(errorcode.ErrorRBACAccessDenied)
		return false
	}
	return true
}
