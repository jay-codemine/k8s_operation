package image

import "context"

// ImageRepository 镜像域仓储接口
type ImageRepository interface {
	RegistrySave(ctx context.Context, r *ImageRegistry) error
	RegistryFindByID(ctx context.Context, id int64) (*ImageRegistry, error)
	RegistryFindByName(ctx context.Context, name string) (*ImageRegistry, error)
	RegistryUpdate(ctx context.Context, r *ImageRegistry) error
	RegistryDelete(ctx context.Context, id int64) error
	RegistryList(ctx context.Context, keyword, typ string, page, pageSize int) ([]ImageRegistry, int64, error)
	RegistryListAll(ctx context.Context) ([]ImageRegistry, error)
	RegistryUpdateStatus(ctx context.Context, id int64, status, lastErr string, checkTime int64) error

	PolicySave(ctx context.Context, p *ImageCleanupPolicy) error
	PolicyFindByID(ctx context.Context, id int64) (*ImageCleanupPolicy, error)
	PolicyUpdate(ctx context.Context, p *ImageCleanupPolicy) error
	PolicyDelete(ctx context.Context, id int64) error
	PolicyToggle(ctx context.Context, id int64, enabled bool) error
	PolicyList(ctx context.Context, registryID int64, keyword string, page, pageSize int) ([]ImageCleanupPolicy, int64, error)
}
