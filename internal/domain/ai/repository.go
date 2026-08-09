package ai

import "context"

// AIRepository AI 域仓储接口
type AIRepository interface {
	SaveConversation(ctx context.Context, c *AIConversation) error
	FindConversationByID(ctx context.Context, id int64) (*AIConversation, error)
	QueryConversations(ctx context.Context, userID int64, page, limit int) ([]*AIConversation, int64, error)
	DeleteConversation(ctx context.Context, id, userID int64) error
}
