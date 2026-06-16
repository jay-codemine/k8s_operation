package dao

import (
	"k8soperation/internal/app/models"
)

// =========================================================================
// AI 会话
// =========================================================================

func (d *Dao) AIConversationCreate(conv *models.AIConversation) error {
	return conv.Create(d.db)
}

func (d *Dao) AIConversationList(userID uint32, page, pageSize int) ([]*models.AIConversation, int64, error) {
	m := &models.AIConversation{}
	return m.ListByUser(d.db, userID, page, pageSize)
}

func (d *Dao) AIConversationGet(id, userID uint32) (*models.AIConversation, error) {
	m := &models.AIConversation{}
	return m.GetByID(d.db, id, userID)
}

func (d *Dao) AIConversationDelete(id, userID uint32) error {
	m := &models.AIConversation{}
	return m.Delete(d.db, id, userID)
}

func (d *Dao) AIConversationUpdateTitle(id uint32, title string) error {
	return d.db.Model(&models.AIConversation{}).Where("id = ?", id).Update("title", title).Error
}

// =========================================================================
// AI 消息
// =========================================================================

func (d *Dao) AIMessageCreate(msg *models.AIMessage) error {
	return msg.Create(d.db)
}

func (d *Dao) AIMessageListByConversation(convID uint32) ([]*models.AIMessage, error) {
	m := &models.AIMessage{}
	return m.ListByConversation(d.db, convID)
}

// =========================================================================
// AI 审批请求
// =========================================================================

func (d *Dao) AIApprovalCreate(req *models.AIApprovalRequest) error {
	return req.Create(d.db)
}

func (d *Dao) AIApprovalGetByID(id uint32) (*models.AIApprovalRequest, error) {
	m := &models.AIApprovalRequest{}
	return m.GetByID(d.db, id)
}

func (d *Dao) AIApprovalListPending(page, pageSize int) ([]*models.AIApprovalRequest, int64, error) {
	m := &models.AIApprovalRequest{}
	return m.ListPending(d.db, page, pageSize)
}

func (d *Dao) AIApprovalListAll(status uint8, page, pageSize int) ([]*models.AIApprovalRequest, int64, error) {
	m := &models.AIApprovalRequest{}
	return m.ListAll(d.db, status, page, pageSize)
}

func (d *Dao) AIApprovalListByUser(userID uint32, page, pageSize int) ([]*models.AIApprovalRequest, int64, error) {
	m := &models.AIApprovalRequest{}
	return m.ListByUser(d.db, userID, page, pageSize)
}

// AIApprovalMyPendingCount 获取用户自己提交的待审批数量
func (d *Dao) AIApprovalMyPendingCount(userID uint32) (int64, error) {
	var count int64
	err := d.db.Model(&models.AIApprovalRequest{}).
		Where("request_user_id = ? AND status = ?", userID, models.AIApprovalPending).
		Count(&count).Error
	return count, err
}

func (d *Dao) AIApprovalUpdateStatus(id uint32, status uint8, approverID uint32, comment string) error {
	m := &models.AIApprovalRequest{}
	return m.UpdateStatus(d.db, id, status, approverID, comment)
}

func (d *Dao) AIApprovalUpdateExecuteResult(id uint32, result string) error {
	return d.db.Model(&models.AIApprovalRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"executed":       true,
		"execute_result": result,
	}).Error
}

// AIApprovalDelete 硬删除审批记录（管理员专用）
func (d *Dao) AIApprovalDelete(id uint32) error {
	// 同时删除关联日志
	d.db.Where("approval_id = ?", id).Delete(&models.AIApprovalLog{})
	return d.db.Where("id = ?", id).Delete(&models.AIApprovalRequest{}).Error
}

// AIApprovalUpdate 更新审批备注
func (d *Dao) AIApprovalUpdate(id uint32, updates map[string]interface{}) error {
	return d.db.Model(&models.AIApprovalRequest{}).Where("id = ?", id).Updates(updates).Error
}

// AIApprovalStats 统计各状态数量
func (d *Dao) AIApprovalStats() (map[string]int64, error) {
	stats := make(map[string]int64)
	var results []struct {
		Status uint8
		Count  int64
	}
	err := d.db.Model(&models.AIApprovalRequest{}).
		Select("status, count(*) as count").
		Group("status").Find(&results).Error
	if err != nil {
		return stats, err
	}
	for _, r := range results {
		switch r.Status {
		case models.AIApprovalPending:
			stats["pending"] = r.Count
		case models.AIApprovalApproved:
			stats["approved"] = r.Count
		case models.AIApprovalRejected:
			stats["rejected"] = r.Count
		case models.AIApprovalExpired:
			stats["expired"] = r.Count
		case models.AIApprovalCanceled:
			stats["canceled"] = r.Count
		}
	}
	return stats, nil
}

// =========================================================================
// AI 审批日志
// =========================================================================

func (d *Dao) AIApprovalLogCreate(log *models.AIApprovalLog) error {
	return log.Create(d.db)
}

func (d *Dao) AIApprovalLogList(approvalID uint32) ([]*models.AIApprovalLog, error) {
	m := &models.AIApprovalLog{}
	return m.ListByApproval(d.db, approvalID)
}
