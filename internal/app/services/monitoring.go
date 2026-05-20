package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	prom "k8soperation/pkg/prometheus"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// MonitoringService 监控服务
type MonitoringService struct {
	client        *prom.Client
	enabled       bool
	prometheusURL string
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(prometheusURL string) *MonitoringService {
	// 如果 config.yaml 未配置，尝试从数据库获取默认数据源
	if prometheusURL == "" && global.DB != nil {
		var ds models.MonitorDatasource
		err := global.DB.Where("type IN (?,?,?) AND is_default = 1 AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").First(&ds).Error
		if err == nil && ds.URL != "" {
			prometheusURL = ds.URL
		}
	}

	client := prom.NewClient(prometheusURL, 30*time.Second)
	return &MonitoringService{
		client:        client,
		enabled:       prometheusURL != "",
		prometheusURL: prometheusURL,
	}
}

// GetPrometheusURL 返回当前使用的 Prometheus 地址
func (s *MonitoringService) GetPrometheusURL() string {
	return s.prometheusURL
}

// ===== 数据结构 =====

// ClusterOverview 集群监控总览
type ClusterOverview struct {
	Healthy      bool    `json:"healthy"`
	CPUUsage     float64 `json:"cpu_usage"`      // 集群 CPU 使用率 %
	MemoryUsage  float64 `json:"memory_usage"`   // 集群内存使用率 %
	DiskUsage    float64 `json:"disk_usage"`     // 集群磁盘使用率 %
	NodeCount    int     `json:"node_count"`     // 节点总数
	PodCount     int     `json:"pod_count"`      // Pod 总数
	AlertCount   int     `json:"alert_count"`    // 活跃告警数
	NetworkIn    float64 `json:"network_in"`     // 入站网络 bytes/s
	NetworkOut   float64 `json:"network_out"`    // 出站网络 bytes/s
}

// NodeMetric 节点指标
type NodeMetric struct {
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpu_usage"`     // %
	MemoryUsage float64 `json:"memory_usage"`  // %
	DiskUsage   float64 `json:"disk_usage"`    // %
	PodCount    int     `json:"pod_count"`
	Status      string  `json:"status"`        // Ready/NotReady
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
	CPUUsage    float64 `json:"cpu_usage"`     // cores
	MemoryUsage float64 `json:"memory_usage"`  // bytes
	Status      string  `json:"status"`
}

// ===== 服务方法 =====

// GetClusterOverview 获取集群监控总览
func (s *MonitoringService) GetClusterOverview(ctx context.Context) (*ClusterOverview, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Prometheus 未配置，请在 config.yaml 的 Monitoring.PrometheusURL 中设置正确地址，或在【数据源管理】页面添加数据源")
	}
	overview := &ClusterOverview{Healthy: true}

	// CPU 使用率
	cpuResult, err := s.client.QueryInstant(ctx, `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`)
	if err == nil {
		overview.CPUUsage = s.extractScalar(cpuResult)
	}

	// 内存使用率
	memResult, err := s.client.QueryInstant(ctx, `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`)
	if err == nil {
		overview.MemoryUsage = s.extractAvgVector(memResult)
	}

	// 磁盘使用率
	diskResult, err := s.client.QueryInstant(ctx, `100 - (avg(node_filesystem_avail_bytes{mountpoint="/"}) / avg(node_filesystem_size_bytes{mountpoint="/"}) * 100)`)
	if err == nil {
		overview.DiskUsage = s.extractScalar(diskResult)
	}

	// 节点数量
	nodeResult, err := s.client.QueryInstant(ctx, `count(up{job="node-exporter"})`)
	if err == nil {
		overview.NodeCount = int(s.extractScalar(nodeResult))
	}

	// Pod 数量
	podResult, err := s.client.QueryInstant(ctx, `count(kube_pod_info)`)
	if err == nil {
		overview.PodCount = int(s.extractScalar(podResult))
	}

	// 活跃告警数
	alertResult, err := s.client.QueryInstant(ctx, `count(ALERTS{alertstate="firing"})`)
	if err == nil {
		overview.AlertCount = int(s.extractScalar(alertResult))
	}

	// 网络入站
	netInResult, err := s.client.QueryInstant(ctx, `sum(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*"}[5m]))`)
	if err == nil {
		overview.NetworkIn = s.extractScalar(netInResult)
	}

	// 网络出站
	netOutResult, err := s.client.QueryInstant(ctx, `sum(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*"}[5m]))`)
	if err == nil {
		overview.NetworkOut = s.extractScalar(netOutResult)
	}

	return overview, nil
}

