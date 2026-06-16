package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ImageRegistryService 镜像仓库服务
type ImageRegistryService struct {
	model *models.ImageRegistryModel
}

func NewImageRegistryService() *ImageRegistryService {
	return &ImageRegistryService{
		model: models.NewImageRegistryModel(),
	}
}

// RegistryListRequest 列表请求参数
type RegistryListRequest struct {
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
}

// RegistryCreateRequest 创建请求参数
type RegistryCreateRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=100"`
	Type            string `json:"type" binding:"required,oneof=docker harbor gcr ecr acr quay"`
	URL             string `json:"url" binding:"required,url"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	AccessKeyID     string `json:"access_key_id"`     // 阿里云 ACR 使用
	AccessKeySecret string `json:"access_key_secret"` // 阿里云 ACR 使用
	Region          string `json:"region"`            // 区域（如 cn-hangzhou）
	Insecure        bool   `json:"insecure"`
	Description     string `json:"description"`
	IsDefault       bool   `json:"is_default"`
}

// RegistryUpdateRequest 更新请求参数
type RegistryUpdateRequest struct {
	ID              int64  `json:"id" binding:"required"`
	Name            string `json:"name" binding:"required,min=1,max=100"`
	Type            string `json:"type" binding:"required,oneof=docker harbor gcr ecr acr quay"`
	URL             string `json:"url" binding:"required,url"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region"`
	Insecure        bool   `json:"insecure"`
	Description     string `json:"description"`
	IsDefault       bool   `json:"is_default"`
}

// RegistryResponse 仓库响应
type RegistryResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	Username       string `json:"username"`
	HasPassword    bool   `json:"has_password"`
	AccessKeyID    string `json:"access_key_id"`
	HasAccessKey   bool   `json:"has_access_key"`
	Region         string `json:"region"`
	Insecure       bool   `json:"insecure"`
	Description    string `json:"description"`
	IsDefault      bool   `json:"is_default"`
	Status         string `json:"status"`
	LastCheckAt    int64  `json:"last_check_at"`
	LastError      string `json:"last_error"`
	CreatedAt      int64  `json:"created_at"`
	ModifiedAt     int64  `json:"modified_at"`
}

// RegistryStats 仓库统计
type RegistryStats struct {
	Total       int64 `json:"total"`
	Connected   int64 `json:"connected"`
	Disconnected int64 `json:"disconnected"`
	TypeCounts  map[string]int64 `json:"type_counts"`
}

// toResponse 转换为响应结构
func toResponse(r *models.ImageRegistry) *RegistryResponse {
	return &RegistryResponse{
		ID:           r.ID,
		Name:         r.Name,
		Type:         r.Type,
		URL:          r.URL,
		Username:     r.Username,
		HasPassword:  r.Password != "",
		AccessKeyID:  r.AccessKeyID,
		HasAccessKey: r.AccessKeySecret != "",
		Region:       r.Region,
		Insecure:     r.Insecure,
		Description:  r.Description,
		IsDefault:    r.IsDefault,
		Status:       r.Status,
		LastCheckAt:  r.LastCheckAt,
		LastError:    r.LastError,
		CreatedAt:    r.CreatedAt,
		ModifiedAt:   r.ModifiedAt,
	}
}

// List 获取镜像仓库列表
func (s *ImageRegistryService) List(req *RegistryListRequest) ([]RegistryResponse, int64, error) {
	registries, total, err := s.model.List(req.Keyword, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]RegistryResponse, 0, len(registries))
	for _, r := range registries {
		result = append(result, *toResponse(&r))
	}
	return result, total, nil
}

// ListAll 获取所有镜像仓库
func (s *ImageRegistryService) ListAll() ([]RegistryResponse, error) {
	registries, err := s.model.ListAll()
	if err != nil {
		return nil, err
	}

	result := make([]RegistryResponse, 0, len(registries))
	for _, r := range registries {
		result = append(result, *toResponse(&r))
	}
	return result, nil
}

// GetByID 根据ID获取
func (s *ImageRegistryService) GetByID(id int64) (*RegistryResponse, error) {
	registry, err := s.model.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toResponse(registry), nil
}

// Create 创建镜像仓库
func (s *ImageRegistryService) Create(req *RegistryCreateRequest, userID int64) (*RegistryResponse, error) {
	// 检查名称是否重复
	exists, err := s.model.ExistsByName(req.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("仓库名称 '%s' 已存在", req.Name)
	}

	registry := &models.ImageRegistry{
		Name:            req.Name,
		Type:            req.Type,
		URL:             strings.TrimSuffix(req.URL, "/"),
		Username:        req.Username,
		Password:        req.Password,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		Region:          req.Region,
		Insecure:        req.Insecure,
		Description:     req.Description,
		IsDefault:       req.IsDefault,
		Status:          "unknown",
		CreatedBy:       userID,
	}

	if err := s.model.Create(registry); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他仓库
	if req.IsDefault {
		_ = s.model.SetDefault(registry.ID)
	}

	// 异步检测连接状态
	go s.CheckConnection(registry.ID)

	return toResponse(registry), nil
}

