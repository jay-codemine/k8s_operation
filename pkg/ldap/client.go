package ldap

import (
	"crypto/tls"
	"fmt"
	"strings"

	"k8soperation/global"
	"k8soperation/pkg/setting"

	"go.uber.org/zap"
	goldap "github.com/go-ldap/ldap/v3"
)

// UserInfo LDAP 用户信息
type UserInfo struct {
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DisplayName string   `json:"display_name"`
	DN          string   `json:"dn"`
	Groups      []string `json:"groups"`
}

// Client LDAP 客户端
type Client struct {
	config *setting.LDAPSettingS
}

// NewClient 创建 LDAP 客户端
func NewClient() *Client {
	return &Client{
		config: global.LDAPSetting,
	}
}

// IsEnabled 是否启用 LDAP
func IsEnabled() bool {
	return global.LDAPSetting != nil && global.LDAPSetting.Enabled
}

// connect 创建 LDAP 连接
func (c *Client) connect() (*goldap.Conn, error) {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	var conn *goldap.Conn
	var err error

	if c.config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.SkipVerify,
		}
		conn, err = goldap.DialTLS("tcp", addr, tlsConfig)
	} else {
		conn, err = goldap.Dial("tcp", addr)
	}

	if err != nil {
		return nil, fmt.Errorf("连接 LDAP 服务器失败: %w", err)
	}

	// 如果不是 TLS 但需要 StartTLS
	if !c.config.UseTLS && c.config.Port == 389 {
		// 可选 StartTLS（如果服务器支持）
	}

	return conn, nil
}

// Authenticate 验证用户密码
// 流程: 管理员 Bind → 搜索用户 → 用户 Bind 验证密码
func (c *Client) Authenticate(username, password string) (*UserInfo, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 1. 管理员 Bind
	if err := conn.Bind(c.config.BindDN, c.config.BindPassword); err != nil {
		return nil, fmt.Errorf("管理员 Bind 失败: %w", err)
	}

	// 2. 搜索用户
	userFilter := fmt.Sprintf(c.config.UserFilter, goldap.EscapeFilter(username))
	searchReq := goldap.NewSearchRequest(
		c.config.BaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		userFilter,
		[]string{
			c.getAttrUsername(),
			c.getAttrEmail(),
			c.getAttrPhone(),
			c.getAttrDisplayName(),
			"dn",
		},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索用户失败: %w", err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("用户 %s 在 LDAP 中不存在", username)
	}
	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("找到多个匹配用户，请检查 UserFilter 配置")
	}

	entry := result.Entries[0]
	userDN := entry.DN

	// 3. 用户 Bind 验证密码
	if err := conn.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("密码验证失败")
	}

	// 4. 提取用户信息
	userInfo := &UserInfo{
		Username:    entry.GetAttributeValue(c.getAttrUsername()),
		Email:       entry.GetAttributeValue(c.getAttrEmail()),
		Phone:       entry.GetAttributeValue(c.getAttrPhone()),
		DisplayName: entry.GetAttributeValue(c.getAttrDisplayName()),
		DN:          userDN,
	}

	// 如果用户名为空则用登录名
	if userInfo.Username == "" {
		userInfo.Username = username
	}

	// 5. 查询用户所属的组
	groups, err := c.getUserGroups(conn, username, userDN)
	if err != nil {
		global.Logger.Warn("查询 LDAP 用户组失败", zap.String("username", username), zap.Error(err))
	}
	userInfo.Groups = groups

	return userInfo, nil
}

// getUserGroups 获取用户所属的 LDAP 组
func (c *Client) getUserGroups(conn *goldap.Conn, username, userDN string) ([]string, error) {
	if c.config.GroupBaseDN == "" {
		return nil, nil
	}

	// 重新用管理员 Bind（因为用户 Bind 可能没权限搜索组）
	if err := conn.Bind(c.config.BindDN, c.config.BindPassword); err != nil {
		return nil, fmt.Errorf("管理员 Bind 失败: %w", err)
	}

	// 构建组搜索过滤器
	groupFilter := c.config.GroupFilter
	if groupFilter == "" {
		groupFilter = "(memberUid=%s)"
	}

	// 某些 LDAP 用 DN 查询组成员，某些用 uid
	filter := fmt.Sprintf(groupFilter, goldap.EscapeFilter(username))
	if strings.Contains(groupFilter, "member=") {
		filter = fmt.Sprintf(groupFilter, goldap.EscapeFilter(userDN))
	}

	groupAttr := c.getGroupAttr()
	searchReq := goldap.NewSearchRequest(
		c.config.GroupBaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{groupAttr},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索组失败: %w", err)
	}

	var groups []string
	for _, entry := range result.Entries {
		groupName := entry.GetAttributeValue(groupAttr)
		if groupName != "" {
			groups = append(groups, groupName)
		}
	}

	return groups, nil
}

