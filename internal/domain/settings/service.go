package settings

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"k8soperation/pkg/setting"
)

// SettingsService 平台设置领域服务
type SettingsService struct {
	repo            SettingsRepository
	platformSetting *setting.PlatformSettingsS
}

// NewSettingsService 创建设置服务
func NewSettingsService(repo SettingsRepository, platformSetting *setting.PlatformSettingsS) *SettingsService {
	return &SettingsService{repo: repo, platformSetting: platformSetting}
}


// ========== 业务方法 ==========

// GetAll 获取所有设置（DB 原始数据）
func (s *SettingsService) GetAll(ctx context.Context) ([]*PlatformSettings, error) {
	return s.repo.FindAll(ctx)
}

// Get 获取聚合响应（混合模式：数据库 > config.yaml > 程序默认值）
func (s *SettingsService) Get(ctx context.Context) (*PlatformSettingsResponse, error) {
	resp := s.getConfigDefaults()

	dbSettings, err := s.GetAll(ctx)
	if err == nil && len(dbSettings) > 0 {
		s.mergeDBSettings(resp, dbSettings)
	}

	s.mergeSensitiveFromConfig(resp)
	return resp, nil
}

// Update 更新设置（只存非敏感配置到数据库）
func (s *SettingsService) Update(ctx context.Context, req *PlatformSettingsResponse) error {
	req.Notification.SMTPServer = ""
	req.Notification.DingTalkWebhook = ""
	req.Notification.WebhookURL = ""

	settings := s.requestToSettings(req)
	now := uint32(time.Now().Unix())
	for _, item := range settings {
		item.ModifiedAt = now
	}
	return s.repo.BatchUpsert(ctx, settings)
}

// Reset 重置为 config.yaml 默认设置
func (s *SettingsService) Reset(ctx context.Context) (*PlatformSettingsResponse, error) {
	defaults := s.getConfigDefaults()
	if err := s.Update(ctx, defaults); err != nil {
		return nil, err
	}
	s.mergeSensitiveFromConfig(defaults)
	return defaults, nil
}

// ========== 私有方法 ==========

// getConfigDefaults 从 config.yaml 获取默认配置
func (s *SettingsService) getConfigDefaults() *PlatformSettingsResponse {
	cfg := s.platformSetting
	if cfg == nil {
		return &PlatformSettingsResponse{
			Basic: BasicSettings{
				DefaultPage: "/clusters", DefaultCluster: "auto",
				Language: "zh-CN", Timezone: "Asia/Shanghai",
			},
			Security: SecuritySettings{
				SessionTimeout: 120, Enable2FA: false,
				PasswordPolicy: "medium", AuditRetention: 30,
			},
			Alert: AlertSettings{
				CPUThreshold: 80, MemThreshold: 80,
				DiskThreshold: 85, AlertSilence: 15,
			},
			Notification: NotificationSettings{},
			About: AboutSettings{
				Version: "2.0.0", BuildDate: "2026-03-04",
				GoVersion: "1.21", VueVersion: "3.5.13",
				DBType: "MySQL 8.0", K8sSupport: "v1.25+",
			},
		}
	}

	return &PlatformSettingsResponse{
		Basic: BasicSettings{
			DefaultPage: cfg.Basic.DefaultPage, DefaultCluster: cfg.Basic.DefaultCluster,
			Language: cfg.Basic.Language, Timezone: cfg.Basic.Timezone,
		},
		Security: SecuritySettings{
			SessionTimeout: cfg.Security.SessionTimeout, Enable2FA: cfg.Security.Enable2FA,
			PasswordPolicy: cfg.Security.PasswordPolicy, AuditRetention: cfg.Security.AuditRetention,
		},
		Alert: AlertSettings{
			CPUThreshold: cfg.Alert.CPUThreshold, MemThreshold: cfg.Alert.MemThreshold,
			DiskThreshold: cfg.Alert.DiskThreshold, AlertSilence: cfg.Alert.AlertSilence,
		},
		Notification: NotificationSettings{
			EnableEmail: cfg.Notification.EnableEmail, EnableDingTalk: cfg.Notification.EnableDingTalk,
			EnableWebhook: cfg.Notification.EnableWebhook,
		},
		About: AboutSettings{
			Version: cfg.About.Version, BuildDate: cfg.About.BuildDate,
			GoVersion: cfg.About.GoVersion, VueVersion: cfg.About.VueVersion,
			DBType: cfg.About.DBType, K8sSupport: cfg.About.K8sSupport,
		},
	}
}

