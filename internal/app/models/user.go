package models

import dm "k8soperation/internal/domain/user"

// User 用户（领域定义在 domain/user）
type User = dm.User

// NewUser 创建用户
var NewUser = dm.NewUser
