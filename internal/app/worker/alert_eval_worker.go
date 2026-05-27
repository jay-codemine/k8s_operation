package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	prom "k8soperation/pkg/prometheus"
)

// AlertEvalWorker 告警规则评估 Worker
// 定期从数据库加载启用的告警规则，查询 Prometheus 执行 PromQL，
// 根据评估结果产生告警事件（firing）或恢复事件（resolved），并触发通知。
type AlertEvalWorker struct {
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewAlertEvalWorker 创建告警评估 Worker
// interval: 全局评估循环间隔（各规则有独立 eval_interval，Worker 按最小粒度 30s 轮询）
func NewAlertEvalWorker() *AlertEvalWorker {
	return &AlertEvalWorker{
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动
func (w *AlertEvalWorker) Start() {
	global.Logger.Info("[AlertEvalWorker] 启动告警规则评估引擎", zap.Duration("interval", w.interval))
	w.wg.Add(1)
	go w.loop()
}

// Stop 停止
func (w *AlertEvalWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	global.Logger.Info("[AlertEvalWorker] 已停止")
}

func (w *AlertEvalWorker) loop() {
	defer w.wg.Done()

	// 启动 5 秒后执行首次评估
	select {
	case <-time.After(5 * time.Second):
	case <-w.stopCh:
		return
	}
	w.evalOnce()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.evalOnce()
		}
	}
}

// evalOnce 执行一次全量评估
func (w *AlertEvalWorker) evalOnce() {
	ctx := context.Background()
	if global.DB == nil {
		return
	}

	// 1. 加载启用的告警规则
	var rules []models.MonitorAlertRule
	if err := global.DB.WithContext(ctx).
		Where("enabled = 1 AND is_del = 0").
		Find(&rules).Error; err != nil {
		global.Logger.Error("[AlertEvalWorker] 加载告警规则失败", zap.Error(err))
		return
	}
	if len(rules) == 0 {
		return
	}

	now := time.Now().Unix()

	for i := range rules {
		rule := &rules[i]

		// 按各规则的 eval_interval 判断是否该本次评估
		if rule.EvalInterval > 0 && rule.LastEvalAt > 0 {
			if now-rule.LastEvalAt < int64(rule.EvalInterval) {
				continue
			}
		}

		w.evalRule(ctx, rule, now)
	}
}

// evalRule 评估单条规则
func (w *AlertEvalWorker) evalRule(ctx context.Context, rule *models.MonitorAlertRule, now int64) {
	// 获取 Prometheus 客户端
	client, dsID := w.resolvePromClient(ctx, rule.DatasourceID)
	if client == nil {
		w.updateRuleEval(ctx, rule.ID, now, "error", 0)
		return
	}

	// 执行 PromQL 查询
	result, err := client.QueryInstant(ctx, rule.Expr)
	if err != nil {
		global.Logger.Warn("[AlertEvalWorker] PromQL 查询失败",
			zap.String("rule", rule.Name), zap.Error(err))
		w.updateRuleEval(ctx, rule.ID, now, "error", 0)
		return
	}

	// 解析结果 - 判断是否有触发值
	vectors, _ := prom.ParseVectorResult(result.Data.Result)
	triggered := len(vectors) > 0
	var triggerValue string
	var triggerLabels map[string]string
	if triggered {
		// 取第一条结果的值
		if val, ok := vectors[0].Value[1].(string); ok {
			triggerValue = val
		}
		triggerLabels = vectors[0].Metric
	}

	// duration 解析（如 "5m", "1m", "30s"）
	durationSec := parseDuration(rule.Duration)

	if triggered {
		// 条件满足
		if rule.PendingSince == 0 {
			// 首次满足 → 进入 pending 状态
			w.updateRuleEval(ctx, rule.ID, now, "pending", now)
			global.Logger.Info("[AlertEvalWorker] 规则进入 pending",
				zap.String("rule", rule.Name), zap.String("value", triggerValue))
		} else if now-rule.PendingSince >= int64(durationSec) {
			// 持续时间已达 duration → 触发 firing
			w.updateRuleEval(ctx, rule.ID, now, "firing", rule.PendingSince)
			w.fireAlert(ctx, rule, dsID, triggerValue, triggerLabels, now)
		} else {
			// 仍在 pending 中
			w.updateRuleEval(ctx, rule.ID, now, "pending", rule.PendingSince)
		}
	} else {
		// 条件不满足 → normal
		if rule.LastEvalResult == "firing" || rule.LastEvalResult == "pending" {
			// 从告警状态恢复
			w.updateRuleEval(ctx, rule.ID, now, "normal", 0)
			if rule.LastEvalResult == "firing" {
				w.resolveAlert(ctx, rule, now)
			}
		} else {
			w.updateRuleEval(ctx, rule.ID, now, "normal", 0)
		}
	}
}

// fireAlert 产生告警事件
func (w *AlertEvalWorker) fireAlert(ctx context.Context, rule *models.MonitorAlertRule, dsID int64, value string, labels map[string]string, now int64) {
	// 检查是否已有相同规则的 firing 事件（避免重复）
	var existingCount int64
	global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{}).
		Where("rule_id = ? AND status = 'firing'", rule.ID).
		Count(&existingCount)
	if existingCount > 0 {
		return // 已有未恢复的事件
	}

	// 合并规则标签和触发标签
	mergedLabels := make(map[string]string)
	if rule.Labels != "" {
		_ = json.Unmarshal([]byte(rule.Labels), &mergedLabels)
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}
	mergedLabels["alertname"] = rule.Name
	mergedLabels["severity"] = rule.Severity

	labelsJSON, _ := json.Marshal(mergedLabels)
	annotationsJSON := rule.Annotations

	// 生成摘要（替换简单模板变量）
	summary := rule.Summary
	if summary == "" {
		summary = fmt.Sprintf("告警规则 [%s] 触发，当前值: %s", rule.Name, value)
	}

	event := &models.MonitorAlertEvent{
		RuleID:       rule.ID,
		DatasourceID: dsID,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Status:       "firing",
		Value:        value,
		Labels:       string(labelsJSON),
		Annotations:  annotationsJSON,
		Summary:      summary,
		Description:  rule.Description,
		FiredAt:      now,
	}

	if err := global.DB.WithContext(ctx).Create(event).Error; err != nil {
		global.Logger.Error("[AlertEvalWorker] 创建告警事件失败",
			zap.String("rule", rule.Name), zap.Error(err))
		return
	}

	global.Logger.Info("[AlertEvalWorker] 告警触发",
		zap.String("rule", rule.Name),
		zap.String("severity", rule.Severity),
		zap.String("value", value),
		zap.Int64("event_id", event.ID))

	// 发送通知
	go w.sendNotification(ctx, rule, event, "firing")
}

