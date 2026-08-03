package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	prom "k8soperation/pkg/prometheus"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// MonitoringService 监控服务
//
// 设计：懒加载 + 每次请求实时解析数据源地址
//   - 构造时传入的 staticURL（config.yaml.Monitoring.PrometheusURL）优先级最高
//   - 否则每次调用从 monitor_datasource 表实时查询，用户在【数据源管理】修改后无需重启即可生效
//   - 筛选优先级：context 中的 datasource_id（前端切换） > is_default=1 > 任一 enabled=1 的 prometheus/victoriametrics/thanos
type MonitoringService struct {
	staticURL string // 构造时传入的固定 URL（config.yaml），最高优先级
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(prometheusURL string) *MonitoringService {
	return &MonitoringService{staticURL: prometheusURL}
}

// dsIDCtxKey 用于在 ctx 中传递前端选择的 datasource_id
type dsIDCtxKey struct{}

// WithDatasourceID 把前端切换的 datasource_id 注入 ctx
func WithDatasourceID(ctx context.Context, id int64) context.Context {
	if id <= 0 {
		return ctx
	}
	return context.WithValue(ctx, dsIDCtxKey{}, id)
}

// datasourceIDFromCtx 从 ctx 提取 datasource_id
func datasourceIDFromCtx(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if v := ctx.Value(dsIDCtxKey{}); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// resolveURL 实时解析当前应使用的 Prometheus 地址
// 优先级：ctx.datasource_id（前端指定）> 数据库 is_default=1 > 数据库任一 enabled=1 > config.yaml.staticURL（首次启动兑底）
func (s *MonitoringService) resolveURL(ctx context.Context) string {
	if global.DB != nil {
		// 0) 最高优先：ctx 中携带的 datasource_id（前端切换）
		if id := datasourceIDFromCtx(ctx); id > 0 {
			var ds models.MonitorDatasource
			if err := global.DB.Where("id = ? AND enabled = 1 AND is_del = 0", id).First(&ds).Error; err == nil && ds.URL != "" {
				return ds.URL
			}
		}
		var ds models.MonitorDatasource
		// 1) 优先取默认数据源
		if err := global.DB.Where("type IN (?,?,?) AND is_default = 1 AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").First(&ds).Error; err == nil && ds.URL != "" {
			return ds.URL
		}
		// 2) 回退：取任一启用的 prometheus/victoriametrics/thanos数据源（按 ID DESC，最新优先）
		if err := global.DB.Where("type IN (?,?,?) AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").Order("id DESC").First(&ds).Error; err == nil && ds.URL != "" {
			return ds.URL
		}
	}
	// 3) 兑底：config.yaml 中的 staticURL（首次启动、数据库没任何数据源时使用）
	return s.staticURL
}

// queryTimeout 单条查询的超时，取 config.yaml.Monitoring.QueryTimeout
func queryTimeout() time.Duration {
	if global.MonitoringSetting != nil && global.MonitoringSetting.QueryTimeout > 0 {
		return time.Duration(global.MonitoringSetting.QueryTimeout) * time.Second
	}
	return 30 * time.Second
}

// resolveClient 解析 URL 并返回临时 client（每次调用新建）
func (s *MonitoringService) resolveClient(ctx context.Context) (*prom.Client, string, bool) {
	url := s.resolveURL(ctx)
	if url == "" {
		return nil, "", false
	}
	return prom.NewClient(url, queryTimeout()), url, true
}

// GetPrometheusURL 返回当前使用的 Prometheus 地址（实时解析）
func (s *MonitoringService) GetPrometheusURL(ctx context.Context) string {
	return s.resolveURL(ctx)
}

// IsEnabled 当前是否有可用的 Prometheus 数据源
func (s *MonitoringService) IsEnabled(ctx context.Context) bool {
	return s.resolveURL(ctx) != ""
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
	Load1       float64 `json:"load1"`         // 1分钟负载
	Load5       float64 `json:"load5"`         // 5分钟负载
	NetworkIn   float64 `json:"network_in"`    // bytes/s
	NetworkOut  float64 `json:"network_out"`   // bytes/s
	Status      string  `json:"status"`        // Ready/NotReady
}

// HealthScore 集群健康评分
type HealthScore struct {
	Score       int                `json:"score"`        // 综合评分 0-100
	Level       string             `json:"level"`        // excellent / good / warning / critical
	Factors     map[string]float64 `json:"factors"`      // 各维度扣分明细
	Suggestions []string           `json:"suggestions"`  // 改进建议
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
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Container string  `json:"container"`
	Restarts  int     `json:"restarts"`
	Reason    string  `json:"reason"`
}

// NamespaceMetric Namespace 维度聚合
type NamespaceMetric struct {
	Namespace   string  `json:"namespace"`
	CPUUsage    float64 `json:"cpu_usage"`     // cores
	MemoryUsage float64 `json:"memory_usage"`  // bytes
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
	CPUUsage    float64 `json:"cpu_usage"`     // cores
	MemoryUsage float64 `json:"memory_usage"`  // bytes
	Status      string  `json:"status"`
}

// NodeDetail 单节点详情聚合（当前指标 + 多维趋势 + Top Pod + 元信息）
type NodeDetail struct {
	Instance string                  `json:"instance"`   // 例如 192.168.124.10:9100
	NodeName string                  `json:"node_name"`  // 例如 k8s-master01
	Current  NodeMetric              `json:"current"`    // 当前快照
	Trends   map[string][]TrendPoint `json:"trends"`     // cpu/memory/disk/net_in/net_out
	TopPods  []PodMetric             `json:"top_pods"`   // 该节点上 Top Pod
	Info     map[string]string       `json:"info"`       // os/kernel/kubelet/role/cpu_total/mem_total
}

// ===== 服务方法 =====

// GetClusterOverview 获取集群监控总览
func (s *MonitoringService) GetClusterOverview(ctx context.Context) (*ClusterOverview, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置，请在 config.yaml 的 Monitoring.PrometheusURL 中设置或在【数据源管理】添加并启用数据源")
	}
	overview := &ClusterOverview{Healthy: true}

	// CPU 使用率
	cpuResult, err := client.QueryInstant(ctx, `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`)
	if err == nil {
		overview.CPUUsage = s.extractScalar(cpuResult)
	}

	// 内存使用率
	memResult, err := client.QueryInstant(ctx, `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`)
	if err == nil {
		overview.MemoryUsage = s.extractAvgVector(memResult)
	}

	// 磁盘使用率
	diskResult, err := client.QueryInstant(ctx, `100 - (avg(node_filesystem_avail_bytes{mountpoint="/"}) / avg(node_filesystem_size_bytes{mountpoint="/"}) * 100)`)
	if err == nil {
		overview.DiskUsage = s.extractScalar(diskResult)
	}

	// 节点数量（kube-prometheus-stack 中 node-exporter 的 job 名可能为 node-exporter 或 kube-prometheus-stack-prometheus-node-exporter，用 =~ 兼容）
	nodeResult, err := client.QueryInstant(ctx, `count(up{job=~".*node-exporter.*"})`)
	if err == nil {
		overview.NodeCount = int(s.extractScalar(nodeResult))
	}

	// Pod 数量
	podResult, err := client.QueryInstant(ctx, `count(kube_pod_info)`)
	if err == nil {
		overview.PodCount = int(s.extractScalar(podResult))
	}

	// 活跃告警数
	alertResult, err := client.QueryInstant(ctx, `count(ALERTS{alertstate="firing"})`)
	if err == nil {
		overview.AlertCount = int(s.extractScalar(alertResult))
	}

	// 网络入站
	netInResult, err := client.QueryInstant(ctx, `sum(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*"}[5m]))`)
	if err == nil {
		overview.NetworkIn = s.extractScalar(netInResult)
	}

	// 网络出站
	netOutResult, err := client.QueryInstant(ctx, `sum(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*"}[5m]))`)
	if err == nil {
		overview.NetworkOut = s.extractScalar(netOutResult)
	}

	return overview, nil
}

// GetNodeMetrics 获取节点指标列表
func (s *MonitoringService) GetNodeMetrics(ctx context.Context) ([]NodeMetric, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
	}
	var nodes []NodeMetric

	// 获取节点 CPU
	cpuResult, err := client.QueryInstant(ctx, `100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`)
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
	memResult, _ := client.QueryInstant(ctx, `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`)
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
	diskResult, _ := client.QueryInstant(ctx, `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)`)
	if diskResult != nil {
		diskVectors, _ := prom.ParseVectorResult(diskResult.Data.Result)
		for _, v := range diskVectors {
			instance := v.Metric["instance"]
			if node, ok := nodeMap[instance]; ok {
				node.DiskUsage = s.parseValue(v.Value[1])
			}
		}
	}

	// 节点 Load1 / Load5
	load1Result, _ := client.QueryInstant(ctx, `node_load1`)
	if load1Result != nil {
		for _, v := range mustVector(load1Result) {
			if node, ok := nodeMap[v.Metric["instance"]]; ok {
				node.Load1 = s.parseValue(v.Value[1])
			}
		}
	}
	load5Result, _ := client.QueryInstant(ctx, `node_load5`)
	if load5Result != nil {
		for _, v := range mustVector(load5Result) {
			if node, ok := nodeMap[v.Metric["instance"]]; ok {
				node.Load5 = s.parseValue(v.Value[1])
			}
		}
	}

	// 节点网络
	netInResult, _ := client.QueryInstant(ctx, `sum by(instance)(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*"}[5m]))`)
	if netInResult != nil {
		for _, v := range mustVector(netInResult) {
			if node, ok := nodeMap[v.Metric["instance"]]; ok {
				node.NetworkIn = s.parseValue(v.Value[1])
			}
		}
	}
	netOutResult, _ := client.QueryInstant(ctx, `sum by(instance)(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*"}[5m]))`)
	if netOutResult != nil {
		for _, v := range mustVector(netOutResult) {
			if node, ok := nodeMap[v.Metric["instance"]]; ok {
				node.NetworkOut = s.parseValue(v.Value[1])
			}
		}
	}

	// 节点 Pod 数（按 node label 聚合）
	// kube_pod_info 的 node label 是 K8s 节点名（如 k8s-master01），
	// 而 node_exporter 的 instance 是 IP:9100（如 192.168.124.10:9100），
	// 两者直接对比基本永远不相等，因此需要借助 kube_node_info 建立 node↔internal_ip 映射。
	nodeIPMap := make(map[string]string) // node_name -> internal_ip
	if r, _ := client.QueryInstant(ctx, `kube_node_info`); r != nil {
		for _, v := range mustVector(r) {
			if name := v.Metric["node"]; name != "" {
				if ip := v.Metric["internal_ip"]; ip != "" {
					nodeIPMap[name] = ip
				}
			}
		}
	}
	podResult, _ := client.QueryInstant(ctx, `count by(node)(kube_pod_info)`)
	if podResult != nil {
		for _, v := range mustVector(podResult) {
			nodeName := v.Metric["node"]
			if nodeName == "" {
				continue
			}
			count := int(s.parseValue(v.Value[1]))
			ip := nodeIPMap[nodeName]
			// 三段兜底匹配：1) 完全等于  2) instance 去掉端口==nodeName  3) instance 去掉端口==internal_ip
			for inst, node := range nodeMap {
				host := inst
				if i := strings.IndexByte(inst, ':'); i > 0 {
					host = inst[:i]
				}
				if inst == nodeName || host == nodeName || node.Name == nodeName || (ip != "" && host == ip) {
					node.PodCount = count
					break
				}
			}
		}
	}

	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}
	return nodes, nil
}

