package settings

import "context"

// SettingsRepository 平台设置仓储接口
type SettingsRepository interface {
	FindAll(ctx context.Context) ([]*PlatformSettings, error)
	BatchUpsert(ctx context.Context, settings []*PlatformSettings) error
}
