package models

// =========================================================================
// AIOps 智能运维模型
// 支持: AI 告警分析、AI 日志诊断、智能巡检
// =========================================================================

// AIOpsAnalysisRecord AI 分析记录
// 记录每一次 AI 分析的请求和结果（告警分析 / 日志诊断 / 巡检报告）
type AIOpsAnalysisRecord struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string `gorm:"type:varchar(30);not null;index" json:"type"`         // 类型: alert_analysis/log_diagnosis/inspection
	RefID       int64  `gorm:"type:bigint;index" json:"ref_id"`                     // 关联 ID（告警事件 ID / 巡检报告 ID）
	Title       string `gorm:"type:varchar(300);not null" json:"title"`             // 分析标题
	Input       string `gorm:"type:text" json:"input"`                              // 输入数据（告警详情/日志片段/巡检摘要）
	Result      string `gorm:"type:longtext" json:"result"`                         // AI 分析结果（Markdown）
	Severity    string `gorm:"type:varchar(20)" json:"severity"`                    // AI 判定严重级别
	Suggestions string `gorm:"type:text" json:"suggestions"`                        // AI 建议 JSON
	Model       string `gorm:"type:varchar(100)" json:"model"`                      // 使用的 AI 模型
	TokensUsed  int    `gorm:"type:int;default:0" json:"tokens_used"`               // 消耗 Token 数
	LatencyMs   int64  `gorm:"type:bigint;default:0" json:"latency_ms"`             // 分析耗时(ms)
	Status      string `gorm:"type:varchar(20);default:'success'" json:"status"`    // 状态: success/failed/timeout
	Error       string `gorm:"type:varchar(500)" json:"error,omitempty"`            // 错误信息
	UserID      int64  `gorm:"type:bigint;index" json:"user_id"`                    // 发起人
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
}

func (AIOpsAnalysisRecord) TableName() string { return "aiops_analysis_record" }

// AIOpsInspectionReport 巡检报告
type AIOpsInspectionReport struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type          string `gorm:"type:varchar(30);not null;index" json:"type"`       // 巡检类型: scheduled/manual
	Scope         string `gorm:"type:varchar(30);default:'full'" json:"scope"`      // 巡检范围: full/cluster/namespace
	ScopeID       string `gorm:"type:varchar(100)" json:"scope_id"`                 // 范围 ID（集群 ID 等）
	HealthScore   int    `gorm:"type:int;default:0" json:"health_score"`            // 健康评分 0-100
	Level         string `gorm:"type:varchar(20);not null" json:"level"`            // 健康等级: healthy/warning/critical
	Summary       string `gorm:"type:text" json:"summary"`                          // 巡检摘要（AI 生成）
	Details       string `gorm:"type:longtext" json:"details"`                      // 巡检详情 JSON
	AIAnalysis    string `gorm:"type:longtext" json:"ai_analysis"`                  // AI 综合分析（Markdown）
	Findings      int    `gorm:"type:int;default:0" json:"findings"`                // 发现问题数
	Suggestions   int    `gorm:"type:int;default:0" json:"suggestions_count"`       // 建议数
	Duration      int64  `gorm:"type:bigint;default:0" json:"duration"`             // 巡检耗时(ms)
	Status        string `gorm:"type:varchar(20);default:'running'" json:"status"`  // 状态: running/completed/failed
	Error         string `gorm:"type:varchar(500)" json:"error,omitempty"`          // 错误信息
	TriggeredBy   int64  `gorm:"type:bigint" json:"triggered_by"`                   // 触发人(0=系统定时)
	CreatedAt     int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	CompletedAt   int64  `gorm:"type:bigint" json:"completed_at"`                   // 完成时间
}

func (AIOpsInspectionReport) TableName() string { return "aiops_inspection_report" }

// =========================================================================
// AIOps 分析类型常量
// =========================================================================

const (
	AIOpsTypeAlertAnalysis = "alert_analysis"
	AIOpsTypeLogDiagnosis  = "log_diagnosis"
	AIOpsTypeInspection    = "inspection"
)

const (
	AIOpsStatusSuccess = "success"
	AIOpsStatusFailed  = "failed"
	AIOpsStatusTimeout = "timeout"
)

const (
	InspectionTypeScheduled = "scheduled"
	InspectionTypeManual    = "manual"
)

const (
	InspectionLevelHealthy  = "healthy"
	InspectionLevelWarning  = "warning"
	InspectionLevelCritical = "critical"
)
