package cicd

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/events"
)

// CicdService CICD 领域服务
type CicdService struct {
	db        *gorm.DB // 保留用于复杂查询/聚合/跨域查询，逐步迁移到 repo
	repo      CicdRepository
	publisher events.EventPublisher
	tenantID  uint32 // 缓存的租户 ID，避免依赖 ScopedDB 上下文
}

// SetTenantID 注入租户 ID（由 Services 层调用）
func (s *CicdService) SetTenantID(tid uint32) { s.tenantID = tid }

// NewCicdService 创建 CICD 服务
func NewCicdService(db *gorm.DB, repo CicdRepository, publisher events.EventPublisher) *CicdService {
	return &CicdService{db: db, repo: repo, publisher: publisher}
}


// ====== 查询/统计辅助类型 ======

// PipelineListFilter 流水线列表筛选条件
type PipelineListFilter struct {
	Keyword       string
	Status        string
	Language      string
	DeployEnv     string
	EnvironmentID int64
	CreatorID     int64
	StartTime     int64
	EndTime       int64
	Page          int
	PageSize      int
}

// BuildStatsResult 构建统计结果
type BuildStatsResult struct {
	TotalBuilds   int64   `json:"total_builds"`
	SuccessBuilds int64   `json:"success_builds"`
	FailedBuilds  int64   `json:"failed_builds"`
	RunningBuilds int64   `json:"running_builds"`
	AvgDuration   float64 `json:"avg_duration"`
	SuccessRate   float64 `json:"success_rate"`
	TodayBuilds   int64   `json:"today_builds"`
	WeekBuilds    int64   `json:"week_builds"`
}

// BuildTrendItem 构建趋势项
type BuildTrendItem struct {
	Date    string `json:"date" gorm:"column:date"`
	Success int64  `json:"success" gorm:"column:success"`
	Failed  int64  `json:"failed" gorm:"column:failed"`
	Total   int64  `json:"total" gorm:"column:total"`
}

// ReleaseStatsEnhanced 发布统计增强
type ReleaseStatsEnhanced struct {
	Total       int64            `json:"total"`
	ByStatus    map[string]int64 `json:"by_status"`
	TodayCount  int64            `json:"today_count"`
	WeekCount   int64            `json:"week_count"`
	SuccessRate float64          `json:"success_rate"`
}

// CicdTaskStats 任务统计
type CicdTaskStats struct {
	Pending   int64
	Queued    int64
	Running   int64
	Succeeded int64
	Failed    int64
	Canceled  int64
}

