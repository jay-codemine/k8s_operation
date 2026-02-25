package hello_world

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/valid"
	"net/http"
)

type HelloWorldController struct{}

// @Summary HelloWorld
// @Produce json
// @Tags HelloWorld
// @Success 200
// @Router /api/ [get]
func (s *HelloWorldController) Get(ctx *gin.Context) {
	param := requests.HelloRequest{}
	if ok := valid.Validate(ctx, &param, requests.ValidHelloRequest); !ok {
		return
	}
	ctx.JSON(
		http.StatusOK, gin.H{
			"Data": fmt.Sprintf("Hello %s", param.Name),
		},
	)
}
