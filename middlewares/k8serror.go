package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func K8sError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		ge := c.Errors.Last()
		if ge == nil || ge.Err == nil {
			return
		}

		err := ge.Err

		// 如果已经返回过响应，就不要再覆盖（但要注意：AbortWithStatusJSON 已经返回了）
		if c.Writer.Written() {
			return
		}

		// TODO：这里可以根据 k8s apierrors 分类
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":               500,
			"kube_message_error": err.Error(),
		})
	}
}
