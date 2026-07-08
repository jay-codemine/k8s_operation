package dao

import (
	"context"
	"k8soperation/internal/app/models"
)

// ============================================================
// 应用商城 DAO
// ============================================================

// AppStoreList 分页查询应用列表
func (d *Dao) AppStoreList(ctx context.Context, req *models.AppStoreListRequest) ([]*models.AppStoreApp, int64, error) {
	var list []*models.AppStoreApp
	var total int64

	query := d.db.WithContext(ctx).Model(&models.AppStoreApp{}).Where("is_del = 0")

	// 分类筛选
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	// 关键词搜索
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("(name LIKE ? OR display_name LIKE ? OR description LIKE ? OR tags LIKE ?)", kw, kw, kw, kw)
	}
	// 状态筛选
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	// 推荐筛选
	if req.Featured > 0 {
		query = query.Where("featured = ?", req.Featured)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
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

// AppStoreGetByID 根据 ID 获取应用详情
func (d *Dao) AppStoreGetByID(ctx context.Context, id uint32) (*models.AppStoreApp, error) {
	var app models.AppStoreApp
	if err := d.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// AppStoreCreate 创建应用
func (d *Dao) AppStoreCreate(ctx context.Context, app *models.AppStoreApp) error {
	return d.db.WithContext(ctx).Create(app).Error
}

// AppStoreUpdate 更新应用
func (d *Dao) AppStoreUpdate(ctx context.Context, app *models.AppStoreApp) error {
	return d.db.WithContext(ctx).Model(app).Updates(map[string]interface{}{
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

// AppStoreDelete 软删除应用
func (d *Dao) AppStoreDelete(ctx context.Context, id uint32) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreApp{}).
		Where("id = ?", id).
		Update("is_del", 1).Error
}

// AppStoreCategories 获取所有分类及计数
func (d *Dao) AppStoreCategories(ctx context.Context) ([]models.AppStoreCategoryCount, error) {
	var result []models.AppStoreCategoryCount
	err := d.db.WithContext(ctx).Model(&models.AppStoreApp{}).
		Select("category, COUNT(*) as count").
		Where("is_del = 0 AND status = 1").
		Group("category").
		Order("count DESC").
		Find(&result).Error
	return result, err
}

// AppStoreGetByName 根据名称查找（唯一性检查）
func (d *Dao) AppStoreGetByName(ctx context.Context, name string) (*models.AppStoreApp, error) {
	var app models.AppStoreApp
	if err := d.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// ============================================================
// 安装记录 DAO
// ============================================================

// AppStoreInstallCreate 创建安装记录
func (d *Dao) AppStoreInstallCreate(ctx context.Context, install *models.AppStoreInstall) error {
	return d.db.WithContext(ctx).Create(install).Error
}

// AppStoreInstallUpdate 更新安装记录
func (d *Dao) AppStoreInstallUpdate(ctx context.Context, id uint32, updates map[string]interface{}) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreInstall{}).
		Where("id = ?", id).Updates(updates).Error
}

// AppStoreInstallGetByID 根据 ID 获取安装记录
func (d *Dao) AppStoreInstallGetByID(ctx context.Context, id uint32) (*models.AppStoreInstall, error) {
	var install models.AppStoreInstall
	if err := d.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&install).Error; err != nil {
		return nil, err
	}
	return &install, nil
}

// AppStoreInstallList 分页查询安装记录
func (d *Dao) AppStoreInstallList(ctx context.Context, req *models.AppStoreInstallListRequest) ([]*models.AppStoreInstall, int64, error) {
	var list []*models.AppStoreInstall
	var total int64

	query := d.db.WithContext(ctx).Model(&models.AppStoreInstall{}).Where("is_del = 0")

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

// AppStoreInstallFindActive 查找某应用在某集群的活跃安装（安装中/已安装）
func (d *Dao) AppStoreInstallFindActive(ctx context.Context, appID, clusterID uint32, namespace string) (*models.AppStoreInstall, error) {
	var install models.AppStoreInstall
	err := d.db.WithContext(ctx).
		Where("app_id = ? AND cluster_id = ? AND namespace = ? AND status IN (?, ?) AND is_del = 0",
			appID, clusterID, namespace,
			models.InstallStatusInstalling, models.InstallStatusInstalled).
		First(&install).Error
	if err != nil {
		return nil, err
	}
	return &install, nil
}

// AppStoreInstallDelete 软删除安装记录
func (d *Dao) AppStoreInstallDelete(ctx context.Context, id uint32) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreInstall{}).
		Where("id = ?", id).
		Update("is_del", 1).Error
}

// ============================================================
// 应用组件 DAO
// ============================================================

// AppStoreComponentListByAppID 获取应用的所有组件（按 sort_order 倒序）
func (d *Dao) AppStoreComponentListByAppID(ctx context.Context, appID uint32) ([]*models.AppStoreComponent, error) {
	var list []*models.AppStoreComponent
	err := d.db.WithContext(ctx).
		Where("app_id = ? AND is_del = 0", appID).
		Order("sort_order DESC, id ASC").
		Find(&list).Error
	return list, err
}

// AppStoreComponentCreate 创建组件
func (d *Dao) AppStoreComponentCreate(ctx context.Context, comp *models.AppStoreComponent) error {
	return d.db.WithContext(ctx).Create(comp).Error
}

// AppStoreComponentUpdate 更新组件
func (d *Dao) AppStoreComponentUpdate(ctx context.Context, comp *models.AppStoreComponent) error {
	return d.db.WithContext(ctx).Model(comp).Updates(map[string]interface{}{
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

// AppStoreComponentDelete 删除组件（软删除）
func (d *Dao) AppStoreComponentDelete(ctx context.Context, id uint32) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreComponent{}).
		Where("id = ?", id).
		Update("is_del", 1).Error
}

// AppStoreComponentGetByID 根据ID获取组件
func (d *Dao) AppStoreComponentGetByID(ctx context.Context, id uint32) (*models.AppStoreComponent, error) {
	var comp models.AppStoreComponent
	if err := d.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&comp).Error; err != nil {
		return nil, err
	}
	return &comp, nil
}

// AppStoreComponentCountByAppID 获取应用的组件数量
func (d *Dao) AppStoreComponentCountByAppID(ctx context.Context, appID uint32) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&models.AppStoreComponent{}).
		Where("app_id = ? AND is_del = 0", appID).Count(&count).Error
	return count, err
}

// AppStoreComponentBatchDelete 批量软删除组件
func (d *Dao) AppStoreComponentBatchDelete(ctx context.Context, ids []uint32) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreComponent{}).
		Where("id IN ?", ids).
		Update("is_del", 1).Error
}

// AppStoreComponentUpdateSort 更新组件排序
func (d *Dao) AppStoreComponentUpdateSort(ctx context.Context, id uint32, sortOrder int) error {
	return d.db.WithContext(ctx).Model(&models.AppStoreComponent{}).
		Where("id = ?", id).
		Update("sort_order", sortOrder).Error
}