// mergeDBSettings 用数据库设置覆盖默认值
func (s *SettingsService) mergeDBSettings(resp *PlatformSettingsResponse, dbSettings []*PlatformSettings) {
	dbResp := SettingsToResponse(dbSettings)

	if dbResp.Basic.DefaultPage != "" {
		resp.Basic.DefaultPage = dbResp.Basic.DefaultPage
	}
	if dbResp.Basic.DefaultCluster != "" {
		resp.Basic.DefaultCluster = dbResp.Basic.DefaultCluster
	}
	if dbResp.Basic.Language != "" {
		resp.Basic.Language = dbResp.Basic.Language
	}
	if dbResp.Basic.Timezone != "" {
		resp.Basic.Timezone = dbResp.Basic.Timezone
	}

	if dbResp.Security.SessionTimeout > 0 {
		resp.Security.SessionTimeout = dbResp.Security.SessionTimeout
	}
	resp.Security.Enable2FA = dbResp.Security.Enable2FA
	if dbResp.Security.PasswordPolicy != "" {
		resp.Security.PasswordPolicy = dbResp.Security.PasswordPolicy
	}
	if dbResp.Security.AuditRetention > 0 {
		resp.Security.AuditRetention = dbResp.Security.AuditRetention
	}

	if dbResp.Alert.CPUThreshold > 0 {
		resp.Alert.CPUThreshold = dbResp.Alert.CPUThreshold
	}
	if dbResp.Alert.MemThreshold > 0 {
		resp.Alert.MemThreshold = dbResp.Alert.MemThreshold
	}
	if dbResp.Alert.DiskThreshold > 0 {
		resp.Alert.DiskThreshold = dbResp.Alert.DiskThreshold
	}
	if dbResp.Alert.AlertSilence > 0 {
		resp.Alert.AlertSilence = dbResp.Alert.AlertSilence
	}

	resp.Notification.EnableEmail = dbResp.Notification.EnableEmail
	resp.Notification.EnableDingTalk = dbResp.Notification.EnableDingTalk
	resp.Notification.EnableWebhook = dbResp.Notification.EnableWebhook
}

// mergeSensitiveFromConfig 合并敏感信息（只从 config.yaml 读取）
func (s *SettingsService) mergeSensitiveFromConfig(resp *PlatformSettingsResponse) {
	cfg := s.platformSetting
	if cfg == nil {
		return
	}

	if cfg.Notification.SMTP.Server != "" {
		resp.Notification.SMTPServer = cfg.Notification.SMTP.Server
	}

	if cfg.Notification.DingTalk.Webhook != "" {
		resp.Notification.DingTalkWebhook = maskWebhook(cfg.Notification.DingTalk.Webhook)
	}

	if cfg.Notification.Webhook.URL != "" {
		resp.Notification.WebhookURL = maskWebhook(cfg.Notification.Webhook.URL)
	}
}

// requestToSettings 将前端请求转换为设置列表
func (s *SettingsService) requestToSettings(req *PlatformSettingsResponse) []*PlatformSettings {
	var list []*PlatformSettings

	list = append(list,
		&PlatformSettings{Category: "basic", Key: "default_page", Value: req.Basic.DefaultPage, ValueType: "string"},
		&PlatformSettings{Category: "basic", Key: "default_cluster", Value: req.Basic.DefaultCluster, ValueType: "string"},
		&PlatformSettings{Category: "basic", Key: "language", Value: req.Basic.Language, ValueType: "string"},
		&PlatformSettings{Category: "basic", Key: "timezone", Value: req.Basic.Timezone, ValueType: "string"},
	)

	list = append(list,
		&PlatformSettings{Category: "security", Key: "session_timeout", Value: fmt.Sprintf("%d", req.Security.SessionTimeout), ValueType: "int"},
		&PlatformSettings{Category: "security", Key: "enable_2fa", Value: strconv.FormatBool(req.Security.Enable2FA), ValueType: "bool"},
		&PlatformSettings{Category: "security", Key: "password_policy", Value: req.Security.PasswordPolicy, ValueType: "string"},
		&PlatformSettings{Category: "security", Key: "audit_retention", Value: fmt.Sprintf("%d", req.Security.AuditRetention), ValueType: "int"},
	)

	list = append(list,
		&PlatformSettings{Category: "alert", Key: "cpu_threshold", Value: fmt.Sprintf("%d", req.Alert.CPUThreshold), ValueType: "int"},
		&PlatformSettings{Category: "alert", Key: "mem_threshold", Value: fmt.Sprintf("%d", req.Alert.MemThreshold), ValueType: "int"},
		&PlatformSettings{Category: "alert", Key: "disk_threshold", Value: fmt.Sprintf("%d", req.Alert.DiskThreshold), ValueType: "int"},
		&PlatformSettings{Category: "alert", Key: "alert_silence", Value: fmt.Sprintf("%d", req.Alert.AlertSilence), ValueType: "int"},
	)

	list = append(list,
		&PlatformSettings{Category: "notification", Key: "enable_email", Value: strconv.FormatBool(req.Notification.EnableEmail), ValueType: "bool"},
		&PlatformSettings{Category: "notification", Key: "enable_dingtalk", Value: strconv.FormatBool(req.Notification.EnableDingTalk), ValueType: "bool"},
		&PlatformSettings{Category: "notification", Key: "enable_webhook", Value: strconv.FormatBool(req.Notification.EnableWebhook), ValueType: "bool"},
	)

	return list
}

// maskWebhook 脱敏 URL，只显示前部分
func maskWebhook(url string) string {
	if len(url) <= 20 {
		return "***"
	}
	return url[:20] + "***"
}
