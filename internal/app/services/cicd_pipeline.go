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

	pipeline := &models.CicdPipeline{
		Name:               req.Name,
		Description:        req.Description,
		GitRepo:            req.GitRepo,
		GitBranch:          req.GitBranch,
		JenkinsURL:         req.JenkinsURL,
		JenkinsJob:         req.JenkinsJob,
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
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		params["PLATFORM_CALLBACK_URL"] = global.JenkinsSetting.CallbackURL + "/api/v1/k8s/cicd/pipeline/callback"
	}
	// 注意：HMAC_SECRET 不再通过参数传递，Jenkins 端应使用 credentials 管理
	// 双方需配置相同的密钥：平台 config.yaml 的 HMACSecret 与 Jenkins credentials 中的 hmac-secret
	
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

	// 获取最新运行记录
	latestRun, _ := s.dao.PipelineRunGetLatest(ctx, id)

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
	Status   string `json:"status"` // success, failed, running, pending
	Duration string `json:"duration"`
	Steps    []PipelineStepInfo `json:"steps"`
}

// PipelineStepInfo 步骤信息
type PipelineStepInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
}

// PipelineStages 获取流水线阶段数据
func (s *Services) PipelineStages(ctx context.Context, id int64, buildNumber int) ([]PipelineStageInfo, error) {
	// 获取流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("流水线不存在")
		}
		return nil, fmt.Errorf("查询流水线失败: %w", err)
	}

	// 确定构建号
	if buildNumber == 0 {
		buildNumber = pipeline.LastBuildNumber
	}
	if buildNumber == 0 {
		// 返回默认阶段（未运行状态）
		return s.getDefaultStages(), nil
	}

	// 获取 Jenkins 客户端
	client := s.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return s.getDefaultStages(), nil
	}

	// 从 Jenkins 获取阶段数据
	pipelineRun, err := client.GetPipelineRun(ctx, pipeline.JenkinsJob, buildNumber)
	if err != nil {
		global.Logger.Warn("[流水线] 获取Jenkins阶段数据失败",
			zap.Int64("pipeline_id", id),
			zap.Int("build_number", buildNumber),
			zap.Error(err),
		)
		return s.getDefaultStages(), nil
	}

	// 转换为前端友好格式
	stages := make([]PipelineStageInfo, 0, len(pipelineRun.Stages))
	for _, stage := range pipelineRun.Stages {
		stageInfo := PipelineStageInfo{
			ID:       stage.ID,
			Name:     stage.Name,
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

	return stages, nil
}

// getDefaultStages 获取默认阶段模板
func (s *Services) getDefaultStages() []PipelineStageInfo {
	return []PipelineStageInfo{
		{Name: "Checkout", Status: "pending", Steps: []PipelineStepInfo{{Name: "Git Clone", Status: "pending"}}},
		{Name: "Build", Status: "pending", Steps: []PipelineStepInfo{{Name: "Compile", Status: "pending"}, {Name: "Package", Status: "pending"}}},
		{Name: "Test", Status: "pending", Steps: []PipelineStepInfo{{Name: "Unit Test", Status: "pending"}}},
		{Name: "Deploy", Status: "pending", Steps: []PipelineStepInfo{{Name: "Push Image", Status: "pending"}, {Name: "Deploy to K8s", Status: "pending"}}},
	}
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

	_, err := deployment.PatchDeploymentImage(ctx, global.ManagementKubeClient, namespace, deploymentName, containerName, finalImage)
	if err != nil {
		result.DeployMessage = fmt.Sprintf("部署失败: %v", err)

		// 发送部署失败钉钉通知
		go s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, finalImage, false, err.Error())

		return result
	}

	result.DeploySuccess = true
	result.DeployMessage = fmt.Sprintf("部署成功: %s/%s 镜像已更新", namespace, deploymentName)

	// 发送部署成功钉钉通知
	go s.notifyLegacyDeployResult(ctx, pipeline, namespace, deploymentName, finalImage, true, "")

	return result
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
