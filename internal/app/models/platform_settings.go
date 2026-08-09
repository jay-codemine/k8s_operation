package models

import dm "k8soperation/internal/domain/settings"

// ===== 类型别名 =====

type (
	PlatformSettings         = dm.PlatformSettings
	PlatformSettingsResponse = dm.PlatformSettingsResponse
	BasicSettings            = dm.BasicSettings
	SecuritySettings         = dm.SecuritySettings
	AlertSettings            = dm.AlertSettings
	NotificationSettings     = dm.NotificationSettings
	AboutSettings            = dm.AboutSettings
)

// ===== 函数别名 =====

var (
	SettingsToMap      = dm.SettingsToMap
	SettingsToResponse = dm.SettingsToResponse
)
