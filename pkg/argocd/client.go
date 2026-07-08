// Package argocd 提供 ArgoCD API 客户端实现
// 用于管理 ArgoCD Application、获取同步状态等操作
package argocd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ==================== 高性能连接池 Transport ====================

// sharedTransport 全局共享的 HTTP Transport（连接池）
var sharedTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          200,
	MaxIdleConnsPerHost:   50,
	MaxConnsPerHost:       100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 30 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// ==================== Client 全局缓存 ====================

var (
	clientCache   = make(map[string]*Client) // key: baseURL
	clientCacheMu sync.RWMutex
)

// Client ArgoCD API 客户端
type Client struct {
	BaseURL    string       // ArgoCD API Server 地址，如 https://argocd.example.com
	AuthToken  string       // ArgoCD Auth Token (Bearer)
	HTTPClient *http.Client // HTTP 客户端（共享连接池）
}

// NewClient 创建 ArgoCD 客户端（使用全局共享连接池）
func NewClient(baseURL, authToken string) *Client {
	normalized := strings.TrimSuffix(baseURL, "/")
	return &Client{
		BaseURL:   normalized,
		AuthToken: authToken,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: sharedTransport,
		},
	}
}

// GetOrCreateClient 获取或创建全局缓存的 ArgoCD 客户端（单例）
func GetOrCreateClient(baseURL, authToken string) *Client {
	normalized := strings.TrimSuffix(baseURL, "/")

	clientCacheMu.RLock()
	if c, ok := clientCache[normalized]; ok {
		clientCacheMu.RUnlock()
		return c
	}
	clientCacheMu.RUnlock()

	clientCacheMu.Lock()
	defer clientCacheMu.Unlock()

	if c, ok := clientCache[normalized]; ok {
		return c
	}

	c := NewClient(normalized, authToken)
	clientCache[normalized] = c
	return c
}

// ==================== HTTP 请求封装 ====================

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("argocd API error [%d]: %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

// ==================== ArgoCD API 响应结构体 ====================

// ApplicationInfo ArgoCD Application 信息
type ApplicationInfo struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Source struct {
			RepoURL        string `json:"repoURL"`
			Path           string `json:"path"`
			TargetRevision string `json:"targetRevision"`
		} `json:"source"`
		Destination struct {
			Server    string `json:"server"`
			Namespace string `json:"namespace"`
		} `json:"destination"`
		SyncPolicy *struct {
			Automated *struct {
				Prune    bool `json:"prune"`
				SelfHeal bool `json:"selfHeal"`
			} `json:"automated"`
		} `json:"syncPolicy"`
	} `json:"spec"`
	Status struct {
		Sync struct {
			Status   string `json:"status"`   // Synced / OutOfSync / Unknown
			Revision string `json:"revision"` // Git commit SHA
		} `json:"sync"`
		Health struct {
			Status string `json:"status"` // Healthy / Progressing / Degraded / Missing
		} `json:"health"`
		OperationState *struct {
			Phase      string `json:"phase"` // Running / Succeeded / Failed / Error
			Message    string `json:"message"`
			FinishedAt string `json:"finishedAt"`
		} `json:"operationState"`
		Resources []struct {
			Kind      string `json:"kind"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Status    string `json:"status"`
			Health    struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"resources"`
	} `json:"status"`
}

// SyncStatus 同步状态摘要
type SyncStatus struct {
	SyncStatus   string `json:"sync_status"`
	SyncRevision string `json:"sync_revision"`
	HealthStatus string `json:"health_status"`
	Phase        string `json:"phase,omitempty"`
}

// SyncRequest 同步请求参数
type SyncRequest struct {
	Revision string `json:"revision,omitempty"` // 目标版本（Git SHA/branch）
	Prune    bool   `json:"prune,omitempty"`    // 是否删除不受管理的资源
	DryRun   bool   `json:"dryRun,omitempty"`   // Dry run 模式
}

// ==================== ArgoCD API 方法 ====================

// GetApplication 获取 ArgoCD Application 详情
// GET /api/v1/applications/{name}
func (c *Client) GetApplication(ctx context.Context, appName string) (*ApplicationInfo, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/api/v1/applications/"+appName, nil)
	if err != nil {
		return nil, err
	}

	var app ApplicationInfo
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("unmarshal application: %w", err)
	}
	return &app, nil
}

// GetApplicationStatus 获取 ArgoCD Application 同步状态
func (c *Client) GetApplicationStatus(ctx context.Context, appName string) (*SyncStatus, error) {
	app, err := c.GetApplication(ctx, appName)
	if err != nil {
		return nil, err
	}

	return &SyncStatus{
		SyncStatus:   app.Status.Sync.Status,
		SyncRevision: app.Status.Sync.Revision,
		HealthStatus: app.Status.Health.Status,
		Phase: func() string {
			if app.Status.OperationState != nil {
				return app.Status.OperationState.Phase
			}
			return ""
		}(),
	}, nil
}

// SyncApplication 触发 ArgoCD Application 同步
// POST /api/v1/applications/{name}/sync
func (c *Client) SyncApplication(ctx context.Context, appName string, revision string, prune bool) error {
	req := SyncRequest{
		Revision: revision,
		Prune:    prune,
	}
	_, err := c.doRequest(ctx, http.MethodPost, "/api/v1/applications/"+appName+"/sync", req)
	return err
}

// ListApplications 获取 ArgoCD Application 列表
// GET /api/v1/applications?project={project}
func (c *Client) ListApplications(ctx context.Context, project string) ([]*ApplicationInfo, error) {
	path := "/api/v1/applications"
	if project != "" {
		path += "?project=" + project
	}

	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []*ApplicationInfo `json:"items"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal applications: %w", err)
	}
	return result.Items, nil
}

// DeleteApplication 删除 ArgoCD Application
// DELETE /api/v1/applications/{name}
func (c *Client) DeleteApplication(ctx context.Context, appName string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, "/api/v1/applications/"+appName, nil)
	return err
}

// CheckConnection 检查 ArgoCD 连接是否正常
func (c *Client) CheckConnection(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodGet, "/api/v1/applications", nil)
	return err
}
