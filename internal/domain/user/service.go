package user

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/events"
	"k8soperation/pkg/db"
	"k8soperation/pkg/logger"
	"k8soperation/pkg/utils"
)

// UserService 用户领域服务
type UserService struct {
	repo      UserRepository
	db        *gorm.DB // 仅用于 assignRole 跨域 RBAC 查询，待 RBAC 域提供接口后移除
	logger    *logger.Logger
	publisher events.EventPublisher
}

// NewUserService 创建用户服务
func NewUserService(repo UserRepository, db *gorm.DB, logger *logger.Logger, publisher events.EventPublisher) *UserService {
	return &UserService{repo: repo, db: db, logger: logger, publisher: publisher}
}

// ========== 基础 CRUD ==========

// Create 创建用户（含密码哈希——领域逻辑）
func (s *UserService) Create(name, password string, tenantID uint32) (*User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if tenantID == 0 {
		tenantID = 1
	}
	nowTime := uint32(time.Now().Unix())
	user := &User{
		Username: name,
		Password: hashedPassword,
		Base:     &db.Base{TenantID: tenantID, CreatedAt: nowTime, ModifiedAt: nowTime},
	}
	if err := s.repo.Save(context.Background(), user); err != nil {
		return nil, err
	}
	return user, nil
}

// CreateFull 创建用户（含邮箱、手机、角色）
func (s *UserService) CreateFull(name, password, email, phone, role string, tenantID uint32) (*User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	nowTime := uint32(time.Now().Unix())
	if tenantID == 0 {
		tenantID = 1
	}
	user := &User{
		Username: name,
		Password: hashedPassword,
		Email:    email,
		Phone:    phone,
		Role:     role,
		Base:     &db.Base{TenantID: tenantID, CreatedAt: nowTime, ModifiedAt: nowTime},
	}
	if err := s.repo.Save(context.Background(), user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateFields 按 map 更新用户字段
func (s *UserService) UpdateFields(id uint32, values map[string]interface{}) error {
	return s.repo.Update(context.Background(), id, values)
}

// Delete 删除用户
func (s *UserService) Delete(id uint32) error {
	return s.repo.Delete(context.Background(), id)
}

// Update 更新用户
func (s *UserService) Update(id uint32, name, password, role string, status int8) error {
	nowTime := uint32(time.Now().Unix())
	values := map[string]interface{}{
		"username": name, "modified_at": nowTime, "status": status,
	}
	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return err
		}
		values["password"] = hashedPassword
	}
	if role != "" {
		values["role"] = role
	}
	return s.repo.Update(context.Background(), id, values)
}

// List 用户列表
func (s *UserService) List(username, role, status string, page, limit int) ([]*User, int64, error) {
	return s.repo.Query(context.Background(), username, role, status, page, limit)
}

// GetByStringID 根据字符串 ID 获取用户
func (s *UserService) GetByStringID(id string) (*User, error) {
	return s.repo.FindByStringID(context.Background(), id)
}

// GetByName 根据用户名获取
func (s *UserService) GetByName(username string) (*User, error) {
	return s.repo.FindByName(context.Background(), username)
}

// GetByID 根据ID获取
func (s *UserService) GetByID(id int64) (*User, error) {
	return s.repo.FindByID(context.Background(), id)
}

// ExistsByName 检查用户名是否存在
func (s *UserService) ExistsByName(username string) (bool, error) {
	count, err := s.repo.CountByName(context.Background(), username)
	return count > 0, err
}

// MigratePassword 将用户密码迁移到 bcrypt 格式
func (s *UserService) MigratePassword(userID uint32, plainPassword string) error {
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	return s.repo.Update(context.Background(), userID, map[string]interface{}{
		"password": hashedPassword, "modified_at": uint32(time.Now().Unix()),
	})
}

// ========== 批量导入 ==========

type BatchImportResult = batchImportResult
type BatchImportDetail = batchImportDetail
type BatchImportItem = batchImportItem

type batchImportResult struct {
	Total   int                 `json:"total"`
	Success int                 `json:"success"`
	Failed  int                 `json:"failed"`
	Skipped int                 `json:"skipped"`
	Details []batchImportDetail `json:"details"`
}

type batchImportDetail struct {
	Username string `json:"username"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	UserID   uint32 `json:"user_id,omitempty"`
}

type batchImportItem struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// BatchImport 批量导入用户
func (s *UserService) BatchImport(items []BatchImportItem, defaultPassword, defaultRole string, skipExisting bool, tenantID uint32) *BatchImportResult {
	result := &BatchImportResult{Total: len(items), Details: make([]BatchImportDetail, 0, len(items))}

	for _, item := range items {
		detail := BatchImportDetail{Username: item.Username}
		exists, err := s.ExistsByName(item.Username)
		if err != nil {
			detail.Status = "failed"
			detail.Message = fmt.Sprintf("检查用户失败: %s", err.Error())
			result.Failed++
			result.Details = append(result.Details, detail)
			continue
		}
		if exists {
			if skipExisting {
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
		password := item.Password
		if password == "" {
			password = defaultPassword
		}
		if password == "" {
			password = "123456"
		}
		role := item.Role
		if role == "" {
			role = defaultRole
		}
		if role == "" {
			role = "developer"
		}
		user, err := s.CreateFull(item.Username, password, item.Email, item.Phone, role, tenantID)
		if err != nil {
			detail.Status = "failed"
			detail.Message = fmt.Sprintf("创建失败: %s", err.Error())
			result.Failed++
		} else {
			detail.Status = "success"
			detail.Message = "创建成功"
			detail.UserID = user.ID
			result.Success++
			if role != "" {
				_ = s.assignRole(int64(user.ID), role)
			}
		}
		result.Details = append(result.Details, detail)
	}
	return result
}

// ========== 认证相关 ==========

// UpdatePasswordByName 根据用户名更新密码
func (s *UserService) UpdatePasswordByName(username, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordByName(context.Background(), username, hashedPassword)
}

// ForgotPassword 忘记密码
func (s *UserService) ForgotPassword(username, newPassword, confirmPassword string) error {
	user, err := s.GetByName(username)
	if err != nil || user == nil || user.ID == 0 {
		return fmt.Errorf("user not found")
	}
	if newPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}
	return s.UpdatePasswordByName(username, newPassword)
}

// Register 用户注册
func (s *UserService) Register(username, password string) error {
	exists, err := s.ExistsByName(username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("用户名 %s 已注册", username)
	}
	user, err := s.Create(username, password, 0)
	if err != nil {
		return err
	}
	s.publish(NewUserRegistered(user.ID, username))
	return nil
}

func (s *UserService) publish(event events.DomainEvent) {
	if s.publisher == nil { return }
	defer func() { _ = recover() }()
	s.publisher.Publish(event)
}

// assignRole 分配角色（跨域 RBAC 查询，待 RBAC 域提供接口后迁移）
func (s *UserService) assignRole(userID int64, roleType string) error {
	var role struct{ ID int64 }
	if err := s.db.Table("sys_role").Select("id").Where("name = ? AND is_del = 0", roleType).First(&role).Error; err != nil {
		return fmt.Errorf("角色 %s 不存在", roleType)
	}
	if err := s.db.Table("sys_user_role").Where("user_id = ?", userID).Delete(nil).Error; err != nil {
		return err
	}
	now := uint64(time.Now().Unix())
	return s.db.Exec("INSERT INTO sys_user_role (user_id, role_id, created_at, created_by) VALUES (?, ?, ?, ?)",
		userID, role.ID, now, 0).Error
}
