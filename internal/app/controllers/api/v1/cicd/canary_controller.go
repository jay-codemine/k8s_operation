package cicd

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
)

// CanaryDeployController 金丝雀部署控制器
type CanaryDeployController struct {
	svc     *services.Services
	factory *services.ClusterClientFactory
}

// NewCanaryDeployController 创建金丝雀控制器
func NewCanaryDeployController() *CanaryDeployController {
	svc := services.NewServices()
	return &CanaryDeployController{
		svc:     svc,
		factory: services.NewClusterClientFactory(svc),
	}
}

// getK8sClient 从上下文获取 K8s 客户端（支持多集群）
func (c *CanaryDeployController) getK8sClient(ctx *gin.Context) (*services.K8sClients, error) {
	clusterIDStr := ctx.GetHeader("X-Cluster-ID")
	if clusterIDStr == "" {
		// 使用默认管理集群
		if global.ManagementKubeClient == nil {
			return nil, errors.New("cluster unavailable")
		}
		return &services.K8sClients{Kube: global.ManagementKubeClient}, nil
	}

	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid X-Cluster-ID")
	}

	clients, err := c.factory.GetClient(ctx.Request.Context(), clusterID)
	if err != nil {
		return nil, errors.New("cluster unavailable")
	}
	return clients, nil
}

// Promote 手动晋升金丝雀
// POST /api/v1/k8s/cicd/canary/promote
func (c *CanaryDeployController) Promote(ctx *gin.Context) {
	var req requests.CanaryPromoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	cli, err := c.getK8sClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": err.Error()})
		return
	}

	if err := c.svc.CanaryPromote(ctx.Request.Context(), cli.Kube, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "OK", "data": "金丝雀已晋升为稳定版本"})
}

// Rollback 回滚金丝雀
// POST /api/v1/k8s/cicd/canary/rollback
func (c *CanaryDeployController) Rollback(ctx *gin.Context) {
	var req requests.CanaryRollbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	cli, err := c.getK8sClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": err.Error()})
		return
	}

	if err := c.svc.CanaryRollback(ctx.Request.Context(), cli.Kube, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "OK", "data": "金丝雀已回滚"})
}

// Status 获取金丝雀状态
// GET /api/v1/k8s/cicd/canary/status
func (c *CanaryDeployController) Status(ctx *gin.Context) {
	var req requests.CanaryStatusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	cli, err := c.getK8sClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": err.Error()})
		return
	}

	info, err := c.svc.CanaryGetStatus(ctx.Request.Context(), cli.Kube, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "OK", "data": info})
}

// SetTrafficSplit 调整金丝雀流量比例
// POST /api/v1/k8s/cicd/canary/traffic-split
func (c *CanaryDeployController) SetTrafficSplit(ctx *gin.Context) {
	var req requests.CanaryTrafficSplitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	cli, err := c.getK8sClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": err.Error()})
		return
	}

	if err := c.svc.CanarySetTrafficSplit(ctx.Request.Context(), cli.Kube, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "OK", "data": "流量比例已调整"})
}
