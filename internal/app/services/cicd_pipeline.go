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
	"k8soperation/pkg/cache"
	"k8soperation/pkg/jenkins"
	"k8soperation/pkg/k8s/deployment"
)

// ==================== 流水线 CRUD ====================

// PipelineCheckName 检查应用名称是否可用
// excludeID > 0 时排除该 ID（编辑模式下排除自身）
// 返回: (available bool, conflictName string, err error)
func (s *Services) PipelineCheckName(ctx context.Context, name string, excludeID int64) (bool, string, error) {
	if name == "" {
		return false, "", errors.New("应用名称不能为空")
	}
	p, err := s.dao.PipelineGetByName(ctx, name)
	if err == nil {
		// 找到同名记录
		if excludeID > 0 && p.ID == excludeID {
			// 就是自身，可用
			return true, "", nil
		}
		return false, p.Name, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, "", nil
	}
	return false, "", fmt.Errorf("检查名称失败: %w", err)
}

// PipelineCreate 创建流水线
func (s *Services) PipelineCreate(ctx context.Context, req *requests.PipelineCreateRequest, userID int64) (int64, []string, error) {
	// 检查名称是否已存在
	_, err := s.dao.PipelineGetByName(ctx, req.Name)
	if err == nil {
		return 0, nil, errors.New("应用名称已存在，请更换一个名称")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil, fmt.Errorf("检查名称失败: %w", err)
	}

	// 收集警告信息（不阻止创建）
	var warnings []string

	// 警告 1：Git 仓库+分支冲突
	gitBranchForCheck := req.GitBranch
	if gitBranchForCheck == "" {
		gitBranchForCheck = global.DefaultBranch()
	}
	conflictPipelines, _ := s.dao.PipelineGetByGitRepoBranch(ctx, req.GitRepo, gitBranchForCheck, 0)
	if len(conflictPipelines) > 0 {
		names := make([]string, 0, len(conflictPipelines))
		for _, cp := range conflictPipelines {
			names = append(names, cp.Name)
		}
		warnings = append(warnings, fmt.Sprintf("该仓库分支已被应用 [%s] 使用，可能导致 Jenkins 构建冲突",
			strings.Join(names, ", ")))
	}

	// 警告 2：自动部署工作负载冲突
	if req.AutoDeploy && req.TargetClusterID > 0 {
		wlName := req.TargetWorkloadName
		if wlName == "" {
			wlName = strings.TrimSuffix(strings.TrimSuffix(req.Name, "-pipeline"), "-prod")
			wlName = strings.TrimSuffix(strings.TrimSuffix(wlName, "-dev"), "-test")
		}
		ns := req.TargetNamespace
		if ns == "" {
			ns = "default"
		}
		if wlName != "" {
			conflictDeploys, _ := s.dao.PipelineGetByWorkload(ctx, req.TargetClusterID, ns, wlName, 0)
			if len(conflictDeploys) > 0 {
				dnames := make([]string, 0, len(conflictDeploys))
				for _, cd := range conflictDeploys {
					dnames = append(dnames, cd.Name)
				}
				warnings = append(warnings, fmt.Sprintf("工作负载 [%s/%s] 已被应用 [%s] 的自动部署占用，可能导致部署相互覆盖",
					ns, wlName, strings.Join(dnames, ", ")))
			}
		}
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
			return 0, nil, fmt.Errorf("不支持的语言类型: %s", languageType)
		}
	}
	if jenkinsJob == "" {
		return 0, nil, errors.New("Jenkins Job 名称不能为空，请指定 jenkins_job 或设置 language_type")
	}

	// ==================== 智能默认值：简化首次创建 ====================
	// 分支默认从配置读取
	gitBranch := req.GitBranch
	if gitBranch == "" {
		gitBranch = global.DefaultBranch()
	}
	// 工作负载类型默认 Deployment
	workloadKind := req.TargetWorkloadKind
	if workloadKind == "" {
		workloadKind = "Deployment"
	}
	// 工作负载名称默认取流水线名称（去掉常见后缀）
	workloadName := req.TargetWorkloadName
	if workloadName == "" && req.AutoDeploy {
		workloadName = strings.TrimSuffix(strings.TrimSuffix(req.Name, "-pipeline"), "-prod")
		workloadName = strings.TrimSuffix(strings.TrimSuffix(workloadName, "-dev"), "-test")
	}
	// 容器名称默认取工作负载名称
	containerName := req.TargetContainer
	if containerName == "" && workloadName != "" {
		containerName = workloadName
	}
	// 命名空间默认 default
	targetNamespace := req.TargetNamespace
	if targetNamespace == "" && req.AutoDeploy {
		targetNamespace = "default"
	}
	// 部署环境默认 dev
	deployEnv := req.DeployEnv
	if deployEnv == "" {
		deployEnv = "dev"
	}

	pipeline := &models.CicdPipeline{
		Name:               req.Name,
		Description:        req.Description,
		GitRepo:            req.GitRepo,
		GitBranch:          gitBranch,
		JenkinsURL:         req.JenkinsURL,
		JenkinsJob:         jenkinsJob,
		LanguageType:       languageType,
		Status:             models.PipelineStatusIdle,
		EnvVars:            models.EnvVars(req.EnvVars),
		DeployConfig:       models.JSONMap(req.DeployConfig),
		// 部署配置（含智能默认值）
		AutoDeploy:         req.AutoDeploy,
		TargetClusterID:    req.TargetClusterID,
		TargetNamespace:    targetNamespace,
		TargetWorkloadKind: workloadKind,
		TargetWorkloadName: workloadName,
		TargetContainer:    containerName,
		DeployEnv:          deployEnv,
		RequireApproval:    req.RequireApproval,
		EnableSonar:        req.EnableSonar,
		EnableArtifactUpload: req.EnableArtifactUpload,
		// 发布联动告警静默
		EnableDeploySilence:  req.EnableDeploySilence,
		SilenceBufferMinutes: req.SilenceBufferMinutes,
		SilenceSeverities:    req.SilenceSeverities,
		CreatedUserID:      userID,
	}

	if err := s.dao.PipelineCreate(ctx, pipeline); err != nil {
		return 0, nil, fmt.Errorf("创建流水线失败: %w", err)
	}

	return pipeline.ID, warnings, nil
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
	// 发布联动告警静默
	if req.EnableDeploySilence != nil {
		updates["enable_deploy_silence"] = *req.EnableDeploySilence
	}
	if req.SilenceBufferMinutes != nil {
		updates["silence_buffer_minutes"] = *req.SilenceBufferMinutes
	}
	if req.SilenceSeverities != nil {
		updates["silence_severities"] = *req.SilenceSeverities
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

	// 智能默认值：自动部署启用时，确保容器名称有值
	autoDeployEnabled := false
	if req.AutoDeploy != nil {
		autoDeployEnabled = *req.AutoDeploy
	} else {
		autoDeployEnabled = pipeline.AutoDeploy
	}
	if autoDeployEnabled {
		// 如果容器名为空，默认用工作负载名
		container := ""
		if req.TargetContainer != nil {
			container = *req.TargetContainer
		} else {
			container = pipeline.TargetContainer
		}
		workloadName := ""
		if req.TargetWorkloadName != nil {
			workloadName = *req.TargetWorkloadName
		} else {
			workloadName = pipeline.TargetWorkloadName
		}
		if container == "" && workloadName != "" {
			updates["target_container"] = workloadName
		}
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
			gitBranch = global.DefaultBranch()
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
			// 发布联动告警静默
			EnableDeploySilence:  item.EnableDeploySilence,
			SilenceBufferMinutes: item.SilenceBufferMinutes,
			SilenceSeverities:    item.SilenceSeverities,
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

	// 立即发送「发布开始」通知（用户点击发布按钮时即时触发）
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

	// 如果没有构建号（构建还没触发成功或还在队列中），直接更新平台状态
	if buildNumber == 0 {
		global.Logger.Info("[流水线] 停止流水线：无构建号，直接更新平台状态",
			zap.Int64("pipeline_id", pipeline.ID),
		)
		_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusAborted)
		// 更新最新的运行记录
		latestRun, err := s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
		if err == nil && latestRun != nil && (latestRun.Status == models.PipelineRunStatusPending || latestRun.Status == models.PipelineRunStatusRunning) {
			_ = s.dao.PipelineRunUpdateStatus(ctx, latestRun.ID, models.PipelineRunStatusAborted)
		}
		return nil
	}

	// 尝试通过 Jenkins 停止构建
	jenkinsStopErr := false
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client != nil {
		if err := client.StopBuild(ctx, pipeline.JenkinsJob, buildNumber); err != nil {
			// Jenkins 停止失败不阻断平台状态更新（可能构建已经结束了）
			global.Logger.Warn("[流水线] Jenkins 停止构建失败（构建可能已结束）",
				zap.Int64("pipeline_id", pipeline.ID),
				zap.Int("build_number", buildNumber),
				zap.Error(err),
			)
			jenkinsStopErr = true
		}
	} else {
		global.Logger.Warn("[流水线] Jenkins 客户端不可用，直接更新平台状态",
			zap.Int64("pipeline_id", pipeline.ID),
		)
		jenkinsStopErr = true
	}

	// 无论 Jenkins 停止是否成功，都更新平台状态
	// 如果 Jenkins 停止失败，可能是构建已经完成，先查询 Jenkins 实际状态
	finalStatus := models.PipelineRunStatusAborted
	if jenkinsStopErr && client != nil {
		buildInfo, err := client.GetBuildInfo(ctx, pipeline.JenkinsJob, buildNumber)
		if err == nil && buildInfo != nil && !buildInfo.Building {
			// 构建已经结束，使用实际状态
			finalStatus = jenkins.BuildStatusToRunStatus(false, buildInfo.Result)
			global.Logger.Info("[流水线] 构建已结束，使用实际状态",
				zap.Int64("pipeline_id", pipeline.ID),
				zap.String("actual_status", finalStatus),
			)
		}
	}

	// 更新流水线状态
	_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, finalStatus)

	// 更新运行记录
	latestRun, err := s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
	if err == nil && latestRun != nil {
		if latestRun.BuildNumber == buildNumber || latestRun.Status == models.PipelineRunStatusPending || latestRun.Status == models.PipelineRunStatusRunning {
			_ = s.dao.PipelineRunUpdateStatus(ctx, latestRun.ID, finalStatus)
		}
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
		// 先尝试读缓存
		var cached jenkins.BuildInfo
		if err := cache.GetBuildInfo(ctx, pipeline.ID, pipeline.LastBuildNumber, &cached); err == nil && cached.Number > 0 {
			buildInfo = &cached
		} else {
			client := s.getJenkinsClient(pipeline.JenkinsURL)
			if client != nil {
				buildInfo, _ = client.GetBuildInfo(ctx, pipeline.JenkinsJob, pipeline.LastBuildNumber)
				if buildInfo != nil {
					// 写入缓存：正在构建用短 TTL，已完成用长 TTL
					cache.SetBuildInfo(ctx, pipeline.ID, pipeline.LastBuildNumber, buildInfo, buildInfo.Building)
				}
			}
		}

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

	return pipeline, buildInfo, latestRun, nil
}

