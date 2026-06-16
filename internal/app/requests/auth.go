package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

type AuthLoginRequest struct {
	Username string `json:"username" form:"username" valid:"username"`
	Password string `json:"password" form:"password" valid:"password"`
}

func NewAuthLoginRequest() *AuthLoginRequest {
	return &AuthLoginRequest{}
}

func ValidAuthLoginRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"username": []string{"required"},
		"password": []string{"required", "min:6"},
	}
	messages := govalidator.MapData{
		"username": []string{
			"required: 用户名为必填字段,字段为 username",
		},
		"password": []string{
			"required: 密码为必填字段,字段为 password",
			"min:密码长度需大于 6",
		},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, messages)
}

func NewAuthRegisterRequest() *AuthRegisterRequest {
	return &AuthRegisterRequest{}
}

type AuthRegisterRequest struct {
	Username        string `json:"username" form:"username" valid:"username"`
	Password        string `json:"password" form:"password" valid:"password"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm" valid:"password_confirm"`
}

func ValidAuthRegisterRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"username":         []string{"required"},
		"password":         []string{"required", "min:6"},
		"password_confirm": []string{"required"},
	}

	messages := govalidator.MapData{
		"username": []string{"required: 用户名为必填字段"},
		"password": []string{
			"required: 密码为必填字段",
			"min: 密码长度需大于 6",
		},
		"password_confirm": []string{"required: 确认密码为必填字段"},
	}

	// ① 先做规则校验
	errs := valid.ValidateOptions(data, rules, messages)
	if len(errs) > 0 {
		return errs
	}

	// ② 跨字段校验
	req := data.(*AuthRegisterRequest)
	if req.Password != req.PasswordConfirm {
		return map[string][]string{
			"password_confirm": {"两次输入的密码不一致"},
		}
	}

	return nil
}

func NewAuthForgotPasswordRequest() *AuthForgotPasswordRequest {
	return &AuthForgotPasswordRequest{}
}

type AuthForgotPasswordRequest struct {
	Username    string `json:"username" form:"username" valid:"username"`
	NewPassword string `json:"new_password" form:"new_password" valid:"new_password"`
	Confirm     string `json:"confirm" form:"confirm" valid:"confirm"`
}

func ValidAuthForgotPasswordRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"username":     []string{"required"},
		"new_password": []string{"required", "min:6"},
		"confirm":      []string{"required", "min:6"},
	}
	messages := govalidator.MapData{
		"username": []string{
			"required: 用户名为必填字段,字段为 username",
		},
		"new_password": []string{
			"required: 新密码为必填字段,字段为 new_password",
			"min:新密码长度需大于 6",
		},
		"confirm": []string{
			"required: 确认密码为必填字段,字段为 confirm",
			"min:确认密码长度需大于 6",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}
