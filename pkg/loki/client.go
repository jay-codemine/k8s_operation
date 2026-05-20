// Package loki 提供 Loki HTTP API 查询客户端
package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client Loki HTTP API 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建 Loki 客户端
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ===== 数据结构 =====

// QueryResponse Loki 查询响应
type QueryResponse struct {
	Status string     `json:"status"`
	Data   ResultData `json:"data"`
}

// ResultData 查询结果数据
type ResultData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
	Stats      json.RawMessage `json:"stats,omitempty"`
}

// StreamResult 日志流结果
type StreamResult struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [[timestamp_ns, log_line], ...]
}

// LabelResponse 标签列表响应
type LabelResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// SeriesEntry series 响应
type SeriesEntry struct {
	Labels map[string]string
}

// StatsResponse 统计信息
type StatsResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// ===== 查询方法 =====

// QueryRange 范围查询 (LogQL)
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, limit int, direction string) (*QueryResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	params.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if direction != "" {
		params.Set("direction", direction) // "forward" or "backward"
	}
	return c.doQuery(ctx, "/loki/api/v1/query_range", params)
}

// QueryInstant 即时查询 (LogQL)
func (c *Client) QueryInstant(ctx context.Context, query string, ts time.Time, limit int) (*QueryResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", fmt.Sprintf("%d", ts.UnixNano()))
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return c.doQuery(ctx, "/loki/api/v1/query", params)
}

// GetLabels 获取所有标签名
func (c *Client) GetLabels(ctx context.Context, start, end time.Time) ([]string, error) {
	params := url.Values{}
	if !start.IsZero() {
		params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	}
	if !end.IsZero() {
		params.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	}

	reqURL := fmt.Sprintf("%s/loki/api/v1/labels?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki labels query failed: %w", err)
	}
	defer resp.Body.Close()

	var result LabelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetLabelValues 获取指定标签的所有值
func (c *Client) GetLabelValues(ctx context.Context, label string, start, end time.Time) ([]string, error) {
	params := url.Values{}
	if !start.IsZero() {
		params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	}
	if !end.IsZero() {
		params.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	}

	reqURL := fmt.Sprintf("%s/loki/api/v1/label/%s/values?%s", c.baseURL, url.PathEscape(label), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki label values query failed: %w", err)
	}
	defer resp.Body.Close()

	var result LabelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetSeries 获取日志流 series
func (c *Client) GetSeries(ctx context.Context, matchers []string, start, end time.Time) ([]map[string]string, error) {
	params := url.Values{}
	for _, m := range matchers {
		params.Add("match[]", m)
	}
	if !start.IsZero() {
		params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	}
	if !end.IsZero() {
		params.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	}

	reqURL := fmt.Sprintf("%s/loki/api/v1/series?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki series query failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status string              `json:"status"`
		Data   []map[string]string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// Healthy 检查 Loki 是否健康
func (c *Client) Healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/ready", nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ParseStreamResult 解析 streams 类型结果
func ParseStreamResult(raw json.RawMessage) ([]StreamResult, error) {
	var results []StreamResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// doQuery 执行查询请求
func (c *Client) doQuery(ctx context.Context, path string, params url.Values) (*QueryResponse, error) {
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("loki returned %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("loki query status: %s", result.Status)
	}

	return &result, nil
}
