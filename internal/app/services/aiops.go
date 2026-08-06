package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/openai"
)

// =========================================================================
// AIOps 智能运维服务
// 功能: AI 告警分析、AI 日志诊断、智能巡检
// =========================================================================

// AIOpsService AIOps 服务
type AIOpsService struct {
	db *gorm.DB
}

// NewAIOpsService 创建 AIOps 服务
func NewAIOpsService(db *gorm.DB) *AIOpsService {
	return &AIOpsService{db: db}
}

// =========================================================================
// AI 告警分析
// =========================================================================

// AlertAnalysisRequest 告警分析请求
type AlertAnalysisRequest struct {
	EventID    int64  `json:"event_id"`    // 告警事件 ID
	ProviderID string `json:"provider_id"` // AI 提供商（可选）
	ModelID    string `json:"model_id"`    // AI 模型（可选）
}

// AlertAnalysisResult 告警分析结果
type AlertAnalysisResult struct {
	EventID     int64    `json:"event_id"`
	RuleName    string   `json:"rule_name"`
	Severity    string   `json:"severity"`
	Analysis    string   `json:"analysis"`    // AI 分析（Markdown）
	RootCause   string   `json:"root_cause"`  // 根因分析
	Impact      string   `json:"impact"`      // 影响范围
	Suggestions []string `json:"suggestions"` // 处置建议
	Priority    string   `json:"priority"`    // AI 判定优先级
	LatencyMs   int64    `json:"latency_ms"`
}

// AnalyzeAlert AI 分析告警事件
func (s *AIOpsService) AnalyzeAlert(ctx context.Context, req *AlertAnalysisRequest, userID int64) (*AlertAnalysisResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 1. 查询告警事件
	var event models.MonitorAlertEvent
	if err := s.db.WithContext(ctx).Where("id = ?", req.EventID).First(&event).Error; err != nil {
		return nil, fmt.Errorf("告警事件不存在: %w", err)
	}

	// 2. 查询关联规则
	var rule models.MonitorAlertRule
	s.db.WithContext(ctx).Where("id = ?", event.RuleID).First(&rule)

	// 3. 构建 AI 分析 Prompt
	prompt := buildAlertAnalysisPrompt(&event, &rule)

	// 4. 调用 AI
	start := time.Now()
	client, err := getAIOpsClient(req.ProviderID, req.ModelID)
	if err != nil {
		return nil, fmt.Errorf("获取 AI 客户端失败: %w", err)
	}

	messages := []openai.Message{
		{Role: "system", Content: alertAnalysisSystemPrompt},
		{Role: "user", Content: prompt},
	}

	reply, err := client.Chat(ctx, messages)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		// 记录失败
		s.saveRecord(ctx, &models.AIOpsAnalysisRecord{
			Type:      models.AIOpsTypeAlertAnalysis,
			RefID:     event.ID,
			Title:     fmt.Sprintf("告警分析: %s", event.RuleName),
			Input:     prompt,
			Status:    models.AIOpsStatusFailed,
			Error:     err.Error(),
			LatencyMs: latency,
			UserID:    userID,
		})
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 5. 解析 AI 结果
	result := &AlertAnalysisResult{
		EventID:   event.ID,
		RuleName:  event.RuleName,
		Severity:  event.Severity,
		Analysis:  reply,
		LatencyMs: latency,
	}

	// 尝试从回复中提取结构化内容
	s.parseAlertAnalysis(reply, result)

	// 6. 保存分析记录
	sugJSON, _ := json.Marshal(result.Suggestions)
	s.saveRecord(ctx, &models.AIOpsAnalysisRecord{
		Type:        models.AIOpsTypeAlertAnalysis,
		RefID:       event.ID,
		Title:       fmt.Sprintf("告警分析: %s", event.RuleName),
		Input:       prompt,
		Result:      reply,
		Severity:    result.Priority,
		Suggestions: string(sugJSON),
		LatencyMs:   latency,
		Status:      models.AIOpsStatusSuccess,
		UserID:      userID,
	})

	return result, nil
}

// =========================================================================
// AI 日志诊断
// =========================================================================

