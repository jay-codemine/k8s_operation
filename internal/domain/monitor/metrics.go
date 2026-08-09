package monitor

// =========================================================================
// Prometheus 查询 DTO — Monitor 域公共数据类型
// =========================================================================

// ClusterOverview 集群监控总览
type ClusterOverview struct {
	Healthy      bool    `json:"healthy"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NodeCount    int     `json:"node_count"`
	PodCount     int     `json:"pod_count"`
	AlertCount   int     `json:"alert_count"`
	NetworkIn    float64 `json:"network_in"`
	NetworkOut   float64 `json:"network_out"`
}

// NodeMetric 节点指标
type NodeMetric struct {
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	PodCount    int     `json:"pod_count"`
	Load1       float64 `json:"load1"`
	Load5       float64 `json:"load5"`
	NetworkIn   float64 `json:"network_in"`
	NetworkOut  float64 `json:"network_out"`
	Status      string  `json:"status"`
}

// HealthScore 集群健康评分
type HealthScore struct {
	Score       int                `json:"score"`
	Level       string             `json:"level"`
	Factors     map[string]float64 `json:"factors"`
	Suggestions []string           `json:"suggestions"`
}

// HeatmapCell 热力图单元（节点 × 时间）
type HeatmapCell struct {
	Node      string  `json:"node"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// PodStatusItem Pod 状态分布
type PodStatusItem struct {
	Phase string `json:"phase"`
	Count int    `json:"count"`
}

// AbnormalPod 异常 Pod（重启）
type AbnormalPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	Restarts  int    `json:"restarts"`
	Reason    string `json:"reason"`
}

// NamespaceMetric Namespace 维度聚合
type NamespaceMetric struct {
	Namespace   string  `json:"namespace"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	PodCount    int     `json:"pod_count"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// TrendData 趋势数据
type TrendData struct {
	Label  string       `json:"label"`
	Points []TrendPoint `json:"points"`
}

// PodMetric Pod 指标
type PodMetric struct {
	Name        string  `json:"name"`
	Namespace   string  `json:"namespace"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Status      string  `json:"status"`
}

// NodeDetail 单节点详情聚合（当前指标 + 多维趋势 + Top Pod + 元信息）
type NodeDetail struct {
	Instance string                  `json:"instance"`
	NodeName string                  `json:"node_name"`
	Current  NodeMetric              `json:"current"`
	Trends   map[string][]TrendPoint `json:"trends"`
	TopPods  []PodMetric             `json:"top_pods"`
	Info     map[string]string       `json:"info"`
}
