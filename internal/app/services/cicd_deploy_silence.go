package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	dm "k8soperation/internal/domain/monitor"
)

// ============================================================
// CI/CD 发布联动告警静默
// 部署开始时自动创建静默规则，部署完成/失败后自动解除
// ============================================================

// DeploySilenceInfo 部署静默信息（用于传递上下文）
type DeploySilenceInfo struct {
	SilenceRuleID int64  `json:"silence_rule_id"`
	PipelineID    int64  `json:"pipeline_id"`
	PipelineName  string `json:"pipeline_name"`
	Namespace     string `json:"namespace"`
	WorkloadName  string `json:"workload_name"`
	CreatedAt     int64  `json:"created_at"`
}

// CreateDeploySilence 部署开始时创建告警静默规则
func (s *Services) CreateDeploySilence(ctx context.Context, pipeline *models.CicdPipeline, namespace, workloadName string) (*DeploySilenceInfo, error) {
	if !pipeline.EnableDeploySilence {
		return nil, nil
	}

	bufferMinutes := pipeline.SilenceBufferMinutes
	if bufferMinutes <= 0 {
		bufferMinutes = 10
	}
	maxMinutes := 30
	totalMinutes := 5 + bufferMinutes
	if totalMinutes > maxMinutes {
		totalMinutes = maxMinutes
	}

	severities := pipeline.SilenceSeverities
	if severities == "" {
		severities = "warning,info"
	}

	matchers := []dm.LabelMatcher{
		{Label: "namespace", Op: "=", Value: namespace},
	}
	if workloadName != "" {
		matchers = append(matchers, dm.LabelMatcher{Label: "workload", Op: "=~", Value: workloadName})
	}
	sevList := strings.Split(severities, ",")
	if len(sevList) > 0 && sevList[0] != "" {
		matchers = append(matchers, dm.LabelMatcher{Label: "severity", Op: "=~", Value: strings.Join(sevList, "|")})
	}

	matchersJSON, err := json.Marshal(matchers)
	if err != nil {
		return nil, fmt.Errorf("序列化匹配条件失败: %v", err)
	}

	now := time.Now()
	endsAt := now.Add(time.Duration(totalMinutes) * time.Minute)

	rule := &dm.SilenceRule{
		Name:     fmt.Sprintf("[发布静默] %s - %s/%s", pipeline.Name, namespace, workloadName),
		Type:     "silence",
		Matchers: string(matchersJSON),
		StartsAt: now.Unix(),
		EndsAt:   endsAt.Unix(),
		Duration: fmt.Sprintf("%dm", totalMinutes),
		Comment:  fmt.Sprintf("CI/CD 发布自动静默 | 流水线: %s | 目标: %s/%s | 预计 %d 分钟后自动解除", pipeline.Name, namespace, workloadName, totalMinutes),
		Enabled:  true,
	}

	if err := s.MonitorCRUDSvc().CreateSilenceRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("创建发布静默规则失败: %v", err)
	}

	global.Logger.Info("[发布静默] 静默规则已创建",
		zap.Int64("rule_id", rule.ID),
		zap.String("pipeline", pipeline.Name),
		zap.String("namespace", namespace),
		zap.String("workload", workloadName),
		zap.Int("duration_minutes", totalMinutes),
	)

	return &DeploySilenceInfo{
		SilenceRuleID: rule.ID,
		PipelineID:    pipeline.ID,
		PipelineName:  pipeline.Name,
		Namespace:     namespace,
		WorkloadName:  workloadName,
		CreatedAt:     now.Unix(),
	}, nil
}

// ReleaseDeploySilence 部署完成或失败后解除静默规则
func (s *Services) ReleaseDeploySilence(ctx context.Context, silenceRuleID int64, success bool) {
	if silenceRuleID <= 0 {
		return
	}
	svc := s.MonitorCRUDSvc()

	if success {
		newEndsAt := time.Now().Add(2 * time.Minute).Unix()
		err := svc.UpdateSilenceRuleFields(ctx, silenceRuleID, map[string]interface{}{
			"ends_at": newEndsAt,
			"comment": fmt.Sprintf("部署成功，静默将于 %s 自动解除", time.Unix(newEndsAt, 0).Format("15:04:05")),
		})
		if err != nil {
			global.Logger.Warn("[发布静默] 更新静默规则结束时间失败", zap.Int64("rule_id", silenceRuleID), zap.Error(err))
		} else {
			global.Logger.Info("[发布静默] 部署成功，静默规则将延续2分钟后解除", zap.Int64("rule_id", silenceRuleID))
		}
	} else {
		err := svc.UpdateSilenceRuleFields(ctx, silenceRuleID, map[string]interface{}{
			"enabled": false,
			"ends_at": time.Now().Unix(),
			"comment": "部署失败，静默已立即解除",
		})
		if err != nil {
			global.Logger.Warn("[发布静默] 解除静默规则失败", zap.Int64("rule_id", silenceRuleID), zap.Error(err))
		} else {
			global.Logger.Info("[发布静默] 部署失败，静默规则已立即解除", zap.Int64("rule_id", silenceRuleID))
		}
	}
}

// GetActiveDeploySilences 获取当前活跃的发布静默规则
func (s *Services) GetActiveDeploySilences(ctx context.Context, pipelineID int64) ([]dm.SilenceRule, error) {
	rules, err := s.MonitorCRUDSvc().ListDeploySilences(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	result := make([]dm.SilenceRule, len(rules))
	for i, r := range rules {
		result[i] = *r
	}
	return result, nil
}

// CleanExpiredDeploySilences 清理过期的发布静默规则
func (s *Services) CleanExpiredDeploySilences(ctx context.Context) (int64, error) {
	return s.MonitorCRUDSvc().DeactivateExpiredSilences(ctx)
}
