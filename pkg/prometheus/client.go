// Package prometheus 提供 Prometheus HTTP API 查询客户端
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client Prometheus HTTP API 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建 Prometheus 客户端
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// QueryResult Prometheus 查询返回的数据结构
type QueryResult struct {
	Status string     `json:"status"`
	Data   ResultData `json:"data"`
}

type ResultData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

// VectorResult 即时查询结果（vector 类型）
type VectorResult struct {
	Metric map[string]string `json:"metric"`
	Value  [2]interface{}    `json:"value"` // [timestamp, value_string]
}

// MatrixResult 范围查询结果（matrix 类型）
type MatrixResult struct {
	Metric map[string]string `json:"metric"`
	Values [][2]interface{}  `json:"values"` // [[timestamp, value_string], ...]
}

// QueryInstant 即时查询 (PromQL)
func (c *Client) QueryInstant(ctx context.Context, query string) (*QueryResult, error) {
	params := url.Values{}
	params.Set("query", query)

	return c.doQuery(ctx, "/api/v1/query", params)
}

// QueryRange 范围查询 (PromQL)
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", fmt.Sprintf("%ds", int(step.Seconds())))

	return c.doQuery(ctx, "/api/v1/query_range", params)
}

// ParseVectorResult 解析 vector 类型结果
func ParseVectorResult(raw json.RawMessage) ([]VectorResult, error) {
	var results []VectorResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ParseMatrixResult 解析 matrix 类型结果
func ParseMatrixResult(raw json.RawMessage) ([]MatrixResult, error) {
	var results []MatrixResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// Healthy 检查 Prometheus 是否健康
func (c *Client) Healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/-/healthy", nil)
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

// doQuery 执行查询请求
func (c *Client) doQuery(ctx context.Context, path string, params url.Values) (*QueryResult, error) {
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query status: %s", result.Status)
	}

	return &result, nil
}
