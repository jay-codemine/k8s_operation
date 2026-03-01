// Package jenkins 提供 Jenkins API 客户端实现
// 用于触发构建、获取构建状态、获取控制台日志等操作
package jenkins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Client Jenkins API 客户端
type Client struct {
	BaseURL    string       // Jenkins 服务器地址，如 http://jenkins.example.com
	Username   string       // Jenkins 用户名
	APIToken   string       // Jenkins API Token（在用户设置中生成）
	HTTPClient *http.Client // HTTP 客户端
}

// NewClient 创建 Jenkins 客户端
func NewClient(baseURL, username, apiToken string) *Client {
	return &Client{
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		Username: username,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BuildInfo Jenkins 构建信息
type BuildInfo struct {
	Number          int    `json:"number"`
	URL             string `json:"url"`
	Building        bool   `json:"building"`
	Result          string `json:"result"`          // SUCCESS, FAILURE, ABORTED, null(if building)
	Duration        int64  `json:"duration"`        // 毫秒
	EstimatedDuration int64 `json:"estimatedDuration"`
	Timestamp       int64  `json:"timestamp"`       // 开始时间戳（毫秒）
	DisplayName     string `json:"displayName"`
	FullDisplayName string `json:"fullDisplayName"`
}

// QueueItem Jenkins 队列项信息
type QueueItem struct {
	ID         int64  `json:"id"`
	Blocked    bool   `json:"blocked"`
	Buildable  bool   `json:"buildable"`
	Stuck      bool   `json:"stuck"`
	Why        string `json:"why"`
	Executable struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	} `json:"executable"`
}

// JobInfo Jenkins Job 信息
type JobInfo struct {
	Name              string         `json:"name"`
	URL               string         `json:"url"`
	Color             string         `json:"color"`
	Buildable         bool           `json:"buildable"`
	NextBuildNumber   int            `json:"nextBuildNumber"`
	LastBuild         *BuildInfo     `json:"lastBuild"`
	LastSuccessfulBuild *BuildInfo   `json:"lastSuccessfulBuild"`
	LastFailedBuild   *BuildInfo     `json:"lastFailedBuild"`
	Property          []JobProperty  `json:"property"` // Job 属性（用于判断是否参数化）
	Class             string         `json:"_class"` // Job 类型（用于区分 Pipeline Job 等）
}

// JobProperty Job 属性
type JobProperty struct {
	Class string `json:"_class"`
}

// TriggerBuildResult 触发构建的返回结果
type TriggerBuildResult struct {
	QueueID     int64  // 队列 ID
	QueueURL    string // 队列 URL
	BuildNumber int    // 构建号（可能需要等待队列消费后才有）
	BuildURL    string // 构建 URL
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Basic Auth
	if c.Username != "" && c.APIToken != "" {
		req.SetBasicAuth(c.Username, c.APIToken)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.HTTPClient.Do(req)
}

// TriggerBuild 触发构建
// jobName: Job 名称
// params: 构建参数（可选）
// 返回队列 ID，后续可通过 WaitForBuild 获取构建号
func (c *Client) TriggerBuild(ctx context.Context, jobName string, params map[string]string) (*TriggerBuildResult, error) {
	// 先检查 Job 是否支持参数化构建
	jobInfo, err := c.GetJobInfo(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("获取Job信息失败: %w", err)
	}

	// 检查是否为参数化 Job
	isParameterized := false
	if jobInfo.Property != nil {
		for _, prop := range jobInfo.Property {
			if prop.Class == "hudson.model.ParametersDefinitionProperty" {
				isParameterized = true
				break
			}
		}
	}

	// 检查是否为 Pipeline Job
	isPipelineJob := strings.Contains(jobInfo.Class, "WorkflowJob") || strings.Contains(jobInfo.Class, "org.jenkinsci.plugins.workflow.job.WorkflowJob")

	var path string
	var body io.Reader

	if isParameterized && len(params) > 0 {
		// 参数化构建
		path = fmt.Sprintf("/job/%s/buildWithParameters", url.PathEscape(jobName))
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		body = strings.NewReader(values.Encode())
	} else if isPipelineJob && len(params) > 0 {
		// 对于 Pipeline Job，即使没有参数化属性，也可能支持参数
		// 先尝试使用 buildWithParameters
		path = fmt.Sprintf("/job/%s/buildWithParameters", url.PathEscape(jobName))
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		body = strings.NewReader(values.Encode())
	} else {
		// 无参数构建
		path = fmt.Sprintf("/job/%s/build", url.PathEscape(jobName))
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("触发构建请求失败: %w", err)
	}
	defer resp.Body.Close()

	// Jenkins 返回 201 表示成功加入队列
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := extractJenkinsError(string(bodyBytes))
		// 如果是参数化错误，尝试无参数构建
		if strings.Contains(errMsg, "not parameterized") || strings.Contains(errMsg, "not support parameters") {
			// 重试无参数构建
			newPath := fmt.Sprintf("/job/%s/build", url.PathEscape(jobName))
			newResp, newErr := c.doRequest(ctx, http.MethodPost, newPath, nil)
			if newErr != nil {
				return nil, fmt.Errorf("触发构建失败: %s, 重试无参数构建也失败: %w", errMsg, newErr)
			}
			defer newResp.Body.Close()

			if newResp.StatusCode == http.StatusCreated || newResp.StatusCode == http.StatusOK {
				// 重试成功
				location := newResp.Header.Get("Location")
				if location == "" {
					return nil, errors.New("未获取到队列信息")
				}
				queueID := extractQueueID(location)
				return &TriggerBuildResult{
					QueueID:  queueID,
					QueueURL: location,
				}, nil
			} else {
				newBody, _ := io.ReadAll(newResp.Body)
				newErrMsg := extractJenkinsError(string(newBody))
				return nil, fmt.Errorf("触发构建失败: %s, 重试无参数构建也失败: %s", errMsg, newErrMsg)
			}
		} else {
			return nil, fmt.Errorf("触发构建失败: %s", errMsg)
		}
	}

	// 从 Location header 获取队列 URL
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("未获取到队列信息")
	}

	// 解析队列 ID：形如 http://jenkins/queue/item/123/
	queueID := extractQueueID(location)

	return &TriggerBuildResult{
		QueueID:  queueID,
		QueueURL: location,
	}, nil
}

