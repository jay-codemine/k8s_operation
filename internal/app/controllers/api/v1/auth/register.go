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

// Register godoc
// @Summary 用户注册
// @Description 用户注册接口：校验参数后创建用户
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body requests.AuthRegisterRequest true "注册参数"
// @Success 200 {object} response.Response "注册成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 409 {object} errorcode.Error "用户已存在/冲突"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/auth/register [post]
func (a *AuthController) Register(ctx *gin.Context) {
	param := requests.NewAuthRegisterRequest()
	resp := response.NewResponse(ctx)

	// 1) 参数校验（required/min 等）
	if ok := valid.Validate(ctx, param, requests.ValidAuthRegisterRequest); !ok {
		return
	}

	// 2) 调用业务层
	svc := services.NewServices()
	if err := svc.AuthRegister(param); err != nil {
		global.Logger.Error("注册失败", zap.Error(err))

		// 业务错误（带 details）：透传
		if ec, ok := err.(*errorcode.Error); ok {
			resp.ToErrorResponse(ec)
			return
		}

		// 兜底错误
		resp.ToErrorResponse(errorcode.ErrorUserCreateFail)
		return
	}

	resp.Success("注册成功")
}
