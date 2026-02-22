package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
	"strings"
)

// 从 Header 里取 Bearer token
func GetTokenFromHeader(c *gin.Context) (string, error) {
	auth := c.Request.Header.Get("Authorization")
	if auth == "" {
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
