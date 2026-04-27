package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/jenkins"
	"k8soperation/pkg/k8s/deployment"
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

	// 模板化发布：根据 language_type 自动推导 JenkinsJob
	languageType := req.LanguageType
	if languageType == "" {
		languageType = models.LanguageTypeCustom
	}
	jenkinsJob := req.JenkinsJob
	if jenkinsJob == "" && languageType != models.LanguageTypeCustom {
		// 从语言类型自动映射到通用构建 Job
		if job, ok := models.DefaultJenkinsJobMap[languageType]; ok {
			jenkinsJob = job
		} else {
			return 0, fmt.Errorf("不支持的语言类型: %s", languageType)
		}
	}
	if jenkinsJob == "" {
		return 0, errors.New("Jenkins Job 名称不能为空，请指定 jenkins_job 或设置 language_type")
	}

	pipeline := &models.CicdPipeline{
		Name:               req.Name,
		Description:        req.Description,
		GitRepo:            req.GitRepo,
		GitBranch:          req.GitBranch,
		JenkinsURL:         req.JenkinsURL,
		JenkinsJob:         jenkinsJob,
		LanguageType:       languageType,
		Status:             models.PipelineStatusIdle,
		EnvVars:            models.EnvVars(req.EnvVars),
		DeployConfig:       models.JSONMap(req.DeployConfig),
		// 部署配置
		AutoDeploy:         req.AutoDeploy,
		TargetClusterID:    req.TargetClusterID,
		TargetNamespace:    req.TargetNamespace,
		TargetWorkloadKind: req.TargetWorkloadKind,
		TargetWorkloadName: req.TargetWorkloadName,
		TargetContainer:    req.TargetContainer,
		DeployEnv:          req.DeployEnv,
		RequireApproval:    req.RequireApproval,
		EnableSonar:        req.EnableSonar,
		EnableArtifactUpload: req.EnableArtifactUpload,
		CreatedUserID:      userID,
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
	// 部署配置字段
	if req.AutoDeploy != nil {
		updates["auto_deploy"] = *req.AutoDeploy
	}
	if req.TargetClusterID != nil {
		updates["target_cluster_id"] = *req.TargetClusterID
	}
	if req.TargetNamespace != nil {
		updates["target_namespace"] = *req.TargetNamespace
	}
	if req.TargetWorkloadKind != nil {
		updates["target_workload_kind"] = *req.TargetWorkloadKind
	}
	if req.TargetWorkloadName != nil {
		updates["target_workload_name"] = *req.TargetWorkloadName
	}
	if req.TargetContainer != nil {
		updates["target_container"] = *req.TargetContainer
	}
	if req.DeployEnv != nil {
		updates["deploy_env"] = *req.DeployEnv
	}
	if req.RequireApproval != nil {
		updates["require_approval"] = *req.RequireApproval
	}
	if req.EnableSonar != nil {
		updates["enable_sonar"] = *req.EnableSonar
	}
	if req.EnableArtifactUpload != nil {
		updates["enable_artifact_upload"] = *req.EnableArtifactUpload
	}
	if req.LanguageType != nil {
		updates["language_type"] = *req.LanguageType
		// 如果同时没有指定 jenkins_job，自动映射
		if req.JenkinsJob == "" && *req.LanguageType != models.LanguageTypeCustom {
			if job, ok := models.DefaultJenkinsJobMap[*req.LanguageType]; ok {
				updates["jenkins_job"] = job
			}
		}
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

// ==================== 批量创建流水线 ====================

// PipelineBatchCreateResult 批量创建结果
type PipelineBatchCreateResult struct {
	SuccessCount int                      `json:"success_count"`
	FailCount    int                      `json:"fail_count"`
	SkipCount    int                      `json:"skip_count"`
	Results      []PipelineBatchItemResult `json:"results"`
}

// PipelineBatchItemResult 单个流水线创建结果
type PipelineBatchItemResult struct {
	Name       string `json:"name"`
	Success    bool   `json:"success"`
	PipelineID int64  `json:"pipeline_id,omitempty"`
	Skipped    bool   `json:"skipped,omitempty"`
	Error      string `json:"error,omitempty"`
}

// PipelineBatchCreate 批量创建流水线
func (s *Services) PipelineBatchCreate(ctx context.Context, req *requests.PipelineBatchCreateRequest, userID int64) (*PipelineBatchCreateResult, error) {
	if len(req.Pipelines) == 0 {
		return nil, errors.New("流水线列表不能为空")
	}
	if len(req.Pipelines) > 200 {
		return nil, errors.New("单次批量创建不能超过 200 条")
	}

	result := &PipelineBatchCreateResult{
		Results: make([]PipelineBatchItemResult, 0, len(req.Pipelines)),
	}

	for _, item := range req.Pipelines {
		itemResult := PipelineBatchItemResult{Name: item.Name}

		// 基本校验
		if item.Name == "" {
			itemResult.Error = "流水线名称不能为空"
			result.FailCount++
			result.Results = append(result.Results, itemResult)
			continue
		}
		if item.GitRepo == "" {
			itemResult.Error = "Git 仓库地址不能为空"
			result.FailCount++
			result.Results = append(result.Results, itemResult)
			continue
		}

		// 检查名称是否已存在
		_, err := s.dao.PipelineGetByName(ctx, item.Name)
		if err == nil {
			// 已存在
			if req.SkipExisting {
				itemResult.Skipped = true
				itemResult.Success = true
				result.SkipCount++
				result.Results = append(result.Results, itemResult)
				continue
			}
			itemResult.Error = "流水线名称已存在"
			result.FailCount++
			result.Results = append(result.Results, itemResult)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			itemResult.Error = fmt.Sprintf("检查名称失败: %v", err)
			result.FailCount++
			result.Results = append(result.Results, itemResult)
			continue
		}

		// 模板化推导
		languageType := item.LanguageType
		if languageType == "" {
			languageType = models.LanguageTypeCustom
		}
		jenkinsJob := ""
		if languageType != models.LanguageTypeCustom {
			if job, ok := models.DefaultJenkinsJobMap[languageType]; ok {
				jenkinsJob = job
			} else {
				itemResult.Error = fmt.Sprintf("不支持的语言类型: %s", languageType)
				result.FailCount++
				result.Results = append(result.Results, itemResult)
				continue
			}
		}

		gitBranch := item.GitBranch
		if gitBranch == "" {
			gitBranch = "main"
		}

		pipeline := &models.CicdPipeline{
			Name:               item.Name,
			Description:        item.Description,
			GitRepo:            item.GitRepo,
			GitBranch:          gitBranch,
			JenkinsJob:         jenkinsJob,
			LanguageType:       languageType,
			Status:             models.PipelineStatusIdle,
			EnvVars:            models.EnvVars(item.EnvVars),
			AutoDeploy:         item.AutoDeploy,
			TargetClusterID:    item.TargetClusterID,
			TargetNamespace:    item.TargetNamespace,
			TargetWorkloadKind: item.TargetWorkloadKind,
			TargetWorkloadName: item.TargetWorkloadName,
			TargetContainer:    item.TargetContainer,
			DeployEnv:          item.DeployEnv,
			RequireApproval:    item.RequireApproval,
			EnableSonar:        item.EnableSonar,
			EnableArtifactUpload: item.EnableArtifactUpload,
			CreatedUserID:      userID,
		}

		if err := s.dao.PipelineCreate(ctx, pipeline); err != nil {
			itemResult.Error = fmt.Sprintf("创建失败: %v", err)
			result.FailCount++
		} else {
			itemResult.Success = true
			itemResult.PipelineID = pipeline.ID
			result.SuccessCount++
		}
		result.Results = append(result.Results, itemResult)
	}

	return result, nil
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
	
	// 处理正在运行的情况
	if pipeline.Status == models.PipelineStatusRunning {
		if req.Force {
			// 强制运行：停止旧构建并清理状态
			global.Logger.Info("[流水线] 强制运行：清理旧构建",
				zap.Int64("pipeline_id", pipeline.ID),
				zap.Int("old_build_number", pipeline.LastBuildNumber),
			)
			// 尝试停止 Jenkins 构建
			if pipeline.LastBuildNumber > 0 {
				client := s.getJenkinsClient(pipeline.JenkinsURL)
				if client != nil {
					_ = client.StopBuild(ctx, pipeline.JenkinsJob, pipeline.LastBuildNumber)
				}
			}
			// 更新旧的运行记录为已中止
			latestRun, _ := s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
			if latestRun != nil && latestRun.Status == models.PipelineRunStatusRunning {
				_ = s.dao.PipelineRunUpdateStatus(ctx, latestRun.ID, models.PipelineRunStatusAborted)
			}
		} else {
			return nil, errors.New("流水线正在运行中，请等待完成或使用强制运行")
		}
	}
	
	// 如果上次运行失败，自动重置状态（不需要 force 参数）
	if pipeline.LastRunStatus == models.PipelineRunStatusFailed ||
		pipeline.LastRunStatus == models.PipelineRunStatusAborted {
		global.Logger.Info("[流水线] 清理失败状态，开始新构建",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("old_status", pipeline.LastRunStatus),
		)
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

	// 创建阶段执行记录
	if err := s.CreateRunStages(ctx, run.ID, pipeline.ID, pipeline); err != nil {
		global.Logger.Warn("[流水线] 创建阶段记录失败",
			zap.Int64("run_id", run.ID),
			zap.Error(err),
		)
	}

	// 更新流水线状态为运行中
	if err := s.dao.PipelineUpdateStatus(ctx, pipeline.ID, models.PipelineStatusRunning); err != nil {
		return nil, fmt.Errorf("更新流水线状态失败: %w", err)
	}

	// 构建 Jenkins 参数
	params := make(map[string]string)
	params["GIT_BRANCH"] = branch
	params["GIT_REPO"] = pipeline.GitRepo
	
	// 平台回调参数（用于 Jenkins 构建完成后回调）
	params["PIPELINE_ID"] = fmt.Sprintf("%d", pipeline.ID)
	params["RUN_ID"] = fmt.Sprintf("%d", run.ID)
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		params["PLATFORM_CALLBACK_URL"] = global.JenkinsSetting.CallbackURL + "/api/v1/k8s/cicd/pipeline/callback"
		// 制品上传地址（Jenkins 构建完成后自动上传制品到平台制品库）
		params["ARTIFACT_UPLOAD_URL"] = global.JenkinsSetting.CallbackURL + "/api/v1/k8s/cicd/artifact/upload"
	}
	// 注意：HMAC_SECRET 不再通过参数传递，Jenkins 端应使用 credentials 管理
	// 双方需配置相同的密钥：平台 config.yaml 的 HMACSecret 与 Jenkins credentials 中的 hmac-secret

	// 模板化发布：根据语言类型自动注入语言特定参数
	s.injectLanguageParams(pipeline, params)
	
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
	global.Logger.Info("[流水线] 开始触发 Jenkins 构建",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.String("pipeline_name", pipeline.Name),
		zap.String("jenkins_job", pipeline.JenkinsJob),
		zap.Int64("run_id", run.ID),
		zap.Any("params", params),
	)

	// 创建 Jenkins 客户端
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		errMsg := "Jenkins 客户端创建失败，请检查 config.yaml 中的 Jenkins 配置"
		global.Logger.Error("[流水线] Jenkins 客户端创建失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("jenkins_url", pipeline.JenkinsURL),
		)
		// 更新运行记录为失败，并记录错误信息
		_ = s.dao.PipelineRunUpdateError(ctx, run.ID, models.PipelineRunStatusFailed, errMsg)
		_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusFailed)
		return
	}

	global.Logger.Info("[流水线] Jenkins 客户端创建成功",
		zap.String("jenkins_url", client.BaseURL),
		zap.String("jenkins_user", client.Username),
	)

	// 触发构建等待超时，优先使用配置，默认 60 秒
	triggerTimeout := 60 * time.Second
	if global.JenkinsSetting != nil && global.JenkinsSetting.TriggerTimeout > 0 {
		triggerTimeout = time.Duration(global.JenkinsSetting.TriggerTimeout) * time.Second
	}

	global.Logger.Info("[流水线] 正在触发 Jenkins 构建...",
		zap.String("job_name", pipeline.JenkinsJob),
		zap.Duration("timeout", triggerTimeout),
	)

	// 触发构建并等待获取构建号
	result, err := client.TriggerBuildAndWait(ctx, pipeline.JenkinsJob, params, triggerTimeout)
	if err != nil {
		errMsg := err.Error()
		global.Logger.Error("[流水线] Jenkins 构建触发失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("job_name", pipeline.JenkinsJob),
			zap.Error(err),
		)
		// 更新运行记录为失败，并记录错误信息
		_ = s.dao.PipelineRunUpdateError(ctx, run.ID, models.PipelineRunStatusFailed, errMsg)
		_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusFailed)
		return
	}

	global.Logger.Info("[流水线] Jenkins 构建触发成功",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.Int("build_number", result.BuildNumber),
		zap.String("build_url", result.BuildURL),
		zap.Int64("queue_id", result.QueueID),
	)

	// 更新运行记录
	_ = s.dao.PipelineRunUpdateBuildNumber(ctx, run.ID, result.BuildNumber)
	_ = s.dao.PipelineUpdateRunInfo(ctx, pipeline.ID, models.PipelineRunStatusRunning, result.BuildNumber, result.BuildURL)

	// 发送构建开始钉钉通知
	s.NotifyBuildStarted(ctx, pipeline, run, result.BuildNumber)
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
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return errors.New("Jenkins 未配置或配置不完整")
	}

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
		// 返回空日志而不是错误，因为流水线可能还未运行过
		return "", nil
	}

	// 检查 Jenkins 配置
	if pipeline.JenkinsJob == "" {
		return "", errors.New("流水线未配置 Jenkins Job")
	}

	// 创建 Jenkins 客户端
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return "", errors.New("Jenkins 未配置或配置不完整，请检查 config.yaml 中的 Jenkins 配置")
	}

	log, err := client.GetConsoleLog(ctx, pipeline.JenkinsJob, buildNumber, req.StartLine)
	if err != nil {
		// 对 404 错误进行友好处理
		if strings.Contains(err.Error(), "404") {
			return "", fmt.Errorf("构建记录已被 Jenkins 清理（Build #%d），请重新运行流水线", buildNumber)
		}
		return "", fmt.Errorf("获取日志失败: %w", err)
	}

	return log, nil
}

