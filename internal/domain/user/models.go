package user

import "k8soperation/pkg/db"

// User 用户
type User struct {
	Username string `json:"username" gorm:"column:username" description:"用户名"`
	Password string `json:"-" gorm:"column:password" description:"密码"`
	Role     string `json:"role" gorm:"column:role;default:user" description:"角色"`
	Email    string `json:"email" gorm:"column:email" description:"邮箱"`
	Phone    string `json:"phone" gorm:"column:phone" description:"手机号"`
	Status   int8   `json:"status" gorm:"column:status;default:1" description:"状态:1激活,0禁用"`
	*db.Base
}

// NewUser 创建用户
func NewUser() *User {
	return &User{}
}

// TableName 返回表名
func (u *User) TableName() string {
	return "user"
}

// AggregateID 实现 domain.AggregateRoot 接口
func (u *User) AggregateID() int64 { return int64(u.ID) }
