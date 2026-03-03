package user

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/user"
)

type UserRouter struct{}

func NewUserRouter() *UserRouter {
	return &UserRouter{}
}

func (r *UserRouter) Inject(router *gin.RouterGroup) {
	uc := v1.NewUserController()

	// 建议统一前缀 /user
	g := router.Group("/user")
	{
		g.POST("/create", uc.Create) // /api/v1/user/create
		g.POST("/delete", uc.Delete) // /api/v1/user/delete
		g.POST("/update", uc.Update) // /api/v1/user/update
		g.GET("/list", uc.List)      // /api/v1/user/list
	}
}
