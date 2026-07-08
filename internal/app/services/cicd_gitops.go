package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/argocd"
	"k8soperation/pkg/argoworkflows"
)

// ==================== GitOps Pipeline Run ====================

// gitOpsPipelineRun 触发 Argo Workflow 执行 GitOps CI 流程
func (s *Services) gitOpsPipelineRun(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, req *requests.PipelineRunRequest) error {
	if global.GitOpsSetting == nil || global.GitOpsSetting.ArgoWorkflowsURL == "" {
		return errors.New("GitOps 模式未配置 Argo Workflows，请在 config.yaml 中设置 GitOps.ArgoWorkflowsURL")
	}

	// 解析 GitOps 配置
	gitOpsConfig, err := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	if err != nil || gitOpsConfig == nil {
		return fmt.Errorf("解析 GitOps 配置失败: %w", err)
	}

	// 确定 workflow template
	workflowTemplate := gitOpsConfig.WorkflowTemplate
	if workflowTemplate == "" {
		var ok bool
		workflowTemplate, ok = models.DefaultArgoWorkflowTemplateMap[pipeline.LanguageType]
		if !ok {
			return fmt.Errorf("不支持的语言类型: %s，无法自动选择 Workflow 模板", pipeline.LanguageType)
		}
	}

	// 确定 workflow namespace
	workflowNamespace := lo.Ternary(gitOpsConfig.WorkflowNamespace != "",
		gitOpsConfig.WorkflowNamespace, "argo")

	// 确定分支
	branch := pipeline.GitBranch
	if req.Branch != "" {
		branch = req.Branch
	}

	// 构建 Workflow 参数
	callbackURL := ""
	if global.GitOpsSetting.CallbackURL != "" {
		callbackURL = global.GitOpsSetting.CallbackURL + "/api/v1/k8s/cicd/pipeline/callback"
	}

	params := []string{
		argoworkflows.BuildParameter("git_repo", pipeline.GitRepo),
		argoworkflows.BuildParameter("git_branch", branch),
		argoworkflows.BuildParameter("pipeline_id", fmt.Sprintf("%d", pipeline.ID)),
		argoworkflows.BuildParameter("run_id", fmt.Sprintf("%d", run.ID)),
		argoworkflows.BuildParameter("platform_callback_url", callbackURL),
		argoworkflows.BuildParameter("language_type", pipeline.LanguageType),
	}

	// 镜像构建参数
	if gitOpsConfig.ImageRegistry != "" {
		params = append(params, argoworkflows.BuildParameter("image_registry", gitOpsConfig.ImageRegistry))
	}
	if gitOpsConfig.ImageRepo != "" {
		params = append(params, argoworkflows.BuildParameter("image_repo", gitOpsConfig.ImageRepo))
	}

	// Git manifest 参数（用于 update-manifest 步骤）
	if gitOpsConfig.GitManifestRepo != "" {
		params = append(params, argoworkflows.BuildParameter("git_manifest_repo", gitOpsConfig.GitManifestRepo))
	}
	if gitOpsConfig.ManifestPath != "" {
		params = append(params, argoworkflows.BuildParameter("manifest_path", gitOpsConfig.ManifestPath))
	}

	// ArgoCD Application 名称（用于 sync 步骤）
	if gitOpsConfig.ArgoCDAppName != "" {
		params = append(params, argoworkflows.BuildParameter("argo_app_name", gitOpsConfig.ArgoCDAppName))
	}

	// 环境变量合并
	for _, ev := range pipeline.EnvVars {
		params = append(params, argoworkflows.BuildParameter("env_"+ev.Name, ev.Value))
	}
	for k, v := range req.EnvVars {
		params = append(params, argoworkflows.BuildParameter("env_"+k, v))
	}

	// 标签（用于后续查询）
	labels := fmt.Sprintf("pipeline_id=%d,run_id=%d,deploy_mode=gitops", pipeline.ID, run.ID)

	// 提交 Workflow
	workflowsClient := argoworkflows.GetOrCreateClient(
		global.GitOpsSetting.ArgoWorkflowsURL,
		global.GitOpsSetting.ArgoWorkflowsToken,
	)

	submitReq := &argoworkflows.WorkflowSubmitRequest{
		Namespace:    workflowNamespace,
		ResourceKind: "WorkflowTemplate",
		ResourceName: workflowTemplate,
		Parameters:   params,
		Labels:       labels,
	}

	wf, err := workflowsClient.SubmitWorkflow(ctx, submitReq)
	if err != nil {
		return fmt.Errorf("提交 Argo Workflow 失败: %w", err)
	}

	global.Logger.Infof("[GitOps] Workflow 已提交: name=%s, namespace=%s, template=%s",
		wf.Metadata.Name, wf.Metadata.Namespace, workflowTemplate)

	// 更新 run 记录中的 GitOps 字段
	updates := map[string]interface{}{
		"workflow_name": wf.Metadata.Name,
		"argo_app_name": gitOpsConfig.ArgoCDAppName,
		"sync_status":   models.SyncStatusUnknown,
	}
	_ = s.dao.PipelineRunUpdate(ctx, run.ID, updates)

	// 更新 pipeline 的 build URL
	buildURL := fmt.Sprintf("%s/workflows/%s/%s", global.GitOpsSetting.ArgoWorkflowsURL, wf.Metadata.Namespace, wf.Metadata.Name)
	_ = s.dao.PipelineUpdateRunInfo(ctx, pipeline.ID, models.PipelineRunStatusRunning, 1, buildURL)

	return nil
}

