package persistence

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/user"
	"k8soperation/internal/errorcode"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Save(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) Update(ctx context.Context, id uint32, values interface{}) error {
	res := r.db.WithContext(ctx).Model(&user.User{}).
		Where("id=? AND is_del=?", id, 0).
		Updates(values)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errorcode.ErrorUserUpdateFail
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id uint32) error {
	var usr user.User
	if err := r.db.WithContext(ctx).Where("id=? and is_del=?", id, 0).First(&usr).Error; err != nil {
		return err
	}
	nowTime := uint32(time.Now().Unix())
	usr.IsDel = 1
	usr.DeletedAt = nowTime
	usr.ModifiedAt = nowTime
	return r.db.WithContext(ctx).Updates(&usr).Error
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*user.User, error) {
	var u user.User
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindByStringID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindByName(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	if username == "" {
		return nil, nil
	}
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) Query(ctx context.Context, username, role, status string, page, limit int) ([]*user.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 1000 {
		limit = 20
	}

	q := r.db.WithContext(ctx).Model(&user.User{}).Where("is_del = 0")
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*user.User
	offset := (page - 1) * limit
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepo) CountByName(ctx context.Context, username string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&user.User{}).Where("username = ? AND is_del = 0", username).Count(&count).Error
	return count, err
}

func (r *userRepo) UpdatePasswordByName(ctx context.Context, username, newPassword string) error {
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("username = ?", username).
		Update("password", newPassword).
		Error
}

// GetStringID 返回用户ID的字符串形式（保留在模型层的工具方法）
func userGetStringID(u *user.User) string {
	return strconv.FormatUint(uint64(u.ID), 10)
}
