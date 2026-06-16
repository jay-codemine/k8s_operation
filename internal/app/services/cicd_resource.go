package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8soperation/internal/app/models"
)

// ==================== 资源模板服务 ====================

// ResourceTemplateList 获取资源模板列表
func (s *Services) ResourceTemplateList(ctx context.Context, env, serviceType string) ([]models.CicdResourceTemplate, error) {
	return s.dao.ResourceTemplateList(ctx, env, serviceType)
}

// ResourceTemplateGetByID 根据ID获取模板
func (s *Services) ResourceTemplateGetByID(ctx context.Context, id uint64) (*models.CicdResourceTemplate, error) {
	return s.dao.ResourceTemplateGetByID(ctx, id)
}

// ResourceTemplateGetDefault 获取默认模板
func (s *Services) ResourceTemplateGetDefault(ctx context.Context, env, serviceType string) (*models.CicdResourceTemplate, error) {
	return s.dao.ResourceTemplateGetDefault(ctx, env, serviceType)
}

// ResourceTemplateCreate 创建模板
func (s *Services) ResourceTemplateCreate(ctx context.Context, tpl *models.CicdResourceTemplate) error {
	return s.dao.ResourceTemplateCreate(ctx, tpl)
}

// ResourceTemplateUpdate 更新模板
func (s *Services) ResourceTemplateUpdate(ctx context.Context, tpl *models.CicdResourceTemplate) error {
	return s.dao.ResourceTemplateUpdate(ctx, tpl)
}

// ResourceTemplateDelete 删除模板
func (s *Services) ResourceTemplateDelete(ctx context.Context, id uint64) error {
	return s.dao.ResourceTemplateDelete(ctx, id)
}

// ==================== 环境资源规则服务 ====================

// EnvResourceRuleList 获取环境规则列表
func (s *Services) EnvResourceRuleList(ctx context.Context, env string) ([]models.CicdEnvResourceRule, error) {
	return s.dao.EnvResourceRuleList(ctx, env)
}

// EnvResourceRuleGet 获取环境规则
func (s *Services) EnvResourceRuleGet(ctx context.Context, env, serviceType string) (*models.CicdEnvResourceRule, error) {
	return s.dao.EnvResourceRuleGet(ctx, env, serviceType)
}

// EnvResourceRuleUpdate 更新规则
func (s *Services) EnvResourceRuleUpdate(ctx context.Context, rule *models.CicdEnvResourceRule) error {
	return s.dao.EnvResourceRuleUpdate(ctx, rule)
}

// ==================== 资源校验服务（核心） ====================

// ValidateResourceConfig 校验资源配置
func (s *Services) ValidateResourceConfig(ctx context.Context, env, serviceType string, config models.ResourceConfig) *models.ResourceValidationResult {
	result := &models.ResourceValidationResult{
		Valid:     true,
		RiskLevel: models.RiskLevelLow,
	}

	// 1. 获取环境规则
	rule, err := s.dao.EnvResourceRuleGet(ctx, env, serviceType)
	if err != nil {
		result.Warnings = append(result.Warnings, "未找到环境规则，使用默认校验")
		rule = &models.CicdEnvResourceRule{
			Env:              env,
			CPULimitMax:      "4",
			MemoryLimitMax:   "8Gi",
			ReplicasMax:      20,
			ReplicasMin:      1,
			RequireApproval:  env == "prod",
			ApprovalRole:     "sre",
		}
	}

	// 2. 副本数校验
	if config.Replicas < rule.ReplicasMin {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("%s环境副本数不能小于%d", env, rule.ReplicasMin))
		result.Valid = false
	}
	if config.Replicas > rule.ReplicasMax {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("%s环境副本数不能超过%d", env, rule.ReplicasMax))
		result.Valid = false
	}

	// 3. CPU 校验
	if err := s.validateCPU(config.Resources.Requests.CPU, config.Resources.Limits.CPU, rule, env, serviceType, result); err != nil {
		result.Valid = false
	}

	// 4. 内存校验
	if err := s.validateMemory(config.Resources.Requests.Memory, config.Resources.Limits.Memory, rule, env, serviceType, result); err != nil {
		result.Valid = false
	}

	// 5. request <= limit 校验
	if !s.validateRequestLessThanLimit(config.Resources, result) {
		result.Valid = false
	}

	// 6. 风险提示
	s.addRiskWarnings(env, serviceType, config, result)

	// 7. 审批判断
	if rule.RequireApproval {
		result.NeedApproval = true
		result.ApprovalRole = rule.ApprovalRole
	}

	// 8. 生成建议
	result.Suggestion = s.generateSuggestion(env, serviceType, config, result)

	return result
}

