package services

import (
	"errors"

	"gorm.io/gorm"

	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
)

// UserLogin 用户登录
func (s *Services) UserLogin(param *requests.AuthLoginRequest) (*models.User, error) {
	return s.userSvc().GetByName(param.Username)
}

// UserForgotPassword 忘记密码
func (s *Services) UserForgotPassword(param *requests.AuthForgotPasswordRequest) error {
	err := s.userSvc().ForgotPassword(param.Username, param.NewPassword, param.Confirm)
	if err != nil {
		if err.Error() == "user not found" {
			return errorcode.ErrorUserNotFound
		}
		if err.Error() == "passwords do not match" {
			return errorcode.ErrorUserPasswordNotMatch
		}
		// Could be gorm.ErrRecordNotFound from GetByName
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorcode.ErrorUserNotFound
		}
		return errorcode.ServerError
	}
	return nil
}

// AuthRegister 注册
func (s *Services) AuthRegister(param *requests.AuthRegisterRequest) error {
	return s.userSvc().Register(param.Username, param.Password)
}
