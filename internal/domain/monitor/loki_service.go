package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"k8soperation/pkg/loki"
)

// LokiService Loki 日志查询服务
//
// 设计：懒加载 + 每次请求实时解析数据源
//   - 若构造时传入了显式 URL（来自配置文件），则始终使用该 URL（向后兼容）
//   - 否则每次调用都会从数据库实时查询 monitor_datasource 表，
//     这样用户在前端【数据源管理】页面新增/启用/修改 Loki 数据源后无需重启即可生效。
//   - 数据源筛选优先级：is_default=1 优先；若无默认则取任一 enabled=1 且未删除的 Loki 数据源（按 ID DESC）。
type LokiService struct {
	db        *gorm.DB
	staticURL string        // 构造时传入的固定 URL（来自 config.yaml），优先级最高
	timeout   time.Duration // 查询超时
}

// NewLokiService 创建 Loki 服务
func NewLokiService(db *gorm.DB, lokiURL string, timeoutSec int) *LokiService {
	return &LokiService{db: db, staticURL: lokiURL, timeout: QueryTimeout(timeoutSec)}
}

// resolveURL 实时解析当前应使用的 Loki 地址
// 优先级：staticURL（配置文件）> is_default=1 > 任一 enabled=1 的 Loki 数据源
func (s *LokiService) resolveURL() string {
	if s.staticURL != "" {
		return s.staticURL
	}
	if s.db == nil {
		return ""
	}
	var ds Datasource
	// 1) 优先取默认 Loki 数据源
	if err := s.db.Where("type = ? AND is_default = 1 AND enabled = 1 AND is_del = 0", "loki").
		First(&ds).Error; err == nil && ds.URL != "" {
		return ds.URL
	}
	// 2) 回退：取任一启用的 Loki 数据源（按 ID DESC，最新创建优先）
	if err := s.db.Where("type = ? AND enabled = 1 AND is_del = 0", "loki").
		Order("id DESC").First(&ds).Error; err == nil && ds.URL != "" {
		return ds.URL
	}
	return ""
}

// resolveClient 解析 URL 并返回临时 client（每次调用新建，loki client 是轻量 http 包装）
func (s *LokiService) resolveClient() (*loki.Client, string, bool) {
	url := s.resolveURL()
	if url == "" {
		return nil, "", false
	}
	return loki.NewClient(url, s.timeout), url, true
}

// GetLokiURL 返回当前 Loki 地址（实时解析）
func (s *LokiService) GetLokiURL() string {
	return s.resolveURL()
}

// IsEnabled 当前是否有可用的 Loki 数据源
func (s *LokiService) IsEnabled() bool {
	return s.resolveURL() != ""
}

// IsHealthy 检查 Loki 连通性（实时解析数据源）
func (s *LokiService) IsHealthy(ctx context.Context) bool {
	client, _, ok := s.resolveClient()
	if !ok {
		return false
	}
	return client.Healthy(ctx)
}

// ===== 服务方法 =====

// QueryLogs 查询日志
func (s *LokiService) QueryLogs(ctx context.Context, query string, start, end time.Time, limit int, direction string) (*LogQueryResult, error) {
	client, _, ok := s.resolveClient()
	if !ok {
		return nil, fmt.Errorf("Loki 未配置，请在【数据源管理】页面添加并启用 Loki 数据源")
	}

	if limit <= 0 {
		limit = 100
	}
	if direction == "" {
		direction = "backward"
	}

	resp, err := client.QueryRange(ctx, query, start, end, limit, direction)
	if err != nil {
		return nil, err
	}

	streams, err := loki.ParseStreamResult(resp.Data.Result)
	if err != nil {
		return nil, fmt.Errorf("parse stream result failed: %w", err)
	}

	result := &LogQueryResult{
		Streams: len(streams),
	}

	for _, stream := range streams {
		for _, val := range stream.Values {
			ts, _ := strconv.ParseInt(val[0], 10, 64)
			entry := LogEntry{
				Timestamp: ts / 1_000_000, // ns -> ms
				Line:      val[1],
				Labels:    stream.Stream,
			}
			result.Entries = append(result.Entries, entry)
		}
	}
	result.TotalLines = len(result.Entries)

	return result, nil
}

