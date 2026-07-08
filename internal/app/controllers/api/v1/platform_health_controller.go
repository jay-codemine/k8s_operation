package v1

import (
	"strconv"

	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PlatformHealthController 平台健康检查控制器
type PlatformHealthController struct {
	service *services.PlatformHealthService
}

func NewPlatformHealthController() *PlatformHealthController {
	return &PlatformHealthController{
		service: services.NewPlatformHealthService(),
	}
}

func NewPlatformHealthControllerWithFactory(factory *services.ClusterClientFactory) *PlatformHealthController {
	return &PlatformHealthController{
		service: services.NewPlatformHealthServiceWithFactory(factory),
	}
}

// GetFullHealth 获取完整平台健康状态
// @Summary 获取平台健康状态
// @Description 获取平台、集群、告警、任务队列和组件的完整健康状态
// @Tags Platform
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=services.FullPlatformHealth}
// @Router /api/v1/platform/health [get]
func (c *PlatformHealthController) GetFullHealth(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	health, err := c.service.GetFullHealth(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("获取健康状态失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(health)
}

// CheckComponent 检查单个组件健康状态
// @Summary 检查单个组件
// @Description 检查指定组件的健康状态
// @Tags Platform
// @Accept json
// @Produce json
// @Param component path string true "组件名称" Enums(database, redis, kubernetes)
// @Success 200 {object} response.Response{data=services.ComponentStatus}
// @Router /api/v1/platform/health/component/{component} [get]
func (c *PlatformHealthController) CheckComponent(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	component := ctx.Param("component")
	if component == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	status, err := c.service.CheckComponentHealth(ctx.Request.Context(), component)
	if err != nil {
		global.Logger.Error("检查组件健康失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}
	resp.Success(status)
}

// Ping 简单存活检查
// @Summary 存活检查
// @Description 简单的存活检查接口
// @Tags Platform
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/platform/health/ping [get]
func (c *PlatformHealthController) Ping(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if err := c.service.Ping(ctx.Request.Context()); err != nil {
		global.Logger.Error("健康检查失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(gin.H{
		"status":  "ok",
		"message": "pong",
	})
}

// CheckClusterConnectivity 检查单个集群连通性
// @Summary 检查集群连通性
// @Description 检查指定集群的连通性状态
// @Tags Platform
// @Accept json
// @Produce json
// @Param cluster_id path int true "集群ID"
// @Success 200 {object} response.Response{data=services.ClusterConnectivityResult}
// @Router /api/v1/platform/health/cluster/{cluster_id}/connectivity [get]
func (c *PlatformHealthController) CheckClusterConnectivity(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	clusterIDStr := ctx.Param("cluster_id")
	if clusterIDStr == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil || clusterID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	result, err := c.service.CheckClusterConnectivity(ctx.Request.Context(), clusterID)
	if err != nil {
		global.Logger.Error("检查集群连通性失败", zap.Int64("cluster_id", clusterID), zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(result)
}