// PipelineHistory 获取流水线运行历史
// 状态同步完全依赖回调 + PollWorker，不再在列表查询时实时调用 Jenkins API
// 避免高并发下每次列表请求都打 Jenkins，影响性能
func (s *Services) PipelineHistory(ctx context.Context, req *requests.PipelineHistoryRequest) ([]*models.CicdPipelineRun, int64, error) {
	list, total, err := s.dao.PipelineRunList(ctx, req.ID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// BuildRecordList 获取全量构建记录（跨流水线），返回包含流水线名称的富化数据
func (s *Services) BuildRecordList(ctx context.Context, page, pageSize int, status, keyword string, pipelineID int64) ([]interface{}, int64, error) {
	list, total, err := s.dao.PipelineRunListAll(ctx, page, pageSize, status, keyword, pipelineID)
	if err != nil {
		return nil, 0, err
	}

	// 批量查询流水线名称缓存
	pipelineNames := make(map[int64]string)
	for _, run := range list {
		if _, ok := pipelineNames[run.PipelineID]; !ok {
			pipelineNames[run.PipelineID] = ""
		}
	}
	for pid := range pipelineNames {
		if p, err := s.dao.PipelineGetByID(ctx, pid); err == nil {
			pipelineNames[pid] = p.Name
		}
	}

	// 转换为富化结果
	result := make([]interface{}, 0, len(list))
	for _, run := range list {
		result = append(result, map[string]interface{}{
			"id":              run.ID,
			"pipeline_id":     run.PipelineID,
			"pipeline_name":   pipelineNames[run.PipelineID],
			"build_number":    run.BuildNumber,
			"status":          run.Status,
			"trigger_type":    run.TriggerType,
			"trigger_user_id": run.TriggerUserID,
			"git_branch":      run.GitBranch,
			"image_url":       run.ImageURL,
			"image_digest":    run.ImageDigest,
			"duration_sec":    run.DurationSec,
			"started_at":      run.StartedAt,
			"finished_at":     run.FinishedAt,
			"created_at":      run.CreatedAt,
			"error_message":   run.ErrorMessage,
		})
	}

	return result, total, nil
}

// BuildStats 获取构建统计数据（总数、成功率、平均时长、趋势）
func (s *Services) BuildStats(ctx context.Context, days int) (map[string]interface{}, error) {
	if days < 1 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	stats, err := s.dao.PipelineRunBuildStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询构建统计失败: %w", err)
	}

	trend, err := s.dao.PipelineRunBuildTrend(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("查询构建趋势失败: %w", err)
	}

	return map[string]interface{}{
		"stats": stats,
		"trend": trend,
	}, nil
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

	// 查找运行记录：优先使用 run_id 精确匹配（避免 build_number 重用导致找到旧记录）
	var run *models.CicdPipelineRun
	if req.RunID > 0 {
		run, err = s.dao.PipelineRunGetByID(ctx, req.RunID)
		if err != nil {
			global.Logger.Warn("[回调] 通过 run_id 查找失败，回退到 build_number",
				zap.Int64("run_id", req.RunID),
				zap.Error(err),
			)
			run = nil
		}
	}
	if run == nil {
		run, err = s.dao.PipelineRunGetByBuildNumber(ctx, pipeline.ID, req.BuildNumber)
		if err != nil {
			return nil, fmt.Errorf("未找到对应的运行记录: pipeline=%d, build=%d, err=%w",
				pipeline.ID, req.BuildNumber, err)
		}
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
	// 仅在构建失败/中止时保存错误信息，避免成功消息污染 error_message 字段
	errMsg := ""
	if runStatus != models.PipelineRunStatusSuccess {
		errMsg = req.Message
	}
	if err := s.dao.PipelineRunUpdateCallback(ctx, run.ID, runStatus, image, req.ImageDigest, errMsg, req.Duration); err != nil {
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

	// 回调到达，主动清除该构建的 Redis 缓存，避免老缓存延迟展示最终状态
	cache.InvalidatePipeline(ctx, pipeline.ID, req.BuildNumber)

	// ==================== 更新阶段状态 ====================
	// 构建完成后，更新各阶段状态（包括将审批阶段设为 waiting）
	// 失败时，将错误信息保存到失败的阶段
	if err := s.UpdateBuildStagesComplete(ctx, run.ID, runStatus, image, req.ImageDigest, errMsg); err != nil {
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
	run.ErrorMessage = errMsg

	// ==================== 发送通知 ====================
	// 如果构建成功且需要审批，发送审批提醒（包含构建成功信息）
	// 如果构建成功且配置了自动部署，暂不发送构建结果通知（等部署完成后发更丰富的通知）
	// 其他情况发送构建结果通知
	skipBuildNotify := false
	if runStatus == models.PipelineRunStatusSuccess && pipeline.RequireApproval {
		go s.NotifyApprovalRequired(ctx, pipeline, run)
	} else if runStatus == models.PipelineRunStatusSuccess && pipeline.AutoDeploy {
		// 自动部署场景：先标记跳过，部署真正启动后才不发；若部署未能启动则补发
		skipBuildNotify = true
		global.Logger.Info("[回调] 自动部署已启用，暂缓构建成功通知",
			zap.Int64("pipeline_id", pipeline.ID),
		)
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
		deployResult := s.autoDeployToK8sWithResult(ctx, pipeline, image, req.ImageDigest, run.ID)
		result.DeployEnabled = deployResult.DeployEnabled
		result.DeploySuccess = deployResult.DeploySuccess
		result.DeployMessage = deployResult.DeployMessage
		result.Namespace = deployResult.Namespace
		result.Deployment = deployResult.Deployment
		result.Image = deployResult.Image

		// 如果自动部署未能实际启动（配置不完整/集群不可达），补发构建结果通知
		// 避免用户完全收不到通知
		if skipBuildNotify && !deployResult.DeployEnabled {
			global.Logger.Info("[回调] 自动部署未启动，补发构建成功通知",
				zap.Int64("pipeline_id", pipeline.ID),
				zap.String("reason", deployResult.DeployMessage),
			)
			go s.NotifyBuildResult(ctx, pipeline, run, true)
		}
	} else if skipBuildNotify {
		// AutoDeploy=true 但没有镜像地址，也补发通知
		go s.NotifyBuildResult(ctx, pipeline, run, true)
	}

	// ==================== 自动同步发布记录 ====================
	// 构建完成后自动创建发布单，让发布管理页面可以看到最新的构建记录
	if runStatus == models.PipelineRunStatusSuccess || runStatus == models.PipelineRunStatusFailed {
		go s.syncPipelineRunToRelease(context.Background(), pipeline, run, runStatus, image, req.ImageDigest)
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

	// 先尝试读缓存
	isRunning := pipeline.Status == models.PipelineStatusRunning
	var cachedStages []PipelineStageInfo
	if err := cache.GetStages(ctx, id, buildNumber, &cachedStages); err == nil && len(cachedStages) > 0 {
		return cachedStages, nil
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

	// 写入缓存：运行中用短 TTL，已完成用长 TTL
	cache.SetStages(ctx, id, buildNumber, stages, isRunning)

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

// getJenkinsClient 获取 Jenkins 客户端（全局缓存单例，复用连接池）
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

	// 使用全局缓存单例，复用连接池
	return jenkins.GetOrCreateClient(
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

// ==================== 流水线运行同步到发布记录 ====================

// syncPipelineRunToRelease 构建完成后自动创建发布单，让发布管理页面能看到最新的构建记录
func (s *Services) syncPipelineRunToRelease(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, runStatus, image, imageDigest string) {
	// 防重：检查是否已经存在对应的发布单（以 build_id 关联）
	existing, _ := s.dao.CicdReleaseGetByBuildID(ctx, run.ID)
	if existing != nil {
		global.Logger.Debug("[同步发布] 已存在对应的发布单，跳过",
			zap.Int64("run_id", run.ID),
			zap.Int64("release_id", existing.ID),
		)
		return
	}

	// 转换状态
	releaseStatus := models.CicdReleaseStatusSucceeded
	if runStatus == models.PipelineRunStatusFailed {
		releaseStatus = models.CicdReleaseStatusFailed
	}

	// 解析镜像地址
	imageRepo := image
	imageTag := ""
	if image != "" {
		// 分离 repo 和 tag，如 registry.cn/proj/app:v1.0 -> repo=registry.cn/proj/app, tag=v1.0
		if idx := strings.LastIndex(image, ":"); idx > 0 && !strings.Contains(image[idx:], "/") {
			imageRepo = image[:idx]
			imageTag = image[idx+1:]
		}
	}

	// 确定工作负载信息
	workloadKind := pipeline.TargetWorkloadKind
	if workloadKind == "" {
		workloadKind = "Deployment"
	}
	namespace := pipeline.TargetNamespace
	if namespace == "" {
		namespace = "default"
	}

	now := uint64(time.Now().Unix())

	// 构建发布单消息：失败时使用具体错误信息，成功时使用通用描述
	releaseMessage := fmt.Sprintf("流水线自动同步: %s #%d", pipeline.Name, run.BuildNumber)
	if runStatus == models.PipelineRunStatusFailed && run.ErrorMessage != "" {
		releaseMessage = run.ErrorMessage
	}

	release := &models.CicdRelease{
		AppName:       pipeline.Name,
		Namespace:     namespace,
		WorkloadKind:  workloadKind,
		WorkloadName:  pipeline.TargetWorkloadName,
		ContainerName: pipeline.TargetContainer,
		Strategy:      "rolling",
		TimeoutSec:    300,
		Status:        releaseStatus,
		Message:       releaseMessage,
		CreatedUserID: run.TriggerUserID,
		RequestID:     fmt.Sprintf("pipeline-sync-%d", run.ID),
		BuildID:       run.ID,
		ImageRepo:     imageRepo,
		ImageTag:      imageTag,
		CreatedAt:     now,
		ModifiedAt:    now,
	}
	if imageDigest != "" {
		release.ImageDigest = &imageDigest
	}

	if err := s.dao.CicdReleaseCreate(ctx, release); err != nil {
		global.Logger.Warn("[同步发布] 创建发布单失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.Int64("run_id", run.ID),
			zap.Error(err),
		)
		return
	}

	// 创建关联的部署任务（让批量发布/重试可以获取到集群信息，回查 client-go 部署目标）
	if pipeline.TargetClusterID > 0 {
		targetImage := image // 完整镜像地址，如 registry.cn/proj/app:v1.0
		if imageDigest != "" && targetImage != "" {
			// 优先使用 digest 确保镜像一致性
			if idx := strings.LastIndex(targetImage, ":"); idx > 0 && !strings.Contains(targetImage[idx:], "/") {
				targetImage = targetImage[:idx] + "@" + imageDigest
			}
		}
		// 任务消息：失败时记录具体原因，方便排查
		taskMessage := "流水线同步"
		if runStatus == models.PipelineRunStatusFailed && run.ErrorMessage != "" {
			taskMessage = run.ErrorMessage
		}
		task := &models.CicdReleaseTask{
			ReleaseID:   release.ID,
			ClusterID:   pipeline.TargetClusterID,
			Status:      releaseStatus,
			Message:     taskMessage,
			TargetImage: targetImage,
			StartedAt:   now,
			FinishedAt:  now,
			CreatedAt:   now,
			ModifiedAt:  now,
		}
		if err := s.dao.CicdTasksCreate(ctx, []*models.CicdReleaseTask{task}); err != nil {
			global.Logger.Warn("[同步发布] 创建部署任务失败",
				zap.Int64("release_id", release.ID),
				zap.Error(err),
			)
		}
	}

	global.Logger.Info("[同步发布] 发布单创建成功",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.Int64("run_id", run.ID),
		zap.Int64("release_id", release.ID),
		zap.String("app_name", pipeline.Name),
		zap.String("status", releaseStatus),
	)
}

// ==================== 自动部署 K8s ====================

// autoDeployToK8sWithResult 回调成功后自动部署到 K8s
// 支持新的部署配置字段：target_cluster_id, target_namespace, target_workload_name, target_container
// 支持多集群部署、审批流程
// 返回部署结果给 Jenkins，让用户在 Jenkins 看到最终状态
func (s *Services) autoDeployToK8sWithResult(ctx context.Context, pipeline *models.CicdPipeline, image, imageDigest string, runID int64) *PipelineCallbackResult {
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
	go s.executeAutoDeployAsync(context.Background(), pipeline, kubeClient, finalImage, workloadKind, runID)

	result.DeploySuccess = true
	result.DeployMessage = fmt.Sprintf("部署已启动: %s/%s 正在更新...", pipeline.TargetNamespace, pipeline.TargetWorkloadName)
	return result
}

// executeAutoDeployAsync 异步执行自动部署（等待 Rollout 完成后发钉钉通知）
// 优化：同步更新 deploy stage 状态，让前端实时展示部署进度
func (s *Services) executeAutoDeployAsync(ctx context.Context, pipeline *models.CicdPipeline, kubeClient kubernetes.Interface, image, workloadKind string, runID int64) {
	var err error
	var logs strings.Builder
	var rolloutResult *RolloutResult
	startTime := time.Now()

	logs.WriteString(fmt.Sprintf("[自动部署] 开始更新 %s/%s\n", pipeline.TargetNamespace, pipeline.TargetWorkloadName))
	logs.WriteString(fmt.Sprintf("工作负载类型: %s\n", workloadKind))
	logs.WriteString(fmt.Sprintf("镜像: %s\n\n", image))

	// 获取流水线运行记录（用于通知中的分支、Commit、发布人等信息）
	var run *models.CicdPipelineRun
	if runID > 0 {
		run, _ = s.dao.PipelineRunGetByID(ctx, runID)
	}

	// 获取集群名称和用户名
	clusterName := s.getClusterName(pipeline.TargetClusterID)
	var username string
	if run != nil && run.TriggerUserID > 0 {
		username = s.getUsernameByID(run.TriggerUserID)
	}

	// 记录部署前的旧镜像（用于回滚通知）
	oldImage := pipeline.LastDeployImage

	// 获取 deploy stage 并设为 running
	var deployStageID int64
	if runID > 0 {
		deployStage, stgErr := s.dao.StageGetByRunIDAndType(ctx, runID, models.StageTypeDeploy)
		if stgErr == nil && deployStage != nil {
			deployStageID = deployStage.ID
			_ = s.dao.StageUpdate(ctx, deployStageID, map[string]interface{}{
				"status":     models.StageStatusRunning,
				"started_at": startTime.Unix(),
			})
		}
	}

	// 「发布开始」通知已在触发时发送（triggerJenkinsBuild），此处不再重复发送

	// 使用统一的 Patch + Rollout 逻辑（与 release flow 完全一致）
		timeout := 5 * time.Minute // 统一默认 5 分钟（WaitDeploymentRollout 内部也有兜底）
		switch workloadKind {
		case "Deployment":
			if patchErr := PatchDeploymentImageFn(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, pipeline.TargetContainer, image); patchErr != nil {
				err = patchErr
			} else {
				logs.WriteString("[INFO] 镜像更新已提交，等待 Deployment Rollout 完成...\n")
				rolloutResult, err = WaitDeploymentRollout(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, timeout, &logs)
			}
		case "StatefulSet":
			if patchErr := PatchStatefulSetImageFn(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, pipeline.TargetContainer, image); patchErr != nil {
				err = patchErr
			} else {
				logs.WriteString("[INFO] 镜像更新已提交，等待 StatefulSet Rollout 完成...\n")
				rolloutResult, err = WaitStatefulSetRollout(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, timeout, &logs)
			}
		case "DaemonSet":
			if patchErr := PatchDaemonSetImageFn(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, pipeline.TargetContainer, image); patchErr != nil {
				err = patchErr
			} else {
				logs.WriteString("[INFO] 镜像更新已提交，等待 DaemonSet Rollout 完成...\n")
				rolloutResult, err = WaitDaemonSetRollout(ctx, kubeClient, pipeline.TargetNamespace, pipeline.TargetWorkloadName, timeout, &logs)
			}
		default:
			err = fmt.Errorf("不支持的工作负载类型: %s", workloadKind)
		}

	// 更新流水线部署状态
	now := uint64(time.Now().Unix())
	duration := int(time.Since(startTime).Seconds())

	// 构建通知信息
	notifyInfo := &AutoDeployNotifyInfo{
		Pipeline:    pipeline,
		Run:         run,
		Image:       image,
		OldImage:    oldImage,
		Duration:    duration,
		ClusterName: clusterName,
		Username:    username,
		Rollout:     rolloutResult,
	}

	if err != nil {
		global.Logger.Error("[自动部署] Rollout 失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.Error(err),
		)
		_ = s.dao.PipelineUpdateDeployInfo(ctx, pipeline.ID, image, "", now, "failed", "")
		// 更新 deploy stage 为失败
		if deployStageID > 0 {
			_ = s.dao.StageUpdate(ctx, deployStageID, map[string]interface{}{
				"status":        models.StageStatusFailed,
				"finished_at":   now,
				"duration_sec":  duration,
				"logs":          logs.String(),
				"error_message": err.Error(),
			})
		}
		// Rollout 失败后发送通知
		notifyInfo.Success = false
		notifyInfo.ErrMsg = err.Error()
		s.notifyAutoDeployResult(ctx, notifyInfo)
	} else {
		global.Logger.Info("[自动部署] Rollout 完成",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.String("image", image),
		)
		version := s.getCurrentWorkloadRevision(ctx, kubeClient, workloadKind, pipeline.TargetNamespace, pipeline.TargetWorkloadName)
		_ = s.dao.PipelineUpdateDeployInfo(ctx, pipeline.ID, image, "", now, "success", version)
		// 更新 deploy stage 为成功
		if deployStageID > 0 {
			_ = s.dao.StageUpdate(ctx, deployStageID, map[string]interface{}{
				"status":       models.StageStatusSuccess,
				"finished_at":  now,
				"duration_sec": duration,
				"logs":         logs.String(),
				"deploy_image": image,
			})
		}
		// 更新流水线运行记录为成功
		if runID > 0 {
			_ = s.dao.PipelineRunUpdateStatus(ctx, runID, models.PipelineRunStatusSuccess)
			_ = s.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, models.PipelineRunStatusSuccess)
		}
		// Rollout 完成后发送通知
		notifyInfo.Success = true
		s.notifyAutoDeployResult(ctx, notifyInfo)
	}
}

// getCurrentWorkloadRevision 获取工作负载在 K8s 集群中的当前版本号
// Deployment: 返回最新 ReplicaSet 的 revision（如 "5"）
// StatefulSet/DaemonSet: 返回最新 ControllerRevision 的 revision（如 "3"）
func (s *Services) getCurrentWorkloadRevision(ctx context.Context, client kubernetes.Interface, workloadKind, namespace, workloadName string) string {
	switch workloadKind {
	case "Deployment", "":
		return s.getDeploymentCurrentRevision(ctx, client, namespace, workloadName)
	case "StatefulSet":
		return s.getStatefulSetCurrentRevision(ctx, client, namespace, workloadName)
	case "DaemonSet":
		return s.getDaemonSetCurrentRevision(ctx, client, namespace, workloadName)
	default:
		return ""
	}
}

// getDeploymentCurrentRevision 获取 Deployment 当前版本号（最新 ReplicaSet 的 revision）
func (s *Services) getDeploymentCurrentRevision(ctx context.Context, client kubernetes.Interface, namespace, workloadName string) string {
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	selector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return ""
	}
	var maxRevision int64
	for _, rs := range rsList.Items {
		isOwned := false
		for _, owner := range rs.OwnerReferences {
			if owner.UID == deploy.UID {
				isOwned = true
				break
			}
		}
		if !isOwned {
			continue
		}
		rev := int64(0)
		if revStr, ok := rs.Annotations["deployment.kubernetes.io/revision"]; ok {
			fmt.Sscanf(revStr, "%d", &rev)
		}
		if rev > maxRevision {
			maxRevision = rev
		}
	}
	if maxRevision > 0 {
		return fmt.Sprintf("%d", maxRevision)
	}
	return ""
}

// getStatefulSetCurrentRevision 获取 StatefulSet 当前版本号（最新 ControllerRevision 的 revision）
func (s *Services) getStatefulSetCurrentRevision(ctx context.Context, client kubernetes.Interface, namespace, workloadName string) string {
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	selector := metav1.FormatLabelSelector(sts.Spec.Selector)
	crList, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return ""
	}
	var maxRevision int64
	for _, cr := range crList.Items {
		isOwned := false
		for _, owner := range cr.OwnerReferences {
			if owner.UID == sts.UID {
				isOwned = true
				break
			}
		}
		if !isOwned {
			continue
		}
		if cr.Revision > maxRevision {
			maxRevision = cr.Revision
		}
	}
	if maxRevision > 0 {
		return fmt.Sprintf("%d", maxRevision)
	}
	return ""
}

// getDaemonSetCurrentRevision 获取 DaemonSet 当前版本号（最新 ControllerRevision 的 revision）
func (s *Services) getDaemonSetCurrentRevision(ctx context.Context, client kubernetes.Interface, namespace, workloadName string) string {
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	selector := metav1.FormatLabelSelector(ds.Spec.Selector)
	crList, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return ""
	}
	var maxRevision int64
	for _, cr := range crList.Items {
		isOwned := false
		for _, owner := range cr.OwnerReferences {
			if owner.UID == ds.UID {
				isOwned = true
				break
			}
		}
		if !isOwned {
			continue
		}
		if cr.Revision > maxRevision {
			maxRevision = cr.Revision
		}
	}
	if maxRevision > 0 {
		return fmt.Sprintf("%d", maxRevision)
	}
	return ""
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

	// ===== 发布联动告警静默 =====
	var silenceRuleID int64
	if pipeline.EnableDeploySilence {
		silenceInfo, silenceErr := s.CreateDeploySilence(ctx, pipeline, namespace, deploymentName)
		if silenceErr != nil {
			global.Logger.Warn("[旧配置部署] 创建发布静默失败（不影响部署）", zap.Error(silenceErr))
		} else if silenceInfo != nil {
			silenceRuleID = silenceInfo.SilenceRuleID
			logs.WriteString(fmt.Sprintf("[INFO] 发布静默已生效，规则ID: %d\n", silenceRuleID))
		}
	}

	// 1. 更新镜像
	_, err := deployment.PatchDeploymentImage(ctx, global.ManagementKubeClient, namespace, deploymentName, containerName, image)
	if err != nil {
		global.Logger.Error("[旧配置部署] 更新镜像失败", zap.Error(err))
		s.ReleaseDeploySilence(ctx, silenceRuleID, false)
		s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, image, false, err.Error())
		return
	}

	logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")

	// 2. 等待 Rollout 完成
	_, err = WaitDeploymentRollout(ctx, global.ManagementKubeClient, namespace, deploymentName, 5*time.Minute, &logs)
	if err != nil {
		global.Logger.Error("[旧配置部署] Rollout 失败", zap.Error(err))
		s.ReleaseDeploySilence(ctx, silenceRuleID, false)
		s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, image, false, err.Error())
		return
	}

	// 部署成功，释放静默
	s.ReleaseDeploySilence(ctx, silenceRuleID, true)
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
		// 通用扩展参数（支持私有依赖库、自定义构建）
		setDefault(params, "EXTRA_REPOS", "")              // 额外依赖仓库（格式: url|path;url|path）
		setDefault(params, "BINARY_NAME", "")              // 自定义二进制名（空=自动检测）
		setDefault(params, "USE_PROJECT_DOCKERFILE", "false") // 使用项目自带 Dockerfile
	case models.LanguageTypeJava:
		setDefault(params, "JAVA_VERSION", "17")
		setDefault(params, "MAVEN_GOALS", "clean package -DskipTests -B")
		setDefault(params, "MAVEN_THREADS", "1C")
		setDefault(params, "SKIP_TESTS", "false")
		setDefault(params, "BUILD_DIR", "") // 空=自动检测 pom.xml 位置
		// 私有 Maven 仓库（用于拉取公司内部依赖包，如 Nexus/GitLab Maven Registry）
		setDefault(params, "MAVEN_PRIVATE_REPO_URL", "")           // 空=仅使用阿里云公共仓库
		setDefault(params, "MAVEN_PRIVATE_REPO_CREDENTIAL_ID", "") // 空=使用默认 'maven-private-repo'
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

	// 凭证 ID：优先从 config.yaml 读取，否则用默认值
	gitCredID := "gitee-id"
	registryCredID := "harbor-registry"
	hmacCredID := "hmac-secret"
	if global.JenkinsSetting != nil {
		if global.JenkinsSetting.GitCredentialID != "" {
			gitCredID = global.JenkinsSetting.GitCredentialID
		}
		if global.JenkinsSetting.RegistryCredentialID != "" {
			registryCredID = global.JenkinsSetting.RegistryCredentialID
		}
		if global.JenkinsSetting.HMACCredentialID != "" {
			hmacCredID = global.JenkinsSetting.HMACCredentialID
		}
	}
	setDefault(params, "GIT_CREDENTIAL_ID", gitCredID)
	setDefault(params, "REGISTRY_CREDENTIAL_ID", registryCredID)
	setDefault(params, "HMAC_CREDENTIAL_ID", hmacCredID)

	// SonarQube 代码质量扫描（暂时强制关闭，待 SonarQube 服务部署后再启用）
	setDefault(params, "ENABLE_SONAR", "false")
	setDefault(params, "SONAR_QUALITY_GATE", "false")

	// 并发构建限制：从 config.yaml 读取，传入 Jenkins 模板动态设置 throttleJobProperty
	maxConcurrent := 10 // 默认值
	if global.JenkinsSetting != nil && global.JenkinsSetting.MaxConcurrentBuilds > 0 {
		maxConcurrent = global.JenkinsSetting.MaxConcurrentBuilds
	}
	setDefault(params, "MAX_CONCURRENT_BUILDS", fmt.Sprintf("%d", maxConcurrent))

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
				"MAVEN_THREADS":   "1C",
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
