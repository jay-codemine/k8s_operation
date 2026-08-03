package cicd

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// CanaryDeployController 金丝雀部署控制器
//
// factory 是启动期创建的共享实例（由 router 注入），内部持有集群客户端缓存，
// 不能在这里自建。业务用的 Services 一律按请求从 context 取租户隔离实例，
// 不长期持有 global.DB。
type CanaryDeployController struct {
	factory *services.ClusterClientFactory
}

// NewCanaryDeployController 创建金丝雀控制器
func NewCanaryDeployController(factory *services.ClusterClientFactory) *CanaryDeployController {
	return &CanaryDeployController{factory: factory}
}

// resolveCluster 解析目标集群客户端。返回 false 时响应已写好，调用方直接 return。
// 这些路由没有挂 ClusterMiddleware，所以必须走 ResolveClusterClients 补上
// 授权与租户校验，不能直接调 factory。
func (c *CanaryDeployController) resolveCluster(ctx *gin.Context, rsp *response.Response) (*services.K8sClients, bool) {
	clients, err := middlewares.ResolveClusterClients(ctx, c.factory)
	if err == nil {
		return clients, true
	}

	switch {
	case errors.Is(err, middlewares.ErrClusterMissingID), errors.Is(err, middlewares.ErrClusterInvalidID):
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
	case errors.Is(err, middlewares.ErrClusterNoAuth):
		rsp.ToErrorResponse(errorcode.UserNotLogin)
	case errors.Is(err, middlewares.ErrClusterForbidden):
		rsp.ToErrorResponse(errorcode.ErrorClusterForbidden)
	case errors.Is(err, middlewares.ErrClusterNotFound):
		rsp.ToErrorResponse(errorcode.ErrorClusterNotFound)
	default:
		// 连接/证书/超时等，err 原文只进日志，不回给前端（含 apiserver 地址）
		global.Logger.Warn("canary: 集群客户端获取失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorClusterUnhealthy)
	}
	return nil, false
}

// Promote 手动晋升金丝雀
// POST /api/v1/k8s/cicd/canary/promote
func (c *CanaryDeployController) Promote(ctx *gin.Context) {
	param := &requests.CanaryPromoteRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCanaryPromoteRequest); !ok {
		return
	}

	cli, ok := c.resolveCluster(ctx, rsp)
	if !ok {
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.CanaryPromote(ctx.Request.Context(), cli.Kube, param); err != nil {
		global.Logger.Error("CanaryPromote error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "金丝雀已晋升为稳定版本"})
}

// Rollback 回滚金丝雀
// POST /api/v1/k8s/cicd/canary/rollback
func (c *CanaryDeployController) Rollback(ctx *gin.Context) {
	param := &requests.CanaryRollbackRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCanaryRollbackRequest); !ok {
		return
	}

	cli, ok := c.resolveCluster(ctx, rsp)
	if !ok {
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.CanaryRollback(ctx.Request.Context(), cli.Kube, param); err != nil {
		global.Logger.Error("CanaryRollback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "金丝雀已回滚"})
}

// Status 获取金丝雀状态
// GET /api/v1/k8s/cicd/canary/status
func (c *CanaryDeployController) Status(ctx *gin.Context) {
	param := &requests.CanaryStatusRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCanaryStatusRequest); !ok {
		return
	}

	cli, ok := c.resolveCluster(ctx, rsp)
	if !ok {
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	info, err := svc.CanaryGetStatus(ctx.Request.Context(), cli.Kube, param)
	if err != nil {
		global.Logger.Error("CanaryGetStatus error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(info)
}

// SetTrafficSplit 调整金丝雀流量比例
// POST /api/v1/k8s/cicd/canary/traffic-split
func (c *CanaryDeployController) SetTrafficSplit(ctx *gin.Context) {
	param := &requests.CanaryTrafficSplitRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCanaryTrafficSplitRequest); !ok {
		return
	}

	cli, ok := c.resolveCluster(ctx, rsp)
	if !ok {
		return
	}

	svc := middlewares.NewServicesFromContext(ctx)
	if err := svc.CanarySetTrafficSplit(ctx.Request.Context(), cli.Kube, param); err != nil {
		global.Logger.Error("CanarySetTrafficSplit error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "流量比例已调整"})
}
