package services

import (
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/domain/user"
	"k8soperation/internal/infra/persistence"
)

func (s *Services) userSvc() *user.UserService {
	return user.NewUserService(persistence.NewUserRepository(s.db), s.db, s.logger, s.eventBus)
}

// UserGetByID 根据 ID 获取用户
func (s *Services) UserGetByID(id int64) (*models.User, error) {
	return s.userSvc().GetByID(id)
}

// UserGetByStringID 根据字符串 ID 获取用户（用于 auth 中间件等场景）
func (s *Services) UserGetByStringID(id string) (*models.User, error) {
	return s.userSvc().GetByStringID(id)
}

// UserGetByName 根据用户名获取用户（用于初始化等场景）
func (s *Services) UserGetByName(username string) (*models.User, error) {
	return s.userSvc().GetByName(username)
}

// UserCreate 创建用户
func (s *Services) UserCreate(parm *requests.UserCreateRequest) (*models.User, error) {
	return s.userSvc().Create(parm.Username, parm.Password, parm.TenantID)
}

// UserCreateSimple 直接参数创建用户（用于初始化等场景）
func (s *Services) UserCreateSimple(username, password string, tenantID uint32) (*models.User, error) {
	return s.userSvc().Create(username, password, tenantID)
}

// UserDelete 删除用户
func (s *Services) UserDelete(param *requests.CommonIdRequest) error {
	return s.userSvc().Delete(param.ID)
}

// UserUpdate 更新用户
func (s *Services) UserUpdate(param *requests.UserUpdateRequest) error {
	return s.userSvc().Update(param.ID, param.Username, param.Password, param.Role, param.Status)
}

// UserList 用户列表
func (s *Services) UserList(param *requests.UserListRequest) ([]*models.User, int64, error) {
	return s.userSvc().List(param.Username, param.Role, param.Status, param.Page, param.Limit)
}

// MigrateUserPassword 将用户密码迁移到 bcrypt 格式
func (s *Services) MigrateUserPassword(userID uint32, plainPassword string) error {
	return s.userSvc().MigratePassword(userID, plainPassword)
}

// UserBatchImport 批量导入用户
func (s *Services) UserBatchImport(param *requests.UserBatchImportRequest) *requests.UserBatchImportResult {
	items := make([]user.BatchImportItem, len(param.Users))
	for i, u := range param.Users {
		items[i] = user.BatchImportItem{
			Username: u.Username, Password: u.Password, Email: u.Email, Phone: u.Phone, Role: u.Role,
		}
	}
	result := s.userSvc().BatchImport(items, param.DefaultPassword, param.DefaultRole, param.SkipExisting)

	// Convert back to requests types
	details := make([]requests.UserBatchImportDetail, len(result.Details))
	for i, d := range result.Details {
		details[i] = requests.UserBatchImportDetail{
			Username: d.Username, Status: d.Status, Message: d.Message, UserID: d.UserID,
		}
	}
	return &requests.UserBatchImportResult{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
		Skipped: result.Skipped,
		Details: details,
	}
}
