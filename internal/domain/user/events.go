package user

import "k8soperation/internal/domain/events"

// UserRegistered 用户注册事件
type UserRegistered struct {
	events.BaseEvent
	UserID   uint32
	Username string
}

func NewUserRegistered(id uint32, username string) UserRegistered {
	return UserRegistered{BaseEvent: events.NewBaseEvent("user.registered"), UserID: id, Username: username}
}

// UserPasswordChanged 密码变更事件
type UserPasswordChanged struct {
	events.BaseEvent
	UserID   uint32
	Username string
}

func NewUserPasswordChanged(id uint32, username string) UserPasswordChanged {
	return UserPasswordChanged{BaseEvent: events.NewBaseEvent("user.password_changed"), UserID: id, Username: username}
}
