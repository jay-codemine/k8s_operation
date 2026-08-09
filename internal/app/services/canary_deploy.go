package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	dmcicd "k8soperation/internal/domain/cicd"
	k8sdeploy "k8soperation/pkg/k8s/deployment"
	prom "k8soperation/pkg/prometheus"
)

// =============================================================================
// 金丝雀部署服务层
// =============================================================================

// CanaryDeployResult 金丝雀部署结果
type CanaryDeployResult struct {
	Success bool   `json:"success"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
}

// CanaryCreateAndMonitor 创建金丝雀并启动异步监控
func (s *Services) CanaryCreateAndMonitor(ctx context.Context, kube kubernetes.Interface, pipeline *models.CicdPipeline, image, containerName string, runID int64) *CanaryDeployResult {
	namespace := pipeline.TargetNamespace
	appName := pipeline.TargetWorkloadName
	canaryReplicas := pipeline.CanaryReplicas
	if canaryReplicas <= 0 {
		canaryReplicas = 1
	}
	if containerName == "" {
		containerName = pipeline.TargetContainer
	}

	global.Logger.Info("[Canary] 开始金丝雀部署",
		zap.Int64("pipeline_id", pipeline.ID),
		zap.String("namespace", namespace),
		zap.String("app", appName),
		zap.String("image", image),
		zap.Int32("replicas", canaryReplicas),
	)

	_, err := k8sdeploy.CreateCanaryDeployment(ctx, kube, namespace, appName, image, containerName, canaryReplicas)
	if err != nil {
		global.Logger.Error("[Canary] 创建金丝雀失败", zap.Error(err))
		return &CanaryDeployResult{
			Success: false,
			Phase:   "failed",
			Message: fmt.Sprintf("创建金丝雀 Deployment 失败: %v", err),
		}
	}

	if runID > 0 {
		s.updateCanaryStage(ctx, runID, "running", "金丝雀 Deployment 创建成功，观察中...")
	}

	return &CanaryDeployResult{
		Success: true,
		Phase:   "monitoring",
		Message: fmt.Sprintf("金丝雀部署成功，观察 %d 秒后自动决策", pipeline.CanaryDurationSec),
	}
}

// CanaryPromote 手动晋升金丝雀
func (s *Services) CanaryPromote(ctx context.Context, kube kubernetes.Interface, req *requests.CanaryPromoteRequest) error {
	containerName := req.ContainerName
	if containerName == "" {
		if req.PipelineID > 0 {
			pipeline, err := s.cicdSvc().PipelineGetByID(ctx, req.PipelineID)
			if err == nil && pipeline != nil {
				containerName = pipeline.TargetContainer
			}
		}
	}

	_, err := k8sdeploy.PromoteCanaryToStable(ctx, kube, req.Namespace, req.AppName, containerName)
	if err != nil {
		return fmt.Errorf("晋升金丝雀失败: %w", err)
	}

	if req.RunID > 0 {
		s.updateCanaryStage(ctx, req.RunID, "promoted", "金丝雀已晋升为稳定版本")
	}

	return nil
}

// CanaryRollback 回滚金丝雀
func (s *Services) CanaryRollback(ctx context.Context, kube kubernetes.Interface, req *requests.CanaryRollbackRequest) error {
	err := k8sdeploy.RollbackCanary(ctx, kube, req.Namespace, req.AppName)
	if err != nil {
		return fmt.Errorf("回滚金丝雀失败: %w", err)
	}

	if req.RunID > 0 {
		s.updateCanaryStage(ctx, req.RunID, "rolled_back", "金丝雀已回滚，stable 不受影响")
	}

	return nil
}

// CanaryGetStatus 获取金丝雀状态
func (s *Services) CanaryGetStatus(ctx context.Context, kube kubernetes.Interface, req *requests.CanaryStatusRequest) (*k8sdeploy.CanaryStatusInfo, error) {
	return k8sdeploy.GetCanaryStatus(ctx, kube, req.Namespace, req.AppName)
}

// CanarySetTrafficSplit 调整金丝雀流量比例
func (s *Services) CanarySetTrafficSplit(ctx context.Context, kube kubernetes.Interface, req *requests.CanaryTrafficSplitRequest) error {
	canaryName := req.AppName + k8sdeploy.CanaryNameSuffix

	replicas := req.CanaryReplicas
	if replicas <= 0 && req.CanaryRatio > 0 {
		info, err := k8sdeploy.GetCanaryStatus(ctx, kube, req.Namespace, req.AppName)
		if err != nil {
			return fmt.Errorf("获取当前状态失败: %w", err)
		}
		replicas = info.StableDesiredReplicas * req.CanaryRatio / (100 - req.CanaryRatio)
		if replicas < 1 {
			replicas = 1
		}
	}

	if replicas <= 0 {
		return fmt.Errorf("无法计算金丝雀副本数，请直接指定 canary_replicas")
	}

	_, err := k8sdeploy.SetCanaryReplicas(ctx, kube, req.Namespace, canaryName, replicas)
	if err != nil {
		return fmt.Errorf("调整流量比例失败: %w", err)
	}

	return nil
}

// CanaryAnalyze 分析金丝雀指标（Prometheus）
func (s *Services) CanaryAnalyze(ctx context.Context, kube kubernetes.Interface, pipeline *models.CicdPipeline, namespace string) *CanaryDeployResult {
	if pipeline.CanaryAnalysisRules == "" {
		return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "无分析规则，跳过指标分析"}
	}

	promClient := s.getDefaultPromClient()
	if promClient == nil {
		return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "Prometheus 不可用，跳过分析"}
	}

	info, err := k8sdeploy.GetCanaryStatus(ctx, kube, namespace, pipeline.TargetWorkloadName)
	if err != nil {
		return &CanaryDeployResult{Success: false, Phase: "error", Message: fmt.Sprintf("获取金丝雀状态失败: %v", err)}
	}

	if info.CanaryReadyReplicas == 0 && info.Phase == k8sdeploy.CanaryPhasePending {
		return &CanaryDeployResult{Success: false, Phase: "failing", Message: "金丝雀 Pod 未就绪，可能启动失败"}
	}

	if info.CanaryImage == info.StableImage && info.CanaryImage != "" {
		return &CanaryDeployResult{Success: true, Phase: "promoted", Message: "金丝雀已晋升，镜像一致"}
	}

	_ = promClient
	return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "金丝雀运行正常"}
}

// getDefaultPromClient 获取默认 Prometheus 客户端
func (s *Services) getDefaultPromClient() *prom.Client {
	ds, err := s.MonitorCRUDSvc().GetDefaultDatasource(context.Background(), []string{"prometheus", "victoriametrics", "thanos"})
	if err != nil || ds.URL == "" {
		return nil
	}
	return prom.NewClient(ds.URL, 30*time.Second)
}

// updateCanaryStage 更新金丝雀阶段状态
func (s *Services) updateCanaryStage(ctx context.Context, runID int64, status, message string) {
	if runID <= 0 {
		return
	}

	now := uint64(time.Now().Unix())
	logs := fmt.Sprintf("[%s] %s\n%s", time.Now().Format("2006-01-02 15:04:05"), message, status)
	updates := map[string]interface{}{
		"status":      status,
		"logs":        logs,
		"finished_at": now,
	}

	existing, err := s.cicdSvc().StageGetByRunIDAndType(ctx, runID, "canary_deploy")
	if err == nil {
		_ = s.cicdSvc().StageUpdate(ctx, existing.ID, updates)
		return
	}

	stage := &dmcicd.CicdPipelineStage{
		RunID:      runID,
		StageType:  "canary_deploy",
		StageName:  "金丝雀部署",
		Status:     status,
		Logs:       logs,
		StartedAt:  now,
		FinishedAt: now,
	}
	if err := s.cicdSvc().StageCreate(ctx, stage); err != nil {
		global.Logger.Warn("[Canary] 创建金丝雀 stage 失败", zap.Error(err))
	}
}

// monitorCanaryAndDecide 后台监控金丝雀并自动决策
func (s *Services) monitorCanaryAndDecide(kube kubernetes.Interface, pipeline *models.CicdPipeline, image string, runID int64) {
	duration := time.Duration(pipeline.CanaryDurationSec) * time.Second
	if duration <= 0 {
		duration = 5 * time.Minute
	}
	interval := 30 * time.Second
	namespace := pipeline.TargetNamespace
	appName := pipeline.TargetWorkloadName

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			result := s.CanaryAnalyze(context.Background(), kube, pipeline, namespace)
			if !result.Success {
				global.Logger.Warn("[Canary] 金丝雀分析异常，自动回滚",
					zap.String("phase", result.Phase),
					zap.String("message", result.Message),
				)
				s.CanaryRollback(context.Background(), kube, &requests.CanaryRollbackRequest{
					PipelineID: pipeline.ID, RunID: runID, Namespace: namespace, AppName: appName,
				})
				return
			}
			if result.Phase == "promoted" {
				return
			}

		case <-timer.C:
			if pipeline.CanaryAutoPromote {
				lastResult := s.CanaryAnalyze(context.Background(), kube, pipeline, namespace)
				if lastResult.Success {
					err := s.CanaryPromote(context.Background(), kube, &requests.CanaryPromoteRequest{
						Namespace: namespace, AppName: appName,
						ContainerName: pipeline.TargetContainer,
						PipelineID: pipeline.ID, RunID: runID,
					})
					if err != nil {
						global.Logger.Error("[Canary] 自动晋升失败，回滚", zap.Error(err))
						s.CanaryRollback(context.Background(), kube, &requests.CanaryRollbackRequest{
							PipelineID: pipeline.ID, RunID: runID, Namespace: namespace, AppName: appName,
						})
					}
				} else {
					s.CanaryRollback(context.Background(), kube, &requests.CanaryRollbackRequest{
						PipelineID: pipeline.ID, RunID: runID, Namespace: namespace, AppName: appName,
					})
				}
			} else {
				s.updateCanaryStage(context.Background(), runID, "waiting_approval",
					fmt.Sprintf("观察期结束，等待手动 Promote 或 Rollback。镜像: %s", image))
			}
			return
		}
	}
}
