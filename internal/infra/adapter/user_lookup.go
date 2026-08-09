package adapter

import (
	"context"

	"k8soperation/internal/domain/rbac"
	"k8soperation/internal/domain/user"
)

// UserLookupAdapter 将 user.UserService 适配为 rbac.UserLookup 接口
type UserLookupAdapter struct {
	svc *user.UserService
}

func NewUserLookup(svc *user.UserService) rbac.UserLookup {
	return &UserLookupAdapter{svc: svc}
}

func (a *UserLookupAdapter) FindUsername(ctx context.Context, userID int64) (string, error) {
	u, err := a.svc.GetByID(userID)
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

// ListUsersByRole 暂委托给 RbacService.RoleUserList（跨域查询仍走 s.db）
func (a *UserLookupAdapter) ListUsersByRole(ctx context.Context, roleID int64) ([]rbac.UserInfo, error) {
	return nil, nil
}
