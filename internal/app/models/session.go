package models

import "time"

type LoginSessionInfo struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	LoginTime time.Time `json:"login_time"`
}
