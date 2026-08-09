package services

import (
	"context"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/domain/settings"
	"k8soperation/internal/infra/persistence"
)

func (s *Services) settingsSvc() *settings.SettingsService {
	return settings.NewSettingsService(persistence.NewSettingsRepository(s.db), global.PlatformSetting)
}

// PlatformSettingsGet 获取所有平台设置（混合模式）
func (s *Services) PlatformSettingsGet(ctx context.Context) (*models.PlatformSettingsResponse, error) {
	return s.settingsSvc().Get(ctx)
}

// PlatformSettingsUpdate 更新平台设置
func (s *Services) PlatformSettingsUpdate(ctx context.Context, req *models.PlatformSettingsResponse) error {
	return s.settingsSvc().Update(ctx, req)
}

// PlatformSettingsReset 重置为 config.yaml 默认设置
func (s *Services) PlatformSettingsReset(ctx context.Context) (*models.PlatformSettingsResponse, error) {
	return s.settingsSvc().Reset(ctx)
}
