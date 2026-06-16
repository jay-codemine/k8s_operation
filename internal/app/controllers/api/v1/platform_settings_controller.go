package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// PlatformSettingsController 平台设置控制器
type PlatformSettingsController struct{}

// NewPlatformSettingsController 创建平台设置控制器
func NewPlatformSettingsController() *PlatformSettingsController {
	return &PlatformSettingsController{}
}

// Get 获取平台设置
// @Summary 获取平台设置
// @Description 获取所有平台配置项
// @Tags 平台管理
// @Produce json
// @Success 200 {object} models.PlatformSettingsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/platform/settings [get]
func (c *PlatformSettingsController) Get(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := services.NewServices()

	settings, err := svc.PlatformSettingsGet(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("获取平台设置失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(settings)
}

// Update 更新平台设置
// @Summary 更新平台设置
// @Description 批量更新平台配置项
// @Tags 平台管理
// @Accept json
// @Produce json
// @Param body body models.PlatformSettingsResponse true "设置内容"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/platform/settings [put]
func (c *PlatformSettingsController) Update(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := services.NewServices()

	var req models.PlatformSettingsResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.Logger.Error("解析设置请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	if err := svc.PlatformSettingsUpdate(ctx.Request.Context(), &req); err != nil {
		global.Logger.Error("更新平台设置失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(gin.H{"message": "设置已保存"})
}

// Reset 重置为默认设置
// @Summary 重置平台设置
// @Description 将所有设置恢复为默认值
// @Tags 平台管理
// @Produce json
// @Success 200 {object} models.PlatformSettingsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/platform/settings/reset [post]
func (c *PlatformSettingsController) Reset(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	svc := services.NewServices()

	settings, err := svc.PlatformSettingsReset(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("重置平台设置失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}

	resp.Success(settings)
}