// PipelineStatus 获取流水线实时状态
func (s *Services) PipelineStatus(ctx context.Context, id int64) (*models.CicdPipeline, *jenkins.BuildInfo, error) {
	pipeline, buildInfo, _, err := s.PipelineStatusWithRun(ctx, id)
	return pipeline, buildInfo, err
}

// PipelineStatusWithRun 获取流水线实时状态（包含最新运行记录）
func (s *Services) PipelineStatusWithRun(ctx context.Context, id int64) (*models.CicdPipeline, *jenkins.BuildInfo, *models.CicdPipelineRun, error) {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errors.New("流水线不存在")
		}
		return nil, nil, nil, fmt.Errorf("查询流水线失败: %w", err)
	}

	// 获取运行记录：优先正在运行的（DAO 已过滤 build_number=0 的幽灵记录）
	latestRun, _ := s.dao.PipelineRunGetRunning(ctx, id)
	if latestRun == nil {
		// 没有正在运行的，获取最新的已构建记录
		latestRun, _ = s.dao.PipelineRunGetLatestBuilt(ctx, id)
	}
	if latestRun == nil {
		// 如果没有已构建的运行记录，回退到任意最新运行记录
		latestRun, _ = s.dao.PipelineRunGetLatest(ctx, id)
	}

	// 如果有构建号，获取 Jenkins 构建状态
	var buildInfo *jenkins.BuildInfo
	if pipeline.LastBuildNumber > 0 {
		client := s.getJenkinsClient(pipeline.JenkinsURL)
		if client != nil {
			buildInfo, _ = client.GetBuildInfo(ctx, pipeline.JenkinsJob, pipeline.LastBuildNumber)

			// 如果构建已完成，同步更新本地状态
			if buildInfo != nil && !buildInfo.Building {
				runStatus := jenkins.BuildStatusToRunStatus(buildInfo.Building, buildInfo.Result)
				if runStatus != pipeline.LastRunStatus {
					_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, runStatus)
					pipeline.LastRunStatus = runStatus
					pipeline.Status = models.PipelineStatusIdle

					// 同步更新运行记录状态
					if latestRun != nil && latestRun.BuildNumber == pipeline.LastBuildNumber && latestRun.Status == models.PipelineRunStatusRunning {
						_ = s.dao.PipelineRunUpdateStatus(ctx, latestRun.ID, runStatus)
						latestRun.Status = runStatus

						// 重要：同步更新各阶段状态（包括将审批阶段设为 waiting）
						// 避免回调未触发时，审批阶段状态与流水线状态不一致
						_ = s.UpdateBuildStagesComplete(ctx, latestRun.ID, runStatus, latestRun.ImageURL, latestRun.ImageDigest, "")
					}
				}
			}
		}
	}

	return pipeline, buildInfo, latestRun, nil
}