// mustVector 安全解析 vector 结果，失败返回空
func mustVector(r *prom.QueryResult) []prom.VectorResult {
	if r == nil {
		return nil
	}
	v, _ := prom.ParseVectorResult(r.Data.Result)
	return v
}

// GetResourceTrend 获取资源趋势数据（近1小时）
func (s *MonitoringService) GetResourceTrend(ctx context.Context, resource string, duration time.Duration) ([]TrendData, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
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
		result, err := client.QueryRange(ctx, query, start, end, step)
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
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
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

	result, err := client.QueryInstant(ctx, query)
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

// IsHealthy 检查 Prometheus 连通性（实时解析数据源）
func (s *MonitoringService) IsHealthy(ctx context.Context) bool {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return false
	}
	return client.Healthy(ctx)
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

// ===== 新增大厂级能力 =====

// GetClusterHealthScore 集群健康评分
// 权重：CPU 25 + 内存 25 + 磁盘 20 + 告警 15 + Pod失败率 15
// 评分逻辑：各维度越高扣分越多，总分 100 - 扣分
func (s *MonitoringService) GetClusterHealthScore(ctx context.Context) (*HealthScore, error) {
	ov, err := s.GetClusterOverview(ctx)
	if err != nil {
		return nil, err
	}
	client, _, _ := s.resolveClient(ctx)

	factors := map[string]float64{}
	suggestions := []string{}

	// CPU 扣分：>80% 全扣 25，>60% 扣 15，>40% 扣 5
	cpuPenalty := 0.0
	switch {
	case ov.CPUUsage >= 80:
		cpuPenalty = 25
		suggestions = append(suggestions, "集群 CPU 负载过高，建议扩容或优化高耗 Pod")
	case ov.CPUUsage >= 60:
		cpuPenalty = 15
	case ov.CPUUsage >= 40:
		cpuPenalty = 5
	}
	factors["cpu"] = cpuPenalty

	// 内存扣分
	memPenalty := 0.0
	switch {
	case ov.MemoryUsage >= 85:
		memPenalty = 25
		suggestions = append(suggestions, "内存负载过高，可能触发 OOM，建议检查内存泄露")
	case ov.MemoryUsage >= 70:
		memPenalty = 15
	case ov.MemoryUsage >= 50:
		memPenalty = 5
	}
	factors["memory"] = memPenalty

	// 磁盘扣分
	diskPenalty := 0.0
	switch {
	case ov.DiskUsage >= 85:
		diskPenalty = 20
		suggestions = append(suggestions, "磁盘使用率过高，建议清理旧镜像/日志或扩容")
	case ov.DiskUsage >= 70:
		diskPenalty = 10
	}
	factors["disk"] = diskPenalty

	// 告警扣分：每个 firing 告警扣 5 分，上限 15
	alertPenalty := float64(ov.AlertCount) * 5
	if alertPenalty > 15 {
		alertPenalty = 15
	}
	if ov.AlertCount > 0 {
		suggestions = append(suggestions, fmt.Sprintf("有 %d 个活跃告警未处理", ov.AlertCount))
	}
	factors["alerts"] = alertPenalty

	// Pod 失败率
	podPenalty := 0.0
	if client != nil && ov.PodCount > 0 {
		failedResult, _ := client.QueryInstant(ctx, `count(kube_pod_status_phase{phase=~"Failed|Unknown"} == 1)`)
		failed := int(s.extractScalar(failedResult))
		pendingResult, _ := client.QueryInstant(ctx, `count(kube_pod_status_phase{phase="Pending"} == 1)`)
		pending := int(s.extractScalar(pendingResult))
		abnormalRate := float64(failed+pending) / float64(ov.PodCount) * 100
		switch {
		case abnormalRate >= 20:
			podPenalty = 15
			suggestions = append(suggestions, fmt.Sprintf("异常 Pod %d 个（Failed %d / Pending %d）", failed+pending, failed, pending))
		case abnormalRate >= 10:
			podPenalty = 8
		case abnormalRate >= 5:
			podPenalty = 3
		}
	}
	factors["pods"] = podPenalty

	totalPenalty := cpuPenalty + memPenalty + diskPenalty + alertPenalty + podPenalty
	score := 100 - int(totalPenalty)
	if score < 0 {
		score = 0
	}

	level := "excellent"
	switch {
	case score < 60:
		level = "critical"
	case score < 75:
		level = "warning"
	case score < 90:
		level = "good"
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "集群运行状况良好，继续保持")
	}

	return &HealthScore{
		Score:       score,
		Level:       level,
		Factors:     factors,
		Suggestions: suggestions,
	}, nil
}

