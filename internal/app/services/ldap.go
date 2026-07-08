package services

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	ldapclient "k8soperation/pkg/ldap"
	"k8soperation/pkg/setting"
)

// LDAPLogin LDAP 登录认证
// 返回: 用户对象, 是否是LDAP认证, error
func (s *Services) LDAPLogin(username, password string) (*models.User, bool, error) {
	// 检查 LDAP 是否启用
	if !ldapclient.IsEnabled() {
		return nil, false, fmt.Errorf("LDAP 未启用")
	}

	client := ldapclient.NewClient()

	// 1. LDAP 认证
	ldapUser, err := client.Authenticate(username, password)
	if err != nil {
		global.Logger.Warn("LDAP 认证失败",
			zap.String("username", username),
			zap.Error(err))

		// 如果启用了本地回退，返回特定错误让调用方尝试本地认证
		if global.LDAPSetting.LocalFallback {
			return nil, false, fmt.Errorf("LDAP_AUTH_FAILED")
		}
		return nil, true, fmt.Errorf("LDAP 认证失败: %w", err)
	}

	// 2. 本地用户同步（JIT Provisioning）
	user, err := s.syncLDAPUser(ldapUser)
	if err != nil {
		return nil, true, fmt.Errorf("同步 LDAP 用户失败: %w", err)
	}

	// 3. 同步 LDAP 组到平台角色
	if global.LDAPSetting.SyncOnLogin && len(ldapUser.Groups) > 0 {
		if err := s.syncLDAPRoles(user, ldapUser.Groups); err != nil {
			global.Logger.Warn("同步 LDAP 角色失败",
				zap.String("username", username),
				zap.Error(err))
			// 角色同步失败不阻断登录
		}
	}

	return user, true, nil
}

// syncLDAPUser 同步 LDAP 用户到本地数据库（首次创建/后续更新）
func (s *Services) syncLDAPUser(ldapUser *ldapclient.UserInfo) (*models.User, error) {
	// 查找本地是否已存在
	existing, err := s.dao.UserGetByName(ldapUser.Username)
	if err == nil && existing != nil {
		// 用户已存在，更新邮箱/手机等信息
		nowTime := uint32(time.Now().Unix())
		values := map[string]interface{}{
			"modified_at": nowTime,
		}
		if ldapUser.Email != "" {
			values["email"] = ldapUser.Email
		}
		if ldapUser.Phone != "" {
			values["phone"] = ldapUser.Phone
		}
		// 确保用户状态为激活
		values["status"] = int8(1)

		_ = existing.Update(global.DB, values)
		return existing, nil
	}

	// 用户不存在，需要自动创建
	if !global.LDAPSetting.AutoCreate {
		return nil, fmt.Errorf("用户 %s 在平台中不存在，且未开启自动创建", ldapUser.Username)
	}

	// 创建本地用户（密码使用随机值，因为走 LDAP 认证不需要本地密码）
	randomPass := fmt.Sprintf("LDAP_%d_%s", time.Now().UnixNano(), ldapUser.Username)
	user, err := s.dao.UserCreateFull(
		ldapUser.Username,
		randomPass,
		ldapUser.Email,
		ldapUser.Phone,
		"user", // role 字段用默认值，实际权限通过 RBAC 控制
	)
	if err != nil {
		return nil, fmt.Errorf("创建本地用户失败: %w", err)
	}

	global.Logger.Info("LDAP 用户首次登录，已自动创建平台账号",
		zap.String("username", ldapUser.Username),
		zap.Uint32("user_id", user.ID))

	return user, nil
}

// syncLDAPRoles 根据 LDAP 组同步平台角色
func (s *Services) syncLDAPRoles(user *models.User, groups []string) error {
	client := ldapclient.NewClient()

	// 获取最高优先级的角色映射
	mapping := client.GetRoleMappingForGroups(groups)
	if mapping == nil {
		global.Logger.Debug("未找到匹配的 LDAP 组映射",
			zap.String("username", user.Username),
			zap.Strings("groups", groups))
		return nil
	}

	// 查找平台角色
	role, err := s.dao.RoleGetByName(mapping.PlatformRole)
	if err != nil {
		return fmt.Errorf("角色 %s 不存在: %w", mapping.PlatformRole, err)
	}

	// 分配角色（替换现有角色）
	if err := s.dao.UserRoleAssign(int64(user.ID), []int64{role.ID}, 0); err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}

	global.Logger.Info("LDAP 角色同步完成",
		zap.String("username", user.Username),
		zap.Strings("ldap_groups", groups),
		zap.String("mapped_role", mapping.PlatformRole))

	return nil
}

