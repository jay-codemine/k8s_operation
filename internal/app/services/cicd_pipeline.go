package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/jenkins"
)

// ==================== 流水线 CRUD ====================

// PipelineCreate 创建流水线
func (s *Services) PipelineCreate(ctx context.Context, req *requests.PipelineCreateRequest, userID int64) (int64, error) {
	// 检查名称是否已存在
	_, err := s.dao.PipelineGetByName(ctx, req.Name)
	if err == nil {
		return 0, errors.New("流水线名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("检查名称失败: %w", err)
	}

	pipeline := &models.CicdPipeline{
		Name:          req.Name,
		Description:   req.Description,
		GitRepo:       req.GitRepo,
		GitBranch:     req.GitBranch,
		JenkinsURL:    req.JenkinsURL,
		JenkinsJob:    req.JenkinsJob,
		Status:        models.PipelineStatusIdle,
		EnvVars:       models.EnvVars(req.EnvVars),
		DeployConfig:  models.JSONMap(req.DeployConfig),
		CreatedUserID: userID,
	}

	if err := s.dao.PipelineCreate(ctx, pipeline); err != nil {
		return 0, fmt.Errorf("创建流水线失败: %w", err)
	}

	return pipeline.ID, nil
}

// PipelineDetail 获取流水线详情
func (s *Services) PipelineDetail(ctx context.Context, id int64) (*models.CicdPipeline, error) {
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("流水线不存在")
		}
		return nil, fmt.Errorf("查询流水线失败: %w", err)
	}
	return pipeline, nil
}

// PipelineList 获取流水线列表
func (s *Services) PipelineList(ctx context.Context, req *requests.PipelineListRequest) ([]*models.PipelineListItem, int64, error) {
	list, total, err := s.dao.PipelineList(ctx, req.Keyword, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询流水线列表失败: %w", err)
	}

	// 转换为列表项
	items := make([]*models.PipelineListItem, 0, len(list))
	for _, p := range list {
		items = append(items, p.ToPipelineListItem())
	}

	return items, total, nil
}

// PipelineUpdate 更新流水线
func (s *Services) PipelineUpdate(ctx context.Context, req *requests.PipelineUpdateRequest) error {
	// 检查流水线是否存在
	pipeline, err := s.dao.PipelineGetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("流水线不存在")
		}
		return fmt.Errorf("查询流水线失败: %w", err)
	}

	// 如果修改了名称，检查新名称是否已存在
	if req.Name != "" && req.Name != pipeline.Name {
		_, err := s.dao.PipelineGetByName(ctx, req.Name)
		if err == nil {
			return errors.New("流水线名称已存在")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("检查名称失败: %w", err)
		}
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.GitRepo != "" {
		updates["git_repo"] = req.GitRepo
	}
	if req.GitBranch != "" {
		updates["git_branch"] = req.GitBranch
	}
	if req.JenkinsURL != "" {
		updates["jenkins_url"] = req.JenkinsURL
	}
	if req.JenkinsJob != "" {
		updates["jenkins_job"] = req.JenkinsJob
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.EnvVars != nil {
		updates["env_vars"] = models.EnvVars(req.EnvVars)
	}
	if req.DeployConfig != nil {
		updates["deploy_config"] = models.JSONMap(req.DeployConfig)
	}

	if len(updates) == 0 {
		return nil
	}

	return s.dao.PipelineUpdate(ctx, req.ID, updates)
}

// PipelineDelete 删除流水线
func (s *Services) PipelineDelete(ctx context.Context, id int64) error {
	// 检查是否存在
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("流水线不存在")
		}
		return fmt.Errorf("查询流水线失败: %w", err)
	}

	// 检查是否正在运行
	if pipeline.Status == models.PipelineStatusRunning {
		return errors.New("流水线正在运行中，无法删除")
	}

	return s.dao.PipelineDelete(ctx, id)
}

// ==================== 流水线运行 ====================

