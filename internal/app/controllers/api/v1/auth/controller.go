package auth

import (
	jwtutil "k8soperation/pkg/jwt"
)

type AuthController struct {
	jwt *jwtutil.Manager
}

func NewAuthController() *AuthController {
	return &AuthController{
		jwt: jwtutil.NewManager(),
	}
}