// GetNodeHeatmap 节点热力图（节点 × 时间）
// metric: cpu / memory
func (s *MonitoringService) GetNodeHeatmap(ctx context.Context, metric string, duration time.Duration) ([]HeatmapCell, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
	}
	end := time.Now()
	start := end.Add(-duration)
	step := duration / 30 // 30 个时间片

	var query string
	switch metric {
	case "cpu":
		query = `100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	case "memory":
		query = `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	result, err := client.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, err
	}
	matrices, _ := prom.ParseMatrixResult(result.Data.Result)

	var cells []HeatmapCell
	for _, m := range matrices {
		instance := m.Metric["instance"]
		for _, pair := range m.Values {
			ts, _ := pair[0].(float64)
			cells = append(cells, HeatmapCell{
				Node:      instance,
				Timestamp: int64(ts),
				Value:     s.parseValue(pair[1]),
			})
		}
	}
	return cells, nil
}

// GetPodStatusDistribution Pod 状态分布
// 说明：与 GetClusterOverview 保持一致的容错策略——Prometheus 未配置或查询失败时
// 返回空列表而非错误，避免首页因监控数据源暂不可用而整体报错。
func (s *MonitoringService) GetPodStatusDistribution(ctx context.Context) ([]PodStatusItem, error) {
	items := make([]PodStatusItem, 0)
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return items, nil
	}
	result, err := client.QueryInstant(ctx, `sum by(phase)(kube_pod_status_phase == 1)`)
	if err != nil || result == nil {
		return items, nil
	}
	vectors, _ := prom.ParseVectorResult(result.Data.Result)
	for _, v := range vectors {
		items = append(items, PodStatusItem{
			Phase: v.Metric["phase"],
			Count: int(s.parseValue(v.Value[1])),
		})
	}
	return items, nil
}