// validateCPU CPU校验
func (s *Services) validateCPU(request, limit string, rule *models.CicdEnvResourceRule, env, serviceType string, result *models.ResourceValidationResult) error {
	// 校验 limit 不超过最大值
	if rule.CPULimitMax != "" {
		limitMilli := parseCPUToMilli(limit)
		maxMilli := parseCPUToMilli(rule.CPULimitMax)
		if limitMilli > maxMilli {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("CPU limit(%s)超过%s环境最大值(%s)", limit, env, rule.CPULimitMax))
			return fmt.Errorf("cpu limit exceeded")
		}
	}

	// 校验 request 不小于最小值
	if rule.CPURequestMin != "" {
		requestMilli := parseCPUToMilli(request)
		minMilli := parseCPUToMilli(rule.CPURequestMin)
		if requestMilli < minMilli {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("CPU request(%s)低于%s环境最小值(%s)", request, env, rule.CPURequestMin))
			return fmt.Errorf("cpu request too low")
		}
	}

	return nil
}

// validateMemory 内存校验
func (s *Services) validateMemory(request, limit string, rule *models.CicdEnvResourceRule, env, serviceType string, result *models.ResourceValidationResult) error {
	// 校验 limit 不超过最大值
	if rule.MemoryLimitMax != "" {
		limitBytes := parseMemoryToBytes(limit)
		maxBytes := parseMemoryToBytes(rule.MemoryLimitMax)
		if limitBytes > maxBytes {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("内存limit(%s)超过%s环境最大值(%s)", limit, env, rule.MemoryLimitMax))
			return fmt.Errorf("memory limit exceeded")
		}
	}

	// 校验 request 不小于最小值
	if rule.MemoryRequestMin != "" {
		requestBytes := parseMemoryToBytes(request)
		minBytes := parseMemoryToBytes(rule.MemoryRequestMin)
		if requestBytes < minBytes {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("内存request(%s)低于%s环境最小值(%s)", request, env, rule.MemoryRequestMin))
			return fmt.Errorf("memory request too low")
		}
	}

	return nil
}

// validateRequestLessThanLimit 校验 request <= limit
func (s *Services) validateRequestLessThanLimit(resources models.ResourceRequests, result *models.ResourceValidationResult) bool {
	valid := true

	// CPU: request <= limit
	requestCPU := parseCPUToMilli(resources.Requests.CPU)
	limitCPU := parseCPUToMilli(resources.Limits.CPU)
	if requestCPU > limitCPU {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("CPU request(%s)不能大于limit(%s)", resources.Requests.CPU, resources.Limits.CPU))
		valid = false
	}

	// Memory: request <= limit
	requestMem := parseMemoryToBytes(resources.Requests.Memory)
	limitMem := parseMemoryToBytes(resources.Limits.Memory)
	if requestMem > limitMem {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("内存request(%s)不能大于limit(%s)", resources.Requests.Memory, resources.Limits.Memory))
		valid = false
	}

	return valid
}

// addRiskWarnings 添加风险提示
func (s *Services) addRiskWarnings(env, serviceType string, config models.ResourceConfig, result *models.ResourceValidationResult) {
	// 生产环境单副本风险
	if env == "prod" && config.Replicas == 1 {
		result.Warnings = append(result.Warnings, "生产环境仅1副本，存在高可用风险")
		result.RiskLevel = models.RiskLevelHigh
	}

	// Java 服务内存过小风险
	if serviceType == "java" {
		memBytes := parseMemoryToBytes(config.Resources.Limits.Memory)
		if memBytes < 1024*1024*1024 { // < 1Gi
			result.Warnings = append(result.Warnings, "Java服务内存limit小于1Gi，存在OOM风险")
			if result.RiskLevel != models.RiskLevelHigh {
				result.RiskLevel = models.RiskLevelMedium
			}
		}
	}

	// 生产环境未启用HPA提示
	if env == "prod" && (config.HPA == nil || !config.HPA.Enabled) {
		result.Warnings = append(result.Warnings, "生产环境建议启用HPA自动伸缩")
	}

	// CPU 配置过小
	cpuMilli := parseCPUToMilli(config.Resources.Limits.CPU)
	if env == "prod" && cpuMilli < 500 {
		result.Warnings = append(result.Warnings, "生产环境CPU limit建议不低于500m")
	}
}

