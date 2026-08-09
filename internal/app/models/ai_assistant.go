package models

import dm "k8soperation/internal/domain/ai"

// ===== 类型别名 =====

type (
	AIConversation    = dm.AIConversation
	AIMessage         = dm.AIMessage
	AIApprovalRequest = dm.AIApprovalRequest
	AIApprovalLog     = dm.AIApprovalLog
)

// ===== 审批状态常量 =====

const (
	AIApprovalPending  = dm.AIApprovalPending
	AIApprovalApproved = dm.AIApprovalApproved
	AIApprovalRejected = dm.AIApprovalRejected
	AIApprovalExpired  = dm.AIApprovalExpired
	AIApprovalCanceled = dm.AIApprovalCanceled
)

// ===== 风险等级常量 =====

const (
	AIRiskLow      = dm.AIRiskLow
	AIRiskMedium   = dm.AIRiskMedium
	AIRiskHigh     = dm.AIRiskHigh
	AIRiskCritical = dm.AIRiskCritical
)