// GetAbnormalPods 获取重启超过 N 次的 Pod
func (s *MonitoringService) GetAbnormalPods(ctx context.Context, minRestarts int, limit int) ([]AbnormalPod, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
	}
	if minRestarts < 1 {
		minRestarts = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := fmt.Sprintf(`topk(%d, kube_pod_container_status_restarts_total > %d)`, limit, minRestarts-1)
	result, err := client.QueryInstant(ctx, query)
	if err != nil {
		return nil, err
	}
	vectors, _ := prom.ParseVectorResult(result.Data.Result)
	var pods []AbnormalPod
	for _, v := range vectors {
		pods = append(pods, AbnormalPod{
			Name:      v.Metric["pod"],
			Namespace: v.Metric["namespace"],
			Container: v.Metric["container"],
			Restarts:  int(s.parseValue(v.Value[1])),
			Reason:    "CrashLoopBackOff/OOMKilled可能",
		})
	}
	return pods, nil
}

// GetNamespaceMetrics 按 Namespace 聚合资源使用
func (s *MonitoringService) GetNamespaceMetrics(ctx context.Context, limit int) ([]NamespaceMetric, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	nsMap := make(map[string]*NamespaceMetric)

	// CPU
	cpuQ := fmt.Sprintf(`topk(%d, sum by(namespace)(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[5m])))`, limit)
	cpuResult, _ := client.QueryInstant(ctx, cpuQ)
	for _, v := range mustVector(cpuResult) {
		ns := v.Metric["namespace"]
		if ns == "" {
			continue
		}
		item, ok := nsMap[ns]
		if !ok {
			item = &NamespaceMetric{Namespace: ns}
			nsMap[ns] = item
		}
		item.CPUUsage = s.parseValue(v.Value[1])
	}

	// Memory
	memQ := fmt.Sprintf(`topk(%d, sum by(namespace)(container_memory_working_set_bytes{container!="",pod!=""}))`, limit)
	memResult, _ := client.QueryInstant(ctx, memQ)
	for _, v := range mustVector(memResult) {
		ns := v.Metric["namespace"]
		if ns == "" {
			continue
		}
		item, ok := nsMap[ns]
		if !ok {
			item = &NamespaceMetric{Namespace: ns}
			nsMap[ns] = item
		}
		item.MemoryUsage = s.parseValue(v.Value[1])
	}

	// Pod 数
	podResult, _ := client.QueryInstant(ctx, `count by(namespace)(kube_pod_info)`)
	for _, v := range mustVector(podResult) {
		ns := v.Metric["namespace"]
		if ns == "" {
			continue
		}
		item, ok := nsMap[ns]
		if !ok {
			item = &NamespaceMetric{Namespace: ns}
			nsMap[ns] = item
		}
		item.PodCount = int(s.parseValue(v.Value[1]))
	}

	var list []NamespaceMetric
	for _, v := range nsMap {
		list = append(list, *v)
	}
	return list, nil
}