// LogDiagnosisRequest 日志诊断请求
type LogDiagnosisRequest struct {
	Query      string `json:"query"`       // LogQL 查询
	Namespace  string `json:"namespace"`   // 命名空间
	Pod        string `json:"pod"`         // Pod 名称
	Container  string `json:"container"`   // 容器名称
	TimeRange  string `json:"time_range"`  // 时间范围 (5m/15m/1h/6h)
	ProviderID string `json:"provider_id"` // AI 提供商
	ModelID    string `json:"model_id"`    // AI 模型
}

// LogDiagnosisResult 日志诊断结果
type LogDiagnosisResult struct {
	Query       string   `json:"query"`
	LogLines    int      `json:"log_lines"`    // 分析的日志行数
	ErrorCount  int      `json:"error_count"`  // 错误日志数
	Analysis    string   `json:"analysis"`     // AI 分析（Markdown）
	Pattern     string   `json:"pattern"`      // 异常模式
	RootCause   string   `json:"root_cause"`   // 根因
	Suggestions []string `json:"suggestions"`  // 建议
	Severity    string   `json:"severity"`     // 严重级别
	LatencyMs   int64    `json:"latency_ms"`
}

// DiagnoseLogs AI 诊断日志
func (s *AIOpsService) DiagnoseLogs(ctx context.Context, req *LogDiagnosisRequest, userID int64) (*LogDiagnosisResult, error) {
	// 1. 构建 LogQL 查询
	query := req.Query
	if query == "" {
		// 根据参数自动构建
		parts := []string{}
		if req.Namespace != "" {
			parts = append(parts, fmt.Sprintf(`namespace="%s"`, req.Namespace))
		}
		if req.Pod != "" {
			parts = append(parts, fmt.Sprintf(`pod=~"%s.*"`, req.Pod))
		}
		if req.Container != "" {
			parts = append(parts, fmt.Sprintf(`container="%s"`, req.Container))
		}
		if len(parts) > 0 {
			query = fmt.Sprintf(`{%s}`, strings.Join(parts, ","))
		} else {
			return nil, fmt.Errorf("请指定日志查询条件")
		}
	}

	// 2. 查询 Loki 日志
	lokiSvc := NewLokiService(s.db, "")
	if !lokiSvc.IsEnabled() {
		return nil, fmt.Errorf("Loki 未配置，请先添加 Loki 数据源")
	}

	timeRange := parseLokiTimeRange(req.TimeRange)
	end := time.Now()
	start := end.Add(-timeRange)

	logResult, err := lokiSvc.QueryLogs(ctx, query, start, end, 200, "backward")
	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %w", err)
	}

	if logResult.TotalLines == 0 {
		return &LogDiagnosisResult{
			Query:    query,
			LogLines: 0,
			Analysis: "未查询到日志数据，请确认查询条件和时间范围是否正确。",
			Severity: "info",
		}, nil
	}

	// 3. 提取错误日志 + 采样
	var logSample []string
	errorCount := 0
	for i, entry := range logResult.Entries {
		line := entry.Line
		if isErrorLog(line) {
			errorCount++
		}
		// 最多取 50 行送给 AI
		if i < 50 || isErrorLog(line) {
			if len(logSample) < 80 {
				logSample = append(logSample, line)
			}
		}
	}

	// 4. 构建 AI Prompt
	prompt := buildLogDiagnosisPrompt(query, req.Namespace, req.Pod, logSample, logResult.TotalLines, errorCount)

	// 5. 调用 AI 分析
	aiStart := time.Now()
	client, err := getAIOpsClient(req.ProviderID, req.ModelID)
	if err != nil {
		return nil, fmt.Errorf("获取 AI 客户端失败: %w", err)
	}

	messages := []openai.Message{
		{Role: "system", Content: logDiagnosisSystemPrompt},
		{Role: "user", Content: prompt},
	}

	reply, err := client.Chat(ctx, messages)
	latency := time.Since(aiStart).Milliseconds()

	if err != nil {
		s.saveRecord(ctx, &models.AIOpsAnalysisRecord{
			Type:      models.AIOpsTypeLogDiagnosis,
			Title:     fmt.Sprintf("日志诊断: %s/%s", req.Namespace, req.Pod),
			Input:     prompt,
			Status:    models.AIOpsStatusFailed,
			Error:     err.Error(),
			LatencyMs: latency,
			UserID:    userID,
		})
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	result := &LogDiagnosisResult{
		Query:      query,
		LogLines:   logResult.TotalLines,
		ErrorCount: errorCount,
		Analysis:   reply,
		LatencyMs:  latency,
	}
	s.parseLogDiagnosis(reply, result)

	// 6. 保存记录
	sugJSON, _ := json.Marshal(result.Suggestions)
	s.saveRecord(ctx, &models.AIOpsAnalysisRecord{
		Type:        models.AIOpsTypeLogDiagnosis,
		Title:       fmt.Sprintf("日志诊断: %s/%s", req.Namespace, req.Pod),
		Input:       prompt,
		Result:      reply,
		Severity:    result.Severity,
		Suggestions: string(sugJSON),
		LatencyMs:   latency,
		Status:      models.AIOpsStatusSuccess,
		UserID:      userID,
	})

	return result, nil
}

