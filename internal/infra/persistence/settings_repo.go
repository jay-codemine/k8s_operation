package persistence

import (
	"context"

	"gorm.io/gorm"

	"k8soperation/internal/domain/settings"
)

type settingsRepo struct {
	db *gorm.DB
}

// NewSettingsRepository 创建平台设置仓储
func NewSettingsRepository(db *gorm.DB) settings.SettingsRepository {
	return &settingsRepo{db: db}
}

func (r *settingsRepo) FindAll(ctx context.Context) ([]*settings.PlatformSettings, error) {
	var list []*settings.PlatformSettings
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *settingsRepo) BatchUpsert(ctx context.Context, list []*settings.PlatformSettings) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range list {
			var existing settings.PlatformSettings
			err := tx.Where("category = ? AND `key` = ?", s.Category, s.Key).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(s).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"value": s.Value, "value_type": s.ValueType,
					"label": s.Label, "desc": s.Desc, "modified_at": s.ModifiedAt,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
