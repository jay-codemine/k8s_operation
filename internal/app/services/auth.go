package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
)

// UserLogin 用户登录
func (s *Services) UserLogin(param *requests.AuthLoginRequest) (*models.User, error) {
	return s.dao.UserGetByName(param.Username)
}

// UserForgotPassword 忘记密码（按用户名直接重置）
func (s *Services) UserForgotPassword(param *requests.AuthForgotPasswordRequest) error {
	// 1) 查用户是否存在
	user, err := s.dao.UserGetByName(param.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorcode.ErrorUserNotFound
		}
		return errorcode.ServerError
	}
	if user == nil || user.ID == 0 {
		return errorcode.ErrorUserNotFound
	}

	// 2) 校验两次密码一致
	if param.NewPassword != param.Confirm {
		return errorcode.ErrorUserPasswordNotMatch
	}

	// 3) 更新密码（当前先按明文）
	if err := s.dao.UserUpdatePasswordByName(param.Username, param.NewPassword); err != nil {
		return errorcode.ServerError
	}

	return nil
}

// 注册
func (s *Services) AuthRegister(param *requests.AuthRegisterRequest) error {
	// 查重：用户名是否已存在
	exists, err := s.dao.UserExistsByUsername(param.Username)
	if err != nil {
		return err
	}
	if exists {
		return errorcode.InvalidParams.WithDetails(
			fmt.Sprintf("用户名 %s 已注册", param.Username),
		)
	}

	// 创建用户（这里建议密码做 hash：bcrypt）
	return s.dao.UserCreate(param.Username, param.Password)
}
