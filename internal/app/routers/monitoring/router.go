package monitoring

import (
	"context"
	"net/http"
	"strconv"
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

// ctxWithDS 从请求 query 参数 datasource_id 提取并注入 ctx，供 service 层优先使用
func ctxWithDS(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if idStr := c.Query("datasource_id"); idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			return services.WithDatasourceID(ctx, id)
		}
	}
	return ctx
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

		// 大厂级能力扩展
		g.GET("/score", r.GetHealthScore)             // 集群健康评分
		g.GET("/heatmap", r.GetHeatmap)               // 节点热力图
		g.GET("/pod-status", r.GetPodStatus)          // Pod 状态分布
		g.GET("/abnormal-pods", r.GetAbnormalPods)    // 异常 Pod
		g.GET("/namespaces", r.GetNamespaceMetrics)   // Namespace 聚合
		g.GET("/node-detail", r.GetNodeDetail)        // 单节点详情聚合
	}

	// CRUD 路由（数据源、告警规则、告警事件）
	NewMonitorCRUDRouter().Inject(g)

	// Loki 日志查询路由
	NewLokiRouter("").Inject(g)
}

// GetOverview 获取集群监控总览
func (r *MonitoringRouter) GetOverview(c *gin.Context) {
	ctx := ctxWithDS(c)
	overview, err := r.svc.GetClusterOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": overview})
}

// GetNodeMetrics 获取节点指标
func (r *MonitoringRouter) GetNodeMetrics(c *gin.Context) {
	ctx := ctxWithDS(c)
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

	ctx := ctxWithDS(c)
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

	ctx := ctxWithDS(c)
	pods, err := r.svc.GetTopPods(ctx, metric, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pods})
}

// HealthCheck 检查 Prometheus 连通性
func (r *MonitoringRouter) HealthCheck(c *gin.Context) {
	ctx := ctxWithDS(c)
	healthy := r.svc.IsHealthy(ctx)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    gin.H{"prometheus": healthy, "url": r.svc.GetPrometheusURL(ctx)},
		"message": "ok",
	})
}

// GetHealthScore 集群健康评分
func (r *MonitoringRouter) GetHealthScore(c *gin.Context) {
	ctx := ctxWithDS(c)
	score, err := r.svc.GetClusterHealthScore(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": score})
}

// GetHeatmap 节点热力图
func (r *MonitoringRouter) GetHeatmap(c *gin.Context) {
	metric := c.DefaultQuery("metric", "cpu")
	durationStr := c.DefaultQuery("duration", "1h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 1 * time.Hour
	}
	ctx := ctxWithDS(c)
	cells, err := r.svc.GetNodeHeatmap(ctx, metric, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cells})
}

// GetPodStatus Pod 状态分布
func (r *MonitoringRouter) GetPodStatus(c *gin.Context) {
	ctx := ctxWithDS(c)
	items, err := r.svc.GetPodStatusDistribution(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// GetAbnormalPods 重启异常的 Pod
func (r *MonitoringRouter) GetAbnormalPods(c *gin.Context) {
	ctx := ctxWithDS(c)
	pods, err := r.svc.GetAbnormalPods(ctx, 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pods})
}

// GetNamespaceMetrics Namespace 聚合指标
func (r *MonitoringRouter) GetNamespaceMetrics(c *gin.Context) {
	ctx := ctxWithDS(c)
	items, err := r.svc.GetNamespaceMetrics(ctx, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// GetNodeDetail 单节点详情聚合（当前指标 + 5 维趋势 + Top Pod + 元信息）
func (r *MonitoringRouter) GetNodeDetail(c *gin.Context) {
	instance := c.Query("instance")
	if instance == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "instance 参数不能为空"})
		return
	}
	durationStr := c.DefaultQuery("duration", "1h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 1 * time.Hour
	}
	ctx := ctxWithDS(c)
	detail, err := r.svc.GetNodeDetail(ctx, instance, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": detail})
}
