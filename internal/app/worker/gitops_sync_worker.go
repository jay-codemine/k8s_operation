package worker

import (
	"context"
	"sync"
	"time"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/argocd"
	"k8soperation/pkg/argoworkflows"
)

// GitOpsSyncWorker GitOps 同步状态轮询 Worker
// 定期检查 Argo Workflow 执行状态和 ArgoCD 同步状态
// 防止 Webhook 丢失导致状态不更新（fallback 机制）
type GitOpsSyncWorker struct {
	svc *services.Services
	pollInterval time.Duration
	maxWaitTime  time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewGitOpsSyncWorker 创建 GitOps 同步状态轮询 Worker
func NewGitOpsSyncWorker() *GitOpsSyncWorker {
	pollInterval := 10 * time.Second
	maxWait := 600 * time.Second
	if global.GitOpsSetting != nil {
		if global.GitOpsSetting.SyncPollInterval > 0 {
			pollInterval = time.Duration(global.GitOpsSetting.SyncPollInterval) * time.Second
		}
		if global.GitOpsSetting.SyncMaxWaitTime > 0 {
			maxWait = time.Duration(global.GitOpsSetting.SyncMaxWaitTime) * time.Second
		}
	}

	return &GitOpsSyncWorker{
		svc:          services.NewBackgroundServices(),
		pollInterval: pollInterval,
		maxWaitTime:  maxWait,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动轮询
func (w *GitOpsSyncWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.pollLoop(ctx)
	global.Logger.Info("[GitOps] Sync Worker 已启动")
}

// Stop 停止轮询
func (w *GitOpsSyncWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	global.Logger.Info("[GitOps] Sync Worker 已停止")
}

// pollLoop 轮询循环
func (w *GitOpsSyncWorker) pollLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.pollRunningWorkflows(ctx)
		}
	}
}

// pollRunningWorkflows 轮询所有运行中的 GitOps 流水线
func (w *GitOpsSyncWorker) pollRunningWorkflows(ctx context.Context) {
	// 查询 deploy_mode=gitops 且 status=running 的 pipeline runs
	runs, err := w.svc.CicdSvc().PipelineRunListPendingForPoll(ctx, 30, 100)
	if err != nil {
		global.Logger.Warnf("[GitOps] 查询运行中的流水线失败: %v", err)
		return
	}

	for _, run := range runs {
		// 只处理 GitOps 模式且有 WorkflowName 的记录
		if run.WorkflowName == "" {
			continue
		}

		pipeline, err := w.svc.CicdSvc().PipelineGetByID(ctx, run.PipelineID)
		if err != nil || pipeline.DeployMode != models.DeployModeGitOps {
			continue
		}

		w.checkWorkflowAndSyncStatus(ctx, pipeline, run)
	}
}

// checkWorkflowAndSyncStatus 检查 Workflow 状态和 ArgoCD 同步状态
func (w *GitOpsSyncWorker) checkWorkflowAndSyncStatus(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) {
	if global.GitOpsSetting == nil {
		return
	}

	// 1. 检查 Argo Workflow 状态
	gitOpsConfig, err := models.GitOpsConfigFromJSONMap(pipeline.GitOpsConfig)
	if err != nil {
		return
	}
	namespace := "argo"
	if gitOpsConfig != nil && gitOpsConfig.WorkflowNamespace != "" {
		namespace = gitOpsConfig.WorkflowNamespace
	}

	if global.GitOpsSetting.ArgoWorkflowsURL != "" {
		workflowsClient := argoworkflows.GetOrCreateClient(
			global.GitOpsSetting.ArgoWorkflowsURL,
			global.GitOpsSetting.ArgoWorkflowsToken,
		)
		wf, err := workflowsClient.GetWorkflow(ctx, namespace, run.WorkflowName)
		if err == nil {
			switch wf.Status.Phase {
			case models.WorkflowStatusSucceeded:
				// Workflow 构建成功，触发审批流程（与 Jenkins 回调逻辑对齐）
				// UpdateBuildStagesComplete 会将构建阶段标记完成，若有审批则设为 waiting 并创建审批记录、发送飞书通知
				if err := w.svc.UpdateBuildStagesComplete(ctx, run.ID,
					models.PipelineRunStatusSuccess,
					run.ImageURL, run.ImageDigest, ""); err != nil {
					global.Logger.Warnf("[GitOps] 更新构建阶段/审批状态失败: run=%d, err=%v", run.ID, err)
				}

				_ = w.svc.CicdSvc().PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusSuccess)
				_ = w.svc.CicdSvc().PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusSuccess)

				// Workflow 成功，检查 ArgoCD sync 状态
				if run.ArgoAppName != "" && run.SyncStatus != models.SyncStatusSynced {
					w.checkArgoCDStatus(ctx, run.ArgoAppName, run)
				}
			case models.WorkflowStatusFailed, models.WorkflowStatusError:
				// 构建失败也更新阶段状态
				if err := w.svc.UpdateBuildStagesComplete(ctx, run.ID,
					models.PipelineRunStatusFailed,
					run.ImageURL, run.ImageDigest,
					wf.Status.Message); err != nil {
					global.Logger.Warnf("[GitOps] 更新失败阶段状态失败: run=%d, err=%v", run.ID, err)
				}

				_ = w.svc.CicdSvc().PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusFailed)
				_ = w.svc.CicdSvc().PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed)
			}
		}
	}

	// 2. 检查 ArgoCD 同步状态
	if run.ArgoAppName != "" && global.GitOpsSetting.ArgoCDURL != "" {
		w.checkArgoCDStatus(ctx, run.ArgoAppName, run)
	}
}

// checkArgoCDStatus 检查 ArgoCD Application 同步状态
func (w *GitOpsSyncWorker) checkArgoCDStatus(ctx context.Context, appName string, run *models.CicdPipelineRun) {
	argoClient := argocd.GetOrCreateClient(
		global.GitOpsSetting.ArgoCDURL,
		global.GitOpsSetting.ArgoCDAuthToken,
	)

	status, err := argoClient.GetApplicationStatus(ctx, appName)
	if err != nil {
		return
	}

	// 更新 sync 状态
	updates := map[string]interface{}{
		"sync_status":   status.SyncStatus,
		"sync_revision": status.SyncRevision,
	}
	_ = w.svc.CicdSvc().PipelineRunUpdate(ctx, run.ID, updates)

	// 同步完成
	if status.SyncStatus == models.SyncStatusSynced && status.HealthStatus == "Healthy" {
		_ = w.svc.CicdSvc().PipelineRunUpdateStatus(ctx, run.ID, models.PipelineRunStatusSuccess)
		_ = w.svc.CicdSvc().PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusSuccess)
	}
}
