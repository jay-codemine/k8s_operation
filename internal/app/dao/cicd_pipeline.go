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

// PipelineRunUpdateError 更新运行错误信息
func (d *Dao) PipelineRunUpdateError(ctx context.Context, id int64, status string, errMsg string) error {
	return d.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
		"finished_at":   time.Now().Unix(),
	})
}

// ==================== 回调 & 轮询支持 ====================

// PipelineGetByJenkinsJob 根据 Jenkins Job 名称查找流水线
func (d *Dao) PipelineGetByJenkinsJob(ctx context.Context, jobName string) (*models.CicdPipeline, error) {
	var p models.CicdPipeline
	err := d.db.WithContext(ctx).
		Where("jenkins_job = ? AND is_del = 0", jobName).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PipelineRunGetByBuildNumber 根据幂等键查找运行记录 (pipeline_id + build_number)
func (d *Dao) PipelineRunGetByBuildNumber(ctx context.Context, pipelineID int64, buildNumber int) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND build_number = ?", pipelineID, buildNumber).
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// PipelineRunUpdateCallback 回调更新运行记录
func (d *Dao) PipelineRunUpdateCallback(ctx context.Context, id int64, status, imageURL, imageDigest, errMsg string, duration int) error {
	updates := map[string]interface{}{
		"status":            status,
		"callback_received": 1,
		"finished_at":       time.Now().Unix(),
		"duration_sec":      duration,
	}
	if imageURL != "" {
		updates["image_url"] = imageURL
	}
	if imageDigest != "" {
		updates["image_digest"] = imageDigest
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	return d.PipelineRunUpdate(ctx, id, updates)
}

// PipelineRunListPendingForPoll 获取需要轮询的运行记录
// 条件：未终态(pending/running) && 回调未收到 && 未超时
func (d *Dao) PipelineRunListPendingForPoll(ctx context.Context, maxAgeMinutes int, limit int) ([]*models.CicdPipelineRun, error) {
	var list []*models.CicdPipelineRun
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())
	
	err := d.db.WithContext(ctx).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at > ?",
			models.PipelineRunStatusPending, models.PipelineRunStatusRunning, cutoffTime).
		Order("created_at ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PipelineRunMarkTimeout 批量标记超时的运行记录为失败
func (d *Dao) PipelineRunMarkTimeout(ctx context.Context, maxAgeMinutes int) (int64, error) {
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())
	
	result := d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at <= ?",
			models.PipelineRunStatusPending, models.PipelineRunStatusRunning, cutoffTime).
		Updates(map[string]interface{}{
			"status":        models.PipelineRunStatusFailed,
			"error_message": "构建超时（未收到回调）",
			"finished_at":   time.Now().Unix(),
			"modified_at":   time.Now().Unix(),
		})
	return result.RowsAffected, result.Error
}

// PipelineUpdateDeployInfo 更新流水线最新部署信息
func (d *Dao) PipelineUpdateDeployInfo(ctx context.Context, id int64, image, digest string, deployTime uint64, status string) error {
	return d.PipelineUpdate(ctx, id, map[string]interface{}{
		"last_deploy_image":  image,
		"last_deploy_digest": digest,
		"last_deploy_time":   deployTime,
		"last_deploy_status": status,
	})
}

// ==================== Environment CRUD ====================

