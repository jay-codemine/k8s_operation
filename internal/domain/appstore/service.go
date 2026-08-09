package appstore

import (
	"context"

	"gorm.io/gorm"
)

// AppStoreService 应用商城领域服务
type AppStoreService struct {
	db *gorm.DB
}

// NewAppStoreService 创建应用商城服务
func NewAppStoreService(db *gorm.DB) *AppStoreService {
	return &AppStoreService{db: db}
}


// ============================================================
// 应用 CRUD
// ============================================================

// AppList 分页查询应用列表
func (s *AppStoreService) AppList(ctx context.Context, req *AppStoreListRequest) ([]*AppStoreApp, int64, error) {
	var list []*AppStoreApp
	var total int64

	query := s.db.WithContext(ctx).Model(&AppStoreApp{}).Where("is_del = 0")

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("(name LIKE ? OR display_name LIKE ? OR description LIKE ? OR tags LIKE ?)", kw, kw, kw, kw)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Featured > 0 {
		query = query.Where("featured = ?", req.Featured)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Order("sort_order DESC, featured DESC, id ASC").
		Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// AppGetByID 根据ID获取应用
func (s *AppStoreService) AppGetByID(ctx context.Context, id uint32) (*AppStoreApp, error) {
	var app AppStoreApp
	if err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// AppGetByName 根据名称获取应用
func (s *AppStoreService) AppGetByName(ctx context.Context, name string) (*AppStoreApp, error) {
	var app AppStoreApp
	if err := s.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// AppCreate 创建应用
func (s *AppStoreService) AppCreate(ctx context.Context, app *AppStoreApp) error {
	return s.db.WithContext(ctx).Create(app).Error
}

// AppUpdate 更新应用
func (s *AppStoreService) AppUpdate(ctx context.Context, app *AppStoreApp) error {
	return s.db.WithContext(ctx).Model(app).Updates(map[string]interface{}{
		"name":         app.Name,
		"display_name": app.DisplayName,
		"category":     app.Category,
		"version":      app.Version,
		"icon":         app.Icon,
		"description":  app.Description,
		"provider":     app.Provider,
		"chart_url":    app.ChartURL,
		"doc_url":      app.DocURL,
		"status":       app.Status,
		"featured":     app.Featured,
		"sort_order":   app.SortOrder,
		"tags":         app.Tags,
		"min_k8s":      app.MinK8s,
		"namespace":    app.Namespace,
		"values_yaml":  app.ValuesYAML,
	}).Error
}

// AppDelete 软删除应用
func (s *AppStoreService) AppDelete(ctx context.Context, id uint32) error {
	return s.db.WithContext(ctx).Model(&AppStoreApp{}).
		Where("id = ?", id).Update("is_del", 1).Error
}

// AppCategories 获取所有分类及计数
func (s *AppStoreService) AppCategories(ctx context.Context) ([]AppStoreCategoryCount, error) {
	var result []AppStoreCategoryCount
	err := s.db.WithContext(ctx).Model(&AppStoreApp{}).
		Select("category, COUNT(*) as count").
		Where("is_del = 0 AND status = 1").
		Group("category").
		Order("count DESC").
		Find(&result).Error
	return result, err
}

// ============================================================
// 安装记录 CRUD
// ============================================================

// InstallCreate 创建安装记录
func (s *AppStoreService) InstallCreate(ctx context.Context, install *AppStoreInstall) error {
	return s.db.WithContext(ctx).Create(install).Error
}

// InstallUpdate 更新安装记录
func (s *AppStoreService) InstallUpdate(ctx context.Context, id uint32, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&AppStoreInstall{}).
		Where("id = ?", id).Updates(updates).Error
}

// InstallGetByID 根据ID获取安装记录
func (s *AppStoreService) InstallGetByID(ctx context.Context, id uint32) (*AppStoreInstall, error) {
	var install AppStoreInstall
	if err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&install).Error; err != nil {
		return nil, err
	}
	return &install, nil
}

// InstallList 分页查询安装记录
func (s *AppStoreService) InstallList(ctx context.Context, req *AppStoreInstallListRequest) ([]*AppStoreInstall, int64, error) {
	var list []*AppStoreInstall
	var total int64

	query := s.db.WithContext(ctx).Model(&AppStoreInstall{}).Where("is_del = 0")

	if req.AppID > 0 {
		query = query.Where("app_id = ?", req.AppID)
	}
	if req.ClusterID > 0 {
		query = query.Where("cluster_id = ?", req.ClusterID)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// InstallFindActive 查找活跃安装记录
func (s *AppStoreService) InstallFindActive(ctx context.Context, appID, clusterID uint32, namespace string) (*AppStoreInstall, error) {
	var install AppStoreInstall
	err := s.db.WithContext(ctx).
		Where("app_id = ? AND cluster_id = ? AND namespace = ? AND status IN (?, ?) AND is_del = 0",
			appID, clusterID, namespace, InstallStatusInstalling, InstallStatusInstalled).
		First(&install).Error
	if err != nil {
		return nil, err
	}
	return &install, nil
}

// InstallDelete 软删除安装记录
func (s *AppStoreService) InstallDelete(ctx context.Context, id uint32) error {
	return s.db.WithContext(ctx).Model(&AppStoreInstall{}).
		Where("id = ?", id).Update("is_del", 1).Error
}

// ============================================================
// 应用组件 CRUD
// ============================================================

// ComponentListByAppID 获取应用的所有组件
func (s *AppStoreService) ComponentListByAppID(ctx context.Context, appID uint32) ([]*AppStoreComponent, error) {
	var list []*AppStoreComponent
	err := s.db.WithContext(ctx).
		Where("app_id = ? AND is_del = 0", appID).
		Order("sort_order DESC, id ASC").
		Find(&list).Error
	return list, err
}

// ComponentCreate 创建组件
func (s *AppStoreService) ComponentCreate(ctx context.Context, comp *AppStoreComponent) error {
	return s.db.WithContext(ctx).Create(comp).Error
}

// ComponentUpdate 更新组件
func (s *AppStoreService) ComponentUpdate(ctx context.Context, comp *AppStoreComponent) error {
	return s.db.WithContext(ctx).Model(comp).Updates(map[string]interface{}{
		"name":       comp.Name,
		"image":      comp.Image,
		"replicas":   comp.Replicas,
		"ports":      comp.Ports,
		"args":       comp.Args,
		"cpu_req":    comp.CPUReq,
		"cpu_lim":    comp.CPULim,
		"mem_req":    comp.MemReq,
		"mem_lim":    comp.MemLim,
		"sort_order": comp.SortOrder,
	}).Error
}

// ComponentDelete 软删除组件
func (s *AppStoreService) ComponentDelete(ctx context.Context, id uint32) error {
	return s.db.WithContext(ctx).Model(&AppStoreComponent{}).
		Where("id = ?", id).Update("is_del", 1).Error
}

// ComponentGetByID 根据ID获取组件
func (s *AppStoreService) ComponentGetByID(ctx context.Context, id uint32) (*AppStoreComponent, error) {
	var comp AppStoreComponent
	if err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&comp).Error; err != nil {
		return nil, err
	}
	return &comp, nil
}

// ComponentCountByAppID 获取应用的组件数量
func (s *AppStoreService) ComponentCountByAppID(ctx context.Context, appID uint32) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&AppStoreComponent{}).
		Where("app_id = ? AND is_del = 0", appID).Count(&count).Error
	return count, err
}

// ComponentBatchDelete 批量软删除组件
func (s *AppStoreService) ComponentBatchDelete(ctx context.Context, ids []uint32) error {
	return s.db.WithContext(ctx).Model(&AppStoreComponent{}).
		Where("id IN ?", ids).Update("is_del", 1).Error
}

// ComponentUpdateSort 更新组件排序
func (s *AppStoreService) ComponentUpdateSort(ctx context.Context, id uint32, sortOrder int) error {
	return s.db.WithContext(ctx).Model(&AppStoreComponent{}).
		Where("id = ?", id).Update("sort_order", sortOrder).Error
}
