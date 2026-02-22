package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// @Summary 忘记密码
// @Description 根据用户名重置密码
// @Tags 认证管理
// @Produce json
// @Param body body requests.AuthForgotPasswordRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/auth/forgot_password [post]
func (u *AuthController) ForgotPassword(ctx *gin.Context) {
	param := requests.NewAuthForgotPasswordRequest()
	resp := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidAuthForgotPasswordRequest); !ok {
		return
	}

	svc := services.NewServices()

	// 调用业务层
	if err := svc.UserForgotPassword(param); err != nil {
		global.Logger.Error("忘记密码失败", zap.String("error", err.Error()))

		// 业务错误：透传
		if ec, ok := err.(*errorcode.Error); ok {
			resp.ToErrorResponse(ec)
			return
		}

		// 兜底错误
		resp.ToErrorResponse(errorcode.ErrorUserUpdateFail)
		return
	}

	resp.Success("密码重置成功")
}