// GetLabels 获取所有标签
func (s *LokiService) GetLabels(ctx context.Context, start, end time.Time) ([]string, error) {
	client, _, ok := s.resolveClient()
	if !ok {
		return nil, fmt.Errorf("Loki 未配置")
	}
	return client.GetLabels(ctx, start, end)
}

// GetLabelValues 获取指定标签的值列表
func (s *LokiService) GetLabelValues(ctx context.Context, label string, start, end time.Time) ([]string, error) {
	client, _, ok := s.resolveClient()
	if !ok {
		return nil, fmt.Errorf("Loki 未配置")
	}
	return client.GetLabelValues(ctx, label, start, end)
}

// GetStreams 获取日志流列表
func (s *LokiService) GetStreams(ctx context.Context, matcher string, start, end time.Time) ([]StreamInfo, error) {
	client, _, ok := s.resolveClient()
	if !ok {
		return nil, fmt.Errorf("Loki 未配置")
	}

	matchers := []string{matcher}
	if matcher == "" {
		matchers = []string{`{job=~".+"}`}
	}

	series, err := client.GetSeries(ctx, matchers, start, end)
	if err != nil {
		return nil, err
	}

	var streams []StreamInfo
	for _, s := range series {
		labelStr := "{"
		first := true
		for k, v := range s {
			if !first {
				labelStr += ", "
			}
			labelStr += fmt.Sprintf(`%s="%s"`, k, v)
			first = false
		}
		labelStr += "}"
		streams = append(streams, StreamInfo{
			Labels:   s,
			LabelStr: labelStr,
		})
	}
	return streams, nil
}

// GetLogVolume 获取日志量趋势（使用 count_over_time）
func (s *LokiService) GetLogVolume(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]LogVolumeSeries, error) {
	client, _, ok := s.resolveClient()
	if !ok {
		return nil, fmt.Errorf("Loki 未配置")
	}

	// 如果没有指定查询，使用通用查询
	if query == "" {
		query = `sum(count_over_time({job=~".+"}[1m]))`
	} else {
		// 将原始 stream selector 包装为 count_over_time
		query = fmt.Sprintf(`sum by (job) (count_over_time(%s[1m]))`, query)
	}

	resp, err := client.QueryRange(ctx, query, start, end, 0, "")
	if err != nil {
		return nil, err
	}

	// 结果可能是 matrix 类型
	var matrixResults []struct {
		Metric map[string]string `json:"metric"`
		Values [][2]interface{}  `json:"values"`
	}
	if err := parseJSON(resp.Data.Result, &matrixResults); err != nil {
		return nil, fmt.Errorf("parse matrix result failed: %w", err)
	}

	var series []LogVolumeSeries
	for _, mr := range matrixResults {
		s := LogVolumeSeries{Labels: mr.Metric}
		for _, v := range mr.Values {
			ts, _ := v[0].(float64)
			valStr, _ := v[1].(string)
			count, _ := strconv.ParseInt(valStr, 10, 64)
			s.Points = append(s.Points, LogVolumePoint{
				Timestamp: int64(ts),
				Count:     count,
			})
		}
		series = append(series, s)
	}
	return series, nil
}

// HealthCheck 执行健康检查（实时解析数据源）
func (s *LokiService) HealthCheck(ctx context.Context) *LokiHealthCheck {
	url := s.resolveURL()
	healthy := false
	if url != "" {
		healthy = loki.NewClient(url, 30*time.Second).Healthy(ctx)
	}
	return &LokiHealthCheck{
		Healthy: healthy,
		URL:     url,
	}
}

func parseJSON(raw []byte, v interface{}) error {
	return json.Unmarshal(raw, v)
}
