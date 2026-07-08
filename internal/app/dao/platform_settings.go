package dao

import (
	"context"
	"fmt"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/utils"
	"strconv"
)

// PlatformSettingsGetAll 获取所有平台设置
func (d *Dao) PlatformSettingsGetAll(ctx context.Context) ([]*models.PlatformSettings, error) {
	var m models.PlatformSettings
	return m.GetAll(d.db.WithContext(ctx))
}

// PlatformSettingsGetByCategory 获取指定分类的设置
func (d *Dao) PlatformSettingsGetByCategory(ctx context.Context, category string) ([]*models.PlatformSettings, error) {
	var m models.PlatformSettings
	return m.GetByCategory(d.db.WithContext(ctx), category)
}

// PlatformSettingsUpdate 批量更新设置
func (d *Dao) PlatformSettingsUpdate(ctx context.Context, settings []*models.PlatformSettings) error {
	now := uint32(utils.NowUnix())
	for _, s := range settings {
		s.ModifiedAt = now
	}
	return models.BatchUpsertSettings(d.db.WithContext(ctx), settings)
}

// PlatformSettingsUpsert 更新或插入单个设置
func (d *Dao) PlatformSettingsUpsert(ctx context.Context, category, key, value, valueType, label, desc string) error {
	now := uint32(utils.NowUnix())
	s := &models.PlatformSettings{
		Category:   category,
		Key:        key,
		Value:      value,
		ValueType:  valueType,
		Label:      label,
		Desc:       desc,
		ModifiedAt: now,
	}
	return s.Upsert(d.db.WithContext(ctx))
}

// PlatformSettingsInitDefaults 初始化默认设置（首次启动时调用）
func (d *Dao) PlatformSettingsInitDefaults(ctx context.Context) error {
	// 检查是否已有设置
	existing, _ := d.PlatformSettingsGetAll(ctx)
	if len(existing) > 0 {
		return nil // 已有设置，跳过初始化
	}

	defaults := []*models.PlatformSettings{
		// 基础设置
		{Category: "basic", Key: "default_page", Value: "/clusters", ValueType: "string", Label: "默认进入页", Desc: "用户登录后默认跳转的页面"},
		{Category: "basic", Key: "default_cluster", Value: "auto", ValueType: "string", Label: "默认集群", Desc: "进入集群相关页面时的默认选择"},
		{Category: "basic", Key: "language", Value: "zh-CN", ValueType: "string", Label: "界面语言", Desc: "平台显示语言"},
		{Category: "basic", Key: "timezone", Value: "Asia/Shanghai", ValueType: "string", Label: "时区设置", Desc: "影响日志和告警的时间显示"},
		// 安全设置
		{Category: "security", Key: "session_timeout", Value: "120", ValueType: "int", Label: "会话超时", Desc: "用户无操作后自动登出的时间（分钟）"},
		{Category: "security", Key: "enable_2fa", Value: "false", ValueType: "bool", Label: "双因素认证", Desc: "强制用户使用2FA登录"},
		{Category: "security", Key: "password_policy", Value: "medium", ValueType: "string", Label: "密码强度要求", Desc: "设置密码复杂度规则"},
		{Category: "security", Key: "audit_retention", Value: "30", ValueType: "int", Label: "审计日志保留", Desc: "审计日志的保留时间（天）"},
		// 告警设置
		{Category: "alert", Key: "cpu_threshold", Value: "80", ValueType: "int", Label: "CPU使用率告警", Desc: "超过阈值时触发告警"},
		{Category: "alert", Key: "mem_threshold", Value: "80", ValueType: "int", Label: "内存使用率告警", Desc: "超过阈值时触发告警"},
		{Category: "alert", Key: "disk_threshold", Value: "85", ValueType: "int", Label: "磁盘使用率告警", Desc: "超过阈值时触发告警"},
		{Category: "alert", Key: "alert_silence", Value: "15", ValueType: "int", Label: "告警静默期", Desc: "相同告警的重复通知间隔（分钟）"},
		// 通知设置
		{Category: "notification", Key: "enable_email", Value: "false", ValueType: "bool", Label: "邮件通知", Desc: "通过邮件发送告警通知"},
		{Category: "notification", Key: "smtp_server", Value: "", ValueType: "string", Label: "SMTP服务器", Desc: "邮件服务器地址"},
		{Category: "notification", Key: "enable_dingtalk", Value: "false", ValueType: "bool", Label: "钉钉通知", Desc: "通过钉钉机器人推送告警"},
		{Category: "notification", Key: "dingtalk_webhook", Value: "", ValueType: "string", Label: "钉钉Webhook", Desc: "钉钉机器人Webhook地址"},
		{Category: "notification", Key: "enable_webhook", Value: "false", ValueType: "bool", Label: "Webhook通知", Desc: "发送到自定义Webhook端点"},
		{Category: "notification", Key: "webhook_url", Value: "", ValueType: "string", Label: "Webhook URL", Desc: "自定义Webhook地址"},
	}

	return d.PlatformSettingsUpdate(ctx, defaults)
}

