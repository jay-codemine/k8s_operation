package monitor

// =========================================================================
// 监控数据源 (monitor_datasource)
// 支持: Prometheus, Loki, Alertmanager, Grafana 等
// =========================================================================

// Datasource 监控数据源
type Datasource struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	Name           string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Type           string `gorm:"type:varchar(30);not null;index" json:"type"`
	URL            string `gorm:"type:varchar(500);not null" json:"url"`
	Description    string `gorm:"type:varchar(500)" json:"description"`
	ClusterID      int64  `gorm:"column:cluster_id;type:bigint;default:0;index" json:"cluster_id"`
	AccessMode     string `gorm:"type:varchar(20);default:'proxy'" json:"access_mode"`
	AuthType       string `gorm:"type:varchar(20);default:'none'" json:"auth_type"`
	AuthUser       string `gorm:"type:varchar(100)" json:"auth_user"`
	AuthPass       string `gorm:"type:varchar(500)" json:"-"`
	TLSCert        string `gorm:"type:text" json:"-"`
	TLSKey         string `gorm:"type:text" json:"-"`
	CACert         string `gorm:"type:text" json:"-"`
	IsDefault      bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	Timeout        int    `gorm:"type:int;default:30" json:"timeout"`
	ScrapeInterval int    `gorm:"type:int;default:15" json:"scrape_interval"`
	Status         string `gorm:"type:varchar(20);default:'unknown'" json:"status"`
	LastCheckAt    int64  `gorm:"type:bigint" json:"last_check_at"`
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (Datasource) TableName() string { return "monitor_datasource" }

// =========================================================================
// 告警规则 (monitor_alert_rule)
// =========================================================================

// AlertRule 告警规则
type AlertRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	DatasourceID   int64  `gorm:"type:bigint;not null;index" json:"datasource_id"`
	Name           string `gorm:"type:varchar(200);not null" json:"name"`
	Group          string `gorm:"type:varchar(100);default:'default'" json:"group"`
	Severity       string `gorm:"type:varchar(20);not null;index" json:"severity"`
	Expr           string `gorm:"type:text;not null" json:"expr"`
	Duration       string `gorm:"type:varchar(20);default:'5m'" json:"duration"`
	Summary        string `gorm:"type:varchar(500)" json:"summary"`
	Description    string `gorm:"type:text" json:"description"`
	Labels         string `gorm:"type:text" json:"labels"`
	Annotations    string `gorm:"type:text" json:"annotations"`
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	NotifyChannels string `gorm:"type:varchar(500)" json:"notify_channels"`
	NotifyURL      string `gorm:"type:varchar(500)" json:"notify_url"`
	EvalInterval   int    `gorm:"type:int;default:60" json:"eval_interval"`
	LastEvalAt     int64  `gorm:"type:bigint" json:"last_eval_at"`
	LastEvalResult string `gorm:"type:varchar(20)" json:"last_eval_result"`
	PendingSince   int64  `gorm:"type:bigint;default:0" json:"pending_since"`
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (AlertRule) TableName() string { return "monitor_alert_rule" }

// =========================================================================
// 告警事件 (monitor_alert_event)
// =========================================================================

// AlertEvent 告警事件
type AlertEvent struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
	RuleID        int64  `gorm:"type:bigint;not null;index" json:"rule_id"`
	DatasourceID  int64  `gorm:"type:bigint;not null;index" json:"datasource_id"`
	RuleName      string `gorm:"type:varchar(200)" json:"rule_name"`
	Severity      string `gorm:"type:varchar(20);not null;index" json:"severity"`
	Status        string `gorm:"type:varchar(20);not null;index" json:"status"`
	Value         string `gorm:"type:varchar(100)" json:"value"`
	Labels        string `gorm:"type:text" json:"labels"`
	Annotations   string `gorm:"type:text" json:"annotations"`
	Summary       string `gorm:"type:varchar(500)" json:"summary"`
	Description   string `gorm:"type:text" json:"description"`
	FiredAt       int64  `gorm:"type:bigint;not null;index" json:"fired_at"`
	ResolvedAt    int64  `gorm:"type:bigint" json:"resolved_at"`
	AckedBy       int64  `gorm:"type:bigint" json:"acked_by"`
	AckedAt       int64  `gorm:"type:bigint" json:"acked_at"`
	SilencedUntil int64  `gorm:"type:bigint" json:"silenced_until"`
	NotifyResult  string `gorm:"type:varchar(200)" json:"notify_result"`
	CreatedAt     int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
}

func (AlertEvent) TableName() string { return "monitor_alert_event" }

// =========================================================================
// 通知渠道 (monitor_notify_channel)
// =========================================================================

