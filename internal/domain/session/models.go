package session

import "time"

// LoginSessionInfo 登录会话信息
type LoginSessionInfo struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	LoginTime time.Time `json:"login_time"`
}