// ==================== GitOps Pipeline Stop ====================

// gitOpsPipelineStop 终止 Argo Workflow
func (s *Services) gitOpsPipelineStop(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) error {
	if global.GitOpsSetting == nil {
		return errors.New("GitOps 模式未配置")
	}

	// 获取最新的 run（如果未传入）
	if run == nil {
		var err error
		run, err = s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
		if err != nil {
			return fmt.Errorf("获取运行记录失败: %w", err)
		}
	}

	if run.WorkflowName == "" {
		return errors.New("未找到关联的 Argo Workflow")
	}

	// 确定 namespace（从 pipeline 的 GitOpsConfig 获取）
	gitOpsConfig, _ := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	namespace := "argo"
	if gitOpsConfig != nil && gitOpsConfig.WorkflowNamespace != "" {
		namespace = gitOpsConfig.WorkflowNamespace
	}

	workflowsClient := argoworkflows.GetOrCreateClient(
		global.GitOpsSetting.ArgoWorkflowsURL,
		global.GitOpsSetting.ArgoWorkflowsToken,
	)

	if err := workflowsClient.StopWorkflow(ctx, namespace, run.WorkflowName); err != nil {
		return fmt.Errorf("终止 Argo Workflow 失败: %w", err)
	}

	_ = s.dao.PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusAborted)
	_ = s.dao.PipelineUpdateStatus(ctx, pipeline.ID, models.PipelineStatusIdle)

	global.Logger.Infof("[GitOps] Workflow 已终止: %s", run.WorkflowName)
	return nil
}

// ==================== GitOps Pipeline Logs ====================

// gitOpsPipelineLogs 获取 Argo Workflow 执行日志
func (s *Services) gitOpsPipelineLogs(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) (string, error) {
	if global.GitOpsSetting == nil {
		return "", errors.New("GitOps 模式未配置")
	}

	if run == nil {
		var err error
		run, err = s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
		if err != nil {
			return "", fmt.Errorf("获取运行记录失败: %w", err)
		}
	}

	if run.WorkflowName == "" {
		return "", errors.New("未找到关联的 Argo Workflow")
	}

	gitOpsConfig, _ := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	namespace := "argo"
	if gitOpsConfig != nil && gitOpsConfig.WorkflowNamespace != "" {
		namespace = gitOpsConfig.WorkflowNamespace
	}

	workflowsClient := argoworkflows.GetOrCreateClient(
		global.GitOpsSetting.ArgoWorkflowsURL,
		global.GitOpsSetting.ArgoWorkflowsToken,
	)

	return workflowsClient.GetWorkflowLogs(ctx, namespace, run.WorkflowName)
}

// ==================== GitOps Pipeline Stages ====================

// gitOpsPipelineStages 将 Argo Workflow 节点转换为阶段展示信息
func (s *Services) gitOpsPipelineStages(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) ([]PipelineStageInfo, error) {
	if global.GitOpsSetting == nil {
		return nil, errors.New("GitOps 模式未配置")
	}

	if run == nil {
		var err error
		run, err = s.dao.PipelineRunGetLatest(ctx, pipeline.ID)
		if err != nil {
			return nil, fmt.Errorf("获取运行记录失败: %w", err)
		}
	}

	if run.WorkflowName == "" {
		return nil, errors.New("未找到关联的 Argo Workflow")
	}

	gitOpsConfig, _ := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	namespace := "argo"
	if gitOpsConfig != nil && gitOpsConfig.WorkflowNamespace != "" {
		namespace = gitOpsConfig.WorkflowNamespace
	}

	workflowsClient := argoworkflows.GetOrCreateClient(
		global.GitOpsSetting.ArgoWorkflowsURL,
		global.GitOpsSetting.ArgoWorkflowsToken,
	)

	wf, err := workflowsClient.GetWorkflow(ctx, namespace, run.WorkflowName)
	if err != nil {
		return nil, fmt.Errorf("获取 Workflow 状态失败: %w", err)
	}

	// 将 workflow nodes 转换为 PipelineStageInfo
	var stages []PipelineStageInfo

	// Argo Workflow nodes 是 map[id]node，需要按顺序排序
	type nodeWithID struct {
		id   string
		node argoworkflows.WorkflowNodeStatus
	}
	var nodes []nodeWithID
	for id, node := range wf.Status.Nodes {
		// 跳过虚拟节点（type != "Pod" 的节点通常为步骤组）
		nodes = append(nodes, nodeWithID{id: id, node: node})
	}

	for _, n := range nodes {
		status := mapWorkflowPhase(n.node.Phase)
		dur := parseArgoTime(n.node.FinishedAt).Sub(parseArgoTime(n.node.StartedAt))
		stage := PipelineStageInfo{
			ID:       n.id,
			Name:     lo.Ternary(n.node.DisplayName != "", n.node.DisplayName, n.node.TemplateName),
			Type:     n.node.TemplateName,
			Status:   status,
			Duration: formatDurationSeconds(int(dur.Seconds())),
		}
		stages = append(stages, stage)
	}

	return stages, nil
}

