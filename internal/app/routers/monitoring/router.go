package monitoring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/services"
)

// MonitoringRouter 监控路由
type MonitoringRouter struct {
	svc *services.MonitoringService
}

// NewMonitoringRouter 创建监控路由（传入 Prometheus URL）
func NewMonitoringRouter(prometheusURL string) *MonitoringRouter {
	return &MonitoringRouter{
		svc: services.NewMonitoringService(prometheusURL),
	}
}

// Inject 注入路由
func (r *MonitoringRouter) Inject(router *gin.RouterGroup) {
	g := router.Group("/monitoring")
	{
		// 监控指标查询（原有 Prometheus 路由）
		g.GET("/overview", r.GetOverview)          // 集群监控总览
		g.GET("/nodes", r.GetNodeMetrics)          // 节点指标列表
		g.GET("/trend/:resource", r.GetTrend)      // 资源趋势图
		g.GET("/top-pods", r.GetTopPods)           // Top N Pods
		g.GET("/health", r.HealthCheck)            // Prometheus 健康检查
	}

	// CRUD 路由（数据源、告警规则、告警事件）
	NewMonitorCRUDRouter().Inject(g)

	// Loki 日志查询路由
	NewLokiRouter("").Inject(g)
}

// GetOverview 获取集群监控总览
func (r *MonitoringRouter) GetOverview(c *gin.Context) {
	ctx := c.Request.Context()
	overview, err := r.svc.GetClusterOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": overview})
}

// GetNodeMetrics 获取节点指标
func (r *MonitoringRouter) GetNodeMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	nodes, err := r.svc.GetNodeMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
}

// GetTrend 获取资源趋势数据
func (r *MonitoringRouter) GetTrend(c *gin.Context) {
	resource := c.Param("resource")
	durationStr := c.DefaultQuery("duration", "1h")

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 1 * time.Hour
	}

	ctx := c.Request.Context()
	trends, err := r.svc.GetResourceTrend(ctx, resource, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": trends})
}

// GetTopPods 获取 Top N Pods
func (r *MonitoringRouter) GetTopPods(c *gin.Context) {
	metric := c.DefaultQuery("metric", "cpu")
	limit := 10

	ctx := c.Request.Context()
	pods, err := r.svc.GetTopPods(ctx, metric, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pods})
}

// HealthCheck 检查 Prometheus 连通性
func (r *MonitoringRouter) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	healthy := r.svc.IsHealthy(ctx)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    gin.H{"prometheus": healthy, "url": r.svc.GetPrometheusURL()},
		"message": "ok",
	})
}