// GetNodeDetail 获取单个节点详情（当前指标 + 5 维趋势 + Top Pod + 元信息）
// instance: node_exporter 实例，例如 192.168.124.10:9100
func (s *MonitoringService) GetNodeDetail(ctx context.Context, instance string, duration time.Duration) (*NodeDetail, error) {
	client, _, ok := s.resolveClient(ctx)
	if !ok {
		return nil, fmt.Errorf("Prometheus 未配置")
	}
	if instance == "" {
		return nil, fmt.Errorf("instance 不能为空")
	}
	detail := &NodeDetail{
		Instance: instance,
		Current:  NodeMetric{Name: instance, Status: "Ready"},
		Trends:   map[string][]TrendPoint{},
		Info:     map[string]string{},
	}

	// 1) 通过 node_uname_info 拿 nodename（与 kube-state-metrics 的 node label 对齐）
	unameQ := fmt.Sprintf(`node_uname_info{instance="%s"}`, instance)
	if r, err := client.QueryInstant(ctx, unameQ); err == nil {
		for _, v := range mustVector(r) {
			if n := v.Metric["nodename"]; n != "" {
				detail.NodeName = n
			}
			if n := v.Metric["release"]; n != "" {
				detail.Info["kernel"] = n
			}
			if n := v.Metric["sysname"]; n != "" {
				detail.Info["os"] = n
			}
			if n := v.Metric["machine"]; n != "" {
				detail.Info["arch"] = n
			}
		}
	}
	// 回退：用 IP 部分作为 NodeName　以免后续 PromQL 拿不到
	if detail.NodeName == "" {
		for i := 0; i < len(instance); i++ {
			if instance[i] == ':' {
				detail.NodeName = instance[:i]
				break
			}
		}
		if detail.NodeName == "" {
			detail.NodeName = instance
		}
	}

	// 2) 当前指标快照
	instFilter := fmt.Sprintf(`instance="%s"`, instance)
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle",%s}[5m])) * 100)`, instFilter)); err == nil {
		detail.Current.CPUUsage = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`100 - (node_memory_MemAvailable_bytes{%s} / node_memory_MemTotal_bytes{%s} * 100)`, instFilter, instFilter)); err == nil {
		detail.Current.MemoryUsage = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`100 - (node_filesystem_avail_bytes{mountpoint="/",%s} / node_filesystem_size_bytes{mountpoint="/",%s} * 100)`, instFilter, instFilter)); err == nil {
		detail.Current.DiskUsage = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`node_load1{%s}`, instFilter)); err == nil {
		detail.Current.Load1 = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`node_load5{%s}`, instFilter)); err == nil {
		detail.Current.Load5 = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`sum(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*",%s}[5m]))`, instFilter)); err == nil {
		detail.Current.NetworkIn = s.extractScalar(r)
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`sum(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*",%s}[5m]))`, instFilter)); err == nil {
		detail.Current.NetworkOut = s.extractScalar(r)
	}

	// 3) 节点容量元信息
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`count(count by(cpu)(node_cpu_seconds_total{%s}))`, instFilter)); err == nil {
		detail.Info["cpu_total"] = fmt.Sprintf("%d 核", int(s.extractScalar(r)))
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`node_memory_MemTotal_bytes{%s}`, instFilter)); err == nil {
		detail.Info["memory_total"] = formatBytesHuman(s.extractScalar(r))
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`node_filesystem_size_bytes{mountpoint="/",%s}`, instFilter)); err == nil {
		detail.Info["disk_total"] = formatBytesHuman(s.extractScalar(r))
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`time() - node_boot_time_seconds{%s}`, instFilter)); err == nil {
		detail.Info["uptime"] = formatUptime(int64(s.extractScalar(r)))
	}
	// kubelet 版本 + role
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`kubernetes_build_info{node="%s"}`, detail.NodeName)); err == nil {
		for _, v := range mustVector(r) {
			if n := v.Metric["git_version"]; n != "" {
				detail.Info["kubelet_version"] = n
			}
		}
	}
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`kube_node_role{node="%s"}`, detail.NodeName)); err == nil {
		roles := []string{}
		for _, v := range mustVector(r) {
			if n := v.Metric["role"]; n != "" {
				roles = append(roles, n)
			}
		}
		if len(roles) > 0 {
			detail.Info["role"] = joinStrings(roles, ",")
		}
	}
	if _, ok := detail.Info["role"]; !ok {
		detail.Info["role"] = "worker"
	}

	// 4) Pod 数（从 kube_pod_info 按 node label 过滤）
	if r, err := client.QueryInstant(ctx, fmt.Sprintf(`count(kube_pod_info{node="%s"})`, detail.NodeName)); err == nil {
		detail.Current.PodCount = int(s.extractScalar(r))
	}

	// 5) 5 维趋势
	end := time.Now()
	start := end.Add(-duration)
	step := duration / 60
	trendQueries := map[string]string{
		"cpu":     fmt.Sprintf(`100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle",%s}[5m])) * 100)`, instFilter),
		"memory":  fmt.Sprintf(`100 - (node_memory_MemAvailable_bytes{%s} / node_memory_MemTotal_bytes{%s} * 100)`, instFilter, instFilter),
		"disk":    fmt.Sprintf(`100 - (node_filesystem_avail_bytes{mountpoint="/",%s} / node_filesystem_size_bytes{mountpoint="/",%s} * 100)`, instFilter, instFilter),
		"net_in":  fmt.Sprintf(`sum(rate(node_network_receive_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*",%s}[5m]))`, instFilter),
		"net_out": fmt.Sprintf(`sum(rate(node_network_transmit_bytes_total{device!~"lo|veth.*|docker.*|br.*|cali.*|cni.*",%s}[5m]))`, instFilter),
	}
	for key, q := range trendQueries {
		r, err := client.QueryRange(ctx, q, start, end, step)
		if err != nil {
			continue
		}
		matrices, _ := prom.ParseMatrixResult(r.Data.Result)
		if len(matrices) == 0 {
			continue
		}
		pts := make([]TrendPoint, 0, len(matrices[0].Values))
		for _, pair := range matrices[0].Values {
			ts, _ := pair[0].(float64)
			pts = append(pts, TrendPoint{Timestamp: int64(ts), Value: s.parseValue(pair[1])})
		}
		detail.Trends[key] = pts
	}

	// 6) 该节点上 Top 10 Pod（CPU）
	topQ := fmt.Sprintf(`topk(10, sum by(pod, namespace) (rate(container_cpu_usage_seconds_total{container!="",pod!=""}[5m]) * on(pod, namespace) group_left() kube_pod_info{node="%s"}))`, detail.NodeName)
	if r, err := client.QueryInstant(ctx, topQ); err == nil {
		cpuMap := map[string]float64{}
		podOrder := []PodMetric{}
		for _, v := range mustVector(r) {
			key := v.Metric["namespace"] + "/" + v.Metric["pod"]
			cpuMap[key] = s.parseValue(v.Value[1])
			podOrder = append(podOrder, PodMetric{
				Name:      v.Metric["pod"],
				Namespace: v.Metric["namespace"],
				CPUUsage:  cpuMap[key],
			})
		}
		// 補上内存
		memQ := fmt.Sprintf(`sum by(pod, namespace) (container_memory_working_set_bytes{container!="",pod!=""} * on(pod, namespace) group_left() kube_pod_info{node="%s"})`, detail.NodeName)
		memMap := map[string]float64{}
		if mr, mErr := client.QueryInstant(ctx, memQ); mErr == nil {
			for _, v := range mustVector(mr) {
				memMap[v.Metric["namespace"]+"/"+v.Metric["pod"]] = s.parseValue(v.Value[1])
			}
		}
		for i := range podOrder {
			podOrder[i].MemoryUsage = memMap[podOrder[i].Namespace+"/"+podOrder[i].Name]
		}
		detail.TopPods = podOrder
	}

	return detail, nil
}

// formatBytesHuman 人读字节数
func formatBytesHuman(b float64) string {
	const k = 1024
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for b >= k && i < len(units)-1 {
		b /= k
		i++
	}
	return fmt.Sprintf("%.1f %s", b, units[i])
}

// formatUptime 秒 → “Xd Yh”
func formatUptime(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	d := sec / 86400
	h := (sec % 86400) / 3600
	if d > 0 {
		return fmt.Sprintf("%dd %dh", d, h)
	}
	m := (sec % 3600) / 60
	return fmt.Sprintf("%dh %dm", h, m)
}

// joinStrings 简易 string slice 拼接（避免额外引入 strings）
func joinStrings(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	out := arr[0]
	for i := 1; i < len(arr); i++ {
		out += sep + arr[i]
	}
	return out
}
