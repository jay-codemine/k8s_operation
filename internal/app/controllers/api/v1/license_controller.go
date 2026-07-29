package v1

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/license"
)

// LicenseController License 授权控制器
type LicenseController struct{}

// NewLicenseController 创建 License 授权控制器
func NewLicenseController() *LicenseController {
	return &LicenseController{}
}

// activateRequest 激活请求体
type activateRequest struct {
	License string `json:"license" binding:"required"`
}

// Status 查询当前授权状态
// @Summary 查询 License 授权状态
// @Description 返回授权状态、被授权方、到期时间与本机机器码（机器码用于向供应商申请 License）
// @Tags License授权
// @Produce json
// @Success 200 {object} license.Status
// @Router /api/v1/platform/license/status [get]
func (c *LicenseController) Status(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	resp.Success(license.GetStatus())
}

// Activate 激活 License
// @Summary 激活 License
// @Description 校验 License（签名/机器码/有效期）并持久化，激活后立即生效
// @Tags License授权
// @Accept json
// @Produce json
// @Param body body activateRequest true "License 文本"
// @Success 200 {object} license.Status
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/platform/license/activate [post]
func (c *LicenseController) Activate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req activateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	p, err := license.Activate(req.License)
	if err != nil {
		global.Logger.Warn("License 激活失败", zap.Error(err))
		switch {
		case errors.Is(err, license.ErrMachineMismatch):
			resp.ToErrorResponse(errorcode.LicenseMachineErr)
		case errors.Is(err, license.ErrExpired):
			resp.ToErrorResponse(errorcode.LicenseExpired)
		case errors.Is(err, license.ErrFormat), errors.Is(err, license.ErrSignature),
			errors.Is(err, license.ErrNotYetValid):
			resp.ToErrorResponse(errorcode.LicenseInvalid.WithDetails(err.Error()))
		default:
			resp.ToErrorResponse(errorcode.LicenseActivateFailed.WithDetails(err.Error()))
		}
		return
	}

	global.Logger.Info("License 激活成功",
		zap.String("licensee", p.Licensee),
		zap.String("edition", p.Edition),
		zap.Int64("expire_at", p.ExpireAt))
	resp.Success(license.GetStatus())
}
