package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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

// canaryAnalysisRule 金丝雀指标分析规则（来自 pipeline.canary_analysis_rules JSON）
type canaryAnalysisRule struct {
	ErrorRate      float64 `json:"error_rate"`       // 错误率阈值（0-1），超过则回滚
	P99LatencyMs   float64 `json:"p99_latency_ms"`   // P99 延迟阈值（ms），超过则回滚
	ErrorRateQuery string  `json:"error_rate_query"` // 可选：自定义错误率 PromQL
	LatencyQuery   string  `json:"latency_query"`    // 可选：自定义 P99 延迟 PromQL
}

// CanaryAnalyze 分析金丝雀指标，判断是否正常、可晋升或需回滚
func (s *Services) CanaryAnalyze(ctx context.Context, kube kubernetes.Interface, pipeline *models.CicdPipeline, namespace string) *CanaryDeployResult {
	info, err := k8sdeploy.GetCanaryStatus(ctx, kube, namespace, pipeline.TargetWorkloadName)
	if err != nil {
		return &CanaryDeployResult{Success: false, Phase: "error", Message: fmt.Sprintf("获取金丝雀状态失败: %v", err)}
	}

	// canary 已被晋升/回滚清理，无需再监控
	if !info.CanaryExists {
		return &CanaryDeployResult{Success: true, Phase: "promoted", Message: "金丝雀已清理（晋升或回滚完成）"}
	}

	// 金丝雀 Pod 未就绪，可能启动失败
	if info.CanaryReadyReplicas == 0 && info.Phase == k8sdeploy.CanaryPhasePending {
		return &CanaryDeployResult{Success: false, Phase: "failing", Message: "金丝雀 Pod 未就绪，可能启动失败"}
	}

	// 镜像已一致，说明已晋升
	if info.CanaryImage == info.StableImage && info.CanaryImage != "" {
		return &CanaryDeployResult{Success: true, Phase: "promoted", Message: "金丝雀已晋升，镜像一致"}
	}

	// 未配置分析规则：只做上述基础检查
	if pipeline.CanaryAnalysisRules == "" {
		return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "无分析规则，跳过指标分析"}
	}

	return s.analyzeCanaryMetrics(ctx, pipeline)
}

// analyzeCanaryMetrics 用 Prometheus 查询金丝雀错误率与 P99 延迟，判断是否超阈值
func (s *Services) analyzeCanaryMetrics(ctx context.Context, pipeline *models.CicdPipeline) *CanaryDeployResult {
	rule, err := parseCanaryAnalysisRule(pipeline.CanaryAnalysisRules)
	if err != nil {
		return &CanaryDeployResult{Success: false, Phase: "error", Message: fmt.Sprintf("解析分析规则失败: %v", err)}
	}

	promClient := s.getDefaultPromClient()
	if promClient == nil {
		return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "Prometheus 不可用，跳过指标分析"}
	}

	// 错误率检查
	if rule.ErrorRate > 0 || rule.ErrorRateQuery != "" {
		q := rule.ErrorRateQuery
		if q == "" {
			q = fmt.Sprintf(`sum(rate(http_requests_total{version=%q,status=~"5.."}[5m])) / clamp_min(sum(rate(http_requests_total{version=%q}[5m])), 1)`,
				k8sdeploy.VersionCanary, k8sdeploy.VersionCanary)
		}
		if val, found, qerr := queryScalar(ctx, promClient, q); qerr != nil {
			global.Logger.Warn("[Canary] 错误率查询失败", zap.Error(qerr))
		} else if found && val > rule.ErrorRate {
			return &CanaryDeployResult{Success: false, Phase: "failing", Message: fmt.Sprintf("金丝雀错误率 %.4f 超过阈值 %.4f", val, rule.ErrorRate)}
		}
	}

	// P99 延迟检查
	if rule.P99LatencyMs > 0 || rule.LatencyQuery != "" {
		q := rule.LatencyQuery
		if q == "" {
			q = fmt.Sprintf(`histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{version=%q}[5m])) by (le))`,
				k8sdeploy.VersionCanary)
		}
		if val, found, qerr := queryScalar(ctx, promClient, q); qerr != nil {
			global.Logger.Warn("[Canary] P99 延迟查询失败", zap.Error(qerr))
		} else if found && val > rule.P99LatencyMs {
			return &CanaryDeployResult{Success: false, Phase: "failing", Message: fmt.Sprintf("金丝雀 P99 延迟 %.0fms 超过阈值 %.0fms", val, rule.P99LatencyMs)}
		}
	}

	return &CanaryDeployResult{Success: true, Phase: "monitoring", Message: "金丝雀指标正常"}
}

// parseCanaryAnalysisRule 解析 canary_analysis_rules JSON
func parseCanaryAnalysisRule(raw string) (*canaryAnalysisRule, error) {
	rule := &canaryAnalysisRule{}
	if err := json.Unmarshal([]byte(raw), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// queryScalar 执行 PromQL 即时查询并取第一个标量值。返回 (值, 是否有有效数据, 错误)。
// 无数据、NaN、Inf 均视为「没有有效数据」，调用方据此跳过该指标的判断。
func queryScalar(ctx context.Context, c *prom.Client, promql string) (float64, bool, error) {
	res, err := c.QueryInstant(ctx, promql)
	if err != nil {
		return 0, false, err
	}
	vec, err := prom.ParseVectorResult(res.Data.Result)
	if err != nil || len(vec) == 0 {
		return 0, false, err
	}
	str, ok := vec[0].Value[1].(string)
	if !ok {
		return 0, false, nil
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false, nil
	}
	return v, true, nil
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
						PipelineID:    pipeline.ID, RunID: runID,
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
