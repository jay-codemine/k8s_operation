package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ImageService 镜像服务
type ImageService struct {
	registryModel *models.ImageRegistryModel
	policyModel   *models.ImageCleanupPolicyModel
	logModel      *models.ImageCleanupLogModel
}

func NewImageService() *ImageService {
	return &ImageService{
		registryModel: models.NewImageRegistryModel(),
		policyModel:   models.NewImageCleanupPolicyModel(),
		logModel:      models.NewImageCleanupLogModel(),
	}
}

// ========================================
// 镜像浏览相关
// ========================================

// ListRepositories 列出仓库中的镜像项目
func (s *ImageService) ListRepositories(registryID int64) ([]Repository, error) {
	registry, err := s.registryModel.GetByID(registryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	client, err := NewRegistryClient(registry)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.ListRepositories(ctx)
}

// ListTags 列出镜像的所有标签
func (s *ImageService) ListTags(registryID int64, repository string) ([]ImageTag, error) {
	registry, err := s.registryModel.GetByID(registryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	client, err := NewRegistryClient(registry)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.ListTags(ctx, repository)
}

// GetImageDetail 获取镜像详情
func (s *ImageService) GetImageDetail(registryID int64, repository, tag string) (*ImageManifest, error) {
	registry, err := s.registryModel.GetByID(registryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	client, err := NewRegistryClient(registry)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.GetManifest(ctx, repository, tag)
}

// DeleteTag 删除镜像标签
func (s *ImageService) DeleteTag(registryID int64, repository, tag string) error {
	registry, err := s.registryModel.GetByID(registryID)
	if err != nil {
		return fmt.Errorf("仓库不存在")
	}

	client, err := NewRegistryClient(registry)
	if err != nil {
		return fmt.Errorf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.DeleteTag(ctx, repository, tag)
}

// ========================================
// 清理策略相关
// ========================================

// CleanupPolicyRequest 清理策略请求
type CleanupPolicyRequest struct {
	ID                int64  `json:"id"`
	RegistryID        int64  `json:"registry_id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Enabled           bool   `json:"enabled"`
	RepositoryPattern string `json:"repository_pattern"`
	TagPattern        string `json:"tag_pattern"`
	KeepLastCount     int    `json:"keep_last_count"`
	KeepDays          int    `json:"keep_days"`
	CronExpression    string `json:"cron_expression"`
	Description       string `json:"description"`
}

// CleanupPolicyResponse 清理策略响应
type CleanupPolicyResponse struct {
	ID                int64  `json:"id"`
	RegistryID        int64  `json:"registry_id"`
	RegistryName      string `json:"registry_name"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	RepositoryPattern string `json:"repository_pattern"`
	TagPattern        string `json:"tag_pattern"`
	KeepLastCount     int    `json:"keep_last_count"`
	KeepDays          int    `json:"keep_days"`
	CronExpression    string `json:"cron_expression"`
	LastRunAt         int64  `json:"last_run_at"`
	LastRunResult     string `json:"last_run_result"`
	DeletedCount      int64  `json:"deleted_count"`
	Description       string `json:"description"`
	CreatedAt         int64  `json:"created_at"`
}

// ListCleanupPolicies 列出清理策略
func (s *ImageService) ListCleanupPolicies(registryID int64, keyword string, page, pageSize int) ([]CleanupPolicyResponse, int64, error) {
	policies, total, err := s.policyModel.List(registryID, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 获取所有仓库信息用于关联名称
	registryMap := make(map[int64]string)
	for _, p := range policies {
		if _, ok := registryMap[p.RegistryID]; !ok {
			if registry, err := s.registryModel.GetByID(p.RegistryID); err == nil {
				registryMap[p.RegistryID] = registry.Name
			}
		}
	}

	result := make([]CleanupPolicyResponse, 0, len(policies))
	for _, p := range policies {
		result = append(result, CleanupPolicyResponse{
			ID:                p.ID,
			RegistryID:        p.RegistryID,
			RegistryName:      registryMap[p.RegistryID],
			Name:              p.Name,
			Enabled:           p.Enabled,
			RepositoryPattern: p.RepositoryPattern,
			TagPattern:        p.TagPattern,
			KeepLastCount:     p.KeepLastCount,
			KeepDays:          p.KeepDays,
			CronExpression:    p.CronExpression,
			LastRunAt:         p.LastRunAt,
			LastRunResult:     p.LastRunResult,
			DeletedCount:      p.DeletedCount,
			Description:       p.Description,
			CreatedAt:         p.CreatedAt,
		})
	}

	return result, total, nil
}

// CreateCleanupPolicy 创建清理策略
func (s *ImageService) CreateCleanupPolicy(req *CleanupPolicyRequest, userID int64) (*CleanupPolicyResponse, error) {
	// 验证仓库是否存在
	registry, err := s.registryModel.GetByID(req.RegistryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	policy := &models.ImageCleanupPolicy{
		RegistryID:        req.RegistryID,
		Name:              req.Name,
		Enabled:           req.Enabled,
		RepositoryPattern: req.RepositoryPattern,
		TagPattern:        req.TagPattern,
		KeepLastCount:     req.KeepLastCount,
		KeepDays:          req.KeepDays,
		CronExpression:    req.CronExpression,
		Description:       req.Description,
		CreatedBy:         userID,
	}

	// 设置默认值
	if policy.RepositoryPattern == "" {
		policy.RepositoryPattern = "*"
	}
	if policy.TagPattern == "" {
		policy.TagPattern = "*"
	}
	if policy.KeepLastCount <= 0 {
		policy.KeepLastCount = 5
	}
	if policy.KeepDays <= 0 {
		policy.KeepDays = 30
	}
	if policy.CronExpression == "" {
		policy.CronExpression = "0 2 * * *"
	}

	if err := s.policyModel.Create(policy); err != nil {
		return nil, err
	}

	return &CleanupPolicyResponse{
		ID:                policy.ID,
		RegistryID:        policy.RegistryID,
		RegistryName:      registry.Name,
		Name:              policy.Name,
		Enabled:           policy.Enabled,
		RepositoryPattern: policy.RepositoryPattern,
		TagPattern:        policy.TagPattern,
		KeepLastCount:     policy.KeepLastCount,
		KeepDays:          policy.KeepDays,
		CronExpression:    policy.CronExpression,
		Description:       policy.Description,
		CreatedAt:         policy.CreatedAt,
	}, nil
}

// UpdateCleanupPolicy 更新清理策略
func (s *ImageService) UpdateCleanupPolicy(req *CleanupPolicyRequest) (*CleanupPolicyResponse, error) {
	existing, err := s.policyModel.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("策略不存在")
	}

	existing.Name = req.Name
	existing.Enabled = req.Enabled
	existing.RepositoryPattern = req.RepositoryPattern
	existing.TagPattern = req.TagPattern
	existing.KeepLastCount = req.KeepLastCount
	existing.KeepDays = req.KeepDays
	existing.CronExpression = req.CronExpression
	existing.Description = req.Description

	if err := s.policyModel.Update(existing); err != nil {
		return nil, err
	}

	registryName := ""
	if registry, err := s.registryModel.GetByID(existing.RegistryID); err == nil {
		registryName = registry.Name
	}

	return &CleanupPolicyResponse{
		ID:                existing.ID,
		RegistryID:        existing.RegistryID,
		RegistryName:      registryName,
		Name:              existing.Name,
		Enabled:           existing.Enabled,
		RepositoryPattern: existing.RepositoryPattern,
		TagPattern:        existing.TagPattern,
		KeepLastCount:     existing.KeepLastCount,
		KeepDays:          existing.KeepDays,
		CronExpression:    existing.CronExpression,
		LastRunAt:         existing.LastRunAt,
		LastRunResult:     existing.LastRunResult,
		DeletedCount:      existing.DeletedCount,
		Description:       existing.Description,
		CreatedAt:         existing.CreatedAt,
	}, nil
}

// DeleteCleanupPolicy 删除清理策略
func (s *ImageService) DeleteCleanupPolicy(id int64) error {
	_, err := s.policyModel.GetByID(id)
	if err != nil {
		return fmt.Errorf("策略不存在")
	}
	return s.policyModel.Delete(id)
}

// ToggleCleanupPolicy 启用/禁用清理策略
func (s *ImageService) ToggleCleanupPolicy(id int64, enabled bool) error {
	policy, err := s.policyModel.GetByID(id)
	if err != nil {
		return fmt.Errorf("策略不存在")
	}
	policy.Enabled = enabled
	return s.policyModel.Update(policy)
}

// RunCleanupPolicy 手动执行清理策略
func (s *ImageService) RunCleanupPolicy(id int64) (*models.ImageCleanupLog, error) {
	policy, err := s.policyModel.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("策略不存在")
	}

	registry, err := s.registryModel.GetByID(policy.RegistryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	// 创建执行日志
	log := &models.ImageCleanupLog{
		PolicyID:   policy.ID,
		RegistryID: registry.ID,
		StartTime:  time.Now().Unix(),
		Status:     "running",
	}
	if err := s.logModel.Create(log); err != nil {
		return nil, err
	}

	// 异步执行清理
	go s.executeCleanup(policy, registry, log)

	return log, nil
}

// executeCleanup 执行清理任务
func (s *ImageService) executeCleanup(policy *models.ImageCleanupPolicy, registry *models.ImageRegistry, log *models.ImageCleanupLog) {
	defer func() {
		if r := recover(); r != nil {
			global.Logger.Error("清理任务 panic", zap.Any("error", r))
			log.Status = "failed"
			log.ErrorMessage = fmt.Sprintf("panic: %v", r)
			log.EndTime = time.Now().Unix()
			s.logModel.Update(log)
		}
	}()

	client, err := NewRegistryClient(registry)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = fmt.Sprintf("创建客户端失败: %v", err)
		log.EndTime = time.Now().Unix()
		s.logModel.Update(log)
		return
	}

	ctx := context.Background()

	// 获取仓库列表
	repos, err := client.ListRepositories(ctx)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = fmt.Sprintf("获取仓库列表失败: %v", err)
		log.EndTime = time.Now().Unix()
		s.logModel.Update(log)
		return
	}

	deletedCount := 0
	var freedSize int64
	scannedCount := 0

	cutoffTime := time.Now().AddDate(0, 0, -policy.KeepDays).Unix()

	for _, repo := range repos {
		// 匹配仓库模式
		if !matchPattern(policy.RepositoryPattern, repo.FullName) {
			continue
		}

		tags, err := client.ListTags(ctx, repo.FullName)
		if err != nil {
			global.Logger.Warn("获取标签失败", zap.String("repo", repo.FullName), zap.Error(err))
			continue
		}

		scannedCount += len(tags)

		// 过滤要保留的标签
		var tagsToDelete []ImageTag
		keepCount := 0

		for _, tag := range tags {
			// 匹配标签模式
			if !matchPattern(policy.TagPattern, tag.Name) {
				continue
			}

			// 保留最近 N 个
			if keepCount < policy.KeepLastCount {
				keepCount++
				continue
			}

			// 保留 N 天内的
			if tag.PushedAt > 0 && tag.PushedAt > cutoffTime {
				continue
			}

			tagsToDelete = append(tagsToDelete, tag)
		}

		// 删除过期标签
		for _, tag := range tagsToDelete {
			if err := client.DeleteTag(ctx, repo.FullName, tag.Name); err != nil {
				global.Logger.Warn("删除标签失败",
					zap.String("repo", repo.FullName),
					zap.String("tag", tag.Name),
					zap.Error(err))
				continue
			}
			deletedCount++
			freedSize += tag.Size
			global.Logger.Info("已删除镜像",
				zap.String("repo", repo.FullName),
				zap.String("tag", tag.Name))
		}
	}

	// 更新日志
	log.Status = "success"
	log.ScannedCount = scannedCount
	log.DeletedCount = deletedCount
	log.FreedSize = freedSize
	log.EndTime = time.Now().Unix()
	s.logModel.Update(log)

	// 更新策略执行结果
	result := fmt.Sprintf("扫描 %d 个标签，删除 %d 个", scannedCount, deletedCount)
	s.policyModel.UpdateRunResult(policy.ID, result, int64(deletedCount))
}

// GetCleanupLogs 获取清理日志
func (s *ImageService) GetCleanupLogs(policyID int64, limit int) ([]models.ImageCleanupLog, error) {
	if policyID > 0 {
		return s.logModel.ListByPolicy(policyID, limit)
	}
	return s.logModel.ListRecent(limit)
}

// matchPattern 简单的通配符匹配
func matchPattern(pattern, name string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}
