package services

import (
	"context"
	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
)

// PlatformSettingsGet 获取所有平台设置（混合模式）
// 优先级：数据库 > config.yaml > 程序默认值
func (s *Services) PlatformSettingsGet(ctx context.Context) (*models.PlatformSettingsResponse, error) {
	// 1. 从 config.yaml 获取默认值
	resp := getConfigDefaults()

	// 2. 从数据库获取用户设置，覆盖默认值
	dbSettings, err := s.dao.PlatformSettingsGetAll(ctx)
	if err == nil && len(dbSettings) > 0 {
		mergeDBSettings(resp, dbSettings)
	}

	// 3. 敏感信息只从 config.yaml 读取（不存数据库）
	mergeSensitiveFromConfig(resp)

	return resp, nil
}

// PlatformSettingsUpdate 更新平台设置（只存非敏感配置到数据库）
func (s *Services) PlatformSettingsUpdate(ctx context.Context, req *models.PlatformSettingsResponse) error {
	// 敏感信息不写数据库，清空后再存
	req.Notification.SMTPServer = ""    // SMTP 密码不存
	req.Notification.DingTalkWebhook = "" // Webhook 不存
	req.Notification.WebhookURL = ""     // 自定义 Webhook 不存

	settings := dao.RequestToSettings(req)
	return s.dao.PlatformSettingsUpdate(ctx, settings)
}

// PlatformSettingsReset 重置为 config.yaml 默认设置
func (s *Services) PlatformSettingsReset(ctx context.Context) (*models.PlatformSettingsResponse, error) {
	// 从 config.yaml 获取默认值
	defaults := getConfigDefaults()

	// 更新到数据库
	if err := s.PlatformSettingsUpdate(ctx, defaults); err != nil {
		return nil, err
	}

	// 返回时合并敏感信息
	mergeSensitiveFromConfig(defaults)
	return defaults, nil
}

// getConfigDefaults 从 config.yaml 获取默认配置
func getConfigDefaults() *models.PlatformSettingsResponse {
	cfg := global.PlatformSetting
	if cfg == nil {
		// config.yaml 未配置，使用程序内默认值
		return &models.PlatformSettingsResponse{
			Basic: models.BasicSettings{
				DefaultPage:    "/clusters",
				DefaultCluster: "auto",
				Language:       "zh-CN",
				Timezone:       "Asia/Shanghai",
			},
			Security: models.SecuritySettings{
				SessionTimeout: 120,
				Enable2FA:      false,
				PasswordPolicy: "medium",
				AuditRetention: 30,
			},
			Alert: models.AlertSettings{
				CPUThreshold:  80,
				MemThreshold:  80,
				DiskThreshold: 85,
				AlertSilence:  15,
			},
			Notification: models.NotificationSettings{},
			About: models.AboutSettings{
				Version:    "2.0.0",
				BuildDate:  "2026-03-04",
				GoVersion:  "1.21",
				VueVersion: "3.5.13",
				DBType:     "MySQL 8.0",
				K8sSupport: "v1.25+",
			},
		}
	}

	return &models.PlatformSettingsResponse{
		Basic: models.BasicSettings{
			DefaultPage:    cfg.Basic.DefaultPage,
			DefaultCluster: cfg.Basic.DefaultCluster,
			Language:       cfg.Basic.Language,
			Timezone:       cfg.Basic.Timezone,
		},
		Security: models.SecuritySettings{
			SessionTimeout: cfg.Security.SessionTimeout,
			Enable2FA:      cfg.Security.Enable2FA,
			PasswordPolicy: cfg.Security.PasswordPolicy,
			AuditRetention: cfg.Security.AuditRetention,
		},
		Alert: models.AlertSettings{
			CPUThreshold:  cfg.Alert.CPUThreshold,
			MemThreshold:  cfg.Alert.MemThreshold,
			DiskThreshold: cfg.Alert.DiskThreshold,
			AlertSilence:  cfg.Alert.AlertSilence,
		},
		Notification: models.NotificationSettings{
			EnableEmail:    cfg.Notification.EnableEmail,
			EnableDingTalk: cfg.Notification.EnableDingTalk,
			EnableWebhook:  cfg.Notification.EnableWebhook,
		},
		About: models.AboutSettings{
			Version:    cfg.About.Version,
			BuildDate:  cfg.About.BuildDate,
			GoVersion:  cfg.About.GoVersion,
			VueVersion: cfg.About.VueVersion,
			DBType:     cfg.About.DBType,
			K8sSupport: cfg.About.K8sSupport,
		},
	}
}

// mergeDBSettings 用数据库设置覆盖默认值
func mergeDBSettings(resp *models.PlatformSettingsResponse, dbSettings []*models.PlatformSettings) {
	// 转换为响应格式，但保留原有的敏感信息
	dbResp := models.SettingsToResponse(dbSettings)

	// Basic
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

	// Security
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

	// Alert
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

	// Notification (开关从数据库，敏感信息从 config.yaml)
	resp.Notification.EnableEmail = dbResp.Notification.EnableEmail
	resp.Notification.EnableDingTalk = dbResp.Notification.EnableDingTalk
	resp.Notification.EnableWebhook = dbResp.Notification.EnableWebhook
}

// mergeSensitiveFromConfig 合并敏感信息（只从 config.yaml 读取）
func mergeSensitiveFromConfig(resp *models.PlatformSettingsResponse) {
	cfg := global.PlatformSetting
	if cfg == nil {
		return
	}

	// SMTP 服务器信息（不返回密码给前端，只显示服务器地址）
	if cfg.Notification.SMTP.Server != "" {
		resp.Notification.SMTPServer = cfg.Notification.SMTP.Server
	}

	// 钉钉 Webhook（脱敏显示，只显示前后部分）
	if cfg.Notification.DingTalk.Webhook != "" {
		resp.Notification.DingTalkWebhook = maskWebhook(cfg.Notification.DingTalk.Webhook)
	}

	// 自定义 Webhook（脱敏显示）
	if cfg.Notification.Webhook.URL != "" {
		resp.Notification.WebhookURL = maskWebhook(cfg.Notification.Webhook.URL)
	}
}

// maskWebhook 脱敏 URL，只显示前 20 字符 + ...
func maskWebhook(url string) string {
	if len(url) <= 25 {
		return url
	}
	return url[:20] + "..."
}
