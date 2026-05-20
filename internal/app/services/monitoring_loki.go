package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/loki"
)

// LokiService Loki 日志查询服务
type LokiService struct {
	client  *loki.Client
	enabled bool
	lokiURL string
}

// NewLokiService 创建 Loki 服务
func NewLokiService(lokiURL string) *LokiService {
	// 如果未指定 URL，从数据库获取默认 Loki 数据源
	if lokiURL == "" && global.DB != nil {
		var ds models.MonitorDatasource
		err := global.DB.Where("type = ? AND is_default = 1 AND enabled = 1 AND is_del = 0", "loki").First(&ds).Error
		if err == nil && ds.URL != "" {
			lokiURL = ds.URL
		}
	}

	client := loki.NewClient(lokiURL, 30*time.Second)
	return &LokiService{
		client:  client,
		enabled: lokiURL != "",
		lokiURL: lokiURL,
	}
}

// GetLokiURL 返回当前 Loki 地址
func (s *LokiService) GetLokiURL() string {
	return s.lokiURL
}

// IsHealthy 检查 Loki 连通性
func (s *LokiService) IsHealthy(ctx context.Context) bool {
	if !s.enabled {
		return false
	}
	return s.client.Healthy(ctx)
}

// ===== 数据结构 =====

// LogEntry 日志条目
type LogEntry struct {
	Timestamp int64             `json:"timestamp"`  // unix 毫秒
	Line      string            `json:"line"`       // 日志内容
	Labels    map[string]string `json:"labels"`     // 标签
}

// LogQueryResult 日志查询结果
type LogQueryResult struct {
	Streams    int        `json:"streams"`     // 流数量
	TotalLines int        `json:"total_lines"` // 总行数
	Entries    []LogEntry `json:"entries"`     // 日志条目
}

// LogVolumePoint 日志量数据点
type LogVolumePoint struct {
	Timestamp int64 `json:"timestamp"` // unix 秒
	Count     int64 `json:"count"`     // 日志行数
}

// LogVolumeSeries 日志量序列
type LogVolumeSeries struct {
	Labels map[string]string `json:"labels"`
	Points []LogVolumePoint  `json:"points"`
}

// StreamInfo 日志流信息
type StreamInfo struct {
	Labels   map[string]string `json:"labels"`
	LabelStr string            `json:"label_str"` // 格式化的标签字符串
}

// ===== 服务方法 =====

// QueryLogs 查询日志
func (s *LokiService) QueryLogs(ctx context.Context, query string, start, end time.Time, limit int, direction string) (*LogQueryResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Loki 未配置，请在【数据源管理】页面添加 Loki 数据源")
	}

	if limit <= 0 {
		limit = 100
	}
	if direction == "" {
		direction = "backward"
	}

	resp, err := s.client.QueryRange(ctx, query, start, end, limit, direction)
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
	if !s.enabled {
		return nil, fmt.Errorf("Loki 未配置")
	}
	return s.client.GetLabels(ctx, start, end)
}

// GetLabelValues 获取指定标签的值列表
func (s *LokiService) GetLabelValues(ctx context.Context, label string, start, end time.Time) ([]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Loki 未配置")
	}
	return s.client.GetLabelValues(ctx, label, start, end)
}

// GetStreams 获取日志流列表
func (s *LokiService) GetStreams(ctx context.Context, matcher string, start, end time.Time) ([]StreamInfo, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Loki 未配置")
	}

	matchers := []string{matcher}
	if matcher == "" {
		matchers = []string{`{job=~".+"}`}
	}

	series, err := s.client.GetSeries(ctx, matchers, start, end)
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
	if !s.enabled {
		return nil, fmt.Errorf("Loki 未配置")
	}

	// 如果没有指定查询，使用通用查询
	if query == "" {
		query = `sum(count_over_time({job=~".+"}[1m]))`
	} else {
		// 将原始 stream selector 包装为 count_over_time
		query = fmt.Sprintf(`sum by (job) (count_over_time(%s[1m]))`, query)
	}

	resp, err := s.client.QueryRange(ctx, query, start, end, 0, "")
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

// LokiHealthCheck Loki 健康检查结果
type LokiHealthCheck struct {
	Healthy bool   `json:"healthy"`
	URL     string `json:"url"`
}

// HealthCheck 执行健康检查
func (s *LokiService) HealthCheck(ctx context.Context) *LokiHealthCheck {
	return &LokiHealthCheck{
		Healthy: s.IsHealthy(ctx),
		URL:     s.lokiURL,
	}
}

func parseJSON(raw []byte, v interface{}) error {
	return json.Unmarshal(raw, v)
}
