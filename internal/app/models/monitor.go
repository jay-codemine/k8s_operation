package models

// =========================================================================
// 监控数据源 (monitor_datasource)
// 支持: Prometheus, Loki, Alertmanager, Grafana 等
// =========================================================================

// MonitorDatasource 监控数据源
type MonitorDatasource struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`       // 所属租户
	Name           string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`              // 数据源名称
	Type           string `gorm:"type:varchar(30);not null;index" json:"type"`                     // 类型: prometheus/loki/alertmanager/grafana/victoriametrics
	URL            string `gorm:"type:varchar(500);not null" json:"url"`                           // 连接地址
	Description    string `gorm:"type:varchar(500)" json:"description"`                            // 描述
	ClusterID      int64  `gorm:"column:cluster_id;type:bigint;default:0;index" json:"cluster_id"` // 关联 K8s 集群 ID（0=全局/未关联）
	AccessMode     string `gorm:"type:varchar(20);default:'proxy'" json:"access_mode"`             // 访问模式: proxy/direct
	AuthType       string `gorm:"type:varchar(20);default:'none'" json:"auth_type"`                // 认证方式: none/basic/bearer/tls
	AuthUser       string `gorm:"type:varchar(100)" json:"auth_user"`                              // 认证用户名
	AuthPass       string `gorm:"type:varchar(500)" json:"-"`                                      // 认证密码/Token（不返回前端）
	TLSCert        string `gorm:"type:text" json:"-"`                                              // TLS 证书
	TLSKey         string `gorm:"type:text" json:"-"`                                              // TLS 密钥
	CACert         string `gorm:"type:text" json:"-"`                                              // CA 证书
	IsDefault      bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`                     // 是否默认数据源
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`                        // 是否启用
	Timeout        int    `gorm:"type:int;default:30" json:"timeout"`                              // 查询超时(秒)
	ScrapeInterval int    `gorm:"type:int;default:15" json:"scrape_interval"`                      // 采集间隔(秒)
	Status         string `gorm:"type:varchar(20);default:'unknown'" json:"status"`                // 连接状态: connected/disconnected/unknown
	LastCheckAt    int64  `gorm:"type:bigint" json:"last_check_at"`                                // 最后健康检查时间
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`                                   // 创建人
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorDatasource) TableName() string { return "monitor_datasource" }

// =========================================================================
// 告警规则 (monitor_alert_rule)
// 基于 PromQL 表达式定义告警条件
// =========================================================================

// MonitorAlertRule 告警规则
type MonitorAlertRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`       // 所属租户
	DatasourceID   int64  `gorm:"type:bigint;not null;index" json:"datasource_id"`  // 关联数据源
	Name           string `gorm:"type:varchar(200);not null" json:"name"`           // 规则名称
	Group          string `gorm:"type:varchar(100);default:'default'" json:"group"` // 规则分组
	Severity       string `gorm:"type:varchar(20);not null;index" json:"severity"`  // 严重级别: critical/warning/info
	Expr           string `gorm:"type:text;not null" json:"expr"`                   // PromQL 表达式
	Duration       string `gorm:"type:varchar(20);default:'5m'" json:"duration"`    // 持续时间 (for)
	Summary        string `gorm:"type:varchar(500)" json:"summary"`                 // 告警摘要模板
	Description    string `gorm:"type:text" json:"description"`                     // 告警描述模板
	Labels         string `gorm:"type:text" json:"labels"`                          // 额外标签 JSON
	Annotations    string `gorm:"type:text" json:"annotations"`                     // 额外注解 JSON
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`         // 是否启用
	NotifyChannels string `gorm:"type:varchar(500)" json:"notify_channels"`         // 通知渠道: webhook/email/dingtalk/feishu
	NotifyURL      string `gorm:"type:varchar(500)" json:"notify_url"`              // webhook URL
	EvalInterval   int    `gorm:"type:int;default:60" json:"eval_interval"`         // 评估间隔(秒)
	LastEvalAt     int64  `gorm:"type:bigint" json:"last_eval_at"`                  // 最后评估时间
	LastEvalResult string `gorm:"type:varchar(20)" json:"last_eval_result"`         // 最后评估结果: normal/firing/pending/error
	PendingSince   int64  `gorm:"type:bigint;default:0" json:"pending_since"`        // 告警条件首次满足时间
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorAlertRule) TableName() string { return "monitor_alert_rule" }

// =========================================================================
// 告警事件 (monitor_alert_event)
// 告警触发记录
// =========================================================================

// MonitorAlertEvent 告警事件
type MonitorAlertEvent struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id"` // 所属租户
	RuleID        int64  `gorm:"type:bigint;not null;index" json:"rule_id"`       // 关联告警规则
	DatasourceID  int64  `gorm:"type:bigint;not null;index" json:"datasource_id"` // 关联数据源
	RuleName      string `gorm:"type:varchar(200)" json:"rule_name"`              // 冗余规则名（快速查询）
	Severity      string `gorm:"type:varchar(20);not null;index" json:"severity"` // 严重级别
	Status        string `gorm:"type:varchar(20);not null;index" json:"status"`   // 状态: firing/resolved/silenced
	Value         string `gorm:"type:varchar(100)" json:"value"`                  // 触发时的值
	Labels        string `gorm:"type:text" json:"labels"`                         // 标签 JSON
	Annotations   string `gorm:"type:text" json:"annotations"`                    // 注解 JSON
	Summary       string `gorm:"type:varchar(500)" json:"summary"`                // 告警摘要
	Description   string `gorm:"type:text" json:"description"`                    // 告警描述
	FiredAt       int64  `gorm:"type:bigint;not null;index" json:"fired_at"`      // 触发时间
	ResolvedAt    int64  `gorm:"type:bigint" json:"resolved_at"`                  // 恢复时间
	AckedBy       int64  `gorm:"type:bigint" json:"acked_by"`                     // 确认人
	AckedAt       int64  `gorm:"type:bigint" json:"acked_at"`                     // 确认时间
	SilencedUntil int64  `gorm:"type:bigint" json:"silenced_until"`               // 静默截止
	NotifyResult  string `gorm:"type:varchar(200)" json:"notify_result"`          // 通知结果
	CreatedAt     int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
}

