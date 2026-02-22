package helloword

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/hello_world"
)

type HelloWorldRouter struct {
}

func NewHelloWorldRouter() *HelloWorldRouter {
	return &HelloWorldRouter{}
}

func (r *HelloWorldRouter) Inject(router *gin.RouterGroup) {
	hc := new(hello_world.HelloWorldController)

	router.GET("/hello", hc.Get)
}
