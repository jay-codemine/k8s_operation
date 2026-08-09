package rbac

import (
	"fmt"
	"strings"
)

// RoleName 角色名称值对象
type RoleName struct{ val string }

func NewRoleName(s string) (RoleName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return RoleName{}, fmt.Errorf("角色名称不能为空")
	}
	if len(s) > 50 {
		return RoleName{}, fmt.Errorf("角色名称最长 50 字符")
	}
	return RoleName{val: s}, nil
}

func (n RoleName) String() string { return n.val }

// PermissionName 权限名值对象
type PermissionName struct{ val string }

func NewPermissionName(s string) (PermissionName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return PermissionName{}, fmt.Errorf("权限名不能为空")
	}
	return PermissionName{val: s}, nil
}

func (n PermissionName) String() string { return n.val }

// AccessLevel 访问级别值对象（none / read / write / admin）
type AccessLevel struct{ val string }

var accessLevels = map[string]int{"none": 0, "read": 1, "write": 2, "admin": 3}

func NewAccessLevel(s string) AccessLevel {
	s = strings.TrimSpace(strings.ToLower(s))
	if _, ok := accessLevels[s]; !ok {
		s = "none"
	}
	return AccessLevel{val: s}
}

func (l AccessLevel) String() string { return l.val }

// Gte 比较是否 >= minLevel
func (l AccessLevel) Gte(min string) bool {
	return accessLevels[l.val] >= accessLevels[min]
}
