package audit

import "context"

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	Save(ctx context.Context, log *AuditLog) error
	BatchSave(ctx context.Context, logs []*AuditLog) error
	FindByID(ctx context.Context, id int64) (*AuditLog, error)
	DeleteBefore(ctx context.Context, beforeTimestamp int64) (int64, error)
	Query(ctx context.Context, query *AuditLogQuery) (*AuditLogListResponse, error)
	QueryStatistics(ctx context.Context) (*AuditStatistics, error)
}
