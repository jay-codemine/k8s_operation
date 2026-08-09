package persistence

import (
	"context"

	"gorm.io/gorm"

	"k8soperation/internal/domain/appstore"
)

type appstoreRepoImpl struct{ db *gorm.DB }

func NewAppStoreRepository(db *gorm.DB) appstore.AppStoreRepository { return &appstoreRepoImpl{db: db} }

func (r *appstoreRepoImpl) SaveApp(ctx context.Context, app *appstore.AppStoreApp) error {
	return r.db.WithContext(ctx).Create(app).Error
}
func (r *appstoreRepoImpl) FindAppByID(ctx context.Context, id int64) (*appstore.AppStoreApp, error) {
	var app appstore.AppStoreApp
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}
func (r *appstoreRepoImpl) QueryApps(ctx context.Context, keyword, category string, page, limit int) ([]*appstore.AppStoreApp, int64, error) {
	db := r.db.WithContext(ctx).Model(&appstore.AppStoreApp{}).Where("is_del = 0")
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	var list []*appstore.AppStoreApp
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 20 }
	if err := db.Order("id DESC").Offset((page-1)*limit).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
func (r *appstoreRepoImpl) DeleteApp(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&appstore.AppStoreApp{}).Where("id = ?", id).Update("is_del", 1).Error
}