// LDAPTestConnection 测试 LDAP 连接
func (s *Services) LDAPTestConnection() error {
	if !ldapclient.IsEnabled() {
		return fmt.Errorf("LDAP 未启用")
	}
	client := ldapclient.NewClient()
	return client.TestConnection()
}

// LDAPSyncAllUsers 全量同步 LDAP 用户
func (s *Services) LDAPSyncAllUsers() (*LDAPSyncResult, error) {
	if !ldapclient.IsEnabled() {
		return nil, fmt.Errorf("LDAP 未启用")
	}

	client := ldapclient.NewClient()
	ldapUsers, err := client.SearchUsers("", 1000)
	if err != nil {
		return nil, fmt.Errorf("搜索 LDAP 用户失败: %w", err)
	}

	result := &LDAPSyncResult{
		Total: len(ldapUsers),
	}

	for _, ldapUser := range ldapUsers {
		user, err := s.syncLDAPUser(ldapUser)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", ldapUser.Username, err.Error()))
			continue
		}

		// 同步组
		if len(ldapUser.Groups) > 0 {
			_ = s.syncLDAPRoles(user, ldapUser.Groups)
		}

		result.Synced++
	}

	return result, nil
}

// LDAPGetConfig 获取当前 LDAP 配置（脱敏）
func (s *Services) LDAPGetConfig() *LDAPConfigResponse {
	cfg := global.LDAPSetting
	if cfg == nil {
		return &LDAPConfigResponse{Enabled: false}
	}

	return &LDAPConfigResponse{
		Enabled:       cfg.Enabled,
		Host:          cfg.Host,
		Port:          cfg.Port,
		UseTLS:        cfg.UseTLS,
		BindDN:        cfg.BindDN,
		BaseDN:        cfg.BaseDN,
		UserFilter:    cfg.UserFilter,
		GroupBaseDN:   cfg.GroupBaseDN,
		GroupFilter:   cfg.GroupFilter,
		GroupAttr:     cfg.GroupAttr,
		AttrUsername:  cfg.AttrUsername,
		AttrEmail:     cfg.AttrEmail,
		AttrPhone:     cfg.AttrPhone,
		SyncOnLogin:   cfg.SyncOnLogin,
		AutoCreate:    cfg.AutoCreate,
		LocalFallback: cfg.LocalFallback,
		GroupRoleMapping: cfg.GroupRoleMapping,
	}
}

// ==================== 响应结构 ====================

// LDAPSyncResult 同步结果
type LDAPSyncResult struct {
	Total  int      `json:"total"`
	Synced int      `json:"synced"`
	Failed int      `json:"failed"`
	Errors []string `json:"errors,omitempty"`
}

// LDAPConfigResponse LDAP 配置响应（脱敏）
type LDAPConfigResponse struct {
	Enabled       bool                          `json:"enabled"`
	Host          string                        `json:"host"`
	Port          int                           `json:"port"`
	UseTLS        bool                          `json:"use_tls"`
	BindDN        string                        `json:"bind_dn"`
	BaseDN        string                        `json:"base_dn"`
	UserFilter    string                        `json:"user_filter"`
	GroupBaseDN   string                        `json:"group_base_dn"`
	GroupFilter   string                        `json:"group_filter"`
	GroupAttr     string                        `json:"group_attr"`
	AttrUsername  string                        `json:"attr_username"`
	AttrEmail     string                        `json:"attr_email"`
	AttrPhone     string                        `json:"attr_phone"`
	SyncOnLogin   bool                          `json:"sync_on_login"`
	AutoCreate    bool                          `json:"auto_create"`
	LocalFallback bool                          `json:"local_fallback"`
	GroupRoleMapping []setting.LDAPGroupRoleMapping `json:"group_role_mapping"`
}
