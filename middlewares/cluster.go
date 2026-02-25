package middlewares

import (
	"errors"
	"go.uber.org/zap"
	"k8soperation/global"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/services"
)

const (
	CtxClusterID  = "cluster_id"
	CtxK8sClients = "k8s_clients"
)

// 可选：定义一些可识别的错误（建议你在 services/dao 里返回这些）
var (
	ErrClusterNotFound  = errors.New("cluster not found")
	ErrClusterForbidden = errors.New("cluster forbidden")
)

func ClusterMiddleware(factory *services.ClusterClientFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) 取 clusterId
		idStr := c.GetHeader("X-Cluster-ID")
		if idStr == "" {
			idStr = c.Query("clusterId")
		}
		if idStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code": 40001,
				"msg":  "missing X-Cluster-ID",
			})
			return
		}

		// 2) 校验 clusterId
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id64 == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code": 40002,
				"msg":  "invalid clusterId",
			})
			return
		}
		clusterID := uint32(id64)

		// 3) 权限校验（强烈建议）
		// 你 JWT middleware 一般会把 user_id 放进 ctx
		// userID := c.GetUint("user_id")
		// if !factory.CanAccess(userID, clusterID) { ... }
		// 这里我先留 TODO，不阻断逻辑

		// 4) 获取 client
		clients, err := factory.Get(c.Request.Context(), clusterID)
		if err != nil {
			// 5) 错误映射：尽量用“可预期错误”返回更准确状态码
			//    如果你 dao/service 还没做错误类型，这里先做“文本兜底”
			msg := "cluster init failed"
			low := strings.ToLower(err.Error())

			// 404：集群不存在
			if errors.Is(err, ErrClusterNotFound) || strings.Contains(low, "not found") || strings.Contains(low, "no rows") {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"code": 40401,
					"msg":  "cluster not found",
				})
				return
			}

			// 403：无权限（等你实现 user_cluster 后启用）
			if errors.Is(err, ErrClusterForbidden) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code": 403,
					"msg":  "cluster forbidden",
				})
				return
			}

			// 503：连接/证书/超时等（集群不可用）
			// 注意：不要把 err 原文全吐给前端，容易泄露 apiserver/证书信息
			// 可以仅返回简化信息（前端显示），详细 err 记录到日志里
			_ = msg
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code": 503,
				"msg":  "cluster unavailable",
			})
			return
		}

		// 6) 注入 context
		c.Set(CtxClusterID, clusterID)
		c.Set(CtxK8sClients, clients)

		// 可选：如果你经常用到 clients.Kube / clients.Dynamic，可以拆开 set
		// c.Set("kube", clients.Kube)
		// c.Set("dynamic", clients.Dynamic)

		host := clients.Config.Host
		global.Logger.Info("bind k8s_clients for request",
			zap.String("cluster_id", c.GetHeader("X-Cluster-ID")),
			zap.String("apiserver", host),
			zap.Bool("has_metrics", clients.Metrics != nil),
			zap.Bool("supports_ev_v1", clients.SupportsEvV1),
		)

		c.Next()
	}
}