// =========================================================================
// 巡检相关
// =========================================================================

// InspectionSummary 巡检输入摘要（供 AI 分析）
type InspectionSummary struct {
	ClustersTotal    int      `json:"clusters_total"`
	ClustersHealthy  int      `json:"clusters_healthy"`
	NodesTotal       int      `json:"nodes_total"`
	NodesReady       int      `json:"nodes_ready"`
	WorkloadsTotal   int      `json:"workloads_total"`
	WorkloadsHealthy int      `json:"workloads_healthy"`
	PodsTotal        int      `json:"pods_total"`
	PodsRunning      int      `json:"pods_running"`
	AlertsFiring     int      `json:"alerts_firing"`
	AlertsCritical   int      `json:"alerts_critical"`
	WarningEvents    []string `json:"warning_events"`
	ErrorPods        []string `json:"error_pods"`
}

// RunInspection 执行一次巡检（会创建巡检报告并调用 AI 分析）
func (s *AIOpsService) RunInspection(ctx context.Context, triggerBy int64) (*models.AIOpsInspectionReport, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	inspType := models.InspectionTypeManual
	if triggerBy == 0 {
		inspType = models.InspectionTypeScheduled
	}

	// 创建巡检报告（running 状态）
	report := &models.AIOpsInspectionReport{
		Type:        inspType,
		Scope:       "full",
		Status:      "running",
		Level:       models.InspectionLevelHealthy,
		TriggeredBy: triggerBy,
	}
	if err := s.db.WithContext(ctx).Create(report).Error; err != nil {
		return nil, fmt.Errorf("创建巡检报告失败: %w", err)
	}

	// 异步执行巡检
	go s.executeInspection(context.Background(), report.ID)

	return report, nil
}

