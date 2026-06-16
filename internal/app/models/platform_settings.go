package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// PlatformSettings 平台系统设置表
// 使用 key-value 模式存储，支持动态扩展
type PlatformSettings struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	Category   string `gorm:"size:50;not null;index:idx_category_key" json:"category"`    // 分类: basic, security, alert, notification
	Key        string `gorm:"size:100;not null;index:idx_category_key" json:"key"`        // 设置键
	Value      string `gorm:"type:text" json:"value"`                                     // 设置值
	ValueType  string `gorm:"size:20;default:'string'" json:"value_type"`                 // 值类型: string, int, bool, json
	Label      string `gorm:"size:100" json:"label"`                                      // 显示名称
	Desc       string `gorm:"size:500" json:"desc"`                                       // 描述
	CreatedAt  uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (p *PlatformSettings) TableName() string {
	return "platform_settings"
}

// ========== CRUD 方法 ==========

// GetByKey 根据分类和键获取设置
func (p *PlatformSettings) GetByKey(db *gorm.DB, category, key string) (*PlatformSettings, error) {
	var setting PlatformSettings
	err := db.Where("category = ? AND `key` = ?", category, key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetByCategory 获取某分类下所有设置
func (p *PlatformSettings) GetByCategory(db *gorm.DB, category string) ([]*PlatformSettings, error) {
	var settings []*PlatformSettings
	err := db.Where("category = ?", category).Find(&settings).Error
	return settings, err
}

// GetAll 获取所有设置
func (p *PlatformSettings) GetAll(db *gorm.DB) ([]*PlatformSettings, error) {
	var settings []*PlatformSettings
	err := db.Find(&settings).Error
	return settings, err
}

// Upsert 更新或插入设置（基于 category + key 唯一约束）
func (p *PlatformSettings) Upsert(db *gorm.DB) error {
	// 先查询是否存在
	var existing PlatformSettings
	err := db.Where("category = ? AND `key` = ?", p.Category, p.Key).First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// 不存在，执行插入
		return db.Create(p).Error
	} else if err != nil {
		return err
	}
	
	// 存在，执行更新
	return db.Model(&existing).Updates(map[string]interface{}{
		"value":       p.Value,
		"value_type":  p.ValueType,
		"label":       p.Label,
		"desc":        p.Desc,
		"modified_at": p.ModifiedAt,
	}).Error
}

// BatchUpsert 批量更新或插入
func BatchUpsertSettings(db *gorm.DB, settings []*PlatformSettings) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			if err := s.Upsert(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// ========== 聚合响应结构 ==========

// PlatformSettingsResponse 前端使用的聚合响应
type PlatformSettingsResponse struct {
	Basic        BasicSettings        `json:"basic"`
	Security     SecuritySettings     `json:"security"`
	Alert        AlertSettings        `json:"alert"`
	Notification NotificationSettings `json:"notification"`
	About        AboutSettings        `json:"about"`
}

type BasicSettings struct {
	DefaultPage    string `json:"default_page"`
	DefaultCluster string `json:"default_cluster"`
	Language       string `json:"language"`
	Timezone       string `json:"timezone"`
}

type SecuritySettings struct {
	SessionTimeout  int    `json:"session_timeout"`
	Enable2FA       bool   `json:"enable_2fa"`
	PasswordPolicy  string `json:"password_policy"`
	AuditRetention  int    `json:"audit_retention"`
}

type AlertSettings struct {
	CPUThreshold  int `json:"cpu_threshold"`
	MemThreshold  int `json:"mem_threshold"`
	DiskThreshold int `json:"disk_threshold"`
	AlertSilence  int `json:"alert_silence"`
}

type NotificationSettings struct {
	EnableEmail     bool   `json:"enable_email"`
	SMTPServer      string `json:"smtp_server"`
	EnableDingTalk  bool   `json:"enable_dingtalk"`
	DingTalkWebhook string `json:"dingtalk_webhook"`
	EnableWebhook   bool   `json:"enable_webhook"`
	WebhookURL      string `json:"webhook_url"`
}

type AboutSettings struct {
	Version     string `json:"version"`
	BuildDate   string `json:"build_date"`
	GoVersion   string `json:"go_version"`
	VueVersion  string `json:"vue_version"`
	DBType      string `json:"db_type"`
	K8sSupport  string `json:"k8s_support"`
}

// ToMap 将设置列表转换为 map
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

// ToResponse 将设置列表转换为响应结构
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
