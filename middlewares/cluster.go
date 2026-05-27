package middlewares

import (
	"errors"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
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
		// 1) 取 clusterId（支持 header / query 多种参数名）
		idStr := c.GetHeader("X-Cluster-ID")
		if idStr == "" {
			idStr = c.Query("clusterId")
		}
		if idStr == "" {
			idStr = c.Query("cluster_id")
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

		// 3) 权限校验：前端菜单隐藏不能替代服务端拦截。
		userID := currentUserID(c)
		if userID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "unauthorized",
			})
			return
		}

		action := inferClusterAction(c)
		svc := services.NewServices()
		if !svc.CheckClusterPermission(userID, int64(clusterID), action) {
			global.Logger.Warn("cluster request forbidden",
				zap.Int64("user_id", userID),
				zap.Uint32("cluster_id", clusterID),
				zap.String("action", action),
				zap.String("path", c.FullPath()),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "cluster forbidden",
			})
			return
		}

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
		c.Set("cluster_action", action)

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

func currentUserID(c *gin.Context) int64 {
	if v, ok := c.Get("user_id"); ok {
		switch id := v.(type) {
		case int64:
			return id
		case int:
			return int64(id)
		case int32:
			return int64(id)
		case uint:
			return int64(id)
		case uint32:
			return int64(id)
		case uint64:
			if id <= uint64(^uint(0)>>1) {
				return int64(id)
			}
		case string:
			parsed, err := strconv.ParseInt(id, 10, 64)
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}

func inferClusterAction(c *gin.Context) string {
	path := strings.ToLower(c.FullPath())
	if path == "" {
		path = strings.ToLower(c.Request.URL.Path)
	}

	switch {
	case strings.Contains(path, "terminal") || strings.Contains(path, "exec"):
		return models.PermissionActionExec
	case strings.Contains(path, "delete") || strings.Contains(path, "drain") || strings.Contains(path, "evict"):
		return models.PermissionActionDelete
	case strings.Contains(path, "create"):
		return models.PermissionActionCreate
	case strings.Contains(path, "update") ||
		strings.Contains(path, "patch") ||
		strings.Contains(path, "apply") ||
		strings.Contains(path, "scale") ||
		strings.Contains(path, "restart") ||
		strings.Contains(path, "rollback") ||
		strings.Contains(path, "cordon") ||
		strings.Contains(path, "uncordon") ||
		strings.Contains(path, "suspend"):
		return models.PermissionActionUpdate
	}

	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return models.PermissionActionView
	case http.MethodPost:
		return models.PermissionActionCreate
	case http.MethodPut, http.MethodPatch:
		return models.PermissionActionUpdate
	case http.MethodDelete:
		return models.PermissionActionDelete
	default:
		return models.PermissionActionView
	}
}