// executeInspection 实际执行巡检逻辑
func (s *AIOpsService) executeInspection(ctx context.Context, reportID int64) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			global.Logger.Error("[AIOps] 巡检执行 panic", zap.Any("panic", r))
			s.updateReport(ctx, reportID, map[string]interface{}{
				"status": "failed",
				"error":  fmt.Sprintf("巡检执行异常: %v", r),
			})
		}
	}()

	// 1. 收集平台健康数据
	healthSvc := NewPlatformHealthService(s.db)
	health, err := healthSvc.GetFullHealth(ctx)
	if err != nil {
		s.updateReport(ctx, reportID, map[string]interface{}{
			"status": "failed",
			"error":  "获取平台健康数据失败: " + err.Error(),
		})
		return
	}

	// 2. 收集告警数据
	var firingAlerts []models.MonitorAlertEvent
	var firingCount, criticalCount int64
	if s.db != nil {
		s.db.Model(&models.MonitorAlertEvent{}).Where("status = 'firing'").Count(&firingCount)
		s.db.Model(&models.MonitorAlertEvent{}).Where("status = 'firing' AND severity = 'critical'").Count(&criticalCount)
		s.db.Where("status = 'firing'").Order("fired_at DESC").Limit(10).Find(&firingAlerts)
	}

	// 3. 构建巡检摘要
	summary := &InspectionSummary{
		ClustersTotal:    health.Clusters.Total,
		ClustersHealthy:  health.Clusters.Online,
		NodesTotal:       health.Nodes.Total,
		NodesReady:       health.Nodes.Ready,
		WorkloadsTotal:   health.Workloads.Deployments.Total + health.Workloads.StatefulSets.Total + health.Workloads.DaemonSets.Total,
		WorkloadsHealthy: health.Workloads.Deployments.Running + health.Workloads.StatefulSets.Running + health.Workloads.DaemonSets.Running,
		PodsTotal:        health.Workloads.Pods.Total,
		PodsRunning:      health.Workloads.Pods.Running,
		AlertsFiring:     int(firingCount),
		AlertsCritical:   int(criticalCount),
	}

	// 提取异常事件
	for _, alert := range firingAlerts {
		summary.WarningEvents = append(summary.WarningEvents,
			fmt.Sprintf("[%s] %s: %s", alert.Severity, alert.RuleName, alert.Summary))
	}

	// 4. 计算健康评分
	score := calculateHealthScore(summary)
	level := models.InspectionLevelHealthy
	if score < 60 {
		level = models.InspectionLevelCritical
	} else if score < 80 {
		level = models.InspectionLevelWarning
	}

	// 5. 调用 AI 生成巡检分析
	var aiAnalysis string
	findings := 0

	client, aiErr := getAIOpsClient("", "")
	if aiErr == nil {
		prompt := buildInspectionPrompt(summary, score)
		messages := []openai.Message{
			{Role: "system", Content: inspectionSystemPrompt},
			{Role: "user", Content: prompt},
		}
		aiAnalysis, _ = client.Chat(ctx, messages)
	}

	// 统计发现问题数
	if summary.AlertsFiring > 0 {
		findings += summary.AlertsFiring
	}
	if summary.NodesTotal > summary.NodesReady {
		findings += summary.NodesTotal - summary.NodesReady
	}
	if summary.WorkloadsTotal > summary.WorkloadsHealthy {
		findings += summary.WorkloadsTotal - summary.WorkloadsHealthy
	}

	detailsJSON, _ := json.Marshal(summary)
	duration := time.Since(start).Milliseconds()

	// 6. 更新巡检报告
	s.updateReport(ctx, reportID, map[string]interface{}{
		"health_score": score,
		"level":        level,
		"summary":      fmt.Sprintf("健康评分 %d/100 | 集群 %d/%d | 节点 %d/%d | 告警 %d 条", score, summary.ClustersHealthy, summary.ClustersTotal, summary.NodesReady, summary.NodesTotal, summary.AlertsFiring),
		"details":      string(detailsJSON),
		"ai_analysis":  aiAnalysis,
		"findings":     findings,
		"suggestions":  countSuggestions(aiAnalysis),
		"duration":     duration,
		"status":       "completed",
		"completed_at": time.Now().Unix(),
	})

	global.Logger.Info("[AIOps] 巡检完成",
		zap.Int64("report_id", reportID),
		zap.Int("score", score),
		zap.String("level", level),
		zap.Int("findings", findings),
		zap.Int64("duration_ms", duration))
}

// =========================================================================
// 报告导出 & 通知
// =========================================================================

