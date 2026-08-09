package persistence

import (
	"context"

	"gorm.io/gorm"

	"k8soperation/internal/domain/image"
)

type imageRepoImpl struct{ db *gorm.DB }

func NewImageRepository(db *gorm.DB) image.ImageRepository { return &imageRepoImpl{db: db} }

func (r *imageRepoImpl) RegistrySave(ctx context.Context, reg *image.ImageRegistry) error {
	return r.db.WithContext(ctx).Create(reg).Error
}
func (r *imageRepoImpl) RegistryFindByID(ctx context.Context, id int64) (*image.ImageRegistry, error) {
	var reg image.ImageRegistry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}
func (r *imageRepoImpl) RegistryFindByName(ctx context.Context, name string) (*image.ImageRegistry, error) {
	var reg image.ImageRegistry
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}
func (r *imageRepoImpl) RegistryUpdate(ctx context.Context, reg *image.ImageRegistry) error {
	return r.db.WithContext(ctx).Model(reg).Where("id = ?", reg.ID).Updates(map[string]interface{}{
		"name": reg.Name, "type": reg.Type, "url": reg.URL, "username": reg.Username,
		"password": reg.Password, "insecure": reg.Insecure, "description": reg.Description,
	}).Error
}
func (r *imageRepoImpl) RegistryDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&image.ImageRegistry{}).Where("id = ?", id).Update("is_del", 1).Error
}
func (r *imageRepoImpl) RegistryList(ctx context.Context, keyword, typ string, page, pageSize int) ([]image.ImageRegistry, int64, error) {
	db := r.db.WithContext(ctx).Model(&image.ImageRegistry{}).Where("is_del = 0")
	if keyword != "" { db = db.Where("name LIKE ? OR url LIKE ?", "%"+keyword+"%", "%"+keyword+"%") }
	if typ != "" { db = db.Where("type = ?", typ) }
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	var list []image.ImageRegistry
	if err := db.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
func (r *imageRepoImpl) RegistryListAll(ctx context.Context) ([]image.ImageRegistry, error) {
	var list []image.ImageRegistry
	err := r.db.WithContext(ctx).Where("is_del = 0").Find(&list).Error
	return list, err
}
func (r *imageRepoImpl) RegistryUpdateStatus(ctx context.Context, id int64, status, lastErr string, checkTime int64) error {
	return r.db.WithContext(ctx).Model(&image.ImageRegistry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status, "last_error": lastErr, "last_check_at": checkTime,
	}).Error
}

func (r *imageRepoImpl) PolicySave(ctx context.Context, p *image.ImageCleanupPolicy) error {
	return r.db.WithContext(ctx).Create(p).Error
}
func (r *imageRepoImpl) PolicyFindByID(ctx context.Context, id int64) (*image.ImageCleanupPolicy, error) {
	var p image.ImageCleanupPolicy
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil { return nil, err }
	return &p, nil
}
func (r *imageRepoImpl) PolicyUpdate(ctx context.Context, p *image.ImageCleanupPolicy) error {
	return r.db.WithContext(ctx).Model(p).Where("id = ?", p.ID).Updates(p).Error
}
func (r *imageRepoImpl) PolicyDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&image.ImageCleanupPolicy{}).Where("id = ?", id).Update("is_del", 1).Error
}
func (r *imageRepoImpl) PolicyToggle(ctx context.Context, id int64, enabled bool) error {
	return r.db.WithContext(ctx).Model(&image.ImageCleanupPolicy{}).Where("id = ?", id).Update("enabled", enabled).Error
}
func (r *imageRepoImpl) PolicyList(ctx context.Context, registryID int64, keyword string, page, pageSize int) ([]image.ImageCleanupPolicy, int64, error) {
	db := r.db.WithContext(ctx).Model(&image.ImageCleanupPolicy{}).Where("is_del = 0")
	if registryID > 0 { db = db.Where("registry_id = ?", registryID) }
	if keyword != "" { db = db.Where("name LIKE ?", "%"+keyword+"%") }
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	var list []image.ImageCleanupPolicy
	if err := db.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
