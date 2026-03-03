package utils

import "strings"

// nullString 将空字符串转为 nil，非空返回指针
func NullString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