// NotifyChannel 通知渠道
type NotifyChannel struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Type            string `gorm:"type:varchar(30);not null;index" json:"type"`
	Description     string `gorm:"type:varchar(500)" json:"description"`
	WebhookURL      string `gorm:"type:varchar(500)" json:"webhook_url"`
	Secret          string `gorm:"type:varchar(500)" json:"-"`
	SecurityKeyword string `gorm:"type:varchar(100)" json:"security_keyword"`
	AtMobiles       string `gorm:"type:varchar(500)" json:"at_mobiles"`
	AtAll           bool   `gorm:"type:tinyint(1);default:0" json:"at_all"`
	SMTPHost        string `gorm:"type:varchar(200)" json:"smtp_host"`
	SMTPPort        int    `gorm:"type:int;default:465" json:"smtp_port"`
	SMTPUser        string `gorm:"type:varchar(200)" json:"smtp_user"`
	SMTPPass        string `gorm:"type:varchar(200)" json:"-"`
	SMTPTo          string `gorm:"type:text" json:"smtp_to"`
	MsgTemplate     string `gorm:"type:text" json:"msg_template"`
	Enabled         bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	SendResolved    bool   `gorm:"type:tinyint(1);default:1" json:"send_resolved"`
	RateLimit       int    `gorm:"type:int;default:10" json:"rate_limit"`
	CreatedBy       int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt       int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt      int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel           int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (NotifyChannel) TableName() string { return "monitor_notify_channel" }

// =========================================================================
// 告警静默规则 (monitor_silence_rule)
// =========================================================================

// SilenceRule 告警静默规则
type SilenceRule struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string `gorm:"type:varchar(200);not null" json:"name"`
	Type       string `gorm:"type:varchar(30);not null;index" json:"type"`
	Matchers   string `gorm:"type:text;not null" json:"matchers"`
	StartsAt   int64  `gorm:"type:bigint" json:"starts_at"`
	EndsAt     int64  `gorm:"type:bigint" json:"ends_at"`
	Duration   string `gorm:"type:varchar(30)" json:"duration"`
	RepeatType string `gorm:"type:varchar(20);default:'once'" json:"repeat_type"`
	RepeatCron string `gorm:"type:varchar(100)" json:"repeat_cron"`
	Comment    string `gorm:"type:varchar(500)" json:"comment"`
	Enabled    bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy  int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt  int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel      int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (SilenceRule) TableName() string { return "monitor_silence_rule" }

// =========================================================================
// 告警抑制规则 (monitor_inhibit_rule)
// =========================================================================

// InhibitRule 告警抑制规则
type InhibitRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string `gorm:"type:varchar(200);not null" json:"name"`
	SourceMatchers string `gorm:"type:text;not null" json:"source_matchers"`
	TargetMatchers string `gorm:"type:text;not null" json:"target_matchers"`
	EqualLabels    string `gorm:"type:varchar(500)" json:"equal_labels"`
	Description    string `gorm:"type:varchar(500)" json:"description"`
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (InhibitRule) TableName() string { return "monitor_inhibit_rule" }

// =========================================================================
// 告警聚合规则 (monitor_aggregate_rule)
// =========================================================================

// AggregateRule 告警聚合规则
type AggregateRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string `gorm:"type:varchar(200);not null" json:"name"`
	GroupBy        string `gorm:"type:varchar(500);not null" json:"group_by"`
	GroupWait      string `gorm:"type:varchar(20);default:'30s'" json:"group_wait"`
	GroupInterval  string `gorm:"type:varchar(20);default:'5m'" json:"group_interval"`
	RepeatInterval string `gorm:"type:varchar(20);default:'4h'" json:"repeat_interval"`
	Matchers       string `gorm:"type:text" json:"matchers"`
	ChannelIDs     string `gorm:"type:varchar(200)" json:"channel_ids"`
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (AggregateRule) TableName() string { return "monitor_aggregate_rule" }

// =========================================================================
// 通知模板 (monitor_notify_template)
// =========================================================================

// NotifyTemplate 通知模板
type NotifyTemplate struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Type        string `gorm:"type:varchar(30);not null;index" json:"type"`
	Scene       string `gorm:"type:varchar(30);not null;default:'alert'" json:"scene"`
	Title       string `gorm:"type:varchar(200)" json:"title"`
	Content     string `gorm:"type:text;not null" json:"content"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	IsDefault   bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`
	Enabled     bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy   int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt  int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel       int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (NotifyTemplate) TableName() string { return "monitor_notify_template" }

// =========================================================================
// 通知路由策略 (monitor_notify_route_policy)
// =========================================================================

// NotifyRoutePolicy 通知路由策略
type NotifyRoutePolicy struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(200);not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	Priority    int    `gorm:"type:int;default:100;index" json:"priority"`
	ChannelIDs  string `gorm:"type:varchar(500);not null" json:"channel_ids"`
	MatchMode   string `gorm:"type:varchar(20);default:'any'" json:"match_mode"`
	Severities  string `gorm:"type:varchar(100)" json:"severities"`
	Groups      string `gorm:"type:varchar(500)" json:"groups"`
	LabelMatch  string `gorm:"type:text" json:"label_match"`
	IsDefault   bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`
	Enabled     bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy   int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt  int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel       int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (NotifyRoutePolicy) TableName() string { return "monitor_notify_route_policy" }
