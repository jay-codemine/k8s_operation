package ai

// =========================================================================
// AIOps 智能运维模型
// =========================================================================

// AIOpsAnalysisRecord AI 分析记录
type AIOpsAnalysisRecord struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string `gorm:"type:varchar(30);not null;index" json:"type"`
	RefID       int64  `gorm:"type:bigint;index" json:"ref_id"`
	Title       string `gorm:"type:varchar(300);not null" json:"title"`
	Input       string `gorm:"type:text" json:"input"`
	Result      string `gorm:"type:longtext" json:"result"`
	Severity    string `gorm:"type:varchar(20)" json:"severity"`
	Suggestions string `gorm:"type:text" json:"suggestions"`
	Model       string `gorm:"type:varchar(100)" json:"model"`
	TokensUsed  int    `gorm:"type:int;default:0" json:"tokens_used"`
	LatencyMs   int64  `gorm:"type:bigint;default:0" json:"latency_ms"`
	Status      string `gorm:"type:varchar(20);default:'success'" json:"status"`
	Error       string `gorm:"type:varchar(500)" json:"error,omitempty"`
	UserID      int64  `gorm:"type:bigint;index" json:"user_id"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
}

func (AIOpsAnalysisRecord) TableName() string { return "aiops_analysis_record" }

// AIOpsInspectionReport 巡检报告
type AIOpsInspectionReport struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string `gorm:"type:varchar(30);not null;index" json:"type"`
	Scope       string `gorm:"type:varchar(30);default:'full'" json:"scope"`
	ScopeID     string `gorm:"type:varchar(100)" json:"scope_id"`
	HealthScore int    `gorm:"type:int;default:0" json:"health_score"`
	Level       string `gorm:"type:varchar(20);not null" json:"level"`
	Summary     string `gorm:"type:text" json:"summary"`
	Details     string `gorm:"type:longtext" json:"details"`
	AIAnalysis  string `gorm:"type:longtext" json:"ai_analysis"`
	Findings    int    `gorm:"type:int;default:0" json:"findings"`
	Suggestions int    `gorm:"type:int;default:0" json:"suggestions_count"`
	Duration    int64  `gorm:"type:bigint;default:0" json:"duration"`
	Status      string `gorm:"type:varchar(20);default:'running'" json:"status"`
	Error       string `gorm:"type:varchar(500)" json:"error,omitempty"`
	TriggeredBy int64  `gorm:"type:bigint" json:"triggered_by"`
	CreatedAt   int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	CompletedAt int64  `gorm:"type:bigint" json:"completed_at"`
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

// =========================================================================
// AI 会话
// =========================================================================

// AIConversation AI 对话会话
type AIConversation struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint32 `gorm:"not null;index" json:"user_id"`
	Title      string `gorm:"size:200;default:'新对话'" json:"title"`
	Status     uint8  `gorm:"default:1" json:"status"`
	CreatedAt  uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (c *AIConversation) TableName() string { return "ai_conversations" }

// AIMessage AI 聊天消息
type AIMessage struct {
	ID             uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uint32 `gorm:"not null;index" json:"conversation_id"`
	Role           string `gorm:"size:20;not null" json:"role"`
	Content        string `gorm:"type:text" json:"content"`
	IntentJSON     string `gorm:"type:text" json:"intent_json"`
	TokenUsed      int    `gorm:"default:0" json:"token_used"`
	CreatedAt      uint32 `gorm:"autoCreateTime" json:"created_at"`
}

func (m *AIMessage) TableName() string { return "ai_messages" }

// AI 审批状态常量
const (
	AIApprovalPending  uint8 = 1
	AIApprovalApproved uint8 = 2
	AIApprovalRejected uint8 = 3
	AIApprovalExpired  uint8 = 4
	AIApprovalCanceled uint8 = 5
)

// AI 风险等级常量
const (
	AIRiskLow      = "low"
	AIRiskMedium   = "medium"
	AIRiskHigh     = "high"
	AIRiskCritical = "critical"
)

// AIApprovalRequest 高危操作审批请求
type AIApprovalRequest struct {
	ID             uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uint32 `gorm:"index" json:"conversation_id"`
	RequestUserID  uint32 `gorm:"not null;index" json:"request_user_id"`
	ApproverUserID uint32 `gorm:"default:0" json:"approver_user_id"`
	Intent         string `gorm:"size:50;not null" json:"intent"`
	Resource       string `gorm:"size:100" json:"resource"`
	ResourceName   string `gorm:"size:200" json:"resource_name"`
	Namespace      string `gorm:"size:100" json:"namespace"`
	ClusterID      uint32 `gorm:"default:0" json:"cluster_id"`
	RiskLevel      string `gorm:"size:20;not null" json:"risk_level"`
	OperationJSON  string `gorm:"type:text" json:"operation_json"`
	ToolName       string `gorm:"size:100" json:"tool_name"`
	ToolArgsJSON   string `gorm:"type:text" json:"tool_args_json"`
	ToolCallID     string `gorm:"size:100" json:"tool_call_id"`
	ExecuteResult  string `gorm:"type:text" json:"execute_result"`
	Executed       bool   `gorm:"default:false" json:"executed"`
	Summary        string `gorm:"size:500" json:"summary"`
	Status         uint8  `gorm:"default:1;index" json:"status"`
	ApproveComment string `gorm:"size:500" json:"approve_comment"`
	ExpireAt       uint32 `gorm:"default:0" json:"expire_at"`
	CreatedAt      uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt     uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (a *AIApprovalRequest) TableName() string { return "ai_approval_requests" }

// AIApprovalLog 审批操作日志
type AIApprovalLog struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ApprovalID uint32 `gorm:"not null;index" json:"approval_id"`
	UserID     uint32 `gorm:"not null" json:"user_id"`
	Action     string `gorm:"size:50;not null" json:"action"`
	Comment    string `gorm:"size:500" json:"comment"`
	CreatedAt  uint32 `gorm:"autoCreateTime" json:"created_at"`
}

func (l *AIApprovalLog) TableName() string { return "ai_approval_logs" }
