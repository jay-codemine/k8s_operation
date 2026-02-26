package dao

import (
	"context"
	"k8soperation/internal/app/models"
	"time"
)

// ==================== Pipeline CRUD ====================

// PipelineCreate 创建流水线
func (d *Dao) PipelineCreate(ctx context.Context, p *models.CicdPipeline) error {
	now := time.Now().Unix()
	p.CreatedAt = uint64(now)
	p.ModifiedAt = uint64(now)
	return d.db.WithContext(ctx).Create(p).Error
}

// PipelineGetByID 根据ID获取流水线
func (d *Dao) PipelineGetByID(ctx context.Context, id int64) (*models.CicdPipeline, error) {
	var p models.CicdPipeline
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PipelineGetByName 根据名称获取流水线（用于唯一性校验）
func (d *Dao) PipelineGetByName(ctx context.Context, name string) (*models.CicdPipeline, error) {
	var p models.CicdPipeline
	err := d.db.WithContext(ctx).
		Where("name = ? AND is_del = 0", name).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PipelineList 获取流水线列表
func (d *Dao) PipelineList(ctx context.Context, keyword, status string, page, pageSize int) ([]*models.CicdPipeline, int64, error) {
	var list []*models.CicdPipeline
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdPipeline{}).Where("is_del = 0")

	// 关键字搜索（名称、描述、Git仓库）
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR git_repo LIKE ?", likeKeyword, likeKeyword, likeKeyword)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 先查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 再查询分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PipelineUpdate 更新流水线
func (d *Dao) PipelineUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipeline{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}

// PipelineDelete 软删除流水线
func (d *Dao) PipelineDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipeline{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// PipelineUpdateStatus 更新流水线状态
func (d *Dao) PipelineUpdateStatus(ctx context.Context, id int64, status string) error {
	return d.PipelineUpdate(ctx, id, map[string]interface{}{
		"status": status,
	})
}

// PipelineUpdateRunInfo 更新流水线运行信息
func (d *Dao) PipelineUpdateRunInfo(ctx context.Context, id int64, runStatus string, buildNumber int, buildURL string) error {
	return d.PipelineUpdate(ctx, id, map[string]interface{}{
		"status":            models.PipelineStatusRunning,
		"last_run_status":   runStatus,
		"last_run_time":     time.Now().Unix(),
		"last_build_number": buildNumber,
		"last_build_url":    buildURL,
	})
}

// PipelineUpdateRunComplete 更新流水线运行完成
func (d *Dao) PipelineUpdateRunComplete(ctx context.Context, id int64, runStatus string) error {
	status := models.PipelineStatusIdle
	if runStatus == models.PipelineRunStatusRunning {
		status = models.PipelineStatusRunning
	}
	return d.PipelineUpdate(ctx, id, map[string]interface{}{
		"status":          status,
		"last_run_status": runStatus,
	})
}

// ==================== PipelineRun CRUD ====================

// PipelineRunCreate 创建流水线运行记录
func (d *Dao) PipelineRunCreate(ctx context.Context, run *models.CicdPipelineRun) error {
	now := time.Now().Unix()
	run.CreatedAt = uint64(now)
	run.ModifiedAt = uint64(now)
	return d.db.WithContext(ctx).Create(run).Error
}

// PipelineRunGetByID 根据ID获取运行记录
func (d *Dao) PipelineRunGetByID(ctx context.Context, id int64) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("id = ?", id).
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// PipelineRunGetLatest 获取流水线最近一次运行记录
func (d *Dao) PipelineRunGetLatest(ctx context.Context, pipelineID int64) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ?", pipelineID).
		Order("id DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// PipelineRunList 获取流水线运行历史
func (d *Dao) PipelineRunList(ctx context.Context, pipelineID int64, page, pageSize int) ([]*models.CicdPipelineRun, int64, error) {
	var list []*models.CicdPipelineRun
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdPipelineRun{}).Where("pipeline_id = ?", pipelineID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PipelineRunUpdate 更新运行记录
func (d *Dao) PipelineRunUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// PipelineRunUpdateStatus 更新运行状态
func (d *Dao) PipelineRunUpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.PipelineRunStatusRunning {
		updates["started_at"] = time.Now().Unix()
	}
	if status == models.PipelineRunStatusSuccess || status == models.PipelineRunStatusFailed || status == models.PipelineRunStatusAborted {
		updates["finished_at"] = time.Now().Unix()
	}
	return d.PipelineRunUpdate(ctx, id, updates)
}

// PipelineRunUpdateBuildNumber 更新构建号
func (d *Dao) PipelineRunUpdateBuildNumber(ctx context.Context, id int64, buildNumber int) error {
	return d.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"build_number": buildNumber,
		"status":       models.PipelineRunStatusRunning,
		"started_at":   time.Now().Unix(),
	})
}

// PipelineRunUpdateLog 更新控制台日志
func (d *Dao) PipelineRunUpdateLog(ctx context.Context, id int64, log string) error {
	return d.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"console_log": log,
	})
}
