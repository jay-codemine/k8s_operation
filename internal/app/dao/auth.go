package dao

import (
	"k8soperation/internal/app/models"
	"k8soperation/pkg/utils"
)

// 判断用户名是否存在
func (d *Dao) UserExistsByUsername(username string) (bool, error) {
	var count int64
	err := d.db.Model(&models.User{}).
		Where("username = ? AND is_del = 0", username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserUpdatePasswordByName 根据用户名更新密码（使用 bcrypt 加密）
func (d *Dao) UserUpdatePasswordByName(username, newPassword string) error {
	// 对新密码进行 bcrypt 加密
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user := models.User{Username: username}
	return user.UpdatePasswordByName(d.db, hashedPassword)
}