// resolveAlert 恢复告警事件
func (w *AlertEvalWorker) resolveAlert(ctx context.Context, rule *models.MonitorAlertRule, now int64) {
	// 将该规则所有 firing 事件标记为 resolved
	result := global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{}).
		Where("rule_id = ? AND status = 'firing'", rule.ID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
		})
	if result.Error != nil {
		global.Logger.Error("[AlertEvalWorker] 恢复告警事件失败",
			zap.String("rule", rule.Name), zap.Error(result.Error))
		return
	}
	if result.RowsAffected > 0 {
		global.Logger.Info("[AlertEvalWorker] 告警恢复",
			zap.String("rule", rule.Name),
			zap.Int64("resolved_count", result.RowsAffected))

		// 查找刚恢复的事件用于发送恢复通知
		var events []models.MonitorAlertEvent
		global.DB.WithContext(ctx).
			Where("rule_id = ? AND status = 'resolved' AND resolved_at = ?", rule.ID, now).
			Limit(5).Find(&events)
		for i := range events {
			go w.sendNotification(ctx, rule, &events[i], "resolved")
		}
	}
}

// sendNotification 发送告警通知
func (w *AlertEvalWorker) sendNotification(ctx context.Context, rule *models.MonitorAlertRule, event *models.MonitorAlertEvent, status string) {
	notifyChannels := rule.NotifyChannels

	// 路由策略兜底：如果规则没有直接绑定渠道，尝试通过路由策略自动匹配
	if notifyChannels == "" {
		svc := services.NewMonitorCRUDService()
		notifyChannels = svc.ResolveRoutePolicyChannels(ctx, rule)
	}

	if notifyChannels == "" {
		return
	}

	// 解析通知渠道 ID 列表
	channelIDs := strings.Split(notifyChannels, ",")
	var channels []models.MonitorNotifyChannel

	for _, idStr := range channelIDs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		chID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		var ch models.MonitorNotifyChannel
		if err := global.DB.WithContext(ctx).
			Where("id = ? AND enabled = 1 AND is_del = 0", chID).
			First(&ch).Error; err == nil {
			// 恢复通知：需检查渠道是否启用发送恢复通知
			if status == "resolved" && !ch.SendResolved {
				continue
			}
			channels = append(channels, ch)
		}
	}

	if len(channels) == 0 {
		return
	}

	// 构建通知数据
	alert := &services.AlertNotification{
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Status:      status,
		Summary:     event.Summary,
		Description: event.Description,
		Value:       event.Value,
		FiredAt:     event.FiredAt,
		ResolvedAt:  event.ResolvedAt,
		Labels:      parseLabelsJSON(event.Labels),
	}

	var notifyResults []string
	for i := range channels {
		ch := &channels[i]
		if err := services.SendNotification(ch, alert); err != nil {
			global.Logger.Warn("[AlertEvalWorker] 通知发送失败",
				zap.String("channel", ch.Name), zap.Error(err))
			notifyResults = append(notifyResults, fmt.Sprintf("%s:失败", ch.Name))
		} else {
			notifyResults = append(notifyResults, fmt.Sprintf("%s:成功", ch.Name))
		}
	}

	// 更新事件通知结果
	notifyStr := strings.Join(notifyResults, "; ")
	if len(notifyStr) > 200 {
		notifyStr = notifyStr[:197] + "..."
	}
	global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{}).
		Where("id = ?", event.ID).
		Update("notify_result", notifyStr)
}

