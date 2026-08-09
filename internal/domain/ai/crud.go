package ai

import "gorm.io/gorm"

// ==================== AIConversation CRUD ====================

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

// ==================== AIMessage CRUD ====================

func (m *AIMessage) Create(db *gorm.DB) error {
	return db.Create(m).Error
}

func (m *AIMessage) ListByConversation(db *gorm.DB, convID uint32) ([]*AIMessage, error) {
	var list []*AIMessage
	err := db.Where("conversation_id = ?", convID).Order("created_at ASC").Find(&list).Error
	return list, err
}

// ==================== AIApprovalRequest CRUD ====================

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
	if err := db.Session(&gorm.Session{}).Where("status = ?", AIApprovalPending).Model(&AIApprovalRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return list, 0, nil
	}
	err := db.Session(&gorm.Session{}).Where("status = ?", AIApprovalPending).Model(&AIApprovalRequest{}).
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (a *AIApprovalRequest) ListAll(db *gorm.DB, st uint8, page, pageSize int) ([]*AIApprovalRequest, int64, error) {
	var list []*AIApprovalRequest
	var total int64
	base := db.Model(&AIApprovalRequest{})
	if st > 0 {
		base = base.Where("status = ?", st)
	}
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return list, 0, nil
	}
	err := base.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
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
	if err := db.Where("request_user_id = ?", userID).Model(&AIApprovalRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return list, 0, nil
	}
	err := db.Where("request_user_id = ?", userID).Model(&AIApprovalRequest{}).
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ==================== AIApprovalLog CRUD ====================

func (l *AIApprovalLog) Create(db *gorm.DB) error {
	return db.Create(l).Error
}

func (l *AIApprovalLog) ListByApproval(db *gorm.DB, approvalID uint32) ([]*AIApprovalLog, error) {
	var list []*AIApprovalLog
	err := db.Where("approval_id = ?", approvalID).Order("created_at ASC").Find(&list).Error
	return list, err
}