// ==================== ArgoCD Sync Operations ====================

// gitOpsTriggerSync 触发 ArgoCD Application 同步（CI 完成后调用）
func (s *Services) gitOpsTriggerSync(ctx context.Context, pipeline *models.CicdPipeline, image, imageDigest string) error {
	if global.GitOpsSetting == nil || global.GitOpsSetting.ArgoCDURL == "" {
		return errors.New("GitOps 模式未配置 ArgoCD")
	}

	gitOpsConfig, err := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	if err != nil || gitOpsConfig == nil {
		return fmt.Errorf("解析 GitOps 配置失败: %w", err)
	}

	if gitOpsConfig.ArgoCDAppName == "" {
		return errors.New("未配置 ArgoCD Application 名称")
	}

	argoClient := argocd.GetOrCreateClient(
		global.GitOpsSetting.ArgoCDURL,
		global.GitOpsSetting.ArgoCDAuthToken,
	)

	// 触发同步
	if err := argoClient.SyncApplication(ctx, gitOpsConfig.ArgoCDAppName,
		gitOpsConfig.TargetRevision, gitOpsConfig.PruneResource); err != nil {
		return fmt.Errorf("触发 ArgoCD 同步失败: %w", err)
	}

	// 更新 deploy 信息
	now := uint64(time.Now().Unix())
	_ = s.dao.PipelineUpdateDeployInfo(ctx, pipeline.ID, image, imageDigest, now, "syncing", "")

	global.Logger.Infof("[GitOps] ArgoCD 同步已触发: app=%s, image=%s",
		gitOpsConfig.ArgoCDAppName, image)

	return nil
}

// GitOpsSyncCallback 处理 ArgoCD 同步状态 Webhook 回调
func (s *Services) GitOpsSyncCallback(ctx context.Context, req *requests.GitOpsSyncCallbackRequest) error {
	// 查找运行记录
	var run *models.CicdPipelineRun
	var err error
	if req.RunID > 0 {
		run, err = s.dao.PipelineRunGetByID(ctx, req.RunID)
	} else if req.PipelineID > 0 {
		run, err = s.dao.PipelineRunGetLatest(ctx, req.PipelineID)
	} else {
		return errors.New("缺少 pipeline_id 或 run_id 参数")
	}
	if err != nil {
		return fmt.Errorf("查询运行记录失败: %w", err)
	}

	// 更新 sync 状态
	updates := map[string]interface{}{
		"sync_status":   req.SyncStatus,
		"sync_revision": req.SyncRevision,
	}
	_ = s.dao.PipelineRunUpdate(ctx, run.ID, updates)

	// 根据同步状态更新 pipeline 和 run 状态
	switch req.SyncStatus {
	case models.SyncStatusSynced:
		if req.HealthStatus == "Healthy" {
			_ = s.dao.PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusSuccess)
			_ = s.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusSuccess)
			global.Logger.Infof("[GitOps] 同步完成: run=%d, app=%s, revision=%s",
				run.ID, req.AppName, req.SyncRevision)
		}
	case models.SyncStatusOutOfSync:
		global.Logger.Warnf("[GitOps] 同步漂移: run=%d, app=%s, revision=%s",
			run.ID, req.AppName, req.SyncRevision)
	}

	return nil
}

// ==================== HMAC 验证 ====================

// GitOpsVerifyHMAC 验证 GitOps Webhook 的 HMAC 签名（公开方法，供 controller 调用）
func (s *Services) GitOpsVerifyHMAC(signature string, payload []byte) bool {
	if global.GitOpsSetting == nil || global.GitOpsSetting.HMACSecret == "" {
		global.Logger.Warn("[GitOps] HMAC 密钥未配置，跳过签名验证")
		return true
	}

	mac := hmac.New(sha256.New, []byte(global.GitOpsSetting.HMACSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 1
}

// ==================== 工具函数 ====================

// mapWorkflowPhase 将 Argo Workflow phase 映射为 Stage 状态
func mapWorkflowPhase(phase string) string {
	switch phase {
	case "Pending":
		return models.StageStatusPending
	case "Running":
		return models.StageStatusRunning
	case "Succeeded":
		return models.StageStatusSuccess
	case "Failed", "Error":
		return models.StageStatusFailed
	case "Omitted", "Skipped":
		return models.StageStatusSkipped
	default:
		return models.StageStatusPending
	}
}

// parseArgoTime 解析 Argo Workflows 的时间格式
func parseArgoTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	// Argo Workflows 使用 ISO 8601 格式
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// 尝试其他常见格式
		t, err = time.Parse("2006-01-02T15:04:05Z", timeStr)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}