// updateRuleEval 更新规则评估状态
func (w *AlertEvalWorker) updateRuleEval(ctx context.Context, ruleID int64, evalAt int64, result string, pendingSince int64) {
	global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
		Where("id = ?", ruleID).
		Updates(map[string]interface{}{
			"last_eval_at":     evalAt,
			"last_eval_result": result,
			"pending_since":    pendingSince,
		})
}

// resolvePromClient 获取 Prometheus 客户端
func (w *AlertEvalWorker) resolvePromClient(ctx context.Context, datasourceID int64) (*prom.Client, int64) {
	if global.DB == nil {
		return nil, 0
	}

	var ds models.MonitorDatasource

	// 1. 优先用规则绑定的数据源
	if datasourceID > 0 {
		if err := global.DB.WithContext(ctx).
			Where("id = ? AND enabled = 1 AND is_del = 0", datasourceID).
			First(&ds).Error; err == nil && ds.URL != "" {
			return prom.NewClient(ds.URL, 30*time.Second), ds.ID
		}
	}

	// 2. 回退到默认数据源
	if err := global.DB.WithContext(ctx).
		Where("type IN (?,?,?) AND is_default = 1 AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").
		First(&ds).Error; err == nil && ds.URL != "" {
		return prom.NewClient(ds.URL, 30*time.Second), ds.ID
	}

	// 3. 回退到任一启用的数据源
	if err := global.DB.WithContext(ctx).
		Where("type IN (?,?,?) AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").
		Order("id DESC").First(&ds).Error; err == nil && ds.URL != "" {
		return prom.NewClient(ds.URL, 30*time.Second), ds.ID
	}

	// 4. 最后回退 config.yaml
	if global.MonitoringSetting != nil && global.MonitoringSetting.PrometheusURL != "" {
		return prom.NewClient(global.MonitoringSetting.PrometheusURL, 30*time.Second), 0
	}

	return nil, 0
}

// parseDuration 解析 duration 字符串为秒数（支持 "30s", "1m", "5m", "1h"）
func parseDuration(d string) int {
	d = strings.TrimSpace(d)
	if d == "" {
		return 0
	}
	// 尝试标准 time.ParseDuration
	if dur, err := time.ParseDuration(d); err == nil {
		return int(dur.Seconds())
	}
	// 兼容 Prometheus 格式（无 "s" 的纯数字默认秒）
	if n, err := strconv.Atoi(d); err == nil {
		return n
	}
	return 300 // 默认 5 分钟
}

// parseLabelsJSON 解析标签 JSON 字符串
func parseLabelsJSON(labelsStr string) map[string]string {
	labels := make(map[string]string)
	if labelsStr != "" {
		_ = json.Unmarshal([]byte(labelsStr), &labels)
	}
	return labels
}