// PipelineRun 运行流水线（触发 Jenkins 构建）
func (s *Services) PipelineRun(ctx context.Context, req *requests.PipelineRunRequest, userID int64) (*models.CicdPipelineRun, error) {
	// 获取流水线配置
	pipeline, err := s.dao.PipelineGetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("流水线不存在")
		}
		return nil, fmt.Errorf("查询流水线失败: %w", err)
	}

	// 检查状态
	if pipeline.Status == models.PipelineStatusDisabled {
		return nil, errors.New("流水线已禁用")
	}
	if pipeline.Status == models.PipelineStatusRunning {
		return nil, errors.New("流水线正在运行中")
	}

	// 确定构建分支
	branch := pipeline.GitBranch
	if req.Branch != "" {
		branch = req.Branch
	}

	// 创建运行记录
	run := &models.CicdPipelineRun{
		PipelineID:    pipeline.ID,
		Status:        models.PipelineRunStatusPending,
		TriggerType:   models.TriggerTypeManual,
		TriggerUserID: userID,
		GitBranch:     branch,
	}
	if err := s.dao.PipelineRunCreate(ctx, run); err != nil {
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	// 更新流水线状态为运行中
	if err := s.dao.PipelineUpdateStatus(ctx, pipeline.ID, models.PipelineStatusRunning); err != nil {
		return nil, fmt.Errorf("更新流水线状态失败: %w", err)
	}

	// 构建 Jenkins 参数
	params := make(map[string]string)
	params["GIT_BRANCH"] = branch
	params["GIT_REPO"] = pipeline.GitRepo
	
	// 合并环境变量
	for _, ev := range pipeline.EnvVars {
		params[ev.Name] = ev.Value
	}
	// 请求中的环境变量优先级更高
	for k, v := range req.EnvVars {
		params[k] = v
	}

	// 异步触发 Jenkins 构建
	go s.triggerJenkinsBuild(context.Background(), pipeline, run, params)

	return run, nil
}

// triggerJenkinsBuild 异步触发 Jenkins 构建
func (s *Services) triggerJenkinsBuild(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, params map[string]string) {
	// 创建 Jenkins 客户端
	// 注意：实际使用时需要从配置或数据库获取 Jenkins 凭据
	jenkinsURL := pipeline.JenkinsURL
	if jenkinsURL == "" {
		jenkinsURL = "http://localhost:8080" // 默认地址，实际应从配置读取
	}

	client := jenkins.NewClient(jenkinsURL, "", "") // TODO: 从配置获取凭据

	// 触发构建并等待获取构建号
	result, err := client.TriggerBuildAndWait(ctx, pipeline.JenkinsJob, params, 60*time.Second)
	if err != nil {
		// 更新运行记录为失败
		_ = s.dao.PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusFailed)
		_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusFailed)
		return
	}

	// 更新运行记录
	_ = s.dao.PipelineRunUpdateBuildNumber(ctx, run.ID, result.BuildNumber)
	_ = s.dao.PipelineUpdateRunInfo(ctx, pipeline.ID, models.PipelineRunStatusRunning, result.BuildNumber, result.BuildURL)
}

// PipelineStop 停止流水线
func (s *Services) PipelineStop(ctx context.Context, req *requests.PipelineStopRequest) error {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("流水线不存在")
		}
		return fmt.Errorf("查询流水线失败: %w", err)
	}

	if pipeline.Status != models.PipelineStatusRunning {
		return errors.New("流水线未在运行")
	}

	// 确定构建号
	buildNumber := req.BuildNumber
	if buildNumber == 0 {
		buildNumber = pipeline.LastBuildNumber
	}
	if buildNumber == 0 {
		return errors.New("无法确定要停止的构建号")
	}

	// 创建 Jenkins 客户端并停止构建
	jenkinsURL := pipeline.JenkinsURL
	if jenkinsURL == "" {
		jenkinsURL = "http://localhost:8080"
	}

	client := jenkins.NewClient(jenkinsURL, "", "") // TODO: 从配置获取凭据

	if err := client.StopBuild(ctx, pipeline.JenkinsJob, buildNumber); err != nil {
		return fmt.Errorf("停止构建失败: %w", err)
	}

	// 更新状态
	_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusAborted)

	// 更新运行记录
	latestRun, err := s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
	if err == nil && latestRun.BuildNumber == buildNumber {
		_ = s.dao.PipelineRunUpdateStatus(ctx, latestRun.ID, models.PipelineRunStatusAborted)
	}

	return nil
}

