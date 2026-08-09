package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	dm "k8soperation/internal/domain/image"
	"k8soperation/internal/infra/persistence"
	"k8soperation/pkg/logger"
)

// ImageService 镜像浏览服务
type ImageService struct {
	imgSvc *dm.ImageService
	logger *logger.Logger
}

func (s *Services) imageSvc() *dm.ImageService {
	return dm.NewImageService(s.db, persistence.NewImageRepository(s.db))
}

// ImageSvc 返回镜像浏览服务（Controller 使用）
func (s *Services) ImageSvc() *ImageService {
	return &ImageService{imgSvc: s.imageSvc(), logger: s.logger}
}

// ========================================
// 镜像浏览相关
// ========================================

// ListRepositories 列出仓库中的镜像项目
func (s *ImageService) ListRepositories(registryID int64) ([]Repository, error) {
	registry, err := s.imgSvc.RegistryGetByID(registryID)
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
	registry, err := s.imgSvc.RegistryGetByID(registryID)
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
	registry, err := s.imgSvc.RegistryGetByID(registryID)
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
	registry, err := s.imgSvc.RegistryGetByID(registryID)
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
	policies, total, err := s.imgSvc.PolicyList(registryID, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	registryMap := make(map[int64]string)
	for _, p := range policies {
		if _, ok := registryMap[p.RegistryID]; !ok {
			if registry, err := s.imgSvc.RegistryGetByID(p.RegistryID); err == nil {
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
	registry, err := s.imgSvc.RegistryGetByID(req.RegistryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	policy := &dm.ImageCleanupPolicy{
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

	if err := s.imgSvc.PolicyCreate(policy); err != nil {
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
	existing, err := s.imgSvc.PolicyGetByID(req.ID)
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

	if err := s.imgSvc.PolicyUpdate(existing); err != nil {
		return nil, err
	}

	registryName := ""
	if registry, err := s.imgSvc.RegistryGetByID(existing.RegistryID); err == nil {
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
	return s.imgSvc.PolicyDeleteWithValidation(id)
}

// ToggleCleanupPolicy 启用/禁用清理策略
func (s *ImageService) ToggleCleanupPolicy(id int64, enabled bool) error {
	return s.imgSvc.PolicyToggle(id, enabled)
}

// RunCleanupPolicy 手动执行清理策略
func (s *ImageService) RunCleanupPolicy(id int64) (*dm.ImageCleanupLog, error) {
	policy, err := s.imgSvc.PolicyGetByID(id)
	if err != nil {
		return nil, fmt.Errorf("策略不存在")
	}

	registry, err := s.imgSvc.RegistryGetByID(policy.RegistryID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	log := &dm.ImageCleanupLog{
		PolicyID:   policy.ID,
		RegistryID: registry.ID,
		StartTime:  time.Now().Unix(),
		Status:     "running",
	}
	if err := s.imgSvc.LogCreate(log); err != nil {
		return nil, err
	}

	go s.executeCleanup(policy, registry, log)

	return log, nil
}

// executeCleanup 执行清理任务
func (s *ImageService) executeCleanup(policy *dm.ImageCleanupPolicy, registry *dm.ImageRegistry, log *dm.ImageCleanupLog) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("清理任务 panic", zap.Any("error", r))
			log.Status = "failed"
			log.ErrorMessage = fmt.Sprintf("panic: %v", r)
			log.EndTime = time.Now().Unix()
			s.imgSvc.LogUpdate(log)
		}
	}()

	client, err := NewRegistryClient(registry)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = fmt.Sprintf("创建客户端失败: %v", err)
		log.EndTime = time.Now().Unix()
		s.imgSvc.LogUpdate(log)
		return
	}

	ctx := context.Background()

	repos, err := client.ListRepositories(ctx)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = fmt.Sprintf("获取仓库列表失败: %v", err)
		log.EndTime = time.Now().Unix()
		s.imgSvc.LogUpdate(log)
		return
	}

	deletedCount := 0
	var freedSize int64
	scannedCount := 0

	cutoffTime := time.Now().AddDate(0, 0, -policy.KeepDays).Unix()

	for _, repo := range repos {
		if !matchPattern(policy.RepositoryPattern, repo.FullName) {
			continue
		}

		tags, err := client.ListTags(ctx, repo.FullName)
		if err != nil {
			s.logger.Warn("获取标签失败", zap.String("repo", repo.FullName), zap.Error(err))
			continue
		}

		scannedCount += len(tags)

		var tagsToDelete []ImageTag
		keepCount := 0

		for _, tag := range tags {
			if !matchPattern(policy.TagPattern, tag.Name) {
				continue
			}

			if keepCount < policy.KeepLastCount {
				keepCount++
				continue
			}

			if tag.PushedAt > 0 && tag.PushedAt > cutoffTime {
				continue
			}

			tagsToDelete = append(tagsToDelete, tag)
		}

		for _, tag := range tagsToDelete {
			if err := client.DeleteTag(ctx, repo.FullName, tag.Name); err != nil {
				s.logger.Warn("删除标签失败",
					zap.String("repo", repo.FullName),
					zap.String("tag", tag.Name),
					zap.Error(err))
				continue
			}
			deletedCount++
			freedSize += tag.Size
			s.logger.Info("已删除镜像",
				zap.String("repo", repo.FullName),
				zap.String("tag", tag.Name))
		}
	}

	log.Status = "success"
	log.ScannedCount = scannedCount
	log.DeletedCount = deletedCount
	log.FreedSize = freedSize
	log.EndTime = time.Now().Unix()
	s.imgSvc.LogUpdate(log)

	result := fmt.Sprintf("扫描 %d 个标签，删除 %d 个", scannedCount, deletedCount)
	s.imgSvc.PolicyUpdateRunResult(policy.ID, result, int64(deletedCount))
}

// GetCleanupLogs 获取清理日志
func (s *ImageService) GetCleanupLogs(policyID int64, limit int) ([]dm.ImageCleanupLog, error) {
	if policyID > 0 {
		return s.imgSvc.LogListByPolicy(policyID, limit)
	}
	return s.imgSvc.LogListRecent(limit)
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
