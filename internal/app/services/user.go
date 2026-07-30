package services

import (
	"fmt"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// UserCreate 创建用户
func (s *Services) UserCreate(parm *requests.UserCreateRequest) (*models.User, error) {
	return s.dao.UserCreate(parm.Username, parm.Password, parm.TenantID)
}

// UserDelete 删除用户
func (s *Services) UserDelete(param *requests.CommonIdRequest) error {
	return s.dao.UserDelete(param.ID)
}

// UserUpdate 更新用户
func (s *Services) UserUpdate(param *requests.UserUpdateRequest) error {
	return s.dao.UserUpdate(param.ID, param.Username, param.Password, param.Role, param.Status)
}

func (s *Services) UserList(param *requests.UserListRequest) ([]*models.User, int64, error) {
	return s.dao.UserList(param.Username, param.Role, param.Status, param.Page, param.Limit)
}

// MigrateUserPassword 将用户密码迁移到 bcrypt 格式
func (s *Services) MigrateUserPassword(userID uint32, plainPassword string) error {
	return s.dao.UserMigratePassword(userID, plainPassword)
}

// UserBatchImport 批量导入用户
func (s *Services) UserBatchImport(param *requests.UserBatchImportRequest) *requests.UserBatchImportResult {
	result := &requests.UserBatchImportResult{
		Total:   len(param.Users),
		Details: make([]requests.UserBatchImportDetail, 0, len(param.Users)),
	}

	for _, item := range param.Users {
		detail := requests.UserBatchImportDetail{
			Username: item.Username,
		}

		// 检查用户是否已存在
		exists, err := s.dao.UserExistsByUsername(item.Username)
		if err != nil {
			detail.Status = "failed"
			detail.Message = fmt.Sprintf("检查用户失败: %s", err.Error())
			result.Failed++
			result.Details = append(result.Details, detail)
			continue
		}

		if exists {
			if param.SkipExisting {
				detail.Status = "skipped"
				detail.Message = "用户已存在，已跳过"
				result.Skipped++
			} else {
				detail.Status = "failed"
				detail.Message = "用户名已存在"
				result.Failed++
			}
			result.Details = append(result.Details, detail)
			continue
		}

		// 确定密码
		password := item.Password
		if password == "" {
			password = param.DefaultPassword
		}
		if password == "" {
			password = "123456"
		}

		// 确定角色
		role := item.Role
		if role == "" {
			role = param.DefaultRole
		}
		if role == "" {
			role = "developer"
		}

		// 创建用户
		user, err := s.dao.UserCreateFull(item.Username, password, item.Email, item.Phone, role)
		if err != nil {
			detail.Status = "failed"
			detail.Message = fmt.Sprintf("创建失败: %s", err.Error())
			result.Failed++
		} else {
			detail.Status = "success"
			detail.Message = "创建成功"
			detail.UserID = user.ID
			result.Success++

			// 如果指定了角色，同步分配 RBAC 角色
			if role != "" {
				_ = s.assignRoleToUser(int64(user.ID), role)
			}
		}
		result.Details = append(result.Details, detail)
	}

	return result
}

// assignRoleToUser 内部方法：根据角色名称给用户分配 RBAC 角色
func (s *Services) assignRoleToUser(userID int64, roleType string) error {
	// 查找角色ID
	role, err := s.dao.RoleGetByName(roleType)
	if err != nil || role == nil {
		return fmt.Errorf("角色 %s 不存在", roleType)
	}
	// 分配角色
	return s.dao.UserRoleAssign(userID, []int64{role.ID}, 0)
}