// TestConnection 测试 LDAP 连接
func (c *Client) TestConnection() error {
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// 尝试管理员 Bind
	if err := conn.Bind(c.config.BindDN, c.config.BindPassword); err != nil {
		return fmt.Errorf("Bind DN 认证失败: %w", err)
	}

	return nil
}

// SearchUsers 搜索 LDAP 用户列表（用于同步）
func (c *Client) SearchUsers(filter string, limit int) ([]*UserInfo, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.Bind(c.config.BindDN, c.config.BindPassword); err != nil {
		return nil, fmt.Errorf("管理员 Bind 失败: %w", err)
	}

	if filter == "" {
		filter = "(objectClass=person)"
	}

	searchReq := goldap.NewSearchRequest(
		c.config.BaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		limit, 0, false,
		filter,
		[]string{
			c.getAttrUsername(),
			c.getAttrEmail(),
			c.getAttrPhone(),
			c.getAttrDisplayName(),
			"dn",
		},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索用户失败: %w", err)
	}

	var users []*UserInfo
	for _, entry := range result.Entries {
		user := &UserInfo{
			Username:    entry.GetAttributeValue(c.getAttrUsername()),
			Email:       entry.GetAttributeValue(c.getAttrEmail()),
			Phone:       entry.GetAttributeValue(c.getAttrPhone()),
			DisplayName: entry.GetAttributeValue(c.getAttrDisplayName()),
			DN:          entry.DN,
		}
		if user.Username != "" {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetRoleMappingForGroups 根据用户的 LDAP 组，返回最高权限的角色映射
func (c *Client) GetRoleMappingForGroups(groups []string) *setting.LDAPGroupRoleMapping {
	if len(c.config.GroupRoleMapping) == 0 {
		return nil
	}

	var bestMapping *setting.LDAPGroupRoleMapping
	bestLevel := 0

	for i := range c.config.GroupRoleMapping {
		mapping := &c.config.GroupRoleMapping[i]
		for _, group := range groups {
			if strings.EqualFold(mapping.LDAPGroup, group) {
				level := rolePriority(mapping.PlatformRole)
				if level > bestLevel {
					bestLevel = level
					bestMapping = mapping
				}
			}
		}
	}

	// 兜底：使用 "default" 映射
	if bestMapping == nil {
		for i := range c.config.GroupRoleMapping {
			if c.config.GroupRoleMapping[i].LDAPGroup == "default" {
				bestMapping = &c.config.GroupRoleMapping[i]
				break
			}
		}
	}

	return bestMapping
}

// rolePriority 角色优先级（越大权限越高）
func rolePriority(role string) int {
	switch role {
	case "super_admin":
		return 6
	case "platform_admin":
		return 5
	case "devops":
		return 4
	case "developer":
		return 3
	case "tester":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

// 属性获取辅助方法（带默认值）
func (c *Client) getAttrUsername() string {
	if c.config.AttrUsername != "" {
		return c.config.AttrUsername
	}
	return "uid"
}

func (c *Client) getAttrEmail() string {
	if c.config.AttrEmail != "" {
		return c.config.AttrEmail
	}
	return "mail"
}

func (c *Client) getAttrPhone() string {
	if c.config.AttrPhone != "" {
		return c.config.AttrPhone
	}
	return "mobile"
}

func (c *Client) getAttrDisplayName() string {
	if c.config.AttrDisplayName != "" {
		return c.config.AttrDisplayName
	}
	return "cn"
}

func (c *Client) getGroupAttr() string {
	if c.config.GroupAttr != "" {
		return c.config.GroupAttr
	}
	return "cn"
}