// EnvironmentList 获取环境列表（支持分页和关键字搜索）
func (d *Dao) EnvironmentList(ctx context.Context, page, pageSize int, keyword string) ([]*models.EnvironmentListItem, int64, error) {
	var list []*models.EnvironmentListItem
	var total int64

	query := d.db.WithContext(ctx).
		Table("cicd_environment AS e").
		Select("e.*, c.name AS cluster_name").
		Joins("LEFT JOIN k8s_cluster c ON c.id = e.cluster_id").
		Where("e.is_del = 0")

	if keyword != "" {
		query = query.Where("e.name LIKE ? OR e.display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("e.sort_order ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// EnvironmentGetByID 根据ID获取环境
func (d *Dao) EnvironmentGetByID(ctx context.Context, id int64) (*models.CicdEnvironment, error) {
	var env models.CicdEnvironment
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&env).Error
	return &env, err
}

// EnvironmentGetByName 根据名称获取环境
func (d *Dao) EnvironmentGetByName(ctx context.Context, name string) (*models.CicdEnvironment, error) {
	var env models.CicdEnvironment
	err := d.db.WithContext(ctx).
		Where("name = ? AND is_del = 0", name).
		First(&env).Error
	return &env, err
}

// EnvironmentCreate 创建环境
func (d *Dao) EnvironmentCreate(ctx context.Context, env *models.CicdEnvironment) (int64, error) {
	now := time.Now().Unix()
	env.CreatedAt = uint64(now)
	env.ModifiedAt = uint64(now)
	err := d.db.WithContext(ctx).Create(env).Error
	return env.ID, err
}

// EnvironmentUpdate 更新环境
func (d *Dao) EnvironmentUpdate(ctx context.Context, env *models.CicdEnvironment) error {
	env.ModifiedAt = uint64(time.Now().Unix())
	return d.db.WithContext(ctx).
		Model(&models.CicdEnvironment{}).
		Where("id = ? AND is_del = 0", env.ID).
		Save(env).Error
}

// EnvironmentDelete 软删除环境
func (d *Dao) EnvironmentDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdEnvironment{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// ==================== Approval CRUD ====================

// ApprovalCreate 创建审批记录
func (d *Dao) ApprovalCreate(ctx context.Context, approval *models.CicdApproval) (int64, error) {
	now := time.Now().Unix()
	approval.CreatedAt = uint64(now)
	approval.ModifiedAt = uint64(now)
	err := d.db.WithContext(ctx).Create(approval).Error
	return approval.ID, err
}

// ApprovalGetByID 根据ID获取审批记录
func (d *Dao) ApprovalGetByID(ctx context.Context, id int64) (*models.CicdApproval, error) {
	var approval models.CicdApproval
	err := d.db.WithContext(ctx).
		Where("id = ?", id).
		First(&approval).Error
	return &approval, err
}

// ApprovalList 获取审批列表
func (d *Dao) ApprovalList(ctx context.Context, page, pageSize int, status string) ([]*models.CicdApproval, int64, error) {
	var list []*models.CicdApproval
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdApproval{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ApprovalUpdateStatus 更新审批状态
func (d *Dao) ApprovalUpdateStatus(ctx context.Context, id int64, status string, approveUserID int64, reason string) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdApproval{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          status,
			"approve_user_id": approveUserID,
			"approve_reason":  reason,
			"approve_time":    now,
			"modified_at":     now,
		}).Error
}

// ApprovalGetPendingByPipeline 获取流水线待审批记录
func (d *Dao) ApprovalGetPendingByPipeline(ctx context.Context, pipelineID int64) (*models.CicdApproval, error) {
	var approval models.CicdApproval
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND status = ?", pipelineID, models.ApprovalStatusPending).
		Order("id DESC").
		First(&approval).Error
	return &approval, err
}

// ==================== Pipeline Stage CRUD ====================

// StageCreate 创建阶段执行记录
func (d *Dao) StageCreate(ctx context.Context, stage *models.CicdPipelineStage) error {
	now := time.Now().Unix()
	stage.CreatedAt = uint64(now)
	stage.ModifiedAt = uint64(now)
	return d.db.WithContext(ctx).Create(stage).Error
}

// StageCreateBatch 批量创建阶段执行记录
func (d *Dao) StageCreateBatch(ctx context.Context, stages []*models.CicdPipelineStage) error {
	if len(stages) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, stage := range stages {
		stage.CreatedAt = uint64(now)
		stage.ModifiedAt = uint64(now)
	}
	return d.db.WithContext(ctx).Create(&stages).Error
}

// StageGetByID 根据ID获取阶段记录
func (d *Dao) StageGetByID(ctx context.Context, id int64) (*models.CicdPipelineStage, error) {
	var stage models.CicdPipelineStage
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&stage).Error
	return &stage, err
}

// StageListByRunID 获取运行记录的所有阶段
func (d *Dao) StageListByRunID(ctx context.Context, runID int64) ([]*models.CicdPipelineStage, error) {
	var list []*models.CicdPipelineStage
	err := d.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("stage_order ASC").
		Find(&list).Error
	return list, err
}

// StageUpdate 更新阶段记录
func (d *Dao) StageUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdPipelineStage{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// StageUpdateStatus 更新阶段状态
func (d *Dao) StageUpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	// 根据状态设置开始/结束时间
	if status == models.StageStatusRunning {
		updates["started_at"] = time.Now().Unix()
	}
	if status == models.StageStatusSuccess || status == models.StageStatusFailed || 
	   status == models.StageStatusSkipped || status == models.StageStatusAborted {
		now := time.Now().Unix()
		updates["finished_at"] = now
	}
	return d.StageUpdate(ctx, id, updates)
}

// StageUpdateWithDuration 更新阶段状态和时长
func (d *Dao) StageUpdateWithDuration(ctx context.Context, id int64, status string, duration int) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":       status,
		"finished_at":  now,
		"duration_sec": duration,
	}
	return d.StageUpdate(ctx, id, updates)
}

