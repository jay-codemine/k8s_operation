package models

import (
	"gorm.io/gorm"
)

// =========================================================================
// AI 会话 (ai_conversations)
// =========================================================================

// AIConversation AI 对话会话
type AIConversation struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint32 `gorm:"not null;index" json:"user_id"`           // 关联用户
	Title     string `gorm:"size:200;default:'新对话'" json:"title"`     // 会话标题
	Status    uint8  `gorm:"default:1" json:"status"`                 // 1=活跃 2=归档
	CreatedAt uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (c *AIConversation) TableName() string { return "ai_conversations" }

// =========================================================================
// AI 聊天消息 (ai_messages)
// =========================================================================

// AIMessage AI 聊天消息
type AIMessage struct {
	ID             uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uint32 `gorm:"not null;index" json:"conversation_id"` // 关联会话
	Role           string `gorm:"size:20;not null" json:"role"`          // system/user/assistant
	Content        string `gorm:"type:text" json:"content"`             // 消息内容
	IntentJSON     string `gorm:"type:text" json:"intent_json"`         // 意图识别结果 JSON（assistant 消息时填写）
	TokenUsed      int    `gorm:"default:0" json:"token_used"`          // Token 消耗
	CreatedAt      uint32 `gorm:"autoCreateTime" json:"created_at"`
}

func (m *AIMessage) TableName() string { return "ai_messages" }

// =========================================================================
// 高危操作审批 (ai_approval_requests)
// =========================================================================

// AI 审批状态常量（uint8，与 CICD 的 string 类型审批状态独立）
const (
	AIApprovalPending  uint8 = 1 // 待审批
	AIApprovalApproved uint8 = 2 // 已通过
	AIApprovalRejected uint8 = 3 // 已拒绝
	AIApprovalExpired  uint8 = 4 // 已过期
	AIApprovalCanceled uint8 = 5 // 已取消
)

// AI 风险等级常量
const (
	AIRiskLow      string = "low"
	AIRiskMedium   string = "medium"
	AIRiskHigh     string = "high"
	AIRiskCritical string = "critical"
)

// AIApprovalRequest 高危操作审批请求
type AIApprovalRequest struct {
	ID             uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uint32 `gorm:"index" json:"conversation_id"`              // 关联 AI 会话
	RequestUserID  uint32 `gorm:"not null;index" json:"request_user_id"`     // 发起人
	ApproverUserID uint32 `gorm:"default:0" json:"approver_user_id"`        // 审批人
	Intent         string `gorm:"size:50;not null" json:"intent"`            // 操作意图: delete/drain/scale 等
	Resource       string `gorm:"size:100" json:"resource"`                  // 资源类型: deployment/namespace/node
	ResourceName   string `gorm:"size:200" json:"resource_name"`             // 资源名称
	Namespace      string `gorm:"size:100" json:"namespace"`                 // 命名空间
	ClusterID      uint32 `gorm:"default:0" json:"cluster_id"`              // 目标集群
	RiskLevel      string `gorm:"size:20;not null" json:"risk_level"`        // 风险等级
	OperationJSON  string `gorm:"type:text" json:"operation_json"`           // 完整操作参数 JSON
	ToolName       string `gorm:"size:100" json:"tool_name"`                 // Function Calling 工具名
	ToolArgsJSON   string `gorm:"type:text" json:"tool_args_json"`           // 工具调用参数 JSON
	ToolCallID     string `gorm:"size:100" json:"tool_call_id"`              // OpenAI tool_call_id
	ExecuteResult  string `gorm:"type:text" json:"execute_result"`           // 执行结果
	Executed       bool   `gorm:"default:false" json:"executed"`             // 是否已执行
	Summary        string `gorm:"size:500" json:"summary"`                   // 操作摘要（AI 生成）
	Status         uint8  `gorm:"default:1;index" json:"status"`             // 审批状态
	ApproveComment string `gorm:"size:500" json:"approve_comment"`           // 审批备注
	ExpireAt       uint32 `gorm:"default:0" json:"expire_at"`               // 过期时间戳
	CreatedAt      uint32 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt     uint32 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (a *AIApprovalRequest) TableName() string { return "ai_approval_requests" }

// =========================================================================
// 审批操作日志 (ai_approval_logs)
// =========================================================================

// AIApprovalLog 审批操作日志
type AIApprovalLog struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement" json:"id"`
	ApprovalID uint32 `gorm:"not null;index" json:"approval_id"`  // 关联审批请求
	UserID     uint32 `gorm:"not null" json:"user_id"`            // 操作人
	Action     string `gorm:"size:50;not null" json:"action"`     // approve/reject/cancel/expire
	Comment    string `gorm:"size:500" json:"comment"`            // 操作说明
	CreatedAt  uint32 `gorm:"autoCreateTime" json:"created_at"`
}

