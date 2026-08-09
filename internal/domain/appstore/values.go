package appstore

import (
	"fmt"
	"strings"
)

// AppName 应用名称值对象
type AppName struct{ val string }

func NewAppName(s string) (AppName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return AppName{}, fmt.Errorf("应用名称不能为空")
	}
	if len(s) > 100 {
		return AppName{}, fmt.Errorf("应用名称最长 100 字符")
	}
	return AppName{val: s}, nil
}

func (n AppName) String() string { return n.val }

// AppVersion 应用版本值对象（semver 格式）
type AppVersion struct{ val string }

func NewAppVersion(s string) (AppVersion, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return AppVersion{val: "0.0.1"}, nil
	}
	return AppVersion{val: s}, nil
}

func (v AppVersion) String() string { return v.val }
