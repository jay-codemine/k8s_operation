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
)

// ============================================================
// CI/CD 发布联动告警静默
// 部署开始时自动创建静默规则，部署完成/失败后自动解除
// ============================================================

// DeploySilenceInfo 部署静默信息（用于传递上下文）
type DeploySilenceInfo struct {
	SilenceRuleID int64  `json:"silence_rule_id"` // 创建的静默规则 ID
	PipelineID    int64  `json:"pipeline_id"`
	PipelineName  string `json:"pipeline_name"`
	Namespace     string `json:"namespace"`
	WorkloadName  string `json:"workload_name"`
	CreatedAt     int64  `json:"created_at"`
}

// CreateDeploySilence 部署开始时创建告警静默规则
// 匹配条件: namespace + workload 相关标签，只静默 warning/info 级别
func (s *Services) CreateDeploySilence(ctx context.Context, pipeline *models.CicdPipeline, namespace, workloadName string) (*DeploySilenceInfo, error) {
	if !pipeline.EnableDeploySilence {
		return nil, nil
	}

	// 计算静默持续时间: 部署最大超时(5min) + 缓冲时间
	bufferMinutes := pipeline.SilenceBufferMinutes
	if bufferMinutes <= 0 {
		bufferMinutes = 10
	}
	// 最长 30 分钟，防止遗忘
	maxMinutes := 30
	totalMinutes := 5 + bufferMinutes // 5min deploy timeout + buffer
	if totalMinutes > maxMinutes {
		totalMinutes = maxMinutes
	}

	// 确定静默的告警级别
	severities := pipeline.SilenceSeverities
	if severities == "" {
		severities = "warning,info"
	}

	// 构建匹配条件 - 匹配目标 namespace + workload 的告警
	matchers := []LabelMatcher{
		{Label: "namespace", Op: "=", Value: namespace},
	}
	// 如果有工作负载名称，增加更精确的匹配
	if workloadName != "" {
		// 匹配常见的工作负载标签（Prometheus 告警规则通常带有这些标签）
		matchers = append(matchers, LabelMatcher{Label: "workload", Op: "=~", Value: workloadName})
	}
	// 按级别匹配（只静默非 critical 级别）
	sevList := strings.Split(severities, ",")
	if len(sevList) > 0 && sevList[0] != "" {
		matchers = append(matchers, LabelMatcher{Label: "severity", Op: "=~", Value: strings.Join(sevList, "|")})
	}

	matchersJSON, err := json.Marshal(matchers)
	if err != nil {
		return nil, fmt.Errorf("序列化匹配条件失败: %v", err)
	}

	now := time.Now()
	endsAt := now.Add(time.Duration(totalMinutes) * time.Minute)

	// 创建静默规则
	rule := &models.MonitorSilenceRule{
		Name:     fmt.Sprintf("[发布静默] %s - %s/%s", pipeline.Name, namespace, workloadName),
		Type:     "silence",
		Matchers: string(matchersJSON),
		StartsAt: now.Unix(),
		EndsAt:   endsAt.Unix(),
		Duration: fmt.Sprintf("%dm", totalMinutes),
		Comment:  fmt.Sprintf("CI/CD 发布自动静默 | 流水线: %s | 目标: %s/%s | 预计 %d 分钟后自动解除", pipeline.Name, namespace, workloadName, totalMinutes),
		Enabled:  true,
	}

	if err := s.dao.DB().WithContext(ctx).Create(rule).Error; err != nil {
		return nil, fmt.Errorf("创建发布静默规则失败: %v", err)
	}

	global.Logger.Info("[发布静默] 静默规则已创建",
		zap.Int64("rule_id", rule.ID),
		zap.String("pipeline", pipeline.Name),
		zap.String("namespace", namespace),
		zap.String("workload", workloadName),
		zap.String("severities", severities),
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

	// 如果部署成功，可以延长一小段静默期（让新 Pod 稳定）
	// 如果部署失败，立即解除静默
	if success {
		// 成功：将结束时间缩短到现在+2分钟（给新 Pod 短暂观察期）
		newEndsAt := time.Now().Add(2 * time.Minute).Unix()
		err := s.dao.DB().WithContext(ctx).Model(&models.MonitorSilenceRule{}).
			Where("id = ? AND is_del = 0", silenceRuleID).
			Updates(map[string]interface{}{
				"ends_at": newEndsAt,
				"comment": fmt.Sprintf("部署成功，静默将于 %s 自动解除", time.Unix(newEndsAt, 0).Format("15:04:05")),
			}).Error
		if err != nil {
			global.Logger.Warn("[发布静默] 更新静默规则结束时间失败", zap.Int64("rule_id", silenceRuleID), zap.Error(err))
		} else {
			global.Logger.Info("[发布静默] 部署成功，静默规则将延续2分钟后解除",
				zap.Int64("rule_id", silenceRuleID),
			)
		}
	} else {
		// 失败：立即解除（禁用规则）
		err := s.dao.DB().WithContext(ctx).Model(&models.MonitorSilenceRule{}).
			Where("id = ? AND is_del = 0", silenceRuleID).
			Updates(map[string]interface{}{
				"enabled": false,
				"ends_at": time.Now().Unix(),
				"comment": "部署失败，静默已立即解除",
			}).Error
		if err != nil {
			global.Logger.Warn("[发布静默] 解除静默规则失败", zap.Int64("rule_id", silenceRuleID), zap.Error(err))
		} else {
			global.Logger.Info("[发布静默] 部署失败，静默规则已立即解除",
				zap.Int64("rule_id", silenceRuleID),
			)
		}
	}
}

// GetActiveDeploySilences 获取当前活跃的发布静默规则
func (s *Services) GetActiveDeploySilences(ctx context.Context, pipelineID int64) ([]models.MonitorSilenceRule, error) {
	now := time.Now().Unix()
	var rules []models.MonitorSilenceRule

	query := s.dao.DB().WithContext(ctx).
		Where("is_del = 0 AND enabled = 1 AND type = 'silence'").
		Where("name LIKE ?", "[发布静默]%").
		Where("ends_at > ?", now)

	if pipelineID > 0 {
		query = query.Where("comment LIKE ?", fmt.Sprintf("%%流水线: %d%%", pipelineID))
	}

	err := query.Order("id DESC").Find(&rules).Error
	return rules, err
}

// CleanExpiredDeploySilences 清理过期的发布静默规则（可选：定期清理）
func (s *Services) CleanExpiredDeploySilences(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	result := s.dao.DB().WithContext(ctx).Model(&models.MonitorSilenceRule{}).
		Where("is_del = 0 AND type = 'silence' AND name LIKE ?", "[发布静默]%").
		Where("ends_at > 0 AND ends_at <= ?", now).
		Update("enabled", false)

	return result.RowsAffected, result.Error
}
