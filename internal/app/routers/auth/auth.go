package auth

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/auth"
)

type AuthRouter struct{}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{}
}

func (r *AuthRouter) Inject(router *gin.RouterGroup) {
	ctl := v1.NewAuthController()

	g := router.Group("/auth")
	{
		g.POST("/login", ctl.Login)
		g.POST("/register", ctl.Register)
		g.POST("/logout", ctl.Logout)
		g.POST("/refresh", ctl.Refresh)
		g.POST("/forgot_password", ctl.ForgotPassword)

	}
}
