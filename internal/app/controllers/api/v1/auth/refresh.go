package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
	jwtutil "k8soperation/pkg/jwt"
)

// Refresh godoc
// @Summary 刷新访问令牌
// @Description 使用旧的 Token 刷新并返回新的 Token
// @Tags 认证管理
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} string "成功"
// @Failure 401 {object} errorcode.Error "Token 无效或已超过最大刷新时间"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/auth/refresh [post]
//
// AuthController 结构体的刷新 Token 方法
// 处理 Token 刷新请求：
// 1. 从 Header 中获取 Bearer Token
// 2. 校验 Token 是否在 MaxRefresh 时间窗口内
// 3. 签发并返回新的 Token
func (a *AuthController) Refresh(c *gin.Context) {
	tokenStr, err := jwtutil.GetTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": errorcode.ErrTokenInvalid.Code(),
			"msg":  errorcode.ErrTokenInvalid.Error(),
		})
		return
	}

	newToken, err := a.jwt.Refresh(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": errorcode.ErrTokenInvalid.Code(),
			"msg":  errorcode.ErrTokenInvalid.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": gin.H{
			"token": newToken,
		},
	})
}