// generateSuggestion 生成建议
func (s *Services) generateSuggestion(env, serviceType string, config models.ResourceConfig, result *models.ResourceValidationResult) string {
	if len(result.Errors) > 0 {
		return "请修正上述错误后重试"
	}

	if result.RiskLevel == models.RiskLevelHigh {
		return fmt.Sprintf("当前配置风险较高，建议：%s环境至少2副本，%s服务内存limit建议1Gi以上", env, serviceType)
	}

	if result.RiskLevel == models.RiskLevelMedium {
		return "当前配置存在一定风险，请根据实际业务情况评估"
	}

	return "配置符合规范"
}

// ==================== 发布审批服务 ====================

// DeployApprovalCreate 创建审批
func (s *Services) DeployApprovalCreate(ctx context.Context, approval *models.CicdDeployApproval) error {
	return s.dao.DeployApprovalCreate(ctx, approval)
}

// DeployApprovalGetByID 获取审批详情
func (s *Services) DeployApprovalGetByID(ctx context.Context, id uint64) (*models.CicdDeployApproval, error) {
	return s.dao.DeployApprovalGetByID(ctx, id)
}

// DeployApprovalList 获取审批列表
func (s *Services) DeployApprovalList(ctx context.Context, status string, page, pageSize int) ([]models.CicdDeployApproval, int64, error) {
	return s.dao.DeployApprovalList(ctx, status, page, pageSize)
}

// DeployApprovalApprove 通过审批
func (s *Services) DeployApprovalApprove(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return s.dao.DeployApprovalApprove(ctx, id, approverID, approverName, comment)
}

// DeployApprovalReject 拒绝审批
func (s *Services) DeployApprovalReject(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return s.dao.DeployApprovalReject(ctx, id, approverID, approverName, comment)
}

// DeployApprovalCancel 取消审批
func (s *Services) DeployApprovalCancel(ctx context.Context, id uint64, applicantID uint64) error {
	return s.dao.DeployApprovalCancel(ctx, id, applicantID)
}

// ==================== 变更日志服务 ====================

// ResourceChangeLogCreate 创建变更日志
func (s *Services) ResourceChangeLogCreate(ctx context.Context, log *models.CicdResourceChangeLog) error {
	return s.dao.ResourceChangeLogCreate(ctx, log)
}

// ResourceChangeLogList 获取变更日志列表
func (s *Services) ResourceChangeLogList(ctx context.Context, pipelineID uint64, env string, page, pageSize int) ([]models.CicdResourceChangeLog, int64, error) {
	return s.dao.ResourceChangeLogList(ctx, pipelineID, env, page, pageSize)
}

// ==================== 工具函数 ====================

// parseCPUToMilli 解析CPU值为毫核
func parseCPUToMilli(cpu string) int64 {
	if cpu == "" {
		return 0
	}
	cpu = strings.TrimSpace(cpu)
	
	if strings.HasSuffix(cpu, "m") {
		val, _ := strconv.ParseInt(strings.TrimSuffix(cpu, "m"), 10, 64)
		return val
	}
	
	// 核心数转毫核
	val, _ := strconv.ParseFloat(cpu, 64)
	return int64(val * 1000)
}

// parseMemoryToBytes 解析内存值为字节
func parseMemoryToBytes(memory string) int64 {
	if memory == "" {
		return 0
	}
	memory = strings.TrimSpace(memory)
	
	units := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
	}
	
	for suffix, multiplier := range units {
		if strings.HasSuffix(memory, suffix) {
			val, _ := strconv.ParseInt(strings.TrimSuffix(memory, suffix), 10, 64)
			return val * multiplier
		}
	}
	
	// 纯数字视为字节
	val, _ := strconv.ParseInt(memory, 10, 64)
	return val
}
