package dao

import (
	"k8soperation/internal/app/models"
	"time"
)

func (d *Dao) UserCreate(name, password string) error {
	// 获取当前时间戳并转换为uint32类型
	nowTime := uint32(time.Now().Unix()) // 将当前时间的Unix时间戳转换为uint32类型并赋值给nowTime变量
	user := models.User{
		Username: name,
		Password: password,
		Base: &models.Base{
			CreatedAt:  nowTime,
			ModifiedAt: nowTime,
			IsDel:      0,
		},
	}
	return user.Create(d.db)
}

// UserDelete 删除用户
func (d *Dao) UserDelete(id uint32) error {
	user := models.User{
		Base: &models.Base{ID: id},
	}
	return user.Delete(d.db)
}

func (d *Dao) UserUpdate(id uint32, name, password string) error {
	nowTime := uint32(time.Now().Unix())
	user := models.User{
		Base: &models.Base{
			ID: id,
		},
	}

	values := map[string]interface{}{
		"username":    name,
		"password":    password,
		"modified_at": nowTime,
	}
	return user.Update(d.db, values)
}

func (d *Dao) UserList(username string, page, limit int) ([]*models.User, error) {
	user := models.User{
		Username: username,
	}

	return user.List(d.db, page, limit)
}

// UserGetByName 根据用户名获取用户信息
func (d *Dao) UserGetByName(username string) (*models.User, error) {
	user := models.User{
		Username: username,
	}
	return user.GetByName(d.db)
}
