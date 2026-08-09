package v1

import (
	"strconv"

	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"

	"github.com/gin-gonic/gin"
)

// PlatformHealthController 平台健康检查控制器
type PlatformHealthController struct {
	factory *services.ClusterClientFactory
}

func NewPlatformHealthController() *PlatformHealthController {
	return &PlatformHealthController{}
}

func NewPlatformHealthControllerWithFactory(factory *services.ClusterClientFactory) *PlatformHealthController {
	return &PlatformHealthController{factory: factory}
}

// GetFullHealth 获取完整平台健康状态
func (c *PlatformHealthController) GetFullHealth(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	health, err := middlewares.NewServicesFromContext(ctx).PlatformHealthSvc(c.factory).GetFullHealth(ctx.Request.Context())
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(health)
}

// CheckComponent 检查单个组件健康状态
func (c *PlatformHealthController) CheckComponent(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	component := ctx.Param("component")
	if component == "" {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	status, err := middlewares.NewServicesFromContext(ctx).PlatformHealthSvc(c.factory).CheckComponentHealth(ctx.Request.Context(), component)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}
	resp.Success(status)
}

// Ping 简单存活检查
func (c *PlatformHealthController) Ping(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if err := middlewares.NewServicesFromContext(ctx).PlatformHealthSvc(c.factory).Ping(ctx.Request.Context()); err != nil {
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(gin.H{
		"status":  "ok",
		"message": "pong",
	})
}

// CheckClusterConnectivity 检查单个集群连通性
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

	result, err := middlewares.NewServicesFromContext(ctx).PlatformHealthSvc(c.factory).CheckClusterConnectivity(ctx.Request.Context(), clusterID)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError)
		return
	}
	resp.Success(result)
}
