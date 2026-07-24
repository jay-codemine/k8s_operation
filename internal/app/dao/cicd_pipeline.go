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

// PipelineGetByGitRepoBranch 根据仓库+分支查找已有流水线（用于冲突警告）
// excludeID != 0 时排除该 ID（编辑模式下排除自身）
func (d *Dao) PipelineGetByGitRepoBranch(ctx context.Context, gitRepo, gitBranch string, excludeID int64) ([]*models.CicdPipeline, error) {
	var list []*models.CicdPipeline
	query := d.db.WithContext(ctx).
		Where("git_repo = ? AND git_branch = ? AND is_del = 0", gitRepo, gitBranch)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Select("id, name").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PipelineGetByWorkload 根据自动部署目标查找已有流水线（用于冲突警告）
// excludeID != 0 时排除该 ID
func (d *Dao) PipelineGetByWorkload(ctx context.Context, clusterID int64, namespace, workloadName string, excludeID int64) ([]*models.CicdPipeline, error) {
	var list []*models.CicdPipeline
	query := d.db.WithContext(ctx).
		Where("target_cluster_id = ? AND target_namespace = ? AND target_workload_name = ? AND auto_deploy = 1 AND is_del = 0",
			clusterID, namespace, workloadName)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Select("id, name").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PipelineList 获取流水线列表
// PipelineListFilter 流水线列表筛选条件（发布中心高级筛选）
type PipelineListFilter struct {
	Keyword   string
	Status    string
	Language  string // language_type
	DeployEnv string // dev/test/staging/prod，兼容旧数据同时匹配 target_namespace
	EnvironmentID int64 // 关联环境ID（>0 精确过滤，优先于 DeployEnv 模糊匹配）
	CreatorID int64  // created_user_id
	StartTime int64  // created_at >= （unix 秒）
	EndTime   int64  // created_at <= （unix 秒）
	Page      int
	PageSize  int
}

func (d *Dao) PipelineList(ctx context.Context, f PipelineListFilter) ([]*models.CicdPipeline, int64, error) {
	var list []*models.CicdPipeline
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdPipeline{}).Where("is_del = 0")

	// 关键字搜索（名称、描述、Git仓库）
	if f.Keyword != "" {
		likeKeyword := "%" + f.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR git_repo LIKE ?", likeKeyword, likeKeyword, likeKeyword)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.Language != "" {
		query = query.Where("language_type = ?", f.Language)
	}
	if f.EnvironmentID > 0 {
		query = query.Where("environment_id = ?", f.EnvironmentID)
	} else if f.DeployEnv != "" {
		// 兼容旧数据：deploy_env 为空时按 target_namespace 模糊匹配环境关键字
		query = query.Where("deploy_env = ? OR (COALESCE(deploy_env, '') = '' AND target_namespace LIKE ?)", f.DeployEnv, "%"+f.DeployEnv+"%")
	}
	if f.CreatorID > 0 {
		query = query.Where("created_user_id = ?", f.CreatorID)
	}
	if f.StartTime > 0 {
		query = query.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime > 0 {
		query = query.Where("created_at <= ?", f.EndTime)
	}

	// 先查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 再查询分页数据
	offset := (f.Page - 1) * f.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&list).Error; err != nil {
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

// PipelineRunGetRunning 获取流水线正在运行的记录
// 过滤 build_number=0 的幽灵记录（创建了但从未实际触发 Jenkins 构建）
func (d *Dao) PipelineRunGetRunning(ctx context.Context, pipelineID int64) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND status IN (?, ?) AND build_number > 0", pipelineID, models.PipelineRunStatusPending, models.PipelineRunStatusRunning).
		Order("id DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// PipelineRunGetLatestBuilt 获取流水线最近一次已构建的运行记录（build_number > 0）
// 跳过创建了但从未实际触发 Jenkins 构建的幽灵记录
func (d *Dao) PipelineRunGetLatestBuilt(ctx context.Context, pipelineID int64) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND build_number > 0", pipelineID).
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

// PipelineRunListAll 获取全量构建记录（跨流水线）支持分页、筛选、搜索
func (d *Dao) PipelineRunListAll(ctx context.Context, page, pageSize int, status, keyword string, pipelineID int64) ([]*models.CicdPipelineRun, int64, error) {
	var list []*models.CicdPipelineRun
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdPipelineRun{})

	if pipelineID > 0 {
		query = query.Where("pipeline_id = ?", pipelineID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("(image_url LIKE ? OR error_message LIKE ? OR trigger_user LIKE ?)", kw, kw, kw)
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
// 注意：同一 pipeline_id + build_number 可能存在多条记录（Jenkins 重用构建号），
// 必须按 id DESC 取最新记录，否则会命中旧记录导致回调被误判为重复
func (d *Dao) PipelineRunGetByBuildNumber(ctx context.Context, pipelineID int64, buildNumber int) (*models.CicdPipelineRun, error) {
	var run models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("pipeline_id = ? AND build_number = ?", pipelineID, buildNumber).
		Order("id DESC").
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

// PipelineRunListTimedOut 查询即将超时的运行记录（用于同步更新 pipeline 表状态）
func (d *Dao) PipelineRunListTimedOut(ctx context.Context, maxAgeMinutes int) ([]*models.CicdPipelineRun, error) {
	var list []*models.CicdPipelineRun
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())

	err := d.db.WithContext(ctx).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at <= ?",
			models.PipelineRunStatusPending, models.PipelineRunStatusRunning, cutoffTime).
		Find(&list).Error
	return list, err
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
// version 为 K8s 集群中的当前版本号（Deployment 的 ReplicaSet revision / StatefulSet/DaemonSet 的 ControllerRevision revision）
func (d *Dao) PipelineUpdateDeployInfo(ctx context.Context, id int64, image, digest string, deployTime uint64, status, version string) error {
	return d.PipelineUpdate(ctx, id, map[string]interface{}{
		"last_deploy_image":   image,
		"last_deploy_digest":  digest,
		"last_deploy_time":    deployTime,
		"last_deploy_status":  status,
		"last_deploy_version": version,
	})
}

// PipelineListStuckRunning 查询卡在 running 状态的流水线（孤儿 pipeline）
// 条件：pipeline.status = 'running' 且没有活跃的 run 记录（pending/running）
func (d *Dao) PipelineListStuckRunning(ctx context.Context) ([]*models.CicdPipeline, error) {
	var list []*models.CicdPipeline
	err := d.db.WithContext(ctx).
		Where("status = ? AND is_del = 0", models.PipelineStatusRunning).
		Find(&list).Error
	return list, err
}

// PipelineRunListCompletedUnsynced 获取已完成但未同步到发布记录的运行记录
func (d *Dao) PipelineRunListCompletedUnsynced(ctx context.Context, limit int) ([]*models.CicdPipelineRun, error) {
	var list []*models.CicdPipelineRun
	err := d.db.WithContext(ctx).
		Where("status IN (?, ?) AND build_number > 0 AND id NOT IN (SELECT build_id FROM cicd_release WHERE is_del = 0 AND build_id > 0)",
			models.PipelineRunStatusSuccess, models.PipelineRunStatusFailed).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ==================== Environment CRUD ====================

// EnvironmentList 获取环境列表（支持分页和关键字搜索）
func (d *Dao) EnvironmentList(ctx context.Context, page, pageSize int, keyword string) ([]*models.EnvironmentListItem, int64, error) {
	var list []*models.EnvironmentListItem
	var total int64

	query := d.db.WithContext(ctx).
		Table("cicd_environment AS e").
		Select("e.*, c.cluster_name AS cluster_name").
		Joins("LEFT JOIN kube_cluster c ON c.id = e.cluster_id").
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

// EnvironmentGetByNamespace 根据命名空间查找环境配置
func (d *Dao) EnvironmentGetByNamespace(ctx context.Context, namespace string) (*models.CicdEnvironment, error) {
	var env models.CicdEnvironment
	err := d.db.WithContext(ctx).
		Where("namespace = ? AND is_del = 0", namespace).
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

// ApprovalList 获取审批列表（含流水线名称和用户名）
func (d *Dao) ApprovalList(ctx context.Context, page, pageSize int, status string, pipelineID int64) ([]*models.ApprovalListItem, int64, error) {
	var list []*models.ApprovalListItem
	var total int64

	query := d.db.WithContext(ctx).
		Table("cicd_approval AS a").
		Select(`a.*, 
			COALESCE(p.name, '') AS pipeline_name,
			COALESCE(u1.username, '') AS request_username,
			COALESCE(u2.username, '') AS approve_username`).
		Joins("LEFT JOIN cicd_pipeline AS p ON a.pipeline_id = p.id").
		Joins("LEFT JOIN `user` AS u1 ON a.request_user_id = u1.id").
		Joins("LEFT JOIN `user` AS u2 ON a.approve_user_id = u2.id")

	if status != "" {
		query = query.Where("a.status = ?", status)
	}
	if pipelineID > 0 {
		query = query.Where("a.pipeline_id = ?", pipelineID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("a.id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ApprovalListByUser 获取指定用户提交的审批列表（开发者只看自己的）
func (d *Dao) ApprovalListByUser(ctx context.Context, userID int64, page, pageSize int, status string, pipelineID int64) ([]*models.ApprovalListItem, int64, error) {
	var list []*models.ApprovalListItem
	var total int64

	query := d.db.WithContext(ctx).
		Table("cicd_approval AS a").
		Select(`a.*, 
			COALESCE(p.name, '') AS pipeline_name,
			COALESCE(u1.username, '') AS request_username,
			COALESCE(u2.username, '') AS approve_username`).
		Joins("LEFT JOIN cicd_pipeline AS p ON a.pipeline_id = p.id").
		Joins("LEFT JOIN `user` AS u1 ON a.request_user_id = u1.id").
		Joins("LEFT JOIN `user` AS u2 ON a.approve_user_id = u2.id").
		Where("a.request_user_id = ?", userID)

	if status != "" {
		query = query.Where("a.status = ?", status)
	}
	if pipelineID > 0 {
		query = query.Where("a.pipeline_id = ?", pipelineID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("a.id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ApprovalStats 获取各状态审批数量统计
func (d *Dao) ApprovalStats(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []StatusCount
	err := d.db.WithContext(ctx).
		Model(&models.CicdApproval{}).
		Select("status, COUNT(*) AS cnt").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"pending":  0,
		"approved": 0,
		"rejected": 0,
		"expired":  0,
		"total":    0,
	}
	for _, r := range rows {
		stats[r.Status] = r.Count
		stats["total"] += r.Count
	}
	return stats, nil
}

// ApprovalStatsByUser 获取指定用户的审批统计
func (d *Dao) ApprovalStatsByUser(ctx context.Context, userID int64) (map[string]int64, error) {
	type StatusCount struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []StatusCount
	err := d.db.WithContext(ctx).
		Model(&models.CicdApproval{}).
		Where("request_user_id = ?", userID).
		Select("status, COUNT(*) AS cnt").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"pending":  0,
		"approved": 0,
		"rejected": 0,
		"expired":  0,
		"total":    0,
	}
	for _, r := range rows {
		stats[r.Status] = r.Count
		stats["total"] += r.Count
	}
	return stats, nil
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

// ApprovalUpdate 更新审批记录字段
func (d *Dao) ApprovalUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdApproval{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ApprovalDelete 删除审批记录
func (d *Dao) ApprovalDelete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.CicdApproval{}).Error
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

// ApprovalGetByStageID 根据StageID获取审批记录
func (d *Dao) ApprovalGetByStageID(ctx context.Context, stageID int64) (*models.CicdApproval, error) {
	var approval models.CicdApproval
	err := d.db.WithContext(ctx).
		Where("stage_id = ? AND status = ?", stageID, models.ApprovalStatusPending).
		First(&approval).Error
	return &approval, err
}

// ApprovalGetByFeishuToken 根据飞书Token获取审批记录
func (d *Dao) ApprovalGetByFeishuToken(ctx context.Context, token string) (*models.CicdApproval, error) {
	var approval models.CicdApproval
	err := d.db.WithContext(ctx).
		Where("feishu_token = ?", token).
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

// StageListApprovalAll 获取所有审批类型的阶段记录（用于数据补全）
func (d *Dao) StageListApprovalAll(ctx context.Context) ([]*models.CicdPipelineStage, error) {
	var list []*models.CicdPipelineStage
	err := d.db.WithContext(ctx).
		Where("stage_type = ?", models.StageTypeApproval).
		Find(&list).Error
	return list, err
}

// ApprovalExistsByStageID 检查指定 stage_id 是否已有审批记录
func (d *Dao) ApprovalExistsByStageID(ctx context.Context, stageID int64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&models.CicdApproval{}).
		Where("stage_id = ?", stageID).
		Count(&count).Error
	return count > 0, err
}

// ==================== 构建统计 ====================

// BuildStatsResult 构建统计结果
type BuildStatsResult struct {
	TotalBuilds    int64   `json:"total_builds"`
	SuccessBuilds  int64   `json:"success_builds"`
	FailedBuilds   int64   `json:"failed_builds"`
	RunningBuilds  int64   `json:"running_builds"`
	AvgDuration    float64 `json:"avg_duration"`
	SuccessRate    float64 `json:"success_rate"`
	TodayBuilds    int64   `json:"today_builds"`
	WeekBuilds     int64   `json:"week_builds"`
}

// PipelineRunBuildStats 获取构建统计数据
func (d *Dao) PipelineRunBuildStats(ctx context.Context) (*BuildStatsResult, error) {
	stats := &BuildStatsResult{}

	// 总数 & 各状态统计
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Select("status, COUNT(*) AS cnt").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	// terminalBuilds 只统计已完结（成功/失败）的构建，作为成功率分母；
	// pending/running 未完结、aborted 为人工中止，均不应拉低成功率（口径与 CicdReleaseStatsEnhanced 一致）。
	var terminalBuilds int64
	for _, r := range rows {
		stats.TotalBuilds += r.Cnt
		switch r.Status {
		case models.PipelineRunStatusSuccess:
			stats.SuccessBuilds = r.Cnt
			terminalBuilds += r.Cnt
		case models.PipelineRunStatusFailed:
			stats.FailedBuilds = r.Cnt
			terminalBuilds += r.Cnt
		case models.PipelineRunStatusRunning, models.PipelineRunStatusPending:
			stats.RunningBuilds += r.Cnt
		}
	}
	if terminalBuilds > 0 {
		stats.SuccessRate = float64(stats.SuccessBuilds) / float64(terminalBuilds) * 100
	}

	// 平均构建时长（仅计算已完成且 duration > 0 的记录）
	var avgDur struct {
		Avg float64 `gorm:"column:avg_dur"`
	}
	d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Select("AVG(duration_sec) AS avg_dur").
		Where("status IN (?, ?) AND duration_sec > 0", models.PipelineRunStatusSuccess, models.PipelineRunStatusFailed).
		Scan(&avgDur)
	stats.AvgDuration = avgDur.Avg

	// 今日构建数
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Where("created_at >= ?", todayStart.Unix()).
		Count(&stats.TodayBuilds)

	// 本周构建数
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Where("created_at >= ?", weekStart.Unix()).
		Count(&stats.WeekBuilds)

	return stats, nil
}

// BuildTrendItem 构建趋势项
type BuildTrendItem struct {
	Date    string `json:"date" gorm:"column:date"`
	Success int64  `json:"success" gorm:"column:success"`
	Failed  int64  `json:"failed" gorm:"column:failed"`
	Total   int64  `json:"total" gorm:"column:total"`
}

// PipelineRunBuildTrend 获取最近N天的构建趋势
func (d *Dao) PipelineRunBuildTrend(ctx context.Context, days int) ([]BuildTrendItem, error) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days+1)

	var results []BuildTrendItem
	err := d.db.WithContext(ctx).
		Model(&models.CicdPipelineRun{}).
		Select(`FROM_UNIXTIME(created_at, '%Y-%m-%d') AS date,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS success,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed,
			COUNT(*) AS total`).
		Where("created_at >= ?", startTime.Unix()).
		Group("date").
		Order("date ASC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
