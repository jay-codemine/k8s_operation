package dao

import (
	"k8soperation/internal/app/models"
	"k8soperation/pkg/utils"
	"time"
)

// UserCreate 创建用户（密码使用 bcrypt 加密存储）
func (d *Dao) UserCreate(name, password string, tenantID uint32) (*models.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if tenantID == 0 {
		tenantID = models.DefaultTenantID
	}

	nowTime := uint32(time.Now().Unix())
	user := &models.User{
		Username: name,
		Password: hashedPassword,
		Base: &models.Base{
			TenantID:   tenantID,
			CreatedAt:  nowTime,
			ModifiedAt: nowTime,
			IsDel:      0,
		},
	}
	if err := user.Create(d.db); err != nil {
		return nil, err
	}
	return user, nil
}

// UserDelete 删除用户
func (d *Dao) UserDelete(id uint32) error {
	user := models.User{
		Base: &models.Base{ID: id},
	}
	return user.Delete(d.db)
}

// UserUpdate 更新用户（密码使用 bcrypt 加密存储）
func (d *Dao) UserUpdate(id uint32, name, password, role string, status int8) error {
	nowTime := uint32(time.Now().Unix())
	user := models.User{
		Base: &models.Base{
			ID: id,
		},
	}

	values := map[string]interface{}{
		"username":    name,
		"modified_at": nowTime,
		"status":      status,
	}

	// 如果提供了新密码，则加密存储
	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return err
		}
		values["password"] = hashedPassword
	}

	// 如果提供了角色，则更新
	if role != "" {
		values["role"] = role
	}

	return user.Update(d.db, values)
}

func (d *Dao) UserList(username, role, status string, page, limit int) ([]*models.User, int64, error) {
	user := models.User{
		Username: username,
	}
	return user.List(d.db, role, status, page, limit)
}

// UserGetByName 根据用户名获取用户信息
func (d *Dao) UserGetByName(username string) (*models.User, error) {
	user := models.User{
		Username: username,
	}
	return user.GetByName(d.db)
}

// UserGetByID 通过ID获取用户
func (d *Dao) UserGetByID(id int64) (*models.User, error) {
	var user models.User
	if err := d.db.Where("id = ? AND is_del = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UserMigratePassword 将用户密码迁移到 bcrypt 格式
func (d *Dao) UserMigratePassword(userID uint32, plainPassword string) error {
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return err
	}

	nowTime := uint32(time.Now().Unix())
	user := models.User{
		Base: &models.Base{ID: userID},
	}

	values := map[string]interface{}{
		"password":    hashedPassword,
		"modified_at": nowTime,
	}

	return user.Update(d.db, values)
}

// UserCreateFull 创建用户（含邮箱、手机、角色）
func (d *Dao) UserCreateFull(name, password, email, phone, role string) (*models.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	nowTime := uint32(time.Now().Unix())
	user := &models.User{
		Username: name,
		Password: hashedPassword,
		Email:    email,
		Phone:    phone,
		Role:     role,
		Base: &models.Base{
			CreatedAt:  nowTime,
			ModifiedAt: nowTime,
			IsDel:      0,
		},
	}
	if err := user.Create(d.db); err != nil {
		return nil, err
	}
	return user, nil
}
