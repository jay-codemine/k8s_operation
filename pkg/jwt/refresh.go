package jwt

import (
	jwtpkg "github.com/golang-jwt/jwt"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/utils"
	"time"
)

// Refresh 刷新并返回新的访问 token（仅在“未过期或仅过期”且在 MaxRefresh 窗口内时成功）
func (m *Manager) Refresh(tokenString string) (string, error) {
	// 显式禁用刷新：配置为 <=0 即不允许刷新
	if m.MaxRefresh <= 0 {
		return "", errorcode.ErrTokenMalformed
	}
	parser := &jwtpkg.Parser{ValidMethods: []string{jwtpkg.SigningMethodHS256.Alg()}}

	// 1) 解析（可能返回过期错误，但 token/claims 仍可用）
	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwtpkg.Token) (interface{}, error) {
		return m.SignKey, nil
	})

	//2) 仅允许“过期”这一类错误；其它错误统一视为无效
	if err != nil {
		if ve, ok := err.(*jwtpkg.ValidationError); !ok || (ve.Errors&^jwtpkg.ValidationErrorExpired) != 0 {
			return "", errorcode.ErrTokenInvalid
		}
	} else {
		// 没有报错时，也要求 token.Valid 为真（更稳妥）
		if !token.Valid {
			return "", errorcode.ErrTokenInvalid
		}
	}
	if token == nil {
		return "", errorcode.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errorcode.ErrTokenInvalid
	}

	// 3) 检查是否还在“最大可刷新窗口”内（从 iat 起算）
	now := utils.TimenowInTimezone()
	// 可选：容忍轻微时钟偏差
	const skew = 30 * time.Second
	cutoff := now.Add(-(m.MaxRefresh + skew)).Unix()
	if claims.IssuedAt < cutoff {
		return "", errorcode.ErrTokenExpiredMaxRefresh
	}
	// 仅更新过期时间
	claims.ExpiresAt = m.expireAtTime()
	return m.createToken(*claims)
}