// Update 更新镜像仓库
func (s *ImageRegistryService) Update(req *RegistryUpdateRequest) (*RegistryResponse, error) {
	// 检查是否存在
	existing, err := s.model.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("仓库不存在")
	}

	// 检查名称是否重复
	exists, err := s.model.ExistsByName(req.Name, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("仓库名称 '%s' 已存在", req.Name)
	}

	existing.Name = req.Name
	existing.Type = req.Type
	existing.URL = strings.TrimSuffix(req.URL, "/")
	existing.Username = req.Username
	if req.Password != "" {
		existing.Password = req.Password
	}
	existing.AccessKeyID = req.AccessKeyID
	if req.AccessKeySecret != "" {
		existing.AccessKeySecret = req.AccessKeySecret
	}
	existing.Region = req.Region
	existing.Insecure = req.Insecure
	existing.Description = req.Description
	existing.IsDefault = req.IsDefault

	if err := s.model.Update(existing); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他仓库
	if req.IsDefault {
		_ = s.model.SetDefault(existing.ID)
	}

	// 异步检测连接状态
	go s.CheckConnection(existing.ID)

	return toResponse(existing), nil
}

// Delete 删除镜像仓库
func (s *ImageRegistryService) Delete(id int64) error {
	_, err := s.model.GetByID(id)
	if err != nil {
		return fmt.Errorf("仓库不存在")
	}
	return s.model.Delete(id)
}

// CheckConnection 检测仓库连接状态
func (s *ImageRegistryService) CheckConnection(id int64) error {
	registry, err := s.model.GetByID(id)
	if err != nil {
		return err
	}

	status := "connected"
	lastError := ""
	checkTime := time.Now().Unix()

	// 阿里云 ACR 使用 OpenAPI 检测
	if registry.Type == "acr" {
		if registry.AccessKeyID == "" || registry.AccessKeySecret == "" {
			status = "disconnected"
			lastError = "未配置 AccessKey"
		} else {
			acrClient, err := NewACRClient(registry)
			if err != nil {
				status = "disconnected"
				lastError = err.Error()
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, err := acrClient.listNamespaces(ctx)
				cancel()
				if err != nil {
					status = "disconnected"
					lastError = err.Error()
				} else {
					status = "connected"
				}
			}
		}
		// 更新状态
		if err := s.model.UpdateStatus(id, status, lastError, checkTime); err != nil {
			global.Logger.Error("更新仓库状态失败", zap.Error(err))
		}
		return nil
	}

	// 其他类型使用 HTTP 检测
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: registry.Insecure,
			},
		},
	}

	// 根据仓库类型检测不同的端点
	checkURL := registry.URL
	switch registry.Type {
	case "docker":
		checkURL = strings.TrimSuffix(registry.URL, "/") + "/v2/"
	case "harbor":
		checkURL = strings.TrimSuffix(registry.URL, "/") + "/api/v2.0/ping"
	default:
		checkURL = strings.TrimSuffix(registry.URL, "/") + "/v2/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		status = "disconnected"
		lastError = err.Error()
	} else {
		// 添加认证
		if registry.Username != "" && registry.Password != "" {
			req.SetBasicAuth(registry.Username, registry.Password)
		}

		resp, err := client.Do(req)
		if err != nil {
			status = "disconnected"
			lastError = err.Error()
		} else {
			defer resp.Body.Close()
			// 2xx 或 401（需要认证但能连接）都算连接成功
			if resp.StatusCode >= 200 && resp.StatusCode < 300 || resp.StatusCode == 401 {
				status = "connected"
			} else {
				status = "disconnected"
				lastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		}
	}

	// 更新状态
	if err := s.model.UpdateStatus(id, status, lastError, checkTime); err != nil {
		global.Logger.Error("更新仓库状态失败", zap.Error(err))
	}

	return nil
}

// CheckAllConnections 检测所有仓库连接状态
func (s *ImageRegistryService) CheckAllConnections() error {
	registries, err := s.model.ListAll()
	if err != nil {
		return err
	}

	for _, r := range registries {
		go s.CheckConnection(r.ID)
	}
	return nil
}

// GetStats 获取仓库统计
func (s *ImageRegistryService) GetStats() (*RegistryStats, error) {
	registries, err := s.model.ListAll()
	if err != nil {
		return nil, err
	}

	stats := &RegistryStats{
		Total:      int64(len(registries)),
		TypeCounts: make(map[string]int64),
	}

	for _, r := range registries {
		if r.Status == "connected" {
			stats.Connected++
		} else {
			stats.Disconnected++
		}
		stats.TypeCounts[r.Type]++
	}

	return stats, nil
}

// SetDefault 设置默认仓库
func (s *ImageRegistryService) SetDefault(id int64) error {
	_, err := s.model.GetByID(id)
	if err != nil {
		return fmt.Errorf("仓库不存在")
	}
	return s.model.SetDefault(id)
}
