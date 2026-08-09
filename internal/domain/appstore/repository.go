package appstore

import "context"

// AppStoreRepository 应用商城仓储接口
type AppStoreRepository interface {
	SaveApp(ctx context.Context, app *AppStoreApp) error
	FindAppByID(ctx context.Context, id int64) (*AppStoreApp, error)
	QueryApps(ctx context.Context, keyword, category string, page, limit int) ([]*AppStoreApp, int64, error)
	DeleteApp(ctx context.Context, id int64) error
}
