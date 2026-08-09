package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	dm "k8soperation/internal/domain/image"
	"k8soperation/pkg/logger"
)

// ImageRegistryService 镜像仓库服务
type ImageRegistryService struct {
	imgSvc *dm.ImageService
	logger *logger.Logger
}

// ImageRegistrySvc 返回镜像仓库服务（Controller 使用）
func (s *Services) ImageRegistrySvc() *ImageRegistryService {
	return &ImageRegistryService{imgSvc: s.imageSvc(), logger: s.logger}
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
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region"`
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
	Total        int64            `json:"total"`
	Connected    int64            `json:"connected"`
	Disconnected int64            `json:"disconnected"`
	TypeCounts   map[string]int64 `json:"type_counts"`
}

// toResponse 转换为响应结构
func toResponse(r *dm.ImageRegistry) *RegistryResponse {
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
	registries, total, err := s.imgSvc.RegistryList(req.Keyword, req.Type, req.Page, req.PageSize)
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
	registries, err := s.imgSvc.RegistryListAll()
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
	registry, err := s.imgSvc.RegistryGetByID(id)
	if err != nil {
		return nil, err
	}
	return toResponse(registry), nil
}

// Create 创建镜像仓库
func (s *ImageRegistryService) Create(req *RegistryCreateRequest, userID int64) (*RegistryResponse, error) {
	registry, err := s.imgSvc.RegistryCreateWithValidation(
		req.Name, req.Type, req.URL, req.Username, req.Password,
		req.AccessKeyID, req.AccessKeySecret, req.Region, req.Description,
		req.Insecure, req.IsDefault, userID,
	)
	if err != nil {
		return nil, err
	}

	// 异步检测连接状态
	go s.CheckConnection(registry.ID)

	return toResponse(registry), nil
}

// Update 更新镜像仓库
func (s *ImageRegistryService) Update(req *RegistryUpdateRequest) (*RegistryResponse, error) {
	registry, err := s.imgSvc.RegistryUpdateWithValidation(
		req.ID, req.Name, req.Type, req.URL, req.Username, req.Password,
		req.AccessKeyID, req.AccessKeySecret, req.Region, req.Description,
		req.Insecure, req.IsDefault,
	)
	if err != nil {
		return nil, err
	}

	// 异步检测连接状态
	go s.CheckConnection(registry.ID)

	return toResponse(registry), nil
}

// Delete 删除镜像仓库
func (s *ImageRegistryService) Delete(id int64) error {
	return s.imgSvc.RegistryDeleteWithValidation(id)
}

// CheckConnection 检测仓库连接状态
func (s *ImageRegistryService) CheckConnection(id int64) error {
	registry, err := s.imgSvc.RegistryGetByID(id)
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
		if err := s.imgSvc.RegistryUpdateStatus(id, status, lastError, checkTime); err != nil {
			s.logger.Error("更新仓库状态失败", zap.Error(err))
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
		if registry.Username != "" && registry.Password != "" {
			req.SetBasicAuth(registry.Username, registry.Password)
		}

		resp, err := client.Do(req)
		if err != nil {
			status = "disconnected"
			lastError = err.Error()
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 || resp.StatusCode == 401 {
				status = "connected"
			} else {
				status = "disconnected"
				lastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		}
	}

	if err := s.imgSvc.RegistryUpdateStatus(id, status, lastError, checkTime); err != nil {
		s.logger.Error("更新仓库状态失败", zap.Error(err))
	}

	return nil
}

// CheckAllConnections 检测所有仓库连接状态
func (s *ImageRegistryService) CheckAllConnections() error {
	registries, err := s.imgSvc.RegistryListAll()
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
	registries, err := s.imgSvc.RegistryListAll()
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
	return s.imgSvc.RegistrySetDefaultWithValidation(id)
}