// ExportReportMarkdown 导出巡检报告为 Markdown 格式
func (s *AIOpsService) ExportReportMarkdown(ctx context.Context, id int64) (string, error) {
	report, err := s.GetInspectionReport(ctx, id)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("# 🛡️ K8s 智能巡检报告\n\n")
	sb.WriteString(fmt.Sprintf("**报告编号**: #%d\n\n", report.ID))
	sb.WriteString(fmt.Sprintf("**巡检类型**: %s\n\n", map[string]string{"scheduled": "定时巡检", "manual": "手动巡检"}[report.Type]))
	sb.WriteString(fmt.Sprintf("**执行时间**: %s\n\n", time.Unix(report.CreatedAt, 0).Format("2006-01-02 15:04:05")))
	if report.CompletedAt > 0 {
		sb.WriteString(fmt.Sprintf("**完成时间**: %s\n\n", time.Unix(report.CompletedAt, 0).Format("2006-01-02 15:04:05")))
	}
	sb.WriteString(fmt.Sprintf("**耗时**: %dms\n\n", report.Duration))
	sb.WriteString("---\n\n")

	// 健康评分
	levelCN := map[string]string{"healthy": "🟢 健康", "warning": "🟡 警告", "critical": "🔴 严重"}
	sb.WriteString("## 📊 健康评分\n\n")
	sb.WriteString(fmt.Sprintf("| 指标 | 值 |\n|------|------|\n"))
	sb.WriteString(fmt.Sprintf("| 健康评分 | **%d / 100** |\n", report.HealthScore))
	sb.WriteString(fmt.Sprintf("| 健康等级 | %s |\n", levelCN[report.Level]))
	sb.WriteString(fmt.Sprintf("| 问题发现 | %d 项 |\n", report.Findings))
	sb.WriteString(fmt.Sprintf("| 优化建议 | %d 条 |\n\n", report.Suggestions))

	// 巡检详情
	if report.Details != "" {
		var summary InspectionSummary
		if json.Unmarshal([]byte(report.Details), &summary) == nil {
			sb.WriteString("## 🔍 巡检维度\n\n")
			sb.WriteString("| 维度 | 总数 | 健康 | 异常 |\n|------|------|------|------|\n")
			sb.WriteString(fmt.Sprintf("| 集群 | %d | %d | %d |\n", summary.ClustersTotal, summary.ClustersHealthy, summary.ClustersTotal-summary.ClustersHealthy))
			sb.WriteString(fmt.Sprintf("| 节点 | %d | %d | %d |\n", summary.NodesTotal, summary.NodesReady, summary.NodesTotal-summary.NodesReady))
			sb.WriteString(fmt.Sprintf("| 工作负载 | %d | %d | %d |\n", summary.WorkloadsTotal, summary.WorkloadsHealthy, summary.WorkloadsTotal-summary.WorkloadsHealthy))
			sb.WriteString(fmt.Sprintf("| 活跃告警 | %d | - | %d |\n\n", summary.AlertsFiring, summary.AlertsCritical))
		}
	}

	// AI 分析
	if report.AIAnalysis != "" {
		sb.WriteString("## 🤖 AI 智能分析\n\n")
		sb.WriteString(report.AIAnalysis)
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("*由 K8s Operation 智能运维平台自动生成 · %s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return sb.String(), nil
}

// NotifyReportRequest 发送巡检报告通知请求
type NotifyReportRequest struct {
	ReportID   int64   `json:"report_id"`
	ChannelIDs []int64 `json:"channel_ids"` // 通知渠道 ID 列表
}

// NotifyReportResult 发送结果
type NotifyReportResult struct {
	ChannelID   int64  `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// NotifyReport 将巡检报告发送到指定通知渠道
func (s *AIOpsService) NotifyReport(ctx context.Context, req *NotifyReportRequest) ([]NotifyReportResult, error) {
	report, err := s.GetInspectionReport(ctx, req.ReportID)
	if err != nil {
		return nil, err
	}

	// 获取通知渠道列表
	var channels []models.MonitorNotifyChannel
	if err := s.db.WithContext(ctx).Where("id IN ? AND is_del = 0 AND enabled = 1", req.ChannelIDs).Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("获取通知渠道失败: %w", err)
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("未找到有效的通知渠道")
	}

	// 构建巡检通知消息
	alert := s.buildReportNotification(report)

	// 逐一发送
	results := make([]NotifyReportResult, 0, len(channels))
	for _, ch := range channels {
		result := NotifyReportResult{
			ChannelID:   ch.ID,
			ChannelName: ch.Name,
			ChannelType: ch.Type,
		}
		if err := SendNotification(&ch, alert); err != nil {
			result.Success = false
			result.Error = err.Error()
			global.Logger.Warn("[AIOps] 巡检报告通知发送失败",
				zap.Int64("report_id", req.ReportID),
				zap.String("channel", ch.Name),
				zap.Error(err))
		} else {
			result.Success = true
			global.Logger.Info("[AIOps] 巡检报告通知发送成功",
				zap.Int64("report_id", req.ReportID),
				zap.String("channel", ch.Name))
		}
		results = append(results, result)
	}

	return results, nil
}

// GetNotifyChannels 获取可用通知渠道（复用监控通知渠道）
func (s *AIOpsService) GetNotifyChannels(ctx context.Context) ([]map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var channels []models.MonitorNotifyChannel
	if err := s.db.WithContext(ctx).Where("is_del = 0 AND enabled = 1").Order("id DESC").Find(&channels).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(channels))
	for _, ch := range channels {
		result = append(result, map[string]interface{}{
			"id":   ch.ID,
			"name": ch.Name,
			"type": ch.Type,
		})
	}
	return result, nil
}

// buildReportNotification 构建巡检报告通知体
func (s *AIOpsService) buildReportNotification(report *models.AIOpsInspectionReport) *AlertNotification {
	levelMap := map[string]string{"healthy": "info", "warning": "warning", "critical": "critical"}
	severity := levelMap[report.Level]
	if severity == "" {
		severity = "info"
	}

	summary := fmt.Sprintf("巡检评分: %d/100 | 等级: %s | 发现问题: %d", report.HealthScore, report.Level, report.Findings)
	description := report.Summary
	if report.AIAnalysis != "" && len(report.AIAnalysis) > 500 {
		description += "\n\n" + report.AIAnalysis[:500] + "..."
	} else if report.AIAnalysis != "" {
		description += "\n\n" + report.AIAnalysis
	}

	return &AlertNotification{
		RuleName:    fmt.Sprintf("智能巡检报告 #%d", report.ID),
		Severity:    severity,
		Status:      "firing",
		Summary:     summary,
		Description: description,
		FiredAt:     report.CreatedAt,
		Labels: map[string]string{
			"type":   "inspection",
			"level":  report.Level,
			"source": "aiops",
		},
	}
}

// =========================================================================
// 查询接口
// =========================================================================

// GetAnalysisRecords 获取分析记录列表
func (s *AIOpsService) GetAnalysisRecords(ctx context.Context, recordType string, page, pageSize int) ([]models.AIOpsAnalysisRecord, int64, error) {
	if s.db == nil {
		return nil, 0, fmt.Errorf("数据库未初始化")
	}

	var records []models.AIOpsAnalysisRecord
	var total int64

	db := s.db.WithContext(ctx).Model(&models.AIOpsAnalysisRecord{})
	if recordType != "" {
		db = db.Where("type = ?", recordType)
	}
	db.Count(&total)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)
	return records, total, nil
}

// GetInspectionReports 获取巡检报告列表
func (s *AIOpsService) GetInspectionReports(ctx context.Context, page, pageSize int) ([]models.AIOpsInspectionReport, int64, error) {
	if s.db == nil {
		return nil, 0, fmt.Errorf("数据库未初始化")
	}

	var reports []models.AIOpsInspectionReport
	var total int64

	db := s.db.WithContext(ctx).Model(&models.AIOpsInspectionReport{})
	db.Count(&total)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&reports)
	return reports, total, nil
}

// GetInspectionReport 获取单个巡检报告
func (s *AIOpsService) GetInspectionReport(ctx context.Context, id int64) (*models.AIOpsInspectionReport, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var report models.AIOpsInspectionReport
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&report).Error; err != nil {
		return nil, fmt.Errorf("巡检报告不存在")
	}
	return &report, nil
}

// GetDashboardStats AIOps 仪表盘统计
func (s *AIOpsService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	stats := map[string]interface{}{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	// 今日分析次数
	var todayAnalysis int64
	s.db.Model(&models.AIOpsAnalysisRecord{}).Where("created_at >= ?", today).Count(&todayAnalysis)
	stats["today_analysis"] = todayAnalysis

	// 总分析次数
	var totalAnalysis int64
	s.db.Model(&models.AIOpsAnalysisRecord{}).Count(&totalAnalysis)
	stats["total_analysis"] = totalAnalysis

	// 最近巡检评分
	var lastReport models.AIOpsInspectionReport
	if s.db.Where("status = 'completed'").Order("created_at DESC").First(&lastReport).Error == nil {
		stats["last_health_score"] = lastReport.HealthScore
		stats["last_health_level"] = lastReport.Level
		stats["last_inspection_at"] = lastReport.CompletedAt
	}

	// 当前 firing 告警数
	var firingCount int64
	s.db.Model(&models.MonitorAlertEvent{}).Where("status = 'firing'").Count(&firingCount)
	stats["firing_alerts"] = firingCount

	// 本周分析次数
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Unix()
	var weekAnalysis int64
	s.db.Model(&models.AIOpsAnalysisRecord{}).Where("created_at >= ?", weekStart).Count(&weekAnalysis)
	stats["week_analysis"] = weekAnalysis

	return stats, nil
}

// =========================================================================
// 内部方法
// =========================================================================

func (s *AIOpsService) saveRecord(ctx context.Context, record *models.AIOpsAnalysisRecord) {
	if s.db == nil {
		return
	}
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		global.Logger.Warn("[AIOps] 保存分析记录失败", zap.Error(err))
	}
}

func (s *AIOpsService) updateReport(ctx context.Context, reportID int64, updates map[string]interface{}) {
	if s.db == nil {
		return
	}
	if err := s.db.WithContext(ctx).Model(&models.AIOpsInspectionReport{}).Where("id = ?", reportID).Updates(updates).Error; err != nil {
		global.Logger.Error("[AIOps] 更新巡检报告失败", zap.Int64("report_id", reportID), zap.Error(err))
	}
}

func (s *AIOpsService) parseAlertAnalysis(reply string, result *AlertAnalysisResult) {
	// 简单解析：从 AI 回复中提取关键段落
	lines := strings.Split(reply, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "根因") || strings.Contains(lower, "root cause") {
			result.RootCause = strings.TrimSpace(line)
		}
		if strings.Contains(lower, "影响") || strings.Contains(lower, "impact") {
			result.Impact = strings.TrimSpace(line)
		}
		if strings.Contains(lower, "建议") || strings.Contains(lower, "suggestion") {
			if strings.HasPrefix(strings.TrimSpace(line), "-") || strings.HasPrefix(strings.TrimSpace(line), "•") {
				result.Suggestions = append(result.Suggestions, strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "-•")))
			}
		}
	}
	// 优先级判断
	result.Priority = result.Severity
	if strings.Contains(reply, "紧急") || strings.Contains(reply, "立即") {
		result.Priority = "critical"
	}
}

func (s *AIOpsService) parseLogDiagnosis(reply string, result *LogDiagnosisResult) {
	if strings.Contains(reply, "critical") || strings.Contains(reply, "严重") {
		result.Severity = "critical"
	} else if strings.Contains(reply, "warning") || strings.Contains(reply, "警告") {
		result.Severity = "warning"
	} else {
		result.Severity = "info"
	}
}

// =========================================================================
// AI 客户端获取
// =========================================================================

func getAIOpsClient(providerID, modelID string) (*openai.Client, error) {
	return getAIClientWithModel(providerID, modelID)
}

// =========================================================================
// Prompt 模板
// =========================================================================

const alertAnalysisSystemPrompt = `你是一个专业的 Kubernetes 运维 AIOps 专家。
你的任务是分析告警事件，给出专业的根因分析、影响范围评估和处置建议。

输出格式要求（Markdown）：
## 🔍 根因分析
简要说明告警触发的根本原因

## 💥 影响范围
说明该告警可能影响的服务和用户

## 🎯 优先级评估
评估处理优先级（P0紧急/P1高/P2中/P3低）

## 📋 处置建议
- 建议1
- 建议2
- 建议3

## ⚡ 快速处置命令
给出可直接执行的 kubectl 命令（如适用）`

const logDiagnosisSystemPrompt = `你是一个专业的 Kubernetes 日志分析专家。
你的任务是分析应用日志，识别异常模式、错误根因，并给出修复建议。

输出格式要求（Markdown）：
## 📊 异常模式
总结日志中发现的异常模式

## 🔍 错误分析
分析关键错误的原因

## 💡 根因判断
给出最可能的根本原因

## 📋 修复建议
- 建议1
- 建议2
- 建议3

## ⚠️ 风险等级
评估当前问题的严重程度（critical/warning/info）`

const inspectionSystemPrompt = `你是一个专业的 Kubernetes 平台巡检专家。
请根据提供的平台健康数据，生成一份专业的巡检分析报告。

输出格式要求（Markdown）：
## 📊 整体评估
对平台当前健康状态的总体评估

## 🔍 问题发现
列出发现的问题（如有）

## 💡 优化建议
- 建议1
- 建议2
- 建议3

## 📈 趋势预测
基于当前数据给出趋势预测和预防性建议

## ✅ 巡检结论
一句话总结巡检结论`

// =========================================================================
// Prompt 构建
// =========================================================================

func buildAlertAnalysisPrompt(event *models.MonitorAlertEvent, rule *models.MonitorAlertRule) string {
	var sb strings.Builder
	sb.WriteString("请分析以下 Kubernetes 告警事件：\n\n")
	sb.WriteString(fmt.Sprintf("**告警名称**: %s\n", event.RuleName))
	sb.WriteString(fmt.Sprintf("**严重级别**: %s\n", event.Severity))
	sb.WriteString(fmt.Sprintf("**触发值**: %s\n", event.Value))
	sb.WriteString(fmt.Sprintf("**告警摘要**: %s\n", event.Summary))
	if event.Description != "" {
		sb.WriteString(fmt.Sprintf("**告警描述**: %s\n", event.Description))
	}
	if event.Labels != "" {
		sb.WriteString(fmt.Sprintf("**标签**: %s\n", event.Labels))
	}
	if rule.Expr != "" {
		sb.WriteString(fmt.Sprintf("**PromQL**: `%s`\n", rule.Expr))
	}
	if rule.Duration != "" {
		sb.WriteString(fmt.Sprintf("**持续时间**: %s\n", rule.Duration))
	}
	firedAt := time.Unix(event.FiredAt, 0).Format("2006-01-02 15:04:05")
	sb.WriteString(fmt.Sprintf("**触发时间**: %s\n", firedAt))
	return sb.String()
}

func buildLogDiagnosisPrompt(query, namespace, pod string, logSample []string, totalLines, errorCount int) string {
	var sb strings.Builder
	sb.WriteString("请分析以下 Kubernetes 应用日志：\n\n")
	sb.WriteString(fmt.Sprintf("**查询条件**: `%s`\n", query))
	if namespace != "" {
		sb.WriteString(fmt.Sprintf("**命名空间**: %s\n", namespace))
	}
	if pod != "" {
		sb.WriteString(fmt.Sprintf("**Pod**: %s\n", pod))
	}
	sb.WriteString(fmt.Sprintf("**日志总行数**: %d\n", totalLines))
	sb.WriteString(fmt.Sprintf("**错误日志数**: %d\n", errorCount))
	sb.WriteString("\n**日志样本**（最近日志）:\n```\n")
	for _, line := range logSample {
		sb.WriteString(line + "\n")
	}
	sb.WriteString("```\n")
	return sb.String()
}

func buildInspectionPrompt(summary *InspectionSummary, score int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("平台巡检数据（健康评分: %d/100）：\n\n", score))
	sb.WriteString(fmt.Sprintf("- 集群: %d 个，健康 %d 个\n", summary.ClustersTotal, summary.ClustersHealthy))
	sb.WriteString(fmt.Sprintf("- 节点: %d 个，就绪 %d 个\n", summary.NodesTotal, summary.NodesReady))
	sb.WriteString(fmt.Sprintf("- 工作负载: %d 个，健康 %d 个\n", summary.WorkloadsTotal, summary.WorkloadsHealthy))
	sb.WriteString(fmt.Sprintf("- 活跃告警: %d 条（其中严重 %d 条）\n", summary.AlertsFiring, summary.AlertsCritical))

	if len(summary.WarningEvents) > 0 {
		sb.WriteString("\n**当前活跃告警**:\n")
		for _, e := range summary.WarningEvents {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}
	return sb.String()
}

// =========================================================================
// 工具函数
// =========================================================================

func calculateHealthScore(summary *InspectionSummary) int {
	score := 100

	// 集群健康扣分
	if summary.ClustersTotal > 0 {
		unhealthyClusters := summary.ClustersTotal - summary.ClustersHealthy
		score -= unhealthyClusters * 20
	}

	// 节点健康扣分
	if summary.NodesTotal > 0 {
		unhealthyNodes := summary.NodesTotal - summary.NodesReady
		score -= unhealthyNodes * 10
	}

	// 工作负载扣分
	if summary.WorkloadsTotal > 0 {
		unhealthyWL := summary.WorkloadsTotal - summary.WorkloadsHealthy
		score -= unhealthyWL * 2
	}

	// 告警扣分
	score -= summary.AlertsFiring * 3
	score -= summary.AlertsCritical * 5

	if score < 0 {
		score = 0
	}
	return score
}

func countSuggestions(analysis string) int {
	count := 0
	lines := strings.Split(analysis, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "• ") {
			count++
		}
	}
	return count
}

func isErrorLog(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, "error") ||
		strings.Contains(lower, "exception") ||
		strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic") ||
		strings.Contains(lower, "fail")
}

func parseLokiTimeRange(tr string) time.Duration {
	switch tr {
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "6h":
		return 6 * time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return 15 * time.Minute
	}
}