// StageUpdateLogs 更新阶段日志
func (d *Dao) StageUpdateLogs(ctx context.Context, id int64, logs string) error {
	return d.StageUpdate(ctx, id, map[string]interface{}{
		"logs": logs,
	})
}

// StageUpdateApproval 更新阶段审批信息
func (d *Dao) StageUpdateApproval(ctx context.Context, id int64, userID int64, decision, comment string) error {
	now := time.Now().Unix()
	status := models.StageStatusSuccess
	if decision == "rejected" {
		status = models.StageStatusFailed
	}
	return d.StageUpdate(ctx, id, map[string]interface{}{
		"status":            status,
		"finished_at":       now,
		"approval_user_id":  userID,
		"approval_decision": decision,
		"approval_comment":  comment,
	})
}

// StageUpdateDeploy 更新阶段部署信息
func (d *Dao) StageUpdateDeploy(ctx context.Context, id int64, clusterID int64, namespace, workloadKind, workloadName, container, image string, replicas int) error {
	return d.StageUpdate(ctx, id, map[string]interface{}{
		"deploy_cluster_id":    clusterID,
		"deploy_namespace":     namespace,
		"deploy_workload_kind": workloadKind,
		"deploy_workload_name": workloadName,
		"deploy_container":     container,
		"deploy_image":         image,
		"deploy_replicas":      replicas,
	})
}

// StageUpdateError 更新阶段错误信息
func (d *Dao) StageUpdateError(ctx context.Context, id int64, errMsg string) error {
	return d.StageUpdate(ctx, id, map[string]interface{}{
		"status":        models.StageStatusFailed,
		"finished_at":   time.Now().Unix(),
		"error_message": errMsg,
	})
}

// StageGetByRunIDAndType 获取指定运行记录的指定类型阶段
func (d *Dao) StageGetByRunIDAndType(ctx context.Context, runID int64, stageType string) (*models.CicdPipelineStage, error) {
	var stage models.CicdPipelineStage
	err := d.db.WithContext(ctx).
		Where("run_id = ? AND stage_type = ?", runID, stageType).
		First(&stage).Error
	return &stage, err
}

// StageGetLogs 获取阶段日志
func (d *Dao) StageGetLogs(ctx context.Context, id int64) (string, error) {
	var stage models.CicdPipelineStage
	err := d.db.WithContext(ctx).
		Select("logs").
		Where("id = ?", id).
		First(&stage).Error
	return stage.Logs, err
}

// StageDeleteByRunID 删除运行记录的所有阶段
func (d *Dao) StageDeleteByRunID(ctx context.Context, runID int64) error {
	return d.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Delete(&models.CicdPipelineStage{}).Error
}
