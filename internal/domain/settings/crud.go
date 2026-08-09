package settings

import "encoding/json"

// 实体 AR 方法 (Create/Update/GetAll/GetByKey/BatchUpsert) 已迁移至
// internal/infra/persistence/settings_repo.go (Repository 模式)
//
// 本文件保留纯数据转换函数（SettingsToMap / SettingsToResponse 等）

// SettingsToMap 将设置列表转换为 map
func SettingsToMap(settings []*PlatformSettings) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, s := range settings {
		if result[s.Category] == nil {
			result[s.Category] = make(map[string]string)
		}
		result[s.Category][s.Key] = s.Value
	}
	return result
}

// SettingsToResponse 将设置列表转换为响应结构
func SettingsToResponse(settings []*PlatformSettings) *PlatformSettingsResponse {
	m := SettingsToMap(settings)

	resp := &PlatformSettingsResponse{
		Basic: BasicSettings{
			DefaultPage:    getOrDefault(m, "basic", "default_page", "/clusters"),
			DefaultCluster: getOrDefault(m, "basic", "default_cluster", "auto"),
			Language:       getOrDefault(m, "basic", "language", "zh-CN"),
			Timezone:       getOrDefault(m, "basic", "timezone", "Asia/Shanghai"),
		},
		Security: SecuritySettings{
			SessionTimeout: getIntOrDefault(m, "security", "session_timeout", 120),
			Enable2FA:      getBoolOrDefault(m, "security", "enable_2fa", false),
			PasswordPolicy: getOrDefault(m, "security", "password_policy", "medium"),
			AuditRetention: getIntOrDefault(m, "security", "audit_retention", 30),
		},
		Alert: AlertSettings{
			CPUThreshold:  getIntOrDefault(m, "alert", "cpu_threshold", 80),
			MemThreshold:  getIntOrDefault(m, "alert", "mem_threshold", 80),
			DiskThreshold: getIntOrDefault(m, "alert", "disk_threshold", 85),
			AlertSilence:  getIntOrDefault(m, "alert", "alert_silence", 15),
		},
		Notification: NotificationSettings{
			EnableEmail:     getBoolOrDefault(m, "notification", "enable_email", false),
			SMTPServer:      getOrDefault(m, "notification", "smtp_server", ""),
			EnableDingTalk:  getBoolOrDefault(m, "notification", "enable_dingtalk", false),
			DingTalkWebhook: getOrDefault(m, "notification", "dingtalk_webhook", ""),
			EnableWebhook:   getBoolOrDefault(m, "notification", "enable_webhook", false),
			WebhookURL:      getOrDefault(m, "notification", "webhook_url", ""),
		},
		About: AboutSettings{
			Version:    "2.0.0",
			BuildDate:  "2026-03-04",
			GoVersion:  "1.21",
			VueVersion: "3.5.13",
			DBType:     "MySQL 8.0",
			K8sSupport: "v1.25+",
		},
	}
	return resp
}

func getOrDefault(m map[string]map[string]string, category, key, defaultVal string) string {
	if cat, ok := m[category]; ok {
		if val, ok := cat[key]; ok && val != "" {
			return val
		}
	}
	return defaultVal
}

func getIntOrDefault(m map[string]map[string]string, category, key string, defaultVal int) int {
	if cat, ok := m[category]; ok {
		if val, ok := cat[key]; ok && val != "" {
			var result int
			json.Unmarshal([]byte(val), &result)
			if result != 0 {
				return result
			}
		}
	}
	return defaultVal
}

func getBoolOrDefault(m map[string]map[string]string, category, key string, defaultVal bool) bool {
	if cat, ok := m[category]; ok {
		if val, ok := cat[key]; ok {
			return val == "true" || val == "1"
		}
	}
	return defaultVal
}