// GetNodeMetrics 获取节点指标列表
func (s *MonitoringService) GetNodeMetrics(ctx context.Context) ([]NodeMetric, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Prometheus 未配置，请在 config.yaml 的 Monitoring.PrometheusURL 中设置正确地址")
	}
	var nodes []NodeMetric

	// 获取节点 CPU
	cpuResult, err := s.client.QueryInstant(ctx, `100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`)
	if err != nil {
		return nil, fmt.Errorf("query node CPU failed: %w", err)
	}

	cpuVectors, _ := prom.ParseVectorResult(cpuResult.Data.Result)
	nodeMap := make(map[string]*NodeMetric)
	for _, v := range cpuVectors {
		instance := v.Metric["instance"]
		node := &NodeMetric{Name: instance, Status: "Ready"}
		node.CPUUsage = s.parseValue(v.Value[1])
		nodeMap[instance] = node
	}

	// 获取节点内存
	memResult, _ := s.client.QueryInstant(ctx, `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`)
	if memResult != nil {
		memVectors, _ := prom.ParseVectorResult(memResult.Data.Result)
		for _, v := range memVectors {
			instance := v.Metric["instance"]
			if node, ok := nodeMap[instance]; ok {
				node.MemoryUsage = s.parseValue(v.Value[1])
			}
		}
	}

	// 获取节点磁盘
	diskResult, _ := s.client.QueryInstant(ctx, `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)`)
	if diskResult != nil {
		diskVectors, _ := prom.ParseVectorResult(diskResult.Data.Result)
		for _, v := range diskVectors {
			instance := v.Metric["instance"]
			if node, ok := nodeMap[instance]; ok {
				node.DiskUsage = s.parseValue(v.Value[1])
			}
		}
	}

	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}
	return nodes, nil
}

// GetResourceTrend 获取资源趋势数据（近1小时）
func (s *MonitoringService) GetResourceTrend(ctx context.Context, resource string, duration time.Duration) ([]TrendData, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Prometheus 未配置，请在 config.yaml 的 Monitoring.PrometheusURL 中设置正确地址")
	}
	end := time.Now()
	start := end.Add(-duration)
	step := duration / 60 // 约60个数据点

	var queries map[string]string
	switch resource {
	case "cpu":
		queries = map[string]string{
			"CPU使用率": `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`,
		}
	case "memory":
		queries = map[string]string{
			"内存使用率": `avg(100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100))`,
		}
	case "disk":
		queries = map[string]string{
			"磁盘使用率": `avg(100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100))`,
		}
	case "network":
		queries = map[string]string{
			"入站流量":  `sum(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*"}[5m]))`,
			"出站流量":  `sum(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*"}[5m]))`,
		}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resource)
	}

	var trends []TrendData
	for label, query := range queries {
		result, err := s.client.QueryRange(ctx, query, start, end, step)
		if err != nil {
			continue
		}
		matrixResults, err := prom.ParseMatrixResult(result.Data.Result)
		if err != nil || len(matrixResults) == 0 {
			continue
		}

		td := TrendData{Label: label}
		for _, pair := range matrixResults[0].Values {
			ts, _ := pair[0].(float64)
			val := s.parseValue(pair[1])
			td.Points = append(td.Points, TrendPoint{
				Timestamp: int64(ts),
				Value:     val,
			})
		}
		trends = append(trends, td)
	}

	return trends, nil
}

// GetTopPods 获取资源占用 Top N Pods
func (s *MonitoringService) GetTopPods(ctx context.Context, metric string, limit int) ([]PodMetric, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Prometheus 未配置，请在 config.yaml 的 Monitoring.PrometheusURL 中设置正确地址")
	}
	var query string
	switch metric {
	case "cpu":
		query = fmt.Sprintf(`topk(%d, sum by(pod, namespace)(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[5m])))`, limit)
	case "memory":
		query = fmt.Sprintf(`topk(%d, sum by(pod, namespace)(container_memory_working_set_bytes{container!="",pod!=""}))`, limit)
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	result, err := s.client.QueryInstant(ctx, query)
	if err != nil {
		return nil, err
	}

	vectors, _ := prom.ParseVectorResult(result.Data.Result)
	var pods []PodMetric
	for _, v := range vectors {
		pod := PodMetric{
			Name:      v.Metric["pod"],
			Namespace: v.Metric["namespace"],
		}
		val := s.parseValue(v.Value[1])
		if metric == "cpu" {
			pod.CPUUsage = val
		} else {
			pod.MemoryUsage = val
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

// IsHealthy 检查 Prometheus 连通性
func (s *MonitoringService) IsHealthy(ctx context.Context) bool {
	if !s.enabled {
		return false
	}
	return s.client.Healthy(ctx)
}

// ===== 辅助方法 =====

func (s *MonitoringService) extractScalar(result *prom.QueryResult) float64 {
	if result == nil {
		return 0
	}
	vectors, err := prom.ParseVectorResult(result.Data.Result)
	if err != nil || len(vectors) == 0 {
		return 0
	}
	return s.parseValue(vectors[0].Value[1])
}

func (s *MonitoringService) extractAvgVector(result *prom.QueryResult) float64 {
	if result == nil {
		return 0
	}
	vectors, err := prom.ParseVectorResult(result.Data.Result)
	if err != nil || len(vectors) == 0 {
		return 0
	}
	var total float64
	for _, v := range vectors {
		total += s.parseValue(v.Value[1])
	}
	return total / float64(len(vectors))
}

func (s *MonitoringService) parseValue(v interface{}) float64 {
	switch val := v.(type) {
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case float64:
		return val
	default:
		return 0
	}
}