func (MonitorAlertEvent) TableName() string { return "monitor_alert_event" }

// =========================================================================
// 通知渠道 (monitor_notify_channel)
// 支持: 钉钉/飞书/Webhook/邮件/企业微信
// =========================================================================

// MonitorNotifyChannel 通知渠道
type MonitorNotifyChannel struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"` // 渠道名称
	Type            string `gorm:"type:varchar(30);not null;index" json:"type"`        // 类型: dingtalk/feishu/webhook/email/wechat
	Description     string `gorm:"type:varchar(500)" json:"description"`               // 描述
	WebhookURL      string `gorm:"type:varchar(500)" json:"webhook_url"`               // Webhook 地址
	Secret          string `gorm:"type:varchar(500)" json:"-"`                         // 签名密钥（钉钉加签等）
	SecurityKeyword string `gorm:"type:varchar(100)" json:"security_keyword"`          // 钉钉安全关键字
	AtMobiles       string `gorm:"type:varchar(500)" json:"at_mobiles"`                // @手机号列表(逗号分隔)
	AtAll           bool   `gorm:"type:tinyint(1);default:0" json:"at_all"`            // 是否@所有人
	SMTPHost        string `gorm:"type:varchar(200)" json:"smtp_host"`                 // SMTP 主机(邮件)
	SMTPPort        int    `gorm:"type:int;default:465" json:"smtp_port"`              // SMTP 端口
	SMTPUser        string `gorm:"type:varchar(200)" json:"smtp_user"`                 // SMTP 用户
	SMTPPass        string `gorm:"type:varchar(200)" json:"-"`                         // SMTP 密码
	SMTPTo          string `gorm:"type:text" json:"smtp_to"`                           // 收件人列表(逗号分隔)
	MsgTemplate     string `gorm:"type:text" json:"msg_template"`                      // 消息模板(支持变量)
	Enabled         bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`           // 是否启用
	SendResolved    bool   `gorm:"type:tinyint(1);default:1" json:"send_resolved"`     // 是否发送恢复通知
	RateLimit       int    `gorm:"type:int;default:10" json:"rate_limit"`              // 限流：每分钟最大通知数
	CreatedBy       int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt       int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt      int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel           int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorNotifyChannel) TableName() string { return "monitor_notify_channel" }

// =========================================================================
// 告警静默规则 (monitor_silence_rule)
// 按标签匹配 + 时间窗口抑制告警通知
// =========================================================================

// MonitorSilenceRule 告警静默规则
type MonitorSilenceRule struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string `gorm:"type:varchar(200);not null" json:"name"`             // 规则名称
	Type       string `gorm:"type:varchar(30);not null;index" json:"type"`        // 类型: silence(静默)/inhibit(抑制)/aggregate(聚合)
	Matchers   string `gorm:"type:text;not null" json:"matchers"`                 // 匹配条件 JSON [{label,op,value}]
	StartsAt   int64  `gorm:"type:bigint" json:"starts_at"`                       // 生效开始时间(0=立即)
	EndsAt     int64  `gorm:"type:bigint" json:"ends_at"`                         // 生效结束时间(0=永久)
	Duration   string `gorm:"type:varchar(30)" json:"duration"`                   // 持续时间(如 2h/1d,与ends_at二选一)
	RepeatType string `gorm:"type:varchar(20);default:'once'" json:"repeat_type"` // 重复类型: once/daily/weekly/cron
	RepeatCron string `gorm:"type:varchar(100)" json:"repeat_cron"`               // Cron 表达式(repeat_type=cron时)
	Comment    string `gorm:"type:varchar(500)" json:"comment"`                   // 备注
	Enabled    bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`           // 是否启用
	CreatedBy  int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt  int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel      int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorSilenceRule) TableName() string { return "monitor_silence_rule" }

