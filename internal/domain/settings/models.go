package settings

// PlatformSettings 平台系统设置表
type PlatformSettings struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	Category   string `gorm:"size:50;not null;index:idx_category_key" json:"category"`
	Key        string `gorm:"size:100;not null;index:idx_category_key" json:"key"`
	Value      string `gorm:"type:text" json:"value"`
	ValueType  string `gorm:"size:20;default:'string'" json:"value_type"`
	Label      string `gorm:"size:100" json:"label"`
	Desc       string `gorm:"size:500" json:"desc"`
	CreatedAt  uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (p *PlatformSettings) TableName() string { return "platform_settings" }

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
