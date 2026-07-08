// Package argoworkflows 提供 Argo Workflows API 客户端实现
// 用于提交 Workflow、获取执行状态、日志等操作
package argoworkflows

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

// Client Argo Workflows API 客户端
type Client struct {
	BaseURL    string       // Argo Workflows Server 地址，如 https://argo-workflows.example.com:2746
	AuthToken  string       // Auth Token (Bearer)
	HTTPClient *http.Client // HTTP 客户端（共享连接池）
}

// NewClient 创建 Argo Workflows 客户端（使用全局共享连接池）
func NewClient(baseURL, authToken string) *Client {
	normalized := strings.TrimSuffix(baseURL, "/")
	return &Client{
		BaseURL:   normalized,
		AuthToken: authToken,
		HTTPClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: sharedTransport,
		},
	}
}

// GetOrCreateClient 获取或创建全局缓存的 Argo Workflows 客户端（单例）
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
		return nil, fmt.Errorf("argo workflows API error [%d]: %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

// ==================== Argo Workflows API 结构体 ====================

// WorkflowSubmitRequest Workflow 提交请求
type WorkflowSubmitRequest struct {
	Namespace     string   `json:"namespace"`
	ResourceKind  string   `json:"resourceKind"`  // "WorkflowTemplate" or "ClusterWorkflowTemplate"
	ResourceName  string   `json:"resourceName"`  // WorkflowTemplate 名称
	Parameters    []string `json:"parameters"`    // ["key1=val1", "key2=val2"]
	Labels        string   `json:"labels,omitempty"`       // "pipeline_id=123,run_id=456"
}

// SubmitOptions 提交选项
type SubmitOptions struct {
	Parameters []string `json:"parameters,omitempty"`
	Labels     string   `json:"labels,omitempty"`
}

// WorkflowInfo Workflow 信息
type WorkflowInfo struct {
	Metadata struct {
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		CreationTimestamp string `json:"creationTimestamp"`
		Labels            map[string]string `json:"labels"`
	} `json:"metadata"`
	Status struct {
		Phase      string                    `json:"phase"` // Pending / Running / Succeeded / Failed / Error
		StartedAt  string                    `json:"startedAt"`
		FinishedAt string                    `json:"finishedAt"`
		Message    string                    `json:"message"`
		Nodes      map[string]WorkflowNodeStatus `json:"nodes"`
	} `json:"status"`
}

// WorkflowNodeStatus 单个 Workflow 节点的状态
type WorkflowNodeStatus struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`        // Pod / Retry / StepGroup
	Phase       string `json:"phase"`
	Message     string `json:"message"`
	StartedAt   string `json:"startedAt"`
	FinishedAt  string `json:"finishedAt"`
	Duration    int    `json:"duration,omitempty"`
	TemplateName string `json:"templateName"`
}

// WorkflowListResponse Workflow 列表响应
type WorkflowListResponse struct {
	Items []WorkflowInfo `json:"items"`
}

// ==================== Argo Workflows API 方法 ====================

// SubmitWorkflow 提交 Workflow（基于 WorkflowTemplate）
// POST /api/v1/workflows/{namespace}/submit
func (c *Client) SubmitWorkflow(ctx context.Context, req *WorkflowSubmitRequest) (*WorkflowInfo, error) {
	submitPayload := map[string]interface{}{
		"namespace":    req.Namespace,
		"resourceKind": req.ResourceKind,
		"resourceName": req.ResourceName,
		"submitOptions": SubmitOptions{
			Parameters: req.Parameters,
			Labels:     req.Labels,
		},
	}

	data, err := c.doRequest(ctx, http.MethodPost, "/api/v1/workflows/"+req.Namespace+"/submit", submitPayload)
	if err != nil {
		return nil, err
	}

	var wf WorkflowInfo
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow: %w", err)
	}
	return &wf, nil
}

// GetWorkflow 获取 Workflow 详情
// GET /api/v1/workflows/{namespace}/{name}
func (c *Client) GetWorkflow(ctx context.Context, namespace, name string) (*WorkflowInfo, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/api/v1/workflows/"+namespace+"/"+name, nil)
	if err != nil {
		return nil, err
	}

	var wf WorkflowInfo
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow: %w", err)
	}
	return &wf, nil
}

// GetWorkflowStatus 获取 Workflow 执行状态摘要
func (c *Client) GetWorkflowStatus(ctx context.Context, namespace, name string) (phase string, message string, err error) {
	wf, err := c.GetWorkflow(ctx, namespace, name)
	if err != nil {
		return "", "", err
	}
	return wf.Status.Phase, wf.Status.Message, nil
}

// ListWorkflows 按标签查询 Workflow 列表
// GET /api/v1/workflows/{namespace}?listOptions.labelSelector={labels}
func (c *Client) ListWorkflows(ctx context.Context, namespace string, labelSelector string) ([]WorkflowInfo, error) {
	path := "/api/v1/workflows/" + namespace
	if labelSelector != "" {
		path += "?listOptions.labelSelector=" + labelSelector
	}

	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result WorkflowListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal workflow list: %w", err)
	}
	return result.Items, nil
}

// StopWorkflow 终止 Workflow
// PUT /api/v1/workflows/{namespace}/{name}/stop
func (c *Client) StopWorkflow(ctx context.Context, namespace, name string) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/api/v1/workflows/"+namespace+"/"+name+"/stop", nil)
	return err
}

// GetWorkflowLogs 获取 Workflow 日志（聚合所有节点）
// GET /api/v1/workflows/{namespace}/{name}/log?logOptions.container=main&logOptions.follow=false
func (c *Client) GetWorkflowLogs(ctx context.Context, namespace, name string) (string, error) {
	data, err := c.doRequest(ctx, http.MethodGet,
		"/api/v1/workflows/"+namespace+"/"+name+"/log?logOptions.follow=false", nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CheckConnection 检查 Argo Workflows 连接是否正常
func (c *Client) CheckConnection(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodGet, "/api/v1/workflows", nil)
	return err
}

// ==================== 参数构建工具 ====================

// BuildParameter 构建 key=value 格式的参数
func BuildParameter(key, value string) string {
	return key + "=" + value
}

// BuildParametersFromMap 从 map 构建参数列表
func BuildParametersFromMap(params map[string]string) []string {
	var result []string
	for k, v := range params {
		result = append(result, k+"="+v)
	}
	return result
}