// =========================================================================
// 告警抑制规则 (monitor_inhibit_rule)
// 当高优先级告警触发时，抑制低优先级同类告警
// =========================================================================

// MonitorInhibitRule 告警抑制规则
type MonitorInhibitRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string `gorm:"type:varchar(200);not null" json:"name"`    // 规则名称
	SourceMatchers string `gorm:"type:text;not null" json:"source_matchers"` // 源告警匹配 JSON (触发抑制的高优告警)
	TargetMatchers string `gorm:"type:text;not null" json:"target_matchers"` // 目标告警匹配 JSON (被抑制的低优告警)
	EqualLabels    string `gorm:"type:varchar(500)" json:"equal_labels"`     // 关联标签(逗号分隔,需一致才抑制)
	Description    string `gorm:"type:varchar(500)" json:"description"`      // 描述
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`  // 是否启用
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorInhibitRule) TableName() string { return "monitor_inhibit_rule" }

// =========================================================================
// 告警聚合规则 (monitor_aggregate_rule)
// 将同类告警合并为一条通知，减少告警风暴
// =========================================================================

// MonitorAggregateRule 告警聚合规则
type MonitorAggregateRule struct {
	ID             int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string `gorm:"type:varchar(200);not null" json:"name"`               // 规则名称
	GroupBy        string `gorm:"type:varchar(500);not null" json:"group_by"`           // 聚合维度(逗号分隔标签名)
	GroupWait      string `gorm:"type:varchar(20);default:'30s'" json:"group_wait"`     // 分组等待时间(首次通知延迟)
	GroupInterval  string `gorm:"type:varchar(20);default:'5m'" json:"group_interval"`  // 分组间隔(同组告警批量通知)
	RepeatInterval string `gorm:"type:varchar(20);default:'4h'" json:"repeat_interval"` // 重复间隔(相同告警多久再通知)
	Matchers       string `gorm:"type:text" json:"matchers"`                            // 匹配条件 JSON
	ChannelIDs     string `gorm:"type:varchar(200)" json:"channel_ids"`                 // 通知渠道ID列表(逗号分隔)
	Enabled        bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	CreatedBy      int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt      int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt     int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel          int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorAggregateRule) TableName() string { return "monitor_aggregate_rule" }

// =========================================================================
// 通知模板 (monitor_notify_template)
// 支持自定义告警通知消息格式，按渠道类型分别配置
// =========================================================================

// MonitorNotifyTemplate 通知模板
type MonitorNotifyTemplate struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`     // 模板名称
	Type        string `gorm:"type:varchar(30);not null;index" json:"type"`            // 渠道类型: dingtalk/feishu/wechat/email/webhook
	Scene       string `gorm:"type:varchar(30);not null;default:'alert'" json:"scene"` // 场景: alert(告警)/resolved(恢复)/test(测试)
	Title       string `gorm:"type:varchar(200)" json:"title"`                         // 模板标题
	Content     string `gorm:"type:text;not null" json:"content"`                      // 模板内容(Markdown/富文本)
	Description string `gorm:"type:varchar(500)" json:"description"`                   // 描述
	IsDefault   bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`            // 是否默认模板
	Enabled     bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`               // 是否启用
	CreatedBy   int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt  int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel       int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorNotifyTemplate) TableName() string { return "monitor_notify_template" }

// =========================================================================
// 通知路由策略 (monitor_notify_route_policy)
// 大厂级告警分发设计：按 severity/group/labels 自动路由到不同渠道
// 支持多策略优先级匹配，解决 YAML 批量导入后需逐条绑定渠道的痛点
// =========================================================================

// MonitorNotifyRoutePolicy 通知路由策略
type MonitorNotifyRoutePolicy struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(200);not null;uniqueIndex" json:"name"`    // 策略名称
	Description string `gorm:"type:varchar(500)" json:"description"`                  // 描述
	Priority    int    `gorm:"type:int;default:100;index" json:"priority"`            // 优先级(越小越优先,0为最高)
	ChannelIDs  string `gorm:"type:varchar(500);not null" json:"channel_ids"`         // 目标通知渠道ID(逗号分隔)
	MatchMode   string `gorm:"type:varchar(20);default:'any'" json:"match_mode"`      // 匹配模式: all(全部满足)/any(任一满足)
	Severities  string `gorm:"type:varchar(100)" json:"severities"`                   // 匹配级别(逗号分隔): critical,warning,info
	Groups      string `gorm:"type:varchar(500)" json:"groups"`                       // 匹配规则分组(逗号分隔)
	LabelMatch  string `gorm:"type:text" json:"label_match"`                          // 标签匹配条件 JSON [{key,op,value}]
	IsDefault   bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`           // 是否为兜底默认策略
	Enabled     bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`              // 是否启用
	CreatedBy   int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt  int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel       int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (MonitorNotifyRoutePolicy) TableName() string { return "monitor_notify_route_policy" }
