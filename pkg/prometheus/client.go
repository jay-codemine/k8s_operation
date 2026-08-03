// Package prometheus 提供 Prometheus HTTP API 查询客户端
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// 拨号超时。默认 Transport 用的是 30s，数据源地址不可达时每条查询都要干等满，
// 而看板一次请求会串行发很多条 PromQL，累计起来远超前端超时。
const dialTimeout = 2 * time.Second

// unreachableTTL 失败负缓存时长
const unreachableTTL = 15 * time.Second

// 不可达数据源的负缓存，按 baseURL 记录最近一次网络级失败。
//
// 为什么必须有：一次看板请求会串行发 8~10 条查询，地址不可达时每条都要等满拨号超时，
// 接口表现为一直挂住直到前端放弃。有了它，只有第一条查询付拨号超时的代价，
// 其余立即失败，整个请求被压到一次拨号超时以内。
// 思路与 services.ClusterClientFactory 的失败负缓存一致。
//
// 放在包级而非 Client 上：调用方（MonitoringService.resolveClient）每次查询都新建 Client，
// 实例上的状态活不过一次调用。
var (
	unreachableMu sync.Mutex
	unreachableAt = map[string]time.Time{}
)

func unreachable(baseURL string) bool {
	unreachableMu.Lock()
	defer unreachableMu.Unlock()
	at, ok := unreachableAt[baseURL]
	return ok && time.Since(at) < unreachableTTL
}

func markUnreachable(baseURL string) {
	unreachableMu.Lock()
	unreachableAt[baseURL] = time.Now()
	unreachableMu.Unlock()
}

func clearUnreachable(baseURL string) {
	unreachableMu.Lock()
	delete(unreachableAt, baseURL)
	unreachableMu.Unlock()
}

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
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: timeout,
				MaxIdleConnsPerHost:   2,
				IdleConnTimeout:       30 * time.Second,
			},
		},
	}
}

// do 发请求，并维护数据源可达性的负缓存
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if unreachable(c.baseURL) {
		return nil, fmt.Errorf("prometheus %s 近期不可达，已跳过本次查询", c.baseURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// 只把网络级失败记为数据源不可达；上层 ctx 被取消（前端断开）不算
		if ctx.Err() == nil {
			markUnreachable(c.baseURL)
		}
		return nil, err
	}
	clearUnreachable(c.baseURL)
	return resp, nil
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
	resp, err := c.do(ctx, req)
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

	resp, err := c.do(ctx, req)
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
