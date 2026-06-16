package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
	"strings"
)

// 从 Header 或 Query 参数里取 Bearer token
// 优先从 Authorization Header 取，fallback 到 query 参数 token（用于文件下载等 window.open 场景）
func GetTokenFromHeader(c *gin.Context) (string, error) {
	auth := c.Request.Header.Get("Authorization")
	if auth == "" {
		// fallback: 尝试从 query 参数取 token
		if qToken := c.Query("token"); qToken != "" {
			return qToken, nil
		}
		return "", errorcode.ErrHeaderEmpty
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errorcode.ErrHeaderMalformed
	}
	if parts[1] == "" {
		return "", errors.New("empty bearer token")
	}
	return parts[1], nil
}