// RequestToSettings 将前端请求转换为设置列表
func RequestToSettings(req *models.PlatformSettingsResponse) []*models.PlatformSettings {
	var settings []*models.PlatformSettings

	// Basic
	settings = append(settings,
		&models.PlatformSettings{Category: "basic", Key: "default_page", Value: req.Basic.DefaultPage, ValueType: "string"},
		&models.PlatformSettings{Category: "basic", Key: "default_cluster", Value: req.Basic.DefaultCluster, ValueType: "string"},
		&models.PlatformSettings{Category: "basic", Key: "language", Value: req.Basic.Language, ValueType: "string"},
		&models.PlatformSettings{Category: "basic", Key: "timezone", Value: req.Basic.Timezone, ValueType: "string"},
	)

	// Security
	settings = append(settings,
		&models.PlatformSettings{Category: "security", Key: "session_timeout", Value: fmt.Sprintf("%d", req.Security.SessionTimeout), ValueType: "int"},
		&models.PlatformSettings{Category: "security", Key: "enable_2fa", Value: strconv.FormatBool(req.Security.Enable2FA), ValueType: "bool"},
		&models.PlatformSettings{Category: "security", Key: "password_policy", Value: req.Security.PasswordPolicy, ValueType: "string"},
		&models.PlatformSettings{Category: "security", Key: "audit_retention", Value: fmt.Sprintf("%d", req.Security.AuditRetention), ValueType: "int"},
	)

	// Alert
	settings = append(settings,
		&models.PlatformSettings{Category: "alert", Key: "cpu_threshold", Value: fmt.Sprintf("%d", req.Alert.CPUThreshold), ValueType: "int"},
		&models.PlatformSettings{Category: "alert", Key: "mem_threshold", Value: fmt.Sprintf("%d", req.Alert.MemThreshold), ValueType: "int"},
		&models.PlatformSettings{Category: "alert", Key: "disk_threshold", Value: fmt.Sprintf("%d", req.Alert.DiskThreshold), ValueType: "int"},
		&models.PlatformSettings{Category: "alert", Key: "alert_silence", Value: fmt.Sprintf("%d", req.Alert.AlertSilence), ValueType: "int"},
	)

	// Notification
	settings = append(settings,
		&models.PlatformSettings{Category: "notification", Key: "enable_email", Value: strconv.FormatBool(req.Notification.EnableEmail), ValueType: "bool"},
		&models.PlatformSettings{Category: "notification", Key: "smtp_server", Value: req.Notification.SMTPServer, ValueType: "string"},
		&models.PlatformSettings{Category: "notification", Key: "enable_dingtalk", Value: strconv.FormatBool(req.Notification.EnableDingTalk), ValueType: "bool"},
		&models.PlatformSettings{Category: "notification", Key: "dingtalk_webhook", Value: req.Notification.DingTalkWebhook, ValueType: "string"},
		&models.PlatformSettings{Category: "notification", Key: "enable_webhook", Value: strconv.FormatBool(req.Notification.EnableWebhook), ValueType: "bool"},
		&models.PlatformSettings{Category: "notification", Key: "webhook_url", Value: req.Notification.WebhookURL, ValueType: "string"},
	)

	return settings
}
