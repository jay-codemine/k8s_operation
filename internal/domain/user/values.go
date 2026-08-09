package user

import (
	"fmt"
	"strings"
)

// Email 邮箱值对象
type Email struct{ val string }

func NewEmail(s string) (Email, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Email{}, nil // 邮箱可选
	}
	if !strings.Contains(s, "@") {
		return Email{}, fmt.Errorf("邮箱格式无效: %s", s)
	}
	return Email{val: s}, nil
}

func (e Email) String() string { return e.val }

// Phone 手机号值对象
type Phone struct{ val string }

func NewPhone(s string) (Phone, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Phone{}, nil // 可选
	}
	if len(s) < 7 {
		return Phone{}, fmt.Errorf("手机号格式无效: %s", s)
	}
	return Phone{val: s}, nil
}

func (p Phone) String() string { return p.val }