// WaitForBuild 等待队列消费，获取构建号
// 会轮询队列状态直到构建开始或超时
func (c *Client) WaitForBuild(ctx context.Context, queueID int64, timeout time.Duration) (int, string, error) {
	deadline := time.Now().Add(timeout)
	path := fmt.Sprintf("/queue/item/%d/api/json", queueID)

	for time.Now().Before(deadline) {
		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return 0, "", err
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// 队列项可能已被消费，返回 404
			if resp.StatusCode == http.StatusNotFound {
				// 可能已经开始构建，尝试获取 Job 的最新构建
				return 0, "", errors.New("队列项已不存在，请检查构建列表")
			}
			return 0, "", fmt.Errorf("查询队列状态失败: HTTP %d", resp.StatusCode)
		}

		var item QueueItem
		if err := json.Unmarshal(bodyBytes, &item); err != nil {
			return 0, "", fmt.Errorf("解析队列信息失败: %w", err)
		}

		// 检查是否已开始执行
		if item.Executable.Number > 0 {
			return item.Executable.Number, item.Executable.URL, nil
		}

		// 检查是否被阻塞或卡住
		if item.Stuck {
			return 0, "", fmt.Errorf("构建被阻塞: %s", item.Why)
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return 0, "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return 0, "", errors.New("等待构建开始超时")
}

// TriggerBuildAndWait 触发构建并等待获取构建号
func (c *Client) TriggerBuildAndWait(ctx context.Context, jobName string, params map[string]string, waitTimeout time.Duration) (*TriggerBuildResult, error) {
	result, err := c.TriggerBuild(ctx, jobName, params)
	if err != nil {
		return nil, err
	}

	buildNumber, buildURL, err := c.WaitForBuild(ctx, result.QueueID, waitTimeout)
	if err != nil {
		return result, err // 返回部分结果
	}

	result.BuildNumber = buildNumber
	result.BuildURL = buildURL
	return result, nil
}

// GetBuildInfo 获取构建信息
func (c *Client) GetBuildInfo(ctx context.Context, jobName string, buildNumber int) (*BuildInfo, error) {
	path := fmt.Sprintf("/job/%s/%d/api/json", url.PathEscape(jobName), buildNumber)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("获取构建信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取构建信息失败: HTTP %d", resp.StatusCode)
	}

	var info BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析构建信息失败: %w", err)
	}

	return &info, nil
}