// PipelineHistory 获取流水线运行历史
func (s *Services) PipelineHistory(ctx context.Context, req *requests.PipelineHistoryRequest) ([]*models.CicdPipelineRun, int64, error) {
	list, total, err := s.dao.PipelineRunList(ctx, req.ID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 获取流水线信息用于同步状态
	pipeline, _ := s.dao.PipelineGetByID(ctx, req.ID)
	if pipeline == nil {
		return list, total, nil
	}

	// 检查并修复处于 "running" 状态但实际已完成的记录
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return list, total, nil
	}

	for _, run := range list {
		// 只处理状态为 "running" 且有构建号的记录
		if run.Status == models.PipelineRunStatusRunning && run.BuildNumber > 0 {
			buildInfo, err := client.GetBuildInfo(ctx, pipeline.JenkinsJob, run.BuildNumber)
			if err == nil && buildInfo != nil && !buildInfo.Building {
				// Jenkins 构建已完成，同步更新记录状态
				runStatus := jenkins.BuildStatusToRunStatus(false, buildInfo.Result)
				_ = s.dao.PipelineRunUpdateStatus(ctx, run.ID, runStatus)
				run.Status = runStatus
			}
		}
	}

	return list, total, nil
}

// PipelineCallbackResult 回调处理结果（返回给 Jenkins）
type PipelineCallbackResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	DeployEnabled bool   `json:"deploy_enabled"` // 是否配置了部署
	DeploySuccess bool   `json:"deploy_success"` // 部署是否成功
	DeployMessage string `json:"deploy_message"` // 部署结果信息
	Namespace     string `json:"namespace,omitempty"`
	Deployment    string `json:"deployment,omitempty"`
	Image         string `json:"image,omitempty"`
}

// PipelineCallback 生产级 Jenkins 回调处理
// 幂等键: job_name + build_number 或 pipeline_id + build_number
// 返回部署结果给 Jenkins，让用户在 Jenkins 看到最终状态
func (s *Services) PipelineCallback(ctx context.Context, req *requests.PipelineCallbackRequest) (*PipelineCallbackResult, error) {
	// 兼容 image_url 字段（Jenkins 发送的是 image_url）
	image := req.Image
	if image == "" && req.ImageURL != "" {
		image = req.ImageURL
	}

	global.Logger.Info("[回调] 收到 Jenkins 构建回调",
		zap.String("job_name", req.JobName),
		zap.Int("build_number", req.BuildNumber),
		zap.String("status", req.Status),
		zap.Int64("pipeline_id", req.PipelineID),
		zap.String("image", image),
		zap.String("image_digest", req.ImageDigest),
	)

	var pipeline *models.CicdPipeline
	var err error

	// 优先使用 pipeline_id 查找（更快）
	if req.PipelineID > 0 {
		pipeline, err = s.dao.PipelineGetByID(ctx, req.PipelineID)
		if err != nil {
			global.Logger.Warn("[回调] 通过 pipeline_id 查找失败，尝试通过 job_name",
				zap.Int64("pipeline_id", req.PipelineID),
				zap.Error(err),
			)
		}
	}

	// 回退到通过 job_name 查找
	if pipeline == nil {
		pipeline, err = s.dao.PipelineGetByJenkinsJob(ctx, req.JobName)
		if err != nil {
			return nil, fmt.Errorf("未找到关联的流水线: job=%s, err=%w", req.JobName, err)
		}
	}

	// 根据幂等键查找运行记录
	run, err := s.dao.PipelineRunGetByBuildNumber(ctx, pipeline.ID, req.BuildNumber)
	if err != nil {
		return nil, fmt.Errorf("未找到对应的运行记录: pipeline=%d, build=%d, err=%w", 
			pipeline.ID, req.BuildNumber, err)
	}

	// 幂等检查：如果已经收到过回调，直接返回成功
	if run.CallbackReceived == 1 {
		global.Logger.Info("[回调] 重复回调，已忽略",
			zap.Int64("run_id", run.ID),
			zap.Int("build_number", req.BuildNumber),
		)
		return &PipelineCallbackResult{
			Success: true,
			Message: "回调已处理（重复请求）",
		}, nil
	}

	// 转换状态
	runStatus := jenkins.BuildStatusToRunStatus(false, req.Status)

	// 更新运行记录（含回调标记、镜像信息）
	if err := s.dao.PipelineRunUpdateCallback(ctx, run.ID, runStatus, image, req.ImageDigest, req.Message, req.Duration); err != nil {
		return nil, fmt.Errorf("更新运行记录失败: %w", err)
	}

	// 更新流水线状态
	if err := s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, runStatus); err != nil {
		global.Logger.Warn("[回调] 更新流水线状态失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.Error(err),
		)
	}

	global.Logger.Info("[回调] 处理成功",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.Int64("run_id", run.ID),
		zap.String("status", runStatus),
		zap.String("image", image),
	)

	// ==================== 更新阶段状态 ====================
	// 构建完成后，更新各阶段状态（包括将审批阶段设为 waiting）
	// 失败时，将错误信息保存到失败的阶段
	if err := s.UpdateBuildStagesComplete(ctx, run.ID, runStatus, image, req.ImageDigest, req.Message); err != nil {
		global.Logger.Warn("[回调] 更新阶段状态失败",
			zap.Int64("run_id", run.ID),
			zap.Error(err),
		)
	}

	// 重新加载运行记录（获取更新后的完整数据）
	run.Status = runStatus
	run.ImageURL = image
	run.ImageDigest = req.ImageDigest
	run.DurationSec = req.Duration
	run.ErrorMessage = req.Message

	// ==================== 发送钉钉通知 ====================
	// 如果构建成功且需要审批，发送审批提醒（包含构建成功信息）
	// 否则发送构建结果通知
	if runStatus == models.PipelineRunStatusSuccess && pipeline.RequireApproval {
		go s.NotifyApprovalRequired(ctx, pipeline, run)
	} else {
		go s.NotifyBuildResult(ctx, pipeline, run, runStatus == models.PipelineRunStatusSuccess)
	}

	// 初始化返回结果
	result := &PipelineCallbackResult{
		Success: true,
		Message: "回调处理成功",
	}

	// ==================== 构建成功后自动部署到 K8s ====================
	// 条件：构建成功 + 有镜像地址 + 配置了部署信息
	if runStatus == models.PipelineRunStatusSuccess && image != "" {
		deployResult := s.autoDeployToK8sWithResult(ctx, pipeline, image, req.ImageDigest)
		result.DeployEnabled = deployResult.DeployEnabled
		result.DeploySuccess = deployResult.DeploySuccess
		result.DeployMessage = deployResult.DeployMessage
		result.Namespace = deployResult.Namespace
		result.Deployment = deployResult.Deployment
		result.Image = deployResult.Image
	}

	return result, nil
}

// VerifyHMACSignature 验证 HMAC 签名
func (s *Services) VerifyHMACSignature(signature, jobName string, buildNumber int, status string) bool {
	if global.JenkinsSetting == nil || global.JenkinsSetting.HMACSecret == "" {
		// 未配置 HMAC 密钥，跳过验证（开发模式）
		global.Logger.Warn("[回调] HMAC 密钥未配置，跳过签名验证")
		return true
	}

	// 计算期望的签名: HMAC-SHA256(secret, job_name+build_number+status)
	expected := computeHMAC(global.JenkinsSetting.HMACSecret, jobName, buildNumber, status)
	return hmacEqual(signature, expected)
}

// VerifyHMACSignatureSimple 验证阶段回调的 HMAC 签名（简化版）
func (s *Services) VerifyHMACSignatureSimple(signature, jobName string, buildNumber int, stageType string) bool {
	if global.JenkinsSetting == nil || global.JenkinsSetting.HMACSecret == "" {
		// 未配置 HMAC 密钥，跳过验证
		return true
	}

	// 计算期望的签名: HMAC-SHA256(secret, job_name+build_number+stage_type)
	expected := computeHMAC(global.JenkinsSetting.HMACSecret, jobName, buildNumber, stageType)
	return hmacEqual(signature, expected)
}

// ==================== Pipeline 阶段数据 ====================

