package persistence

import (
	"context"

	"gorm.io/gorm"

	"k8soperation/internal/domain/ai"
)

type aiRepoImpl struct{ db *gorm.DB }

func NewAIRepository(db *gorm.DB) ai.AIRepository { return &aiRepoImpl{db: db} }

func (r *aiRepoImpl) SaveConversation(ctx context.Context, c *ai.AIConversation) error {
	return r.db.WithContext(ctx).Create(c).Error
}
func (r *aiRepoImpl) FindConversationByID(ctx context.Context, id int64) (*ai.AIConversation, error) {
	var c ai.AIConversation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
func (r *aiRepoImpl) QueryConversations(ctx context.Context, userID int64, page, limit int) ([]*ai.AIConversation, int64, error) {
	db := r.db.WithContext(ctx).Model(&ai.AIConversation{}).Where("is_del = 0")
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*ai.AIConversation
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 20 }
	if err := db.Order("id DESC").Offset((page-1)*limit).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
func (r *aiRepoImpl) DeleteConversation(ctx context.Context, id, userID int64) error {
	return r.db.WithContext(ctx).Model(&ai.AIConversation{}).Where("id = ? AND user_id = ?", id, userID).Update("is_del", 1).Error
}
