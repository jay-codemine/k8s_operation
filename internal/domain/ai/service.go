package ai

import (
	"context"

	"gorm.io/gorm"
)

// AIService AI 助手领域服务
type AIService struct {
	db *gorm.DB
}

// NewAIService 创建 AI 服务
func NewAIService(db *gorm.DB) *AIService {
	return &AIService{db: db}
}


// ==================== 会话 CRUD ====================

// ConversationCreate 创建会话
func (s *AIService) ConversationCreate(ctx context.Context, conv *AIConversation) error {
	return conv.Create(s.db.WithContext(ctx))
}

// ConversationList 用户会话列表
func (s *AIService) ConversationList(ctx context.Context, userID uint32, page, pageSize int) ([]*AIConversation, int64, error) {
	var m AIConversation
	return m.ListByUser(s.db.WithContext(ctx), userID, page, pageSize)
}

// ConversationGetByID 获取会话
func (s *AIService) ConversationGetByID(ctx context.Context, id, userID uint32) (*AIConversation, error) {
	var m AIConversation
	return m.GetByID(s.db.WithContext(ctx), id, userID)
}

// ConversationDelete 归档会话
func (s *AIService) ConversationDelete(ctx context.Context, id, userID uint32) error {
	var m AIConversation
	return m.Delete(s.db.WithContext(ctx), id, userID)
}

// ConversationUpdateTitle 更新会话标题
func (s *AIService) ConversationUpdateTitle(ctx context.Context, id uint32, title string) error {
	return s.db.WithContext(ctx).Model(&AIConversation{}).Where("id = ?", id).Update("title", title).Error
}

// ==================== 消息 CRUD ====================

// MessageCreate 保存消息
func (s *AIService) MessageCreate(ctx context.Context, msg *AIMessage) error {
	return msg.Create(s.db.WithContext(ctx))
}

// MessageListByConversation 获取会话消息
func (s *AIService) MessageListByConversation(ctx context.Context, convID uint32) ([]*AIMessage, error) {
	var m AIMessage
	return m.ListByConversation(s.db.WithContext(ctx), convID)
}

// ==================== 审批请求 CRUD ====================

// ApprovalCreate 创建审批请求
func (s *AIService) ApprovalCreate(ctx context.Context, req *AIApprovalRequest) error {
	return req.Create(s.db.WithContext(ctx))
}

// ApprovalGetByID 获取审批请求
func (s *AIService) ApprovalGetByID(ctx context.Context, id uint32) (*AIApprovalRequest, error) {
	var m AIApprovalRequest
	return m.GetByID(s.db.WithContext(ctx), id)
}

// ApprovalListPending 待审批列表
func (s *AIService) ApprovalListPending(ctx context.Context, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var m AIApprovalRequest
	return m.ListPending(s.db.WithContext(ctx), page, pageSize)
}

// ApprovalListAll 所有审批列表（可按状态筛选）
func (s *AIService) ApprovalListAll(ctx context.Context, status uint8, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var m AIApprovalRequest
	return m.ListAll(s.db.WithContext(ctx), status, page, pageSize)
}

// ApprovalListByUser 用户提交的审批列表
func (s *AIService) ApprovalListByUser(ctx context.Context, userID uint32, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var m AIApprovalRequest
	return m.ListByUser(s.db.WithContext(ctx), userID, page, pageSize)
}

// ApprovalUpdateStatus 更新审批状态
func (s *AIService) ApprovalUpdateStatus(ctx context.Context, id uint32, status uint8, approverID uint32, comment string) error {
	var m AIApprovalRequest
	return m.UpdateStatus(s.db.WithContext(ctx), id, status, approverID, comment)
}

// ApprovalUpdateExecuteResult 更新执行结果
func (s *AIService) ApprovalUpdateExecuteResult(ctx context.Context, id uint32, result string) error {
	return s.db.WithContext(ctx).Model(&AIApprovalRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"executed":       true,
		"execute_result": result,
	}).Error
}

// ApprovalDelete 删除审批记录及关联日志
func (s *AIService) ApprovalDelete(ctx context.Context, id uint32) error {
	s.db.WithContext(ctx).Where("approval_id = ?", id).Delete(&AIApprovalLog{})
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&AIApprovalRequest{}).Error
}

// ApprovalUpdate 更新审批备注
func (s *AIService) ApprovalUpdate(ctx context.Context, id uint32, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&AIApprovalRequest{}).Where("id = ?", id).Updates(updates).Error
}

// ApprovalStats 统计各状态数量
func (s *AIService) ApprovalStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)
	var results []struct {
		Status uint8
		Count  int64
	}
	err := s.db.WithContext(ctx).Model(&AIApprovalRequest{}).
		Select("status, count(*) as count").
		Group("status").Find(&results).Error
	if err != nil {
		return stats, err
	}
	for _, r := range results {
		switch r.Status {
		case AIApprovalPending:
			stats["pending"] = r.Count
		case AIApprovalApproved:
			stats["approved"] = r.Count
		case AIApprovalRejected:
			stats["rejected"] = r.Count
		case AIApprovalExpired:
			stats["expired"] = r.Count
		case AIApprovalCanceled:
			stats["canceled"] = r.Count
		}
	}
	return stats, nil
}

// ApprovalMyPendingCount 用户待审批数量
func (s *AIService) ApprovalMyPendingCount(ctx context.Context, userID uint32) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&AIApprovalRequest{}).
		Where("request_user_id = ? AND status = ?", userID, AIApprovalPending).
		Count(&count).Error
	return count, err
}

// ==================== 审批日志 CRUD ====================

// ApprovalLogCreate 记录审批日志
func (s *AIService) ApprovalLogCreate(ctx context.Context, log *AIApprovalLog) error {
	return log.Create(s.db.WithContext(ctx))
}

// ApprovalLogListByApproval 获取审批日志
func (s *AIService) ApprovalLogListByApproval(ctx context.Context, approvalID uint32) ([]*AIApprovalLog, error) {
	var m AIApprovalLog
	return m.ListByApproval(s.db.WithContext(ctx), approvalID)
}
