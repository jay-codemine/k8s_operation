package jwt

import jwtpkg "github.com/golang-jwt/jwt"

// 自定义载荷
type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	jwtpkg.StandardClaims
}