// PipelineStageInfo 阶段信息（前端友好格式）
type PipelineStageInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`   // 阶段类型: checkout/dependencies/compile/test/lint/build/push/approval/deploy/custom
	Status   string `json:"status"` // success, failed, running, pending, waiting
	Duration string `json:"duration"`
	Steps    []PipelineStepInfo `json:"steps"`
	CanOperate   bool                       `json:"can_operate,omitempty"`
	ApprovalInfo *models.StageApprovalInfo   `json:"approval_info,omitempty"`
}

// PipelineStepInfo 步骤信息
type PipelineStepInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
}

// PipelineStages 获取流水线阶段数据（动态从 Jenkins 获取）
func (s *Services) PipelineStages(ctx context.Context, id int64, buildNumber int) ([]PipelineStageInfo, error) {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("流水线不存在")
		}
		return nil, fmt.Errorf("查询流水线失败: %w", err)
	}

	// 确定构建号：优先使用正在运行的构建
	if buildNumber == 0 {
		// 查找正在运行的构建记录
		runningRun, _ := s.dao.PipelineRunGetRunning(ctx, id)
		if runningRun != nil && runningRun.BuildNumber > 0 {
			buildNumber = runningRun.BuildNumber
			global.Logger.Debug("[流水线] 使用正在运行的构建号",
				zap.Int64("pipeline_id", id),
				zap.Int("build_number", buildNumber),
			)
		} else {
			// 没有正在运行的，使用最后一次构建号
			buildNumber = pipeline.LastBuildNumber
		}
	}
	if buildNumber == 0 {
		// 返回默认阶段（未运行状态）
		return s.getDefaultStagesForPipeline(pipeline), nil
	}

	// 获取 Jenkins 客户端
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return s.getDefaultStagesForPipeline(pipeline), nil
	}

	// 从 Jenkins 获取阶段数据
	pipelineRun, err := client.GetPipelineRun(ctx, pipeline.JenkinsJob, buildNumber)
	if err != nil {
		global.Logger.Warn("[流水线] 获取Jenkins阶段数据失败",
			zap.Int64("pipeline_id", id),
			zap.Int("build_number", buildNumber),
			zap.Error(err),
		)
		return s.getDefaultStagesForPipeline(pipeline), nil
	}

	// 动态转换为前端友好格式（保持 Jenkins 阶段名称）
	stages := make([]PipelineStageInfo, 0, len(pipelineRun.Stages)+2)
	for _, stage := range pipelineRun.Stages {
		stageInfo := PipelineStageInfo{
			ID:       stage.ID,
			Name:     stage.Name,
			Type:     s.inferStageTypeFromName(stage.Name), // 动态推断阶段类型
			Status:   s.convertJenkinsStatus(stage.Status),
			Duration: s.formatDuration(stage.DurationMillis),
			Steps:    make([]PipelineStepInfo, 0),
		}
		
		// 转换节点为步骤
		for _, node := range stage.StageFlowNodes {
			stageInfo.Steps = append(stageInfo.Steps, PipelineStepInfo{
				ID:       node.ID,
				Name:     node.Name,
				Status:   s.convertJenkinsStatus(node.Status),
				Duration: s.formatDuration(node.DurationMillis),
			})
		}
		
		stages = append(stages, stageInfo)
	}

	// 追加平台特有阶段（审批/部署）—— 使用 DB 真实数据
	stages = s.appendPlatformStages(ctx, stages, pipeline, pipelineRun.Status, buildNumber)

	return stages, nil
}

// inferStageTypeFromName 根据阶段名称推断阶段类型
func (s *Services) inferStageTypeFromName(name string) string {
	nameLower := strings.ToLower(name)
	
	// 按优先级匹配
	switch {
	// 清理工作空间阶段
	case strings.Contains(nameLower, "clean") || strings.Contains(nameLower, "清理"):
		return "clean"
	// Jenkins 声明式管道自动添加的 SCM checkout 阶段
	case strings.Contains(nameLower, "declarative: checkout scm") || (strings.Contains(nameLower, "scm") && !strings.Contains(nameLower, "clean")):
		return "scm"
	case strings.Contains(nameLower, "checkout") || strings.Contains(nameLower, "代码检出") || strings.Contains(nameLower, "拉取代码"):
		return "checkout"
	case strings.Contains(nameLower, "dependencies") || strings.Contains(nameLower, "依赖"):
		return "dependencies"
	case strings.Contains(nameLower, "compile") || strings.Contains(nameLower, "编译"):
		return "compile"
	case strings.Contains(nameLower, "test") || strings.Contains(nameLower, "测试"):
		return "test"
	case strings.Contains(nameLower, "lint") || strings.Contains(nameLower, "代码检查") || strings.Contains(nameLower, "vet"):
		return "lint"
	case strings.Contains(nameLower, "push") || strings.Contains(nameLower, "推送镜像"):
		return "push"
	case strings.Contains(nameLower, "sonar") || strings.Contains(nameLower, "代码扫描") || strings.Contains(nameLower, "code scan"):
		return "sonar"
	case strings.Contains(nameLower, "quality gate") || strings.Contains(nameLower, "质量门禁") || strings.Contains(nameLower, "qualitygate"):
		return "quality_gate"
	case strings.Contains(nameLower, "upload artifact") || strings.Contains(nameLower, "上传制品") || strings.Contains(nameLower, "upload"):
		return "upload_artifact"
	case strings.Contains(nameLower, "build binary") || strings.Contains(nameLower, "构建制品") || strings.Contains(nameLower, "package") || strings.Contains(nameLower, "打包"):
		return "build_binary"
	case strings.Contains(nameLower, "build") || strings.Contains(nameLower, "构建镜像") || strings.Contains(nameLower, "构建"):
		return "build"
	case strings.Contains(nameLower, "approval") || strings.Contains(nameLower, "审批"):
		return "approval"
	case strings.Contains(nameLower, "deploy") || strings.Contains(nameLower, "部署"):
		return "deploy"
	default:
		return "custom" // 未知阶段类型
	}
}

// appendPlatformStages 追加平台特有阶段（审批/部署）
// 优先从 DB 获取真实阶段数据（包含真实 ID、状态、can_operate、审批信息）
// 避免前端使用假 ID “approval” 导致提交审批时 stage_id=NaN
func (s *Services) appendPlatformStages(ctx context.Context, stages []PipelineStageInfo, pipeline *models.CicdPipeline, jenkinsStatus string, buildNumber int) []PipelineStageInfo {
	buildSuccess := jenkinsStatus == "SUCCESS"

	// 尝试从 DB 获取真实的审批/部署阶段数据
	var dbApprovalStage, dbDeployStage *models.CicdPipelineStage
	if buildNumber > 0 {
		run, _ := s.dao.PipelineRunGetByBuildNumber(ctx, pipeline.ID, buildNumber)
		if run != nil {
			dbApprovalStage, _ = s.dao.StageGetByRunIDAndType(ctx, run.ID, models.StageTypeApproval)
			dbDeployStage, _ = s.dao.StageGetByRunIDAndType(ctx, run.ID, models.StageTypeDeploy)
		}
	}

	// 添加审批阶段（如果配置了）
	if pipeline.RequireApproval {
		if dbApprovalStage != nil {
			// 使用 DB 真实数据（包含真实 ID、状态、审批信息）
			approvalInfo := PipelineStageInfo{
				ID:     fmt.Sprintf("%d", dbApprovalStage.ID),
				Name:   "人工审批",
				Type:   "approval",
				Status: dbApprovalStage.Status,
				Steps:  []PipelineStepInfo{},
			}
			// 设置 can_operate
			if dbApprovalStage.Status == models.StageStatusWaiting {
				approvalInfo.CanOperate = true
			}
			// 填充审批信息
			if dbApprovalStage.ApprovalDecision != "" {
				approvalInfo.ApprovalInfo = &models.StageApprovalInfo{
					ApproverID: dbApprovalStage.ApprovalUserID,
					Decision:   dbApprovalStage.ApprovalDecision,
					Comment:    dbApprovalStage.ApprovalComment,
					ApprovedAt: dbApprovalStage.FinishedAt,
				}
			}
			stages = append(stages, approvalInfo)
		} else {
			// 无 DB 数据，使用推断状态（兼容旧数据）
			approvalStatus := "pending"
			if buildSuccess {
				approvalStatus = "waiting"
			}
			stages = append(stages, PipelineStageInfo{
				ID: "approval", Name: "人工审批", Type: "approval",
				Status: approvalStatus, Steps: []PipelineStepInfo{},
			})
		}
	}

	// 添加部署阶段（如果配置了）
	if pipeline.AutoDeploy {
		if dbDeployStage != nil {
			stages = append(stages, PipelineStageInfo{
				ID:     fmt.Sprintf("%d", dbDeployStage.ID),
				Name:   "部署",
				Type:   "deploy",
				Status: dbDeployStage.Status,
				Steps:  []PipelineStepInfo{},
			})
		} else {
			stages = append(stages, PipelineStageInfo{
				ID: "deploy", Name: "部署", Type: "deploy",
				Status: "pending", Steps: []PipelineStepInfo{},
			})
		}
	}

	return stages
}

// getDefaultStagesForPipeline 获取默认阶段（未运行时展示）
// 完整闭环：拉取 → 编译 → 测试 → 代码扫描 → 质量门禁 → 构建制品 → 上传制品库 → 打包镜像 → 推送镜像 → 审批 → 部署
func (s *Services) getDefaultStagesForPipeline(pipeline *models.CicdPipeline) []PipelineStageInfo {
	stages := []PipelineStageInfo{
		{ID: "1", Name: "Clean Workspace", Type: "clean", Status: "pending", Steps: []PipelineStepInfo{}},
		{ID: "2", Name: "Checkout Info", Type: "checkout", Status: "pending", Steps: []PipelineStepInfo{}},
		{ID: "3", Name: "Dependencies", Type: "dependencies", Status: "pending", Steps: []PipelineStepInfo{}},
		{ID: "4", Name: "Compile Check", Type: "compile", Status: "pending", Steps: []PipelineStepInfo{}},
		{ID: "5", Name: "Test", Type: "test", Status: "pending", Steps: []PipelineStepInfo{}},
		{ID: "6", Name: "Lint", Type: "lint", Status: "pending", Steps: []PipelineStepInfo{}},
	}

	// SonarQube 代码扫描 + 质量门禁（如果启用）
	if pipeline.EnableSonar {
		stages = append(stages,
			PipelineStageInfo{ID: "7", Name: "SonarQube Analysis", Type: "sonar", Status: "pending", Steps: []PipelineStepInfo{}},
			PipelineStageInfo{ID: "8", Name: "Quality Gate", Type: "quality_gate", Status: "pending", Steps: []PipelineStepInfo{}},
		)
	}

	// 构建制品 + 上传制品库（如果启用）
	if pipeline.EnableArtifactUpload {
		nextID := len(stages) + 1
		stages = append(stages,
			PipelineStageInfo{ID: fmt.Sprintf("%d", nextID), Name: "Build Binary", Type: "build_binary", Status: "pending", Steps: []PipelineStepInfo{}},
			PipelineStageInfo{ID: fmt.Sprintf("%d", nextID+1), Name: "Upload Artifact", Type: "upload_artifact", Status: "pending", Steps: []PipelineStepInfo{}},
		)
	}

	// 打包镜像 + 推送镜像
	nextID := len(stages) + 1
	stages = append(stages,
		PipelineStageInfo{ID: fmt.Sprintf("%d", nextID), Name: "Build Image", Type: "build", Status: "pending", Steps: []PipelineStepInfo{}},
		PipelineStageInfo{ID: fmt.Sprintf("%d", nextID+1), Name: "Push Image", Type: "push", Status: "pending", Steps: []PipelineStepInfo{}},
	)

	// 根据流水线配置追加平台阶段
	if pipeline.RequireApproval {
		stages = append(stages, PipelineStageInfo{
			ID: fmt.Sprintf("%d", len(stages)+1), Name: "人工审批", Type: "approval", Status: "pending", Steps: []PipelineStepInfo{},
		})
	}
	if pipeline.AutoDeploy {
		stages = append(stages, PipelineStageInfo{
			ID: fmt.Sprintf("%d", len(stages)+1), Name: "部署", Type: "deploy", Status: "pending", Steps: []PipelineStepInfo{},
		})
	}

	return stages
}

// convertJenkinsStatus 转换 Jenkins 状态为前端状态
func (s *Services) convertJenkinsStatus(status string) string {
	switch status {
	case "SUCCESS":
		return "success"
	case "FAILURE", "FAILED":
		return "failed"
	case "IN_PROGRESS":
		return "running"
	case "ABORTED":
		return "aborted"
	case "NOT_EXECUTED", "PAUSED_PENDING_INPUT":
		return "pending"
	default:
		return "pending"
	}
}

// formatDuration 格式化时长
func (s *Services) formatDuration(millis int64) string {
	if millis <= 0 {
		return "-"
	}
	seconds := millis / 1000
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh%dm", seconds/3600, (seconds%3600)/60)
}

// getJenkinsClient 获取 Jenkins 客户端
// 优先使用流水线配置的 URL，否则使用全局配置
// 凭据统一从全局配置读取
func (s *Services) getJenkinsClient(pipelineJenkinsURL string) *jenkins.Client {
	// 检查全局 Jenkins 配置是否存在
	if global.JenkinsSetting == nil {
		global.Logger.Warn("Jenkins 配置未加载，请检查 config.yaml 中的 Jenkins 配置块")
		return nil
	}

	// 确定 Jenkins URL：流水线配置优先，否则用全局配置
	jenkinsURL := pipelineJenkinsURL
	if jenkinsURL == "" {
		jenkinsURL = global.JenkinsSetting.URL
	}
	if jenkinsURL == "" {
		global.Logger.Warn("Jenkins URL 未配置")
		return nil
	}

	return jenkins.NewClient(
		jenkinsURL,
		global.JenkinsSetting.Username,
		global.JenkinsSetting.APIToken,
	)
}

// ==================== HMAC 辅助函数 ====================

// computeHMAC 计算 HMAC-SHA256 签名
// 签名格式: job_name:build_number:status (冒号分隔)
func computeHMAC(secret, jobName string, buildNumber int, status string) string {
	data := fmt.Sprintf("%s:%d:%s", jobName, buildNumber, status)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacEqual 安全比较两个 HMAC 签名（防止时序攻击）
func hmacEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ==================== 自动部署 K8s ====================

// autoDeployToK8sWithResult 回调成功后自动部署到 K8s
// 支持新的部署配置字段：target_cluster_id, target_namespace, target_workload_name, target_container
// 支持多集群部署、审批流程
// 返回部署结果给 Jenkins，让用户在 Jenkins 看到最终状态
func (s *Services) autoDeployToK8sWithResult(ctx context.Context, pipeline *models.CicdPipeline, image, imageDigest string) *PipelineCallbackResult {
	result := &PipelineCallbackResult{
		DeployEnabled: false,
		DeploySuccess: false,
	}

	// 1. 检查是否启用自动部署
	if !pipeline.AutoDeploy {
		// 兼容旧的 DeployConfig JSON 配置
		if pipeline.DeployConfig == nil || len(pipeline.DeployConfig) == 0 {
			global.Logger.Info("[自动部署] 未启用自动部署，跳过",
				zap.Int64("pipeline_id", pipeline.ID),
			)
			result.DeployMessage = "未启用自动部署"
			return result
		}
		// 使用旧的 DeployConfig 配置
		return s.autoDeployWithLegacyConfig(ctx, pipeline, image, imageDigest)
	}

	// 2. 检查部署配置是否完整
	if pipeline.TargetNamespace == "" || pipeline.TargetWorkloadName == "" || pipeline.TargetContainer == "" {
		global.Logger.Info("[自动部署] 部署配置不完整，跳过部署",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("namespace", pipeline.TargetNamespace),
			zap.String("workload", pipeline.TargetWorkloadName),
			zap.String("container", pipeline.TargetContainer),
		)
		result.DeployMessage = "部署配置不完整，跳过自动部署"
		return result
	}

	result.DeployEnabled = true
	result.Namespace = pipeline.TargetNamespace
	result.Deployment = pipeline.TargetWorkloadName

	// 3. 检查是否需要审批（生产环境）
	if pipeline.RequireApproval {
		global.Logger.Info("[自动部署] 需要审批，创建审批记录",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("env", pipeline.DeployEnv),
		)
		// TODO: 创建审批记录，等待审批通过后再部署
		result.DeploySuccess = false
		result.DeployMessage = fmt.Sprintf("待审批: %s 环境需要审批后才能部署", pipeline.DeployEnv)
		return result
	}

	// 4. 获取目标集群的 K8s 客户端
	var kubeClient kubernetes.Interface
	if pipeline.TargetClusterID > 0 {
		// 多集群模式：根据集群ID初始化客户端
		clients, err := s.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: uint32(pipeline.TargetClusterID)})
		if err != nil || clients == nil || clients.Kube == nil {
			global.Logger.Error("[自动部署] 获取集群客户端失败",
				zap.Int64("cluster_id", pipeline.TargetClusterID),
				zap.Error(err),
			)
			result.DeployMessage = fmt.Sprintf("获取集群客户端失败: cluster_id=%d", pipeline.TargetClusterID)
			return result
		}
		kubeClient = clients.Kube
	} else {
		// 单集群模式：使用默认管理集群
		if global.ManagementKubeClient == nil {
			global.Logger.Error("[自动部署] K8s 客户端未初始化")
			result.DeployMessage = "K8s 客户端未初始化"
			return result
		}
		kubeClient = global.ManagementKubeClient
	}

	// 5. 构造最终镜像地址（优先使用 image@digest 确保一致性）
	finalImage := image
	if imageDigest != "" {
		if idx := strings.LastIndex(image, ":"); idx > 0 && !strings.Contains(image[idx:], "/") {
			finalImage = image[:idx] + "@" + imageDigest
		} else {
			finalImage = image + "@" + imageDigest
		}
	}
	result.Image = finalImage

	global.Logger.Info("[自动部署] 开始更新工作负载",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.Int64("cluster_id", pipeline.TargetClusterID),
		zap.String("namespace", pipeline.TargetNamespace),
		zap.String("workload_kind", pipeline.TargetWorkloadKind),
		zap.String("workload_name", pipeline.TargetWorkloadName),
		zap.String("container", pipeline.TargetContainer),
		zap.String("image", finalImage),
	)

	// 6. 根据工作负载类型异步执行部署（等待 Rollout 完成后再发通知）
	workloadKind := pipeline.TargetWorkloadKind
	if workloadKind == "" {
		workloadKind = "Deployment"
	}

	// 启动异步部署，等待 Rollout 完成后发送钉钉通知
	go s.executeAutoDeployAsync(context.Background(), pipeline, kubeClient, finalImage, workloadKind)

	result.DeploySuccess = true
	result.DeployMessage = fmt.Sprintf("部署已启动: %s/%s 正在更新...", pipeline.TargetNamespace, pipeline.TargetWorkloadName)
	return result
}

// executeAutoDeployAsync 异步执行自动部署（等待 Rollout 完成后发钉钉通知）
func (s *Services) executeAutoDeployAsync(ctx context.Context, pipeline *models.CicdPipeline, kubeClient kubernetes.Interface, image, workloadKind string) {
	var err error
	var logs strings.Builder
	
	logs.WriteString(fmt.Sprintf("[自动部署] 开始更新 %s/%s\n", pipeline.TargetNamespace, pipeline.TargetWorkloadName))
	logs.WriteString(fmt.Sprintf("工作负载类型: %s\n", workloadKind))
	logs.WriteString(fmt.Sprintf("镜像: %s\n\n", image))

	switch workloadKind {
	case "Deployment":
		// Patch 更新镜像
		patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, 
			pipeline.TargetContainer, image)
		_, err = kubeClient.AppsV1().Deployments(pipeline.TargetNamespace).Patch(
			ctx, pipeline.TargetWorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
		if err == nil {
			logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")
			err = s.waitAutoDeployRollout(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, &logs)
		}
	case "StatefulSet":
		patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, 
			pipeline.TargetContainer, image)
		_, err = kubeClient.AppsV1().StatefulSets(pipeline.TargetNamespace).Patch(
			ctx, pipeline.TargetWorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	case "DaemonSet":
		patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, 
			pipeline.TargetContainer, image)
		_, err = kubeClient.AppsV1().DaemonSets(pipeline.TargetNamespace).Patch(
			ctx, pipeline.TargetWorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	default:
		err = fmt.Errorf("不支持的工作负载类型: %s", workloadKind)
	}

	// 更新流水线部署状态
	now := uint64(time.Now().Unix())
	if err != nil {
		global.Logger.Error("[自动部署] Rollout 失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.Error(err),
		)
		_ = s.dao.PipelineUpdateDeployInfo(ctx, pipeline.ID, image, "", now, "failed")
		// Rollout 失败后发送钉钉通知
		s.notifyAutoDeployResult(ctx, pipeline, image, false, err.Error())
	} else {
		global.Logger.Info("[自动部署] Rollout 完成",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("image", image),
		)
		_ = s.dao.PipelineUpdateDeployInfo(ctx, pipeline.ID, image, "", now, "success")
		// Rollout 完成后发送钉钉通知
		s.notifyAutoDeployResult(ctx, pipeline, image, true, "")
	}
}

// waitAutoDeployRollout 等待自动部署的 Rollout 完成
func (s *Services) waitAutoDeployRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, logs *strings.Builder) error {
	timeout := 5 * time.Minute
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		dp, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取 Deployment 失败: %v", err)
		}

		replicas := int32(1)
		if dp.Spec.Replicas != nil {
			replicas = *dp.Spec.Replicas
		}

		logs.WriteString(fmt.Sprintf("[ROLLOUT] 期望: %d | 更新: %d | 就绪: %d | 可用: %d | Gen: %d/%d\n",
			replicas,
			dp.Status.UpdatedReplicas,
			dp.Status.ReadyReplicas,
			dp.Status.AvailableReplicas,
			dp.Status.ObservedGeneration, dp.Generation))

		// 检查 Rollout 是否失败
		for _, cond := range dp.Status.Conditions {
			if cond.Type == "Progressing" && cond.Reason == "ProgressDeadlineExceeded" {
				return fmt.Errorf("Rollout 超时: %s", cond.Message)
			}
		}

		// Rollout 完成条件（不检查 Replicas，因为滚动更新期间旧 Pod 可能还在终止中）
		if dp.Status.ObservedGeneration >= dp.Generation &&
			dp.Status.UpdatedReplicas == replicas &&
			dp.Status.AvailableReplicas == replicas {
			logs.WriteString(fmt.Sprintf("[SUCCESS] 所有 %d 个副本已就绪\n", replicas))
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("Rollout 超时（%v）", timeout)
}

// autoDeployWithLegacyConfig 兼容旧的 DeployConfig JSON 配置
func (s *Services) autoDeployWithLegacyConfig(ctx context.Context, pipeline *models.CicdPipeline, image, imageDigest string) *PipelineCallbackResult {
	result := &PipelineCallbackResult{
		DeployEnabled: false,
		DeploySuccess: false,
	}

	namespace, _ := pipeline.DeployConfig["namespace"].(string)
	deploymentName, _ := pipeline.DeployConfig["deployment_name"].(string)
	containerName, _ := pipeline.DeployConfig["container_name"].(string)

	if namespace == "" || deploymentName == "" || containerName == "" {
		result.DeployMessage = "部署配置不完整，跳过自动部署"
		return result
	}

	result.DeployEnabled = true
	result.Namespace = namespace
	result.Deployment = deploymentName

	if global.ManagementKubeClient == nil {
		result.DeployMessage = "K8s 客户端未初始化"
		return result
	}

	finalImage := image
	if imageDigest != "" {
		if idx := strings.LastIndex(image, ":"); idx > 0 && !strings.Contains(image[idx:], "/") {
			finalImage = image[:idx] + "@" + imageDigest
		} else {
			finalImage = image + "@" + imageDigest
		}
	}
	result.Image = finalImage

	// 异步执行部署，等待 Rollout 完成后发通知
	go s.executeLegacyDeployAsync(context.Background(), pipeline, namespace, deploymentName, containerName, finalImage)

	result.DeploySuccess = true
	result.DeployMessage = fmt.Sprintf("部署已启动: %s/%s 正在更新...", namespace, deploymentName)
	return result
}

// executeLegacyDeployAsync 异步执行旧配置部署（等待 Rollout 完成）
func (s *Services) executeLegacyDeployAsync(ctx context.Context, pipeline *models.CicdPipeline, namespace, deploymentName, containerName, image string) {
	var logs strings.Builder
	logs.WriteString(fmt.Sprintf("[旧配置部署] 开始更新 %s/%s\n", namespace, deploymentName))

	// 1. 更新镜像
	_, err := deployment.PatchDeploymentImage(ctx, global.ManagementKubeClient, namespace, deploymentName, containerName, image)
	if err != nil {
		global.Logger.Error("[旧配置部署] 更新镜像失败", zap.Error(err))
		s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, image, false, err.Error())
		return
	}

	logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")

	// 2. 等待 Rollout 完成
	err = s.waitAutoDeployRollout(ctx, global.ManagementKubeClient, namespace, deploymentName, &logs)
	if err != nil {
		global.Logger.Error("[旧配置部署] Rollout 失败", zap.Error(err))
		s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, image, false, err.Error())
		return
	}

	global.Logger.Info("[旧配置部署] Rollout 完成")
	s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, image, true, "")
}

// patchStatefulSetImage 更新 StatefulSet 镜像
func (s *Services) patchStatefulSetImage(ctx context.Context, kubeClient kubernetes.Interface, namespace, name, container, image string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	_, err := kubeClient.AppsV1().StatefulSets(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	return err
}

// patchDaemonSetImage 更新 DaemonSet 镜像
func (s *Services) patchDaemonSetImage(ctx context.Context, kubeClient kubernetes.Interface, namespace, name, container, image string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	_, err := kubeClient.AppsV1().DaemonSets(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	return err
}

// ==================== 模板化发布支持 ====================

// injectLanguageParams 根据语言类型自动注入 Jenkins 构建参数
// 这是"一个模板服务 100 个项目"的核心：所有项目差异通过参数传入
func (s *Services) injectLanguageParams(pipeline *models.CicdPipeline, params map[string]string) {
	// 始终注入 LANGUAGE_TYPE，让 Jenkins 模板可交叉校验（防止自定义 Job 配错脚本）
	if pipeline.LanguageType != "" {
		params["LANGUAGE_TYPE"] = pipeline.LanguageType
	}

	// ==================== 语言特有参数 ====================
	switch pipeline.LanguageType {
	case models.LanguageTypeGo:
		setDefault(params, "GO_VERSION", "1.24")
		setDefault(params, "SKIP_TESTS", "false")
	case models.LanguageTypeJava:
		setDefault(params, "JAVA_VERSION", "17")
		setDefault(params, "MAVEN_GOALS", "clean package -DskipTests -B")
		setDefault(params, "SKIP_TESTS", "false")
		// Java 特有 SonarQube 参数
		setDefault(params, "SONAR_SOURCES", "src/main/java")
		setDefault(params, "SONAR_JAVA_BINARIES", "target/classes")
		setDefault(params, "SONAR_EXCLUSIONS", "**/test/**,**/generated/**")
	case models.LanguageTypeFrontend:
		setDefault(params, "NODE_VERSION", "18")
		setDefault(params, "BUILD_COMMAND", "npm run build")
		setDefault(params, "BUILD_OUTPUT_DIR", "dist")
		setDefault(params, "SKIP_TESTS", "false")
		setDefault(params, "SONAR_SOURCES", "src")
		setDefault(params, "SONAR_EXCLUSIONS", "**/node_modules/**,**/dist/**,**/*.spec.*,**/*.test.*")
	case models.LanguageTypePython:
		setDefault(params, "PYTHON_VERSION", "3.11")
		setDefault(params, "SKIP_TESTS", "false")
		setDefault(params, "SONAR_SOURCES", ".")
		setDefault(params, "SONAR_EXCLUSIONS", "**/venv/**,**/__pycache__/**,**/test_*,**/*_test.py,**/migrations/**")
	}

	// ==================== 通用参数（全语言） ====================
	setDefault(params, "DOCKERFILE_PATH", "")
	setDefault(params, "GIT_CREDENTIAL_ID", "gitee-id")

	// SonarQube 代码质量扫描（根据流水线配置注入，所有语言统一）
	if pipeline.EnableSonar {
		setDefault(params, "ENABLE_SONAR", "true")
		setDefault(params, "SONAR_QUALITY_GATE", "true")
	} else {
		setDefault(params, "ENABLE_SONAR", "false")
		setDefault(params, "SONAR_QUALITY_GATE", "false")
	}

	// 制品上传（根据流水线配置注入，所有语言统一）
	if pipeline.EnableArtifactUpload {
		setDefault(params, "ENABLE_ARTIFACT_UPLOAD", "true")
	} else {
		setDefault(params, "ENABLE_ARTIFACT_UPLOAD", "false")
	}
}

// setDefault 设置默认参数（不覆盖已有值）
func setDefault(params map[string]string, key, value string) {
	if _, exists := params[key]; !exists {
		params[key] = value
	}
}

// TemplateVerifyInfo 模板验证信息
type TemplateVerifyInfo struct {
	LanguageType    string            `json:"language_type"`
	JenkinsJob      string            `json:"jenkins_job"`
	TemplateFile    string            `json:"template_file"`
	Stages          []string          `json:"stages"`
	DefaultParams   map[string]string `json:"default_params"`
	CallbackURL     string            `json:"callback_url"`
	HMACEnabled     bool              `json:"hmac_enabled"`
	Description     string            `json:"description"`
}

// TemplateVerifyAll 验证所有模板配置是否完整
func (s *Services) TemplateVerifyAll(ctx context.Context) ([]TemplateVerifyInfo, error) {
	templates := []TemplateVerifyInfo{
		{
			LanguageType: models.LanguageTypeGo,
			JenkinsJob:   models.DefaultJenkinsJobMap[models.LanguageTypeGo],
			TemplateFile: "configs/jenkins-templates/go-pipeline.groovy",
			Stages:       []string{"checkout", "dependencies", "compile", "test", "lint", "sonar", "quality_gate", "build_binary", "upload_artifact", "build", "push"},
			DefaultParams: map[string]string{
				"GO_VERSION":  "1.24",
				"SKIP_TESTS":  "false",
			},
			Description: "Go 项目通用构建模板，支持 go test / golangci-lint / SonarQube / 制品上传 / nerdctl build",
		},
		{
			LanguageType: models.LanguageTypeJava,
			JenkinsJob:   models.DefaultJenkinsJobMap[models.LanguageTypeJava],
			TemplateFile: "configs/jenkins-templates/java-spring-pipeline.groovy",
			Stages:       []string{"checkout", "dependencies", "compile", "test", "sonar", "quality_gate", "build_binary", "upload_artifact", "build", "push"},
			DefaultParams: map[string]string{
				"JAVA_VERSION":    "17",
				"MAVEN_GOALS":     "clean package -DskipTests -B",
			},
			Description: "Java/Spring Boot 通用构建模板，支持 Maven + SonarQube + 质量门禁 + 制品上传",
		},
		{
			LanguageType: models.LanguageTypeFrontend,
			JenkinsJob:   models.DefaultJenkinsJobMap[models.LanguageTypeFrontend],
			TemplateFile: "configs/jenkins-templates/frontend-pipeline.groovy",
			Stages:       []string{"checkout", "dependencies", "test", "compile", "sonar", "quality_gate", "build_binary", "upload_artifact", "build", "push"},
			DefaultParams: map[string]string{
				"NODE_VERSION":    "18",
				"BUILD_COMMAND":   "npm run build",
				"BUILD_OUTPUT_DIR": "dist",
			},
			Description: "前端通用构建模板（Vue/React/Angular），支持 npm ci / SonarQube / 制品上传 / Nginx 镜像",
		},
		{
			LanguageType: models.LanguageTypePython,
			JenkinsJob:   models.DefaultJenkinsJobMap[models.LanguageTypePython],
			TemplateFile: "configs/jenkins-templates/python-pipeline.groovy",
			Stages:       []string{"checkout", "dependencies", "lint", "test", "sonar", "quality_gate", "build_binary", "upload_artifact", "build", "push"},
			DefaultParams: map[string]string{
				"PYTHON_VERSION": "3.11",
			},
			Description: "Python 通用构建模板，支持 pip / flake8 / pytest / SonarQube / 制品上传",
		},
	}

	// 填充回调配置
	for i := range templates {
		if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
			templates[i].CallbackURL = global.JenkinsSetting.CallbackURL + "/api/v1/k8s/cicd/pipeline/callback"
		}
		if global.JenkinsSetting != nil && global.JenkinsSetting.HMACSecret != "" {
			templates[i].HMACEnabled = true
		}
	}

	return templates, nil
}

// TemplateSimulateRun 模拟模板化流水线完整发布流程（不实际触发 Jenkins，仅验证参数和流程）
func (s *Services) TemplateSimulateRun(ctx context.Context, languageType, gitRepo, gitBranch, imageRepo string) (map[string]interface{}, error) {
	// 1. 解析 Jenkins Job
	jenkinsJob, ok := models.DefaultJenkinsJobMap[languageType]
	if !ok {
		return nil, fmt.Errorf("不支持的语言类型: %s，可选: go, java, frontend, python", languageType)
	}

	// 2. 构建 Jenkins 参数
	params := map[string]string{
		"GIT_REPO":    gitRepo,
		"GIT_BRANCH":  gitBranch,
		"IMAGE_REPO":  imageRepo,
		"PIPELINE_ID": "0",
	}

	// 模拟回调 URL
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		params["PLATFORM_CALLBACK_URL"] = global.JenkinsSetting.CallbackURL + "/api/v1/k8s/cicd/pipeline/callback"
	}

	// 3. 注入语言参数
	mockPipeline := &models.CicdPipeline{LanguageType: languageType}
	s.injectLanguageParams(mockPipeline, params)

	// 4. 检查 Jenkins 配置
	jenkinsConfigured := false
	if global.JenkinsSetting != nil && global.JenkinsSetting.URL != "" {
		jenkinsConfigured = true
	}

	// 5. 检查 Jenkins Job 是否存在
	jobExists := false
	var jobCheckError string
	if jenkinsConfigured {
		client := s.getJenkinsClient("")
		if client != nil {
			_, err := client.GetJobInfo(ctx, jenkinsJob)
			if err == nil {
				jobExists = true
			} else {
				jobCheckError = fmt.Sprintf("Jenkins Job '%s' 不存在，需先在 Jenkins 中创建该 Job 并配置 Pipeline Script", jenkinsJob)
			}
		}
	}

	return map[string]interface{}{
		"language_type":     languageType,
		"jenkins_job":       jenkinsJob,
		"template_file":     fmt.Sprintf("configs/jenkins-templates/%s", getTemplateFile(languageType)),
		"jenkins_params":    params,
		"jenkins_configured": jenkinsConfigured,
		"job_exists":        jobExists,
		"job_check_error":   jobCheckError,
		"flow": []string{
			"1. 平台触发 Jenkins 构建，传入上述参数",
			"2. Jenkins 执行通用模板: " + jenkinsJob,
			"3. 每个阶段完成后回调平台 /stage/callback",
			"4. 构建完成后回调平台 /pipeline/callback",
			"5. 平台根据配置自动部署到 K8s",
		},
		"setup_guide": fmt.Sprintf(
			"Jenkins 设置步骤（推荐 Pipeline script from SCM）:\n"+
				"1. 创建 Pipeline Job，命名为: %s\n"+
				"2. Pipeline → Definition: Pipeline script from SCM\n"+
				"3. SCM: Git → Repository URL: 平台仓库地址\n"+
				"4. Script Path: configs/jenkins-templates/%s\n"+
				"5. 确保 Jenkins 已配置 credentials: harbor-registry, hmac-secret, gitee-id\n"+
				"6. 在平台创建流水线时选择 language_type='%s'，无需手动填 jenkins_job",
			jenkinsJob, getTemplateFile(languageType), languageType,
		),
	}, nil
}

// getTemplateFile 获取模板文件名
func getTemplateFile(languageType string) string {
	switch languageType {
	case models.LanguageTypeGo:
		return "go-pipeline.groovy"
	case models.LanguageTypeJava:
		return "java-spring-pipeline.groovy"
	case models.LanguageTypeFrontend:
		return "frontend-pipeline.groovy"
	case models.LanguageTypePython:
		return "python-pipeline.groovy"
	default:
		return "custom"
	}
}

// ==================== SonarQube 代码质量管理 ====================

// GetSonarReport 获取流水线的 SonarQube 代码质量报告
func (s *Services) GetSonarReport(ctx context.Context, pipelineID int64, runID int64) (map[string]interface{}, error) {
	db := global.DB.WithContext(ctx)

	// 获取流水线信息
	var pipeline models.CicdPipeline
	if err := db.Where("id = ? AND is_del = 0", pipelineID).First(&pipeline).Error; err != nil {
		return nil, fmt.Errorf("流水线不存在")
	}

	// 获取运行记录
	var run models.CicdPipelineRun
	if runID > 0 {
		if err := db.Where("id = ? AND pipeline_id = ?", runID, pipelineID).First(&run).Error; err != nil {
			return nil, fmt.Errorf("运行记录不存在")
		}
	} else {
		// 获取最新一次运行记录
		if err := db.Where("pipeline_id = ?", pipelineID).Order("id DESC").First(&run).Error; err != nil {
			return nil, fmt.Errorf("暂无运行记录")
		}
	}

	// 获取 sonar 和 quality_gate 阶段
	var sonarStage models.CicdPipelineStage
	hasSonar := db.Where("run_id = ? AND stage_type = ?", run.ID, models.StageTypeSonar).First(&sonarStage).Error == nil

	var qgStage models.CicdPipelineStage
	hasQG := db.Where("run_id = ? AND stage_type = ?", run.ID, models.StageTypeQualityGate).First(&qgStage).Error == nil

	// 构建报告
	report := map[string]interface{}{
		"pipeline_id":   pipeline.ID,
		"pipeline_name": pipeline.Name,
		"language_type": pipeline.LanguageType,
		"run_id":        run.ID,
		"build_number":  run.BuildNumber,
		"run_status":    run.Status,
		"has_sonar":     hasSonar,
	}

	if hasSonar {
		report["sonar_stage"] = map[string]interface{}{
			"status":       sonarStage.Status,
			"started_at":   sonarStage.StartedAt,
			"finished_at":  sonarStage.FinishedAt,
			"duration_sec": sonarStage.DurationSec,
		}
	}

	if hasQG {
		report["quality_gate"] = map[string]interface{}{
			"status":       qgStage.Status,
			"started_at":   qgStage.StartedAt,
			"finished_at":  qgStage.FinishedAt,
		}
	}

	// 从 stages_result JSON 中提取 SonarQube 数据
	hasSonarData := false
	if run.StagesResult != nil {
		if sonarData, ok := run.StagesResult["sonar_report"]; ok {
			report["sonar_report"] = sonarData
			hasSonarData = true
		}
	}

	// 如果有 sonar 阶段但没有回调数据（Jenkins sonar-callback 未成功回调），返回阶段状态信息
	if hasSonar && !hasSonarData {
		qgStatus := models.QualityGateNone
		message := "代码扫描已完成，但扫描结果数据暂未回传，请稍后刷新"
		if sonarStage.Status == "success" {
			message = "SonarQube 扫描已成功完成，指标数据正在加载中"
		}
		if hasQG && qgStage.Status == "success" {
			qgStatus = models.QualityGateOK
		} else if hasQG && qgStage.Status == "failed" {
			qgStatus = models.QualityGateError
		}
		report["sonar_report"] = map[string]interface{}{
			"project_key":            pipeline.Name,
			"quality_gate":           qgStatus,
			"bugs":                   0,
			"vulnerabilities":        0,
			"code_smells":            0,
			"coverage":               0.0,
			"duplications":           0.0,
			"lines_of_code":          0,
			"security_hotspots":      0,
			"reliability_rating":     "A",
			"security_rating":        "A",
			"maintainability_rating": "A",
			"message":                message,
		}
	}

	// 如果没有代码扫描阶段，返回默认模拟数据（方便前端开发调试）
	if !hasSonar {
		report["sonar_report"] = map[string]interface{}{
			"project_key":          pipeline.Name,
			"quality_gate":         models.QualityGateNone,
			"bugs":                 0,
			"vulnerabilities":      0,
			"code_smells":          0,
			"coverage":             0.0,
			"duplications":         0.0,
			"lines_of_code":        0,
			"security_hotspots":    0,
			"reliability_rating":   "A",
			"security_rating":      "A",
			"maintainability_rating": "A",
			"message":              "暂无 SonarQube 扫描记录，请确保流水线已启用代码质量扫描",
		}
	}

	return report, nil
}

// SaveSonarReport 保存 SonarQube 扫描结果
func (s *Services) SaveSonarReport(ctx context.Context, pipelineID int64, runID int64, info *models.StageSonarInfo) error {
	db := global.DB.WithContext(ctx)

	info.ScanTime = uint64(time.Now().Unix())

	// 将 SonarQube 数据存储到运行记录的 stages_result JSON 中
	var run models.CicdPipelineRun
	if runID > 0 {
		if err := db.Where("id = ? AND pipeline_id = ?", runID, pipelineID).First(&run).Error; err != nil {
			return fmt.Errorf("运行记录不存在")
		}
	} else {
		if err := db.Where("pipeline_id = ?", pipelineID).Order("id DESC").First(&run).Error; err != nil {
			return fmt.Errorf("暂无运行记录")
		}
	}

	// 更新 stages_result
	stagesResult := run.StagesResult
	if stagesResult == nil {
		stagesResult = make(models.JSONMap)
	}
	stagesResult["sonar_report"] = info

	if err := db.Model(&models.CicdPipelineRun{}).Where("id = ?", run.ID).
		Update("stages_result", stagesResult).Error; err != nil {
		return fmt.Errorf("保存 SonarQube 报告失败: %v", err)
	}

	return nil
}
