package user

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/errorcode"
	"k8soperation/pkg/db"
)

// Create 注册用户
func (u *User) Create(tx *gorm.DB) error {
	return tx.Create(u).Error
}

// Delete 软删除用户
func (u *User) Delete(tx *gorm.DB) error {
	var usr User
	if err := tx.Where("id=? and is_del=?", u.ID, 0).First(&usr).Error; err != nil {
		return err
	}

	nowTime := uint32(time.Now().Unix())
	usr.IsDel = 1
	usr.DeletedAt = nowTime
	usr.ModifiedAt = nowTime

	return tx.Updates(&usr).Error
}

// Update 更新用户信息
func (u *User) Update(tx *gorm.DB, values interface{}) error {
	res := tx.Model(u).
		Where("id=? AND is_del=?", u.ID, 0).
		Updates(values)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errorcode.ErrorUserUpdateFail
	}
	return nil
}

// List 查询用户列表
func (u *User) List(tx *gorm.DB, role, status string, page, limit int) ([]*User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 1000 {
		limit = 20
	}

	var users []*User
	var total int64

	q := tx.Model(&User{}).Where("is_del = 0")
	if u.Username != "" {
		q = q.Where("username LIKE ?", "%"+u.Username+"%")
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// GetByName 按用户名查找
func (u *User) GetByName(tx *gorm.DB) (*User, error) {
	var usr User
	if u.Username == "" {
		return nil, nil
	}
	if err := tx.Where("username = ?", u.Username).First(&usr).Error; err != nil {
		return nil, err
	}
	return &usr, nil
}

// GetUserByID 通过ID查找用户
func (u *User) GetUserByID(tx *gorm.DB, id string) User {
	usr, _ := u.GetUserByIDE(tx, id)
	return usr
}

// GetUserByIDE 通过ID查找用户并返回查询错误
func (u *User) GetUserByIDE(tx *gorm.DB, id string) (User, error) {
	usr := User{Base: &db.Base{}}
	if id == "" {
		return usr, gorm.ErrRecordNotFound
	}
	if err := tx.Where("id = ?", id).First(&usr).Error; err != nil {
		return User{Base: &db.Base{}}, err
	}
	return usr, nil
}

// GetStringID 返回用户ID的字符串形式
func (u *User) GetStringID() string {
	return strconv.FormatUint(uint64(u.ID), 10)
}

// UpdatePasswordByName 按用户名更新密码
func (u *User) UpdatePasswordByName(tx *gorm.DB, newPassword string) error {
	return tx.Model(&User{}).
		Where("username = ?", u.Username).
		Update("password", newPassword).
		Error
}
