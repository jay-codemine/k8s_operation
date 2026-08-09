package user

import "context"

// UserRepository 用户仓储接口
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, id uint32, values interface{}) error
	Delete(ctx context.Context, id uint32) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByStringID(ctx context.Context, id string) (*User, error)
	FindByName(ctx context.Context, username string) (*User, error)
	Query(ctx context.Context, username, role, status string, page, limit int) ([]*User, int64, error)
	CountByName(ctx context.Context, username string) (int64, error)
	UpdatePasswordByName(ctx context.Context, username, hashedPassword string) error
}