func (l *AIApprovalLog) TableName() string { return "ai_approval_logs" }

// =========================================================================
// CRUD 方法
// =========================================================================

// --- AIConversation ---

func (c *AIConversation) Create(db *gorm.DB) error {
	return db.Create(c).Error
}

func (c *AIConversation) ListByUser(db *gorm.DB, userID uint32, page, pageSize int) ([]*AIConversation, int64, error) {
	var list []*AIConversation
	var total int64
	query := db.Where("user_id = ? AND status = 1", userID)
	query.Model(&AIConversation{}).Count(&total)
	err := query.Order("modified_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (c *AIConversation) GetByID(db *gorm.DB, id, userID uint32) (*AIConversation, error) {
	var conv AIConversation
	err := db.Where("id = ? AND user_id = ?", id, userID).First(&conv).Error
	return &conv, err
}

func (c *AIConversation) Delete(db *gorm.DB, id, userID uint32) error {
	return db.Where("id = ? AND user_id = ?", id, userID).Update("status", 2).Error
}

// --- AIMessage ---

func (m *AIMessage) Create(db *gorm.DB) error {
	return db.Create(m).Error
}

func (m *AIMessage) ListByConversation(db *gorm.DB, convID uint32) ([]*AIMessage, error) {
	var list []*AIMessage
	err := db.Where("conversation_id = ?", convID).Order("created_at ASC").Find(&list).Error
	return list, err
}

// --- AIApprovalRequest ---

func (a *AIApprovalRequest) Create(db *gorm.DB) error {
	return db.Create(a).Error
}

func (a *AIApprovalRequest) GetByID(db *gorm.DB, id uint32) (*AIApprovalRequest, error) {
	var req AIApprovalRequest
	err := db.Where("id = ?", id).First(&req).Error
	return &req, err
}

func (a *AIApprovalRequest) ListPending(db *gorm.DB, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var list []*AIApprovalRequest
	var total int64
	query := db.Where("status = ?", AIApprovalPending)
	query.Model(&AIApprovalRequest{}).Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (a *AIApprovalRequest) ListAll(db *gorm.DB, status uint8, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var list []*AIApprovalRequest
	var total int64
	query := db.Model(&AIApprovalRequest{})
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (a *AIApprovalRequest) UpdateStatus(db *gorm.DB, id uint32, status uint8, approverID uint32, comment string) error {
	return db.Model(&AIApprovalRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           status,
		"approver_user_id": approverID,
		"approve_comment":  comment,
	}).Error
}

func (a *AIApprovalRequest) ListByUser(db *gorm.DB, userID uint32, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var list []*AIApprovalRequest
	var total int64
	query := db.Where("request_user_id = ?", userID)
	query.Model(&AIApprovalRequest{}).Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// --- AIApprovalLog ---

func (l *AIApprovalLog) Create(db *gorm.DB) error {
	return db.Create(l).Error
}

func (l *AIApprovalLog) ListByApproval(db *gorm.DB, approvalID uint32) ([]*AIApprovalLog, error) {
	var list []*AIApprovalLog
	err := db.Where("approval_id = ?", approvalID).Order("created_at ASC").Find(&list).Error
	return list, err
}
