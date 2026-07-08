package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

type HelloRequest struct {
	Name string `json:"name,omitempty" valid:"name" form:"name"`
}

func ValidHelloRequest(data interface{}, ctx *gin.Context) map[string][]string {
	// 自定义验证规则
	rules := govalidator.MapData{
		"name": []string{"required"},
	}

	// 自定义验证出错时的提示
	messages := govalidator.MapData{
		"name": []string{
			"required: 名字为必填项,参数名称为name",
		},
	}

	// 进行校验
	return valid.ValidateOptions(data, rules, messages)
}
