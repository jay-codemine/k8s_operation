package cicd

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/middlewares"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// QuickOnboardController 快速接入控制器
type QuickOnboardController struct{}

func NewQuickOnboardController() *QuickOnboardController {
	return &QuickOnboardController{}
}

// Onboard godoc
// @Summary 快速接入应用（一键创建 K8s 资源 + 接入流水线 + 可选首次部署）
// @Description 支持 5 种工作负载类型: Deployment/StatefulSet/DaemonSet/CronJob/Job
// @Tags CICD QuickOnboard
// @Accept json
// @Produce json
// @Param body body requests.QuickOnboardRequest true "接入参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/quick-onboard [post]
func (c *QuickOnboardController) Onboard(ctx *gin.Context) {
	param := requests.NewQuickOnboardRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidQuickOnboardRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := middlewares.NewServicesFromContext(ctx)
	result, err := svc.QuickOnboard(ctx.Request.Context(), param, userID)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("QuickOnboard error",
			zap.String("app_name", param.AppName),
			zap.String("workload_kind", param.WorkloadKind),
			zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseCreateFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(result)
}
