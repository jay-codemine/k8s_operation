package monitor

// =========================================================================
// Loki 日志查询 DTO — Monitor 域公共数据类型
// =========================================================================

// LogEntry 日志条目
type LogEntry struct {
	Timestamp int64             `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels"`
}

// LogQueryResult 日志查询结果
type LogQueryResult struct {
	Streams    int        `json:"streams"`
	TotalLines int        `json:"total_lines"`
	Entries    []LogEntry `json:"entries"`
}

// LogVolumePoint 日志量数据点
type LogVolumePoint struct {
	Timestamp int64 `json:"timestamp"`
	Count     int64 `json:"count"`
}

// LogVolumeSeries 日志量序列
type LogVolumeSeries struct {
	Labels map[string]string `json:"labels"`
	Points []LogVolumePoint  `json:"points"`
}

// StreamInfo 日志流信息
type StreamInfo struct {
	Labels   map[string]string `json:"labels"`
	LabelStr string            `json:"label_str"`
}

// LokiHealthCheck Loki 健康检查结果
type LokiHealthCheck struct {
	Healthy bool   `json:"healthy"`
	URL     string `json:"url"`
}
