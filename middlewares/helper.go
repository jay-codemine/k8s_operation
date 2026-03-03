package middlewares

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/services"
	"net/http"
)

func MustGetK8sClients(c *gin.Context) *services.K8sClients {
	v, ok := c.Get(CtxK8sClients)
	if !ok || v == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": 50010,
			"msg":  "k8s clients not found in context (cluster middleware missing?)",
		})
		return nil
	}

	cli, ok := v.(*services.K8sClients)
	if !ok || cli == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": 50011,
			"msg":  "invalid k8s clients in context",
		})
		return nil
	}

	return cli
}