// GetConsoleLog 获取控制台日志
// startLine: 从第几行开始（用于增量获取），0 表示从头开始
func (c *Client) GetConsoleLog(ctx context.Context, jobName string, buildNumber int, startLine int) (string, error) {
	path := fmt.Sprintf("/job/%s/%d/consoleText", url.PathEscape(jobName), buildNumber)

	// 如果需要增量获取，使用 progressiveText
	if startLine > 0 {
		path = fmt.Sprintf("/job/%s/%d/logText/progressiveText?start=%d", url.PathEscape(jobName), buildNumber, startLine)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("获取控制台日志失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取控制台日志失败: HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取日志内容失败: %w", err)
	}

	return string(bodyBytes), nil
}

// StopBuild 停止构建
func (c *Client) StopBuild(ctx context.Context, jobName string, buildNumber int) error {
	path := fmt.Sprintf("/job/%s/%d/stop", url.PathEscape(jobName), buildNumber)

	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("停止构建请求失败: %w", err)
	}
	defer resp.Body.Close()

	// Jenkins 停止构建返回 302 重定向或 200
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("停止构建失败: HTTP %d", resp.StatusCode)
	}

	return nil
}

// GetJobInfo 获取 Job 信息
func (c *Client) GetJobInfo(ctx context.Context, jobName string) (*JobInfo, error) {
	path := fmt.Sprintf("/job/%s/api/json", url.PathEscape(jobName))

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("获取Job信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Job不存在: %s", jobName)
		}
		return nil, fmt.Errorf("获取Job信息失败: HTTP %d", resp.StatusCode)
	}

	var info JobInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析Job信息失败: %w", err)
	}

	return &info, nil
}

// CheckConnection 检查 Jenkins 连接是否正常
func (c *Client) CheckConnection(ctx context.Context) error {
	path := "/api/json"

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("Jenkins连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New("Jenkins认证失败，请检查用户名和API Token")
		}
		return fmt.Errorf("Jenkins连接异常: HTTP %d", resp.StatusCode)
	}

	return nil
}

// 辅助函数：从 URL 中提取队列 ID
func extractQueueID(queueURL string) int64 {
	// URL 格式: http://jenkins/queue/item/123/
	re := regexp.MustCompile(`/queue/item/(\d+)`)
	matches := re.FindStringSubmatch(queueURL)
	if len(matches) > 1 {
		id, _ := strconv.ParseInt(matches[1], 10, 64)
		return id
	}
	return 0
}

