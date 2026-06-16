package monitoring

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/services"
)

// LokiRouter Loki 日志路由
type LokiRouter struct {
	svc *services.LokiService
}

// NewLokiRouter 创建 Loki 路由
func NewLokiRouter(lokiURL string) *LokiRouter {
	return &LokiRouter{
		svc: services.NewLokiService(lokiURL),
	}
}

// Inject 注入 Loki 路由到 monitoring group
func (r *LokiRouter) Inject(g *gin.RouterGroup) {
	loki := g.Group("/loki")
	{
		loki.GET("/health", r.HealthCheck)            // Loki 健康检查
		loki.GET("/query", r.QueryLogs)               // 日志查询
		loki.GET("/labels", r.GetLabels)              // 获取标签列表
		loki.GET("/label/:name/values", r.GetLabelValues) // 获取标签值
		loki.GET("/streams", r.GetStreams)             // 获取日志流列表
		loki.GET("/volume", r.GetLogVolume)           // 获取日志量趋势
	}
}

// HealthCheck Loki 健康检查
func (r *LokiRouter) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	result := r.svc.HealthCheck(ctx)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// QueryLogs 查询日志
func (r *LokiRouter) QueryLogs(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "query 参数不能为空"})
		return
	}

	duration := c.DefaultQuery("duration", "1h")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	direction := c.DefaultQuery("direction", "backward")

	dur, err := time.ParseDuration(duration)
	if err != nil {
		dur = 1 * time.Hour
	}

	end := time.Now()
	start := end.Add(-dur)

	ctx := c.Request.Context()
	result, err := r.svc.QueryLogs(ctx, query, start, end, limit, direction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// GetLabels 获取标签列表
func (r *LokiRouter) GetLabels(c *gin.Context) {
	duration := c.DefaultQuery("duration", "1h")
	dur, _ := time.ParseDuration(duration)
	if dur == 0 {
		dur = 1 * time.Hour
	}

	end := time.Now()
	start := end.Add(-dur)

	ctx := c.Request.Context()
	labels, err := r.svc.GetLabels(ctx, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": labels})
}

// GetLabelValues 获取标签值
func (r *LokiRouter) GetLabelValues(c *gin.Context) {
	name := c.Param("name")
	duration := c.DefaultQuery("duration", "1h")
	dur, _ := time.ParseDuration(duration)
	if dur == 0 {
		dur = 1 * time.Hour
	}

	end := time.Now()
	start := end.Add(-dur)

	ctx := c.Request.Context()
	values, err := r.svc.GetLabelValues(ctx, name, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": values})
}

// GetStreams 获取日志流列表
func (r *LokiRouter) GetStreams(c *gin.Context) {
	matcher := c.DefaultQuery("matcher", "")
	duration := c.DefaultQuery("duration", "1h")
	dur, _ := time.ParseDuration(duration)
	if dur == 0 {
		dur = 1 * time.Hour
	}

	end := time.Now()
	start := end.Add(-dur)

	ctx := c.Request.Context()
	streams, err := r.svc.GetStreams(ctx, matcher, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": streams})
}

// GetLogVolume 获取日志量趋势
func (r *LokiRouter) GetLogVolume(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	duration := c.DefaultQuery("duration", "1h")
	dur, _ := time.ParseDuration(duration)
	if dur == 0 {
		dur = 1 * time.Hour
	}

	end := time.Now()
	start := end.Add(-dur)
	step := dur / 60

	ctx := c.Request.Context()
	volume, err := r.svc.GetLogVolume(ctx, query, start, end, step)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": volume})
}