// WithTx 事务包装
func (s *CicdService) WithTx(ctx context.Context, fn func(tx *CicdService) error) error {
	dbWithCtx := s.db.WithContext(ctx)
	wrapper := func(gormTx *gorm.DB) error {
		svcTx := &CicdService{db: gormTx, publisher: s.publisher, tenantID: s.tenantID}
		return fn(svcTx)
	}
	return dbWithCtx.Transaction(wrapper)
}
func (s *CicdService) PipelineCreate(ctx context.Context, p *CicdPipeline) error {
	now := time.Now().Unix()
	p.CreatedAt = uint64(now)
	p.ModifiedAt = uint64(now)
	return s.repo.PipelineSave(ctx, p)
}
func (s *CicdService) PipelineGetByID(ctx context.Context, id int64) (*CicdPipeline, error) {
	return s.repo.PipelineFindByID(ctx, id)
}
func (s *CicdService) PipelineGetByName(ctx context.Context, name string) (*CicdPipeline, error) {
	return s.repo.PipelineFindByName(ctx, name)
}
func (s *CicdService) PipelineGetByGitRepoBranch(ctx context.Context, gitRepo, gitBranch string, excludeID int64) ([]*CicdPipeline, error) {
	var list []*CicdPipeline
	query := s.db.WithContext(ctx).
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
func (s *CicdService) PipelineGetByWorkload(ctx context.Context, clusterID int64, namespace, workloadName string, excludeID int64) ([]*CicdPipeline, error) {
	var list []*CicdPipeline
	query := s.db.WithContext(ctx).
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
func (s *CicdService) PipelineList(ctx context.Context, f PipelineListFilter) ([]*CicdPipeline, int64, error) {
	var list []*CicdPipeline
	var total int64

	query := s.db.WithContext(ctx).Model(&CicdPipeline{}).Where("is_del = 0")

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
func (s *CicdService) PipelineUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.repo.PipelineUpdate(ctx, id, updates)
}
func (s *CicdService) PipelineDelete(ctx context.Context, id int64) error {
	return s.repo.PipelineDelete(ctx, id)
}
func (s *CicdService) PipelineUpdateStatus(ctx context.Context, id int64, status string) error {
	return s.PipelineUpdate(ctx, id, map[string]interface{}{
		"status": status,
	})
}
func (s *CicdService) PipelineUpdateRunInfo(ctx context.Context, id int64, runStatus string, buildNumber int, buildURL string) error {
	return s.PipelineUpdate(ctx, id, map[string]interface{}{
		"status":            PipelineStatusRunning,
		"last_run_status":   runStatus,
		"last_run_time":     time.Now().Unix(),
		"last_build_number": buildNumber,
		"last_build_url":    buildURL,
	})
}
func (s *CicdService) PipelineUpdateRunComplete(ctx context.Context, id int64, runStatus string) error {
	status := PipelineStatusIdle
	if runStatus == PipelineRunStatusRunning {
		status = PipelineStatusRunning
	}
	return s.PipelineUpdate(ctx, id, map[string]interface{}{
		"status":          status,
		"last_run_status": runStatus,
	})
}
func (s *CicdService) PipelineRunCreate(ctx context.Context, run *CicdPipelineRun) error {
	now := time.Now().Unix()
	run.CreatedAt = uint64(now)
	run.ModifiedAt = uint64(now)
	if err := s.repo.PipelineRunSave(ctx, run); err != nil {
		return err
	}
	s.publishEvent(NewPipelineTriggered(run.PipelineID, run.ID, ""))
	return nil
}

func (s *CicdService) publishEvent(event events.DomainEvent) {
	if s.publisher == nil {
		return
	}
	defer func() { _ = recover() }()
	s.publisher.Publish(event)
}
func (s *CicdService) PipelineRunGetByID(ctx context.Context, id int64) (*CicdPipelineRun, error) {
	return s.repo.PipelineRunFindByID(ctx, id)
}
func (s *CicdService) PipelineRunGetLatest(ctx context.Context, pipelineID int64) (*CicdPipelineRun, error) {
	return s.repo.PipelineRunFindLatest(ctx, pipelineID)
}
func (s *CicdService) PipelineRunGetRunning(ctx context.Context, pipelineID int64) (*CicdPipelineRun, error) {
	var run CicdPipelineRun
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND status IN (?, ?) AND build_number > 0", pipelineID, PipelineRunStatusPending, PipelineRunStatusRunning).
		Order("id DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}
func (s *CicdService) PipelineRunGetLatestBuilt(ctx context.Context, pipelineID int64) (*CicdPipelineRun, error) {
	var run CicdPipelineRun
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND build_number > 0", pipelineID).
		Order("id DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}
func (s *CicdService) PipelineRunList(ctx context.Context, pipelineID int64, page, pageSize int) ([]*CicdPipelineRun, int64, error) {
	var list []*CicdPipelineRun
	var total int64

	query := s.db.WithContext(ctx).Model(&CicdPipelineRun{}).Where("pipeline_id = ?", pipelineID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
func (s *CicdService) PipelineRunListAll(ctx context.Context, page, pageSize int, status, keyword string, pipelineID int64) ([]*CicdPipelineRun, int64, error) {
	var list []*CicdPipelineRun
	var total int64

	query := s.db.WithContext(ctx).Model(&CicdPipelineRun{})

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
func (s *CicdService) PipelineRunUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.repo.PipelineRunUpdate(ctx, id, updates)
}
func (s *CicdService) PipelineRunUpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == PipelineRunStatusRunning {
		updates["started_at"] = time.Now().Unix()
	}
	if status == PipelineRunStatusSuccess || status == PipelineRunStatusFailed || status == PipelineRunStatusAborted {
		updates["finished_at"] = time.Now().Unix()
	}
	return s.PipelineRunUpdate(ctx, id, updates)
}
func (s *CicdService) PipelineRunUpdateBuildNumber(ctx context.Context, id int64, buildNumber int) error {
	return s.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"build_number": buildNumber,
		"status":       PipelineRunStatusRunning,
		"started_at":   time.Now().Unix(),
	})
}
func (s *CicdService) PipelineRunUpdateLog(ctx context.Context, id int64, log string) error {
	return s.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"console_log": log,
	})
}
func (s *CicdService) PipelineRunUpdateError(ctx context.Context, id int64, status string, errMsg string) error {
	return s.PipelineRunUpdate(ctx, id, map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
		"finished_at":   time.Now().Unix(),
	})
}
func (s *CicdService) PipelineGetByJenkinsJob(ctx context.Context, jobName string) (*CicdPipeline, error) {
	var p CicdPipeline
	err := s.db.WithContext(ctx).
		Where("jenkins_job = ? AND is_del = 0", jobName).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (s *CicdService) PipelineRunGetByBuildNumber(ctx context.Context, pipelineID int64, buildNumber int) (*CicdPipelineRun, error) {
	var run CicdPipelineRun
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND build_number = ?", pipelineID, buildNumber).
		Order("id DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}
func (s *CicdService) PipelineRunUpdateCallback(ctx context.Context, id int64, status, imageURL, imageDigest, errMsg string, duration int) error {
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
	return s.PipelineRunUpdate(ctx, id, updates)
}
func (s *CicdService) PipelineRunListPendingForPoll(ctx context.Context, maxAgeMinutes int, limit int) ([]*CicdPipelineRun, error) {
	var list []*CicdPipelineRun
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())
	
	err := s.db.WithContext(ctx).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at > ?",
			PipelineRunStatusPending, PipelineRunStatusRunning, cutoffTime).
		Order("created_at ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
func (s *CicdService) PipelineRunListTimedOut(ctx context.Context, maxAgeMinutes int) ([]*CicdPipelineRun, error) {
	var list []*CicdPipelineRun
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())

	err := s.db.WithContext(ctx).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at <= ?",
			PipelineRunStatusPending, PipelineRunStatusRunning, cutoffTime).
		Find(&list).Error
	return list, err
}
func (s *CicdService) PipelineRunMarkTimeout(ctx context.Context, maxAgeMinutes int) (int64, error) {
	cutoffTime := uint64(time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute).Unix())
	
	result := s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
		Where("status IN (?, ?) AND callback_received = 0 AND created_at <= ?",
			PipelineRunStatusPending, PipelineRunStatusRunning, cutoffTime).
		Updates(map[string]interface{}{
			"status":        PipelineRunStatusFailed,
			"error_message": "构建超时（未收到回调）",
			"finished_at":   time.Now().Unix(),
			"modified_at":   time.Now().Unix(),
		})
	return result.RowsAffected, result.Error
}
func (s *CicdService) PipelineUpdateDeployInfo(ctx context.Context, id int64, image, digest string, deployTime uint64, status, version string) error {
	return s.PipelineUpdate(ctx, id, map[string]interface{}{
		"last_deploy_image":   image,
		"last_deploy_digest":  digest,
		"last_deploy_time":    deployTime,
		"last_deploy_status":  status,
		"last_deploy_version": version,
	})
}
func (s *CicdService) PipelineListStuckRunning(ctx context.Context) ([]*CicdPipeline, error) {
	var list []*CicdPipeline
	err := s.db.WithContext(ctx).
		Where("status = ? AND is_del = 0", PipelineStatusRunning).
		Find(&list).Error
	return list, err
}
func (s *CicdService) PipelineRunListCompletedUnsynced(ctx context.Context, limit int) ([]*CicdPipelineRun, error) {
	var list []*CicdPipelineRun
	err := s.db.WithContext(ctx).
		Where("status IN (?, ?) AND build_number > 0 AND id NOT IN (SELECT build_id FROM cicd_release WHERE is_del = 0 AND build_id > 0)",
			PipelineRunStatusSuccess, PipelineRunStatusFailed).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
func (s *CicdService) EnvironmentList(ctx context.Context, page, pageSize int, keyword string) ([]*EnvironmentListItem, int64, error) {
	var list []*EnvironmentListItem
	var total int64

	tid, ok := s.tenantID, s.tenantID != 0
	if !ok {
		return nil, 0, nil
	}

	query := s.db.WithContext(ctx).
		Table("cicd_environment AS e").
		Select("e.*, c.cluster_name AS cluster_name").
		Joins("LEFT JOIN kube_cluster c ON c.id = e.cluster_id AND c.tenant_id = ? AND c.is_del = 0", tid).
		Where("e.is_del = 0 AND e.tenant_id = ?", tid)

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
func (s *CicdService) EnvironmentGetByID(ctx context.Context, id int64) (*CicdEnvironment, error) {
	return s.repo.EnvironmentFindByID(ctx, id)
}
func (s *CicdService) EnvironmentGetByName(ctx context.Context, name string) (*CicdEnvironment, error) {
	return s.repo.EnvironmentFindByName(ctx, name)
}
func (s *CicdService) EnvironmentGetByNamespace(ctx context.Context, namespace string) (*CicdEnvironment, error) {
	var env CicdEnvironment
	err := s.db.WithContext(ctx).
		Where("namespace = ? AND is_del = 0", namespace).
		First(&env).Error
	return &env, err
}
func (s *CicdService) EnvironmentCreate(ctx context.Context, env *CicdEnvironment) (int64, error) {
	now := time.Now().Unix()
	env.CreatedAt = uint64(now)
	env.ModifiedAt = uint64(now)
	return s.repo.EnvironmentSave(ctx, env)
}
func (s *CicdService) EnvironmentUpdate(ctx context.Context, env *CicdEnvironment) error {
	env.ModifiedAt = uint64(time.Now().Unix())
	return s.repo.EnvironmentUpdate(ctx, env)
}
func (s *CicdService) EnvironmentDelete(ctx context.Context, id int64) error {
	return s.repo.EnvironmentDelete(ctx, id)
}
func (s *CicdService) ApprovalCreate(ctx context.Context, approval *CicdApproval) (int64, error) {
	now := time.Now().Unix()
	approval.CreatedAt = uint64(now)
	approval.ModifiedAt = uint64(now)
	return s.repo.ApprovalSave(ctx, approval)
}
func (s *CicdService) ApprovalGetByID(ctx context.Context, id int64) (*CicdApproval, error) {
	var approval CicdApproval
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&approval).Error
	return &approval, err
}
func (s *CicdService) ApprovalList(ctx context.Context, page, pageSize int, status string, pipelineID int64) ([]*ApprovalListItem, int64, error) {
	var list []*ApprovalListItem
	var total int64
	tid, ok := s.tenantID, s.tenantID != 0
	if !ok {
		return nil, 0, nil
	}


	query := s.db.WithContext(ctx).
		Table("cicd_approval AS a").
		Select(`a.*, 
			COALESCE(p.name, '') AS pipeline_name,
			COALESCE(u1.username, '') AS request_username,
			COALESCE(u2.username, '') AS approve_username`).
		Joins("LEFT JOIN cicd_pipeline AS p ON a.pipeline_id = p.id AND p.tenant_id = ?", tid).
		Joins("LEFT JOIN `user` AS u1 ON a.request_user_id = u1.id AND u1.tenant_id = ?", tid).
		Joins("LEFT JOIN `user` AS u2 ON a.approve_user_id = u2.id AND u2.tenant_id = ?", tid).
		Where("a.tenant_id = ?", tid)

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
func (s *CicdService) ApprovalListByUser(ctx context.Context, userID int64, page, pageSize int, status string, pipelineID int64) ([]*ApprovalListItem, int64, error) {
	var list []*ApprovalListItem
	var total int64
	tid2, ok2 := s.tenantID, s.tenantID != 0
	if !ok2 {
		return nil, 0, nil
	}


	query := s.db.WithContext(ctx).
		Table("cicd_approval AS a").
		Select(`a.*, 
			COALESCE(p.name, '') AS pipeline_name,
			COALESCE(u1.username, '') AS request_username,
			COALESCE(u2.username, '') AS approve_username`).
		Joins("LEFT JOIN cicd_pipeline AS p ON a.pipeline_id = p.id AND p.tenant_id = ?", tid2).
		Joins("LEFT JOIN `user` AS u1 ON a.request_user_id = u1.id AND u1.tenant_id = ?", tid2).
		Joins("LEFT JOIN `user` AS u2 ON a.approve_user_id = u2.id AND u2.tenant_id = ?", tid2).
		Where("a.tenant_id = ? AND a.request_user_id = ?", tid2, userID)

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
func (s *CicdService) ApprovalStats(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []StatusCount
	err := s.db.WithContext(ctx).
		Model(&CicdApproval{}).
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
func (s *CicdService) ApprovalStatsByUser(ctx context.Context, userID int64) (map[string]int64, error) {
	type StatusCount struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []StatusCount
	err := s.db.WithContext(ctx).
		Model(&CicdApproval{}).
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
func (s *CicdService) ApprovalUpdateStatus(ctx context.Context, id int64, status string, approveUserID int64, reason string) error {
	now := time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdApproval{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          status,
			"approve_user_id": approveUserID,
			"approve_reason":  reason,
			"approve_time":    now,
			"modified_at":     now,
		}).Error
}
func (s *CicdService) ApprovalUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdApproval{}).
		Where("id = ?", id).
		Updates(updates).Error
}
func (s *CicdService) ApprovalDelete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&CicdApproval{}).Error
}
func (s *CicdService) ApprovalGetPendingByPipeline(ctx context.Context, pipelineID int64) (*CicdApproval, error) {
	var approval CicdApproval
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND status = ?", pipelineID, ApprovalStatusPending).
		Order("id DESC").
		First(&approval).Error
	return &approval, err
}
func (s *CicdService) ApprovalGetByStageID(ctx context.Context, stageID int64) (*CicdApproval, error) {
	var approval CicdApproval
	err := s.db.WithContext(ctx).
		Where("stage_id = ? AND status = ?", stageID, ApprovalStatusPending).
		First(&approval).Error
	return &approval, err
}
func (s *CicdService) ApprovalGetByFeishuToken(ctx context.Context, token string) (*CicdApproval, error) {
	var approval CicdApproval
	err := s.db.WithContext(ctx).
		Where("feishu_token = ?", token).
		First(&approval).Error
	return &approval, err
}
func (s *CicdService) StageCreate(ctx context.Context, stage *CicdPipelineStage) error {
	now := time.Now().Unix()
	stage.CreatedAt = uint64(now)
	stage.ModifiedAt = uint64(now)
	return s.db.WithContext(ctx).Create(stage).Error
}
func (s *CicdService) StageCreateBatch(ctx context.Context, stages []*CicdPipelineStage) error {
	if len(stages) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, stage := range stages {
		stage.CreatedAt = uint64(now)
		stage.ModifiedAt = uint64(now)
	}
	return s.db.WithContext(ctx).Create(&stages).Error
}
func (s *CicdService) StageGetByID(ctx context.Context, id int64) (*CicdPipelineStage, error) {
	var stage CicdPipelineStage
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&stage).Error
	return &stage, err
}
func (s *CicdService) StageListByRunID(ctx context.Context, runID int64) ([]*CicdPipelineStage, error) {
	var list []*CicdPipelineStage
	err := s.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("stage_order ASC").
		Find(&list).Error
	return list, err
}
func (s *CicdService) StageUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdPipelineStage{}).
		Where("id = ?", id).
		Updates(updates).Error
}
func (s *CicdService) StageUpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	// 根据状态设置开始/结束时间
	if status == StageStatusRunning {
		updates["started_at"] = time.Now().Unix()
	}
	if status == StageStatusSuccess || status == StageStatusFailed || 
	   status == StageStatusSkipped || status == StageStatusAborted {
		now := time.Now().Unix()
		updates["finished_at"] = now
	}
	return s.StageUpdate(ctx, id, updates)
}
func (s *CicdService) StageUpdateWithDuration(ctx context.Context, id int64, status string, duration int) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":       status,
		"finished_at":  now,
		"duration_sec": duration,
	}
	return s.StageUpdate(ctx, id, updates)
}
func (s *CicdService) StageUpdateLogs(ctx context.Context, id int64, logs string) error {
	return s.StageUpdate(ctx, id, map[string]interface{}{
		"logs": logs,
	})
}
func (s *CicdService) StageUpdateApproval(ctx context.Context, id int64, userID int64, decision, comment string) error {
	now := time.Now().Unix()
	status := StageStatusSuccess
	if decision == "rejected" {
		status = StageStatusFailed
	}
	return s.StageUpdate(ctx, id, map[string]interface{}{
		"status":            status,
		"finished_at":       now,
		"approval_user_id":  userID,
		"approval_decision": decision,
		"approval_comment":  comment,
	})
}
func (s *CicdService) StageUpdateDeploy(ctx context.Context, id int64, clusterID int64, namespace, workloadKind, workloadName, container, image string, replicas int) error {
	return s.StageUpdate(ctx, id, map[string]interface{}{
		"deploy_cluster_id":    clusterID,
		"deploy_namespace":     namespace,
		"deploy_workload_kind": workloadKind,
		"deploy_workload_name": workloadName,
		"deploy_container":     container,
		"deploy_image":         image,
		"deploy_replicas":      replicas,
	})
}
func (s *CicdService) StageUpdateError(ctx context.Context, id int64, errMsg string) error {
	return s.StageUpdate(ctx, id, map[string]interface{}{
		"status":        StageStatusFailed,
		"finished_at":   time.Now().Unix(),
		"error_message": errMsg,
	})
}
func (s *CicdService) StageGetByRunIDAndType(ctx context.Context, runID int64, stageType string) (*CicdPipelineStage, error) {
	var stage CicdPipelineStage
	err := s.db.WithContext(ctx).
		Where("run_id = ? AND stage_type = ?", runID, stageType).
		First(&stage).Error
	return &stage, err
}
func (s *CicdService) StageGetLogs(ctx context.Context, id int64) (string, error) {
	var stage CicdPipelineStage
	err := s.db.WithContext(ctx).
		Select("logs").
		Where("id = ?", id).
		First(&stage).Error
	return stage.Logs, err
}
func (s *CicdService) StageDeleteByRunID(ctx context.Context, runID int64) error {
	return s.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Delete(&CicdPipelineStage{}).Error
}
func (s *CicdService) StageListApprovalAll(ctx context.Context) ([]*CicdPipelineStage, error) {
	var list []*CicdPipelineStage
	err := s.db.WithContext(ctx).
		Where("stage_type = ?", StageTypeApproval).
		Find(&list).Error
	return list, err
}
func (s *CicdService) ApprovalExistsByStageID(ctx context.Context, stageID int64) (bool, error) {
	return s.repo.ApprovalExistsByStageID(ctx, stageID)
}
func (s *CicdService) PipelineRunBuildStats(ctx context.Context) (*BuildStatsResult, error) {
	stats := &BuildStatsResult{}

	// 总数 & 各状态统计
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
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
		case PipelineRunStatusSuccess:
			stats.SuccessBuilds = r.Cnt
			terminalBuilds += r.Cnt
		case PipelineRunStatusFailed:
			stats.FailedBuilds = r.Cnt
			terminalBuilds += r.Cnt
		case PipelineRunStatusRunning, PipelineRunStatusPending:
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
	s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
		Select("AVG(duration_sec) AS avg_dur").
		Where("status IN (?, ?) AND duration_sec > 0", PipelineRunStatusSuccess, PipelineRunStatusFailed).
		Scan(&avgDur)
	stats.AvgDuration = avgDur.Avg

	// 今日构建数
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
		Where("created_at >= ?", todayStart.Unix()).
		Count(&stats.TodayBuilds)

	// 本周构建数
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
		Where("created_at >= ?", weekStart.Unix()).
		Count(&stats.WeekBuilds)

	return stats, nil
}
func (s *CicdService) PipelineRunBuildTrend(ctx context.Context, days int) ([]BuildTrendItem, error) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days+1)

	var results []BuildTrendItem
	err := s.db.WithContext(ctx).
		Model(&CicdPipelineRun{}).
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
func (s *CicdService) CicdReleaseCreate(ctx context.Context, rel *CicdRelease) error {
	return s.db.WithContext(ctx).
		Create(rel).
		Error
}
func (s *CicdService) CicdReleaseUpdateStatusCAS(
	ctx context.Context,
	releaseID int64,
	from []string,
	to string,
	message string,
) (bool, error) {

	res := s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Where("status IN ?", from).
		Updates(map[string]any{
			"status":      to,
			"message":     message,
			"modified_at": time.Now().Unix(),
		})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
func (s *CicdService) CicdReleaseGetByRequestID(ctx context.Context, requestID string) (*CicdRelease, error) {
	var rel CicdRelease
	err := s.db.WithContext(ctx).
		Where("request_id = ? AND is_del = 0", requestID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
func (s *CicdService) CicdReleaseGetByID(ctx context.Context, releaseID int64) (*CicdRelease, error) {
	var rel CicdRelease
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", releaseID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
func (s *CicdService) CicdReleaseList(ctx context.Context, keyword, appName, status string, page, pageSize int) ([]*CicdRelease, int64, error) {
	var list []*CicdRelease
	var total int64

	query := s.db.WithContext(ctx).Model(&CicdRelease{}).Where("is_del = 0")

	// keyword 模糊搜索：应用名、工作负载名、镜像等
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where(
			"app_name LIKE ? OR workload_name LIKE ? OR image_repo LIKE ? OR image_tag LIKE ?",
			likePattern, likePattern, likePattern, likePattern,
		)
	}
	// 精确匹配 app_name（兼容旧参数）
	if appName != "" {
		query = query.Where("app_name LIKE ?", "%"+appName+"%")
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
func (s *CicdService) CicdReleaseStats(ctx context.Context) (map[string]int64, error) {
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Select("status, COUNT(*) AS cnt").
		Where("is_del = 0").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	stats := map[string]int64{
		"Pending": 0, "Queued": 0, "Running": 0,
		"Succeeded": 0, "Failed": 0, "Canceled": 0, "Rollback": 0,
		"total": 0,
	}
	for _, r := range rows {
		stats[r.Status] = r.Cnt
		stats["total"] += r.Cnt
	}
	return stats, nil
}
func (s *CicdService) CicdReleaseCancel(ctx context.Context, releaseID int64) (bool, error) {
	return s.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{CicdReleaseStatusPending, CicdReleaseStatusAwaitingApproval, CicdReleaseStatusQueued, CicdReleaseStatusRunning},
		CicdReleaseStatusCanceled,
		"user canceled",
	)
}
func (s *CicdService) CicdReleaseGetByBuildID(ctx context.Context, buildID int64) (*CicdRelease, error) {
	var rel CicdRelease
	err := s.db.WithContext(ctx).
		Where("build_id = ? AND is_del = 0", buildID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
func (s *CicdService) CicdReleaseUpdate(ctx context.Context, releaseID int64, updates map[string]any) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("id = ? AND is_del = 0 AND status IN ?", releaseID, []string{CicdReleaseStatusPending, CicdReleaseStatusAwaitingApproval, CicdReleaseStatusFailed, CicdReleaseStatusCanceled}).
		Updates(updates).Error
}
func (s *CicdService) CicdReleaseDelete(ctx context.Context, releaseID int64) error {
	now := time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Updates(map[string]any{
			"is_del":     1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}
func (s *CicdService) CicdReleaseUpdateImage(ctx context.Context, releaseID int64, imageRepo, imageTag, imageDigest string) error {
	updates := map[string]any{
		"image_repo":  imageRepo,
		"image_tag":   imageTag,
		"modified_at": time.Now().Unix(),
	}
	if imageDigest != "" {
		updates["image_digest"] = imageDigest
	}
	return s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Updates(updates).Error
}
func (s *CicdService) CicdReleaseHistory(ctx context.Context, appName, namespace, status string, startTime, endTime int64, page, pageSize int) ([]*CicdRelease, int64, error) {
	var list []*CicdRelease
	var total int64

	query := s.db.WithContext(ctx).Model(&CicdRelease{}).Where("is_del = 0")

	if appName != "" {
		query = query.Where("app_name LIKE ?", "%"+appName+"%")
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at <= ?", endTime)
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
func (s *CicdService) CicdReleaseStatsEnhanced(ctx context.Context) (*ReleaseStatsEnhanced, error) {
	stats := &ReleaseStatsEnhanced{
		ByStatus: make(map[string]int64),
	}

	// 按状态统计
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Select("status, COUNT(*) AS cnt").
		Where("is_del = 0").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	var successCount int64
	var terminalCount int64
	for _, r := range rows {
		stats.Total += r.Cnt
		stats.ByStatus[r.Status] = r.Cnt
		if r.Status == CicdReleaseStatusSucceeded {
			successCount = r.Cnt
		}
		if r.Status == CicdReleaseStatusSucceeded || r.Status == CicdReleaseStatusFailed {
			terminalCount += r.Cnt
		}
	}
	if terminalCount > 0 {
		stats.SuccessRate = float64(successCount) / float64(terminalCount) * 100
	}

	// 今日发布数
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("is_del = 0 AND created_at >= ?", todayStart.Unix()).
		Count(&stats.TodayCount)

	// 本周发布数
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	s.db.WithContext(ctx).
		Model(&CicdRelease{}).
		Where("is_del = 0 AND created_at >= ?", weekStart.Unix()).
		Count(&stats.WeekCount)

	return stats, nil
}
func (s *CicdService) CicdTasksCreate(ctx context.Context, tasks []*CicdReleaseTask) error {
	if len(tasks) == 0 {
		return nil
	}
	// 逐条创建，确保 GORM 正确回填每条记录的自增 ID
	for _, task := range tasks {
		if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
			return err
		}
	}
	return nil
}
func (s *CicdService) CicdTaskUpdatePrevImage(ctx context.Context, taskID int64, prev string) error {
	now := uint64(time.Now().Unix())
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"prev_image":  prev,
			"modified_at": now,
		}).Error
}
func (s *CicdService) CicdTaskGetByID(ctx context.Context, taskID int64) (*CicdReleaseTask, error) {
	var task CicdReleaseTask
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", taskID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
func (s *CicdService) CicdTaskUpdateStatus(ctx context.Context, taskID int64, status, message string) error {
	now := uint64(time.Now().Unix())
	updates := map[string]any{
		"status":      status,
		"message":     message,
		"modified_at": now,
	}
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(updates).Error
}
func (s *CicdService) CicdTaskUpdateStatusCAS(ctx context.Context, taskID int64, from []string, to, message string) (bool, error) {
	now := uint64(time.Now().Unix())
	res := s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Where("status IN ?", from).
		Updates(map[string]any{
			"status":      to,
			"message":     message,
			"modified_at": now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
func (s *CicdService) CicdTaskMarkStarted(ctx context.Context, taskID int64) error {
	now := uint64(time.Now().Unix())
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"status":      CicdTaskStatusRunning,
			"started_at":  now,
			"modified_at": now,
		}).Error
}
func (s *CicdService) CicdTaskMarkFinished(ctx context.Context, taskID int64, status, message string) error {
	now := uint64(time.Now().Unix())
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"status":      status,
			"message":     message,
			"finished_at": now,
			"modified_at": now,
		}).Error
}
func (s *CicdService) CicdTasksByReleaseID(ctx context.Context, releaseID int64) ([]*CicdReleaseTask, error) {
	var tasks []*CicdReleaseTask
	err := s.db.WithContext(ctx).
		Where("release_id = ? AND is_del = 0", releaseID).
		Order("id ASC").
		Find(&tasks).Error
	return tasks, err
}
func (s *CicdService) CicdTaskUpdateTargetImage(ctx context.Context, taskID int64, targetImage string) error {
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"target_image": targetImage,
			"modified_at":  time.Now().Unix(),
		}).Error
}
func (s *CicdService) CicdTasksFailByRelease(ctx context.Context, releaseID int64, msg string) error {
	return s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Where("release_id = ? AND is_del = 0", releaseID).
		Where("status IN ?", []string{"Pending", "Queued"}).
		Updates(map[string]any{
			"status":      "Failed",
			"message":     msg,
			"modified_at": time.Now().Unix(),
		}).Error
}
func (s *CicdService) CicdTaskStatsByRelease(ctx context.Context, releaseID int64) (*CicdTaskStats, error) {
	var rows []struct {
		Status string
		Cnt    int64
	}

	if err := s.db.WithContext(ctx).
		Model(&CicdReleaseTask{}).
		Select("status AS status, COUNT(1) AS cnt").
		Where("release_id = ? AND is_del = 0", releaseID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	st := &CicdTaskStats{}
	for _, r := range rows {
		switch r.Status {
		case "Pending":
			st.Pending = r.Cnt
		case "Queued":
			st.Queued = r.Cnt
		case "Running":
			st.Running = r.Cnt
		case "Succeeded":
			st.Succeeded = r.Cnt
		case "Failed":
			st.Failed = r.Cnt
		case "Canceled":
			st.Canceled = r.Cnt
		}
	}
	return st, nil
}
func (s *CicdService) ArtifactCreate(ctx context.Context, artifact *CicdArtifact) error {
	now := uint64(time.Now().Unix())
	artifact.CreatedAt = now
	artifact.ModifiedAt = now
	return s.db.WithContext(ctx).Create(artifact).Error
}
func (s *CicdService) ArtifactGetByID(ctx context.Context, id int64) (*CicdArtifact, error) {
	var artifact CicdArtifact
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&artifact).Error
	return &artifact, err
}
func (s *CicdService) ArtifactList(ctx context.Context, pipelineID int64, artifactType, languageType, status string, page, pageSize int) ([]*CicdArtifact, int64, error) {
	db := s.db.WithContext(ctx).Model(&CicdArtifact{}).Where("is_del = 0")

	if pipelineID > 0 {
		db = db.Where("pipeline_id = ?", pipelineID)
	}
	if artifactType != "" {
		db = db.Where("artifact_type = ?", artifactType)
	}
	if languageType != "" {
		db = db.Where("language_type = ?", languageType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*CicdArtifact
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
func (s *CicdService) ArtifactListByRunID(ctx context.Context, runID int64) ([]*CicdArtifact, error) {
	var list []*CicdArtifact
	err := s.db.WithContext(ctx).
		Where("run_id = ? AND is_del = 0", runID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
func (s *CicdService) ArtifactUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdArtifact{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}
func (s *CicdService) ArtifactDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdArtifact{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
			"status":      ArtifactStatusDeleted,
		}).Error
}
func (s *CicdService) ArtifactIncrDownload(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Model(&CicdArtifact{}).
		Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}
func (s *CicdService) ArtifactBatchDelete(ctx context.Context, ids []int64) (int64, error) {
	now := time.Now().Unix()
	result := s.db.WithContext(ctx).
		Model(&CicdArtifact{}).
		Where("id IN ? AND is_del = 0", ids).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
			"status":      ArtifactStatusDeleted,
		})
	return result.RowsAffected, result.Error
}
func (s *CicdService) ArtifactGetByIDs(ctx context.Context, ids []int64) ([]*CicdArtifact, error) {
	var list []*CicdArtifact
	err := s.db.WithContext(ctx).
		Where("id IN ? AND is_del = 0", ids).
		Find(&list).Error
	return list, err
}
func (s *CicdService) ArtifactStats(ctx context.Context, pipelineID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	db := s.db.WithContext(ctx).
		Model(&CicdArtifact{}).
		Select("artifact_type, COUNT(*) as count, SUM(file_size) as total_size").
		Where("is_del = 0 AND status = ?", ArtifactStatusReady)

	if pipelineID > 0 {
		db = db.Where("pipeline_id = ?", pipelineID)
	}

	err := db.Group("artifact_type").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询制品统计失败: %w", err)
	}
	return results, nil
}
func (s *CicdService) BuildAgentCreate(ctx context.Context, agent *CicdBuildAgent) error {
	now := uint64(time.Now().Unix())
	agent.CreatedAt = now
	agent.ModifiedAt = now
	return s.db.WithContext(ctx).Create(agent).Error
}
func (s *CicdService) BuildAgentGetByID(ctx context.Context, id int64) (*CicdBuildAgent, error) {
	var agent CicdBuildAgent
	err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&agent).Error
	return &agent, err
}
func (s *CicdService) BuildAgentGetByName(ctx context.Context, name string) (*CicdBuildAgent, error) {
	var agent CicdBuildAgent
	err := s.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&agent).Error
	return &agent, err
}
func (s *CicdService) BuildAgentList(ctx context.Context, category, scope, status, keyword string, page, pageSize int) ([]*CicdBuildAgent, int64, error) {
	db := s.db.WithContext(ctx).Model(&CicdBuildAgent{}).Where("is_del = 0")

	if category != "" {
		db = db.Where("category = ?", category)
	}
	if scope != "" {
		db = db.Where("scope = ? OR scope = ?", scope, AgentScopeAll)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*CicdBuildAgent
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
func (s *CicdService) BuildAgentListByScope(ctx context.Context, scope string) ([]*CicdBuildAgent, error) {
	var list []*CicdBuildAgent
	err := s.db.WithContext(ctx).
		Where("is_del = 0 AND status = ? AND (scope = ? OR scope = ?)", AgentStatusActive, scope, AgentScopeAll).
		Order("category, name").
		Find(&list).Error
	return list, err
}
func (s *CicdService) BuildAgentUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdBuildAgent{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}
func (s *CicdService) BuildAgentDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdBuildAgent{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}
func (s *CicdService) BuildAgentIncrDownload(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Model(&CicdBuildAgent{}).
		Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}
func (s *CicdService) BuildAgentIncrUsed(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Model(&CicdBuildAgent{}).
		Where("id = ?", id).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}
func (s *CicdService) ResourceTemplateList(ctx context.Context, env, serviceType string) ([]CicdResourceTemplate, error) {
	var list []CicdResourceTemplate
	db := s.db.WithContext(ctx).Where("deleted_at = 0")

	if env != "" {
		db = db.Where("env = ?", env)
	}
	if serviceType != "" {
		db = db.Where("service_type = ?", serviceType)
	}

	err := db.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}
func (s *CicdService) ResourceTemplateGetByID(ctx context.Context, id uint64) (*CicdResourceTemplate, error) {
	var tpl CicdResourceTemplate
	err := s.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}
func (s *CicdService) ResourceTemplateGetDefault(ctx context.Context, env, serviceType string) (*CicdResourceTemplate, error) {
	var tpl CicdResourceTemplate
	err := s.db.WithContext(ctx).
		Where("env = ? AND service_type = ? AND is_default = 1 AND deleted_at = 0", env, serviceType).
		First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}
func (s *CicdService) ResourceTemplateCreate(ctx context.Context, tpl *CicdResourceTemplate) error {
	tpl.CreatedAt = uint64(time.Now().Unix())
	tpl.ModifiedAt = uint64(time.Now().Unix())
	return s.db.WithContext(ctx).Create(tpl).Error
}
func (s *CicdService) ResourceTemplateUpdate(ctx context.Context, tpl *CicdResourceTemplate) error {
	tpl.ModifiedAt = uint64(time.Now().Unix())
	return s.db.WithContext(ctx).Save(tpl).Error
}
func (s *CicdService) ResourceTemplateDelete(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Model(&CicdResourceTemplate{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now().Unix()).Error
}
func (s *CicdService) EnvResourceRuleList(ctx context.Context, env string) ([]CicdEnvResourceRule, error) {
	var list []CicdEnvResourceRule
	db := s.db.WithContext(ctx)

	if env != "" {
		db = db.Where("env = ?", env)
	}

	err := db.Order("env ASC, service_type ASC").Find(&list).Error
	return list, err
}
func (s *CicdService) EnvResourceRuleGet(ctx context.Context, env, serviceType string) (*CicdEnvResourceRule, error) {
	var rule CicdEnvResourceRule

	// 优先匹配特定服务类型的规则
	err := s.db.WithContext(ctx).
		Where("env = ? AND service_type = ?", env, serviceType).
		First(&rule).Error
	if err == nil {
		return &rule, nil
	}

	// 回退到通用规则
	if err == gorm.ErrRecordNotFound {
		err = s.db.WithContext(ctx).
			Where("env = ? AND service_type = ''", env).
			First(&rule).Error
	}

	if err != nil {
		return nil, err
	}
	return &rule, nil
}
func (s *CicdService) EnvResourceRuleGetByID(ctx context.Context, id uint64) (*CicdEnvResourceRule, error) {
	var rule CicdEnvResourceRule
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
func (s *CicdService) EnvResourceRuleUpdate(ctx context.Context, rule *CicdEnvResourceRule) error {
	rule.ModifiedAt = uint64(time.Now().Unix())
	return s.db.WithContext(ctx).Save(rule).Error
}
func (s *CicdService) DeployApprovalCreate(ctx context.Context, approval *CicdDeployApproval) error {
	approval.AppliedAt = uint64(time.Now().Unix())
	approval.ExpiredAt = uint64(time.Now().Add(72 * time.Hour).Unix()) // 72小时过期
	return s.db.WithContext(ctx).Create(approval).Error
}
func (s *CicdService) DeployApprovalGetByID(ctx context.Context, id uint64) (*CicdDeployApproval, error) {
	var approval CicdDeployApproval
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}
func (s *CicdService) DeployApprovalList(ctx context.Context, status string, page, pageSize int) ([]CicdDeployApproval, int64, error) {
	var list []CicdDeployApproval
	var total int64

	db := s.db.WithContext(ctx).Model(&CicdDeployApproval{})
	if status != "" {
		db = db.Where("status = ?", status)
	}

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("applied_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func (s *CicdService) DeployApprovalListByApplicant(ctx context.Context, applicantID uint64, page, pageSize int) ([]CicdDeployApproval, int64, error) {
	var list []CicdDeployApproval
	var total int64

	db := s.db.WithContext(ctx).Model(&CicdDeployApproval{}).Where("applicant_id = ?", applicantID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("applied_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func (s *CicdService) DeployApprovalApprove(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return s.db.WithContext(ctx).Model(&CicdDeployApproval{}).
		Where("id = ? AND status = ?", id, ApprovalStatusPending).
		Updates(map[string]interface{}{
			"status":          ApprovalStatusApproved,
			"approver_id":     approverID,
			"approver_name":   approverName,
			"approve_comment": comment,
			"approved_at":     time.Now().Unix(),
		}).Error
}
func (s *CicdService) DeployApprovalReject(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return s.db.WithContext(ctx).Model(&CicdDeployApproval{}).
		Where("id = ? AND status = ?", id, ApprovalStatusPending).
		Updates(map[string]interface{}{
			"status":          ApprovalStatusRejected,
			"approver_id":     approverID,
			"approver_name":   approverName,
			"approve_comment": comment,
			"approved_at":     time.Now().Unix(),
		}).Error
}
func (s *CicdService) DeployApprovalCancel(ctx context.Context, id uint64, applicantID uint64) error {
	return s.db.WithContext(ctx).Model(&CicdDeployApproval{}).
		Where("id = ? AND applicant_id = ? AND status = ?", id, applicantID, ApprovalStatusPending).
		Update("status", ApprovalStatusCancelled).Error
}
func (s *CicdService) DeployApprovalExpireOld(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Model(&CicdDeployApproval{}).
		Where("status = ? AND expired_at < ?", ApprovalStatusPending, time.Now().Unix()).
		Update("status", ApprovalStatusExpired)
	return result.RowsAffected, result.Error
}
func (s *CicdService) ResourceChangeLogCreate(ctx context.Context, log *CicdResourceChangeLog) error {
	log.CreatedAt = uint64(time.Now().Unix())
	return s.db.WithContext(ctx).Create(log).Error
}
func (s *CicdService) ResourceChangeLogList(ctx context.Context, pipelineID uint64, env string, page, pageSize int) ([]CicdResourceChangeLog, int64, error) {
	var list []CicdResourceChangeLog
	var total int64

	db := s.db.WithContext(ctx).Model(&CicdResourceChangeLog{}).Where("pipeline_id = ?", pipelineID)
	if env != "" {
		db = db.Where("env = ?", env)
	}
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func (s *CicdService) TemplateCreate(ctx context.Context, template *CicdPipelineTemplate) error {
	return s.db.WithContext(ctx).Create(template).Error
}
func (s *CicdService) TemplateGetByID(ctx context.Context, id int64) (*CicdPipelineTemplate, error) {
	var template CicdPipelineTemplate
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}
func (s *CicdService) TemplateGetByName(ctx context.Context, name string) (*CicdPipelineTemplate, error) {
	var template CicdPipelineTemplate
	err := s.db.WithContext(ctx).
		Where("name = ? AND is_del = 0", name).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}
func (s *CicdService) TemplateUpdate(ctx context.Context, template *CicdPipelineTemplate) error {
	return s.db.WithContext(ctx).
		Model(template).
		Select("name", "description", "type", "stages", "default_env_vars", "deploy_config", "jenkins_template", "modified_at").
		Updates(template).Error
}
func (s *CicdService) TemplateDelete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Model(&CicdPipelineTemplate{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_del":     1,
			"deleted_at": s.db.NowFunc().Unix(),
		}).Error
}
func (s *CicdService) TemplateList(ctx context.Context, keyword, templateType string, page, pageSize int) ([]*CicdPipelineTemplate, int64, error) {
	var templates []*CicdPipelineTemplate
	var total int64

	db := s.db.WithContext(ctx).Model(&CicdPipelineTemplate{}).Where("is_del = 0")

	// 关键字搜索
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型筛选
	if templateType != "" {
		db = db.Where("type = ?", templateType)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
func (s *CicdService) TemplateIncrUsageCount(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).
		Model(&CicdPipelineTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", s.db.Raw("usage_count + 1")).Error
}
func (s *CicdService) PipelineTargetListByPipeline(ctx context.Context, pipelineID int64) ([]*CicdPipelineTargetView, error) {
	var list []*CicdPipelineTargetView
	tid, ok := s.tenantID, s.tenantID != 0
	if !ok {
		return nil, nil
	}

	err := s.db.WithContext(ctx).
		Table("cicd_pipeline_target AS t").
		Select("t.*, COALESCE(c.cluster_name, '') AS cluster_name").
		Joins("LEFT JOIN kube_cluster c ON c.id = t.cluster_id AND c.tenant_id = ? AND c.is_del = 0", tid).
		Where("t.pipeline_id = ? AND t.is_del = 0 AND t.tenant_id = ?", pipelineID, tid).
		Order("t.sort_order ASC, t.id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
func (s *CicdService) PipelineTargetGetByPipelineAndEnv(ctx context.Context, pipelineID int64, env string) (*CicdPipelineTarget, error) {
	var t CicdPipelineTarget
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND env = ? AND is_del = 0", pipelineID, env).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}
func (s *CicdService) PipelineTargetCreate(ctx context.Context, t *CicdPipelineTarget) error {
	now := uint64(time.Now().Unix())
	t.CreatedAt = now
	t.ModifiedAt = now
	return s.db.WithContext(ctx).Create(t).Error
}
func (s *CicdService) PipelineTargetUpdate(ctx context.Context, id int64, updates map[string]any) error {
	updates["modified_at"] = time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdPipelineTarget{}).
		Where("id = ? AND is_del = 0", id).
		Updates(updates).Error
}
func (s *CicdService) PipelineTargetDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return s.db.WithContext(ctx).
		Model(&CicdPipelineTarget{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]any{
			"is_del":      1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}
func (s *CicdService) PipelineTargetUpsert(ctx context.Context, t *CicdPipelineTarget) error {
	exist, err := s.PipelineTargetGetByPipelineAndEnv(ctx, t.PipelineID, t.Env)
	if err == nil && exist != nil {
		return s.PipelineTargetUpdate(ctx, exist.ID, map[string]any{
			"cluster_id":       t.ClusterID,
			"namespace":        t.Namespace,
			"workload_kind":    t.WorkloadKind,
			"workload_name":    t.WorkloadName,
			"container":        t.Container,
			"auto_deploy":      t.AutoDeploy,
			"require_approval": t.RequireApproval,
			"promote_from":     t.PromoteFrom,
			"sort_order":       t.SortOrder,
		})
	}
	return s.PipelineTargetCreate(ctx, t)
}
func (s *CicdService) CicdReleaseLatestByPipelineEnv(ctx context.Context, pipelineID int64, env string) (*CicdRelease, error) {
	var rel CicdRelease
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND env = ? AND is_del = 0", pipelineID, env).
		Order("id DESC").
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
func (s *CicdService) CicdReleaseLatestBySourceRunID(ctx context.Context, pipelineID, sourceRunID int64) (*CicdRelease, error) {
	var rel CicdRelease
	err := s.db.WithContext(ctx).
		Where("pipeline_id = ? AND source_run_id = ? AND env <> '' AND is_del = 0", pipelineID, sourceRunID).
		Order("id DESC").
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