// extractJenkinsError 从 Jenkins HTML 错误响应中提取可读错误信息
func extractJenkinsError(htmlBody string) string {
	// 尝试提取 <title>...</title> 中的错误信息
	titleRe := regexp.MustCompile(`<title>(?:Error \d+ )?(.+?)</title>`)
	if matches := titleRe.FindStringSubmatch(htmlBody); len(matches) > 1 {
		return matches[1]
	}
	// 尝试提取 MESSAGE 行
	msgRe := regexp.MustCompile(`<td>MESSAGE</td>\s*</tr>\s*<tr>\s*<td>(.+?)</td>`)
	if matches := msgRe.FindStringSubmatch(htmlBody); len(matches) > 1 {
		return matches[1]
	}
	// 尝试提取 <h2>...</h2>
	h2Re := regexp.MustCompile(`<h2>(.+?)</h2>`)
	if matches := h2Re.FindStringSubmatch(htmlBody); len(matches) > 1 {
		return matches[1]
	}
	// 截取前 200 字符
	if len(htmlBody) > 200 {
		return htmlBody[:200] + "..."
	}
	return htmlBody
}

// BuildStatusToRunStatus 将 Jenkins 构建状态转换为流水线运行状态
func BuildStatusToRunStatus(building bool, result string) string {
	if building {
		return "running"
	}
	switch result {
	case "SUCCESS":
		return "success"
	case "FAILURE":
		return "failed"
	case "ABORTED":
		return "aborted"
	default:
		return "pending"
	}
}

// ==================== Pipeline Workflow API ====================

// PipelineStage Pipeline 阶段信息
type PipelineStage struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Status              string           `json:"status"` // SUCCESS, IN_PROGRESS, FAILED, NOT_EXECUTED, ABORTED
	StartTimeMillis     int64            `json:"startTimeMillis"`
	DurationMillis      int64            `json:"durationMillis"`
	PauseDurationMillis int64            `json:"pauseDurationMillis"`
	StageFlowNodes      []PipelineNode   `json:"stageFlowNodes"`
}

// PipelineNode Pipeline 节点信息
type PipelineNode struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	DurationMillis      int64  `json:"durationMillis"`
	ParentNodes         []string `json:"parentNodes"`
}

// PipelineRun Pipeline 运行信息
type PipelineRun struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Status              string          `json:"status"`
	StartTimeMillis     int64           `json:"startTimeMillis"`
	EndTimeMillis       int64           `json:"endTimeMillis"`
	DurationMillis      int64           `json:"durationMillis"`
	QueueDurationMillis int64           `json:"queueDurationMillis"`
	PauseDurationMillis int64           `json:"pauseDurationMillis"`
	Stages              []PipelineStage `json:"stages"`
}

// GetPipelineRun 获取 Pipeline 运行详情（包含阶段信息）
func (c *Client) GetPipelineRun(ctx context.Context, jobName string, buildNumber int) (*PipelineRun, error) {
	path := fmt.Sprintf("/job/%s/%d/wfapi/describe", url.PathEscape(jobName), buildNumber)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("获取Pipeline运行信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("构建记录不存在: %s #%d", jobName, buildNumber)
		}
		return nil, fmt.Errorf("获取Pipeline运行信息失败: HTTP %d", resp.StatusCode)
	}

	var run PipelineRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("解析Pipeline运行信息失败: %w", err)
	}

	return &run, nil
}

// GetNodeLog 获取 Pipeline 节点日志
func (c *Client) GetNodeLog(ctx context.Context, jobName string, buildNumber int, nodeID string) (string, error) {
	path := fmt.Sprintf("/job/%s/%d/execution/node/%s/wfapi/log", url.PathEscape(jobName), buildNumber, nodeID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("获取节点日志失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取节点日志失败: HTTP %d", resp.StatusCode)
	}

	// 返回的是 JSON 格式，包含 text 字段
	var logData struct {
		NodeID     string `json:"nodeId"`
		NodeStatus string `json:"nodeStatus"`
		Length     int64  `json:"length"`
		HasMore    bool   `json:"hasMore"`
		Text       string `json:"text"`
		ConsoleURL string `json:"consoleUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&logData); err != nil {
		return "", fmt.Errorf("解析节点日志失败: %w", err)
	}

	return logData.Text, nil
}
