package dao

import (
	"k8soperation/internal/app/models"
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

// UserUpdatePasswordByName 根据用户名更新密码
func (d *Dao) UserUpdatePasswordByName(username, newPassword string) error {
	user := models.User{Username: username}
	return user.UpdatePasswordByName(d.db, newPassword)
}