// PipelineLogs 获取流水线日志
func (s *Services) PipelineLogs(ctx context.Context, req *requests.PipelineLogsRequest) (string, error) {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("流水线不存在")
		}
		return "", fmt.Errorf("查询流水线失败: %w", err)
	}

	// 确定构建号
	buildNumber := req.BuildNumber
	if buildNumber == 0 {
		buildNumber = pipeline.LastBuildNumber
	}
	if buildNumber == 0 {
		return "", errors.New("没有可用的构建记录")
	}

	// 创建 Jenkins 客户端
	jenkinsURL := pipeline.JenkinsURL
	if jenkinsURL == "" {
		jenkinsURL = "http://localhost:8080"
	}

	client := jenkins.NewClient(jenkinsURL, "", "") // TODO: 从配置获取凭据

	log, err := client.GetConsoleLog(ctx, pipeline.JenkinsJob, buildNumber, req.StartLine)
	if err != nil {
		return "", fmt.Errorf("获取日志失败: %w", err)
	}

	return log, nil
}

// PipelineStatus 获取流水线实时状态
func (s *Services) PipelineStatus(ctx context.Context, id int64) (*models.CicdPipeline, *jenkins.BuildInfo, error) {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("流水线不存在")
		}
		return nil, nil, fmt.Errorf("查询流水线失败: %w", err)
	}

	// 如果有构建号，获取 Jenkins 构建状态
	var buildInfo *jenkins.BuildInfo
	if pipeline.LastBuildNumber > 0 {
		jenkinsURL := pipeline.JenkinsURL
		if jenkinsURL == "" {
			jenkinsURL = "http://localhost:8080"
		}

		client := jenkins.NewClient(jenkinsURL, "", "")
		buildInfo, _ = client.GetBuildInfo(ctx, pipeline.JenkinsJob, pipeline.LastBuildNumber)

		// 如果构建已完成，同步更新本地状态
		if buildInfo != nil && !buildInfo.Building {
			runStatus := jenkins.BuildStatusToRunStatus(buildInfo.Building, buildInfo.Result)
			if runStatus != pipeline.LastRunStatus {
				_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, runStatus)
				pipeline.LastRunStatus = runStatus
				pipeline.Status = models.PipelineStatusIdle
			}
		}
	}

	return pipeline, buildInfo, nil
}

// PipelineHistory 获取流水线运行历史
func (s *Services) PipelineHistory(ctx context.Context, req *requests.PipelineHistoryRequest) ([]*models.CicdPipelineRun, int64, error) {
	return s.dao.PipelineRunList(ctx, req.ID, req.Page, req.PageSize)
}

// PipelineJenkinsCallback 处理 Jenkins 构建状态回调
func (s *Services) PipelineJenkinsCallback(ctx context.Context, req *requests.JenkinsBuildStatusCallback) error {
	// 根据 Job 名称查找流水线
	list, _, err := s.dao.PipelineList(ctx, "", "", 1, 1000)
	if err != nil {
		return fmt.Errorf("查询流水线列表失败: %w", err)
	}

	var pipeline *models.CicdPipeline
	for _, p := range list {
		if p.JenkinsJob == req.JobName {
			pipeline = p
			break
		}
	}

	if pipeline == nil {
		return fmt.Errorf("未找到关联的流水线: job=%s", req.JobName)
	}

	// 转换状态
	runStatus := jenkins.BuildStatusToRunStatus(false, req.Status)

	// 更新流水线状态
	if err := s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, runStatus); err != nil {
		return fmt.Errorf("更新流水线状态失败: %w", err)
	}

	// 更新运行记录
	latestRun, err := s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
	if err == nil && latestRun.BuildNumber == req.BuildNumber {
		updates := map[string]interface{}{
			"status":       runStatus,
			"duration_sec": req.Duration,
		}
		_ = s.dao.PipelineRunUpdate(ctx, latestRun.ID, updates)
	}

	return nil
}
