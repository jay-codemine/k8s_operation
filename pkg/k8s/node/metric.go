package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"time"
)

// GetNodeMetrics 获取节点指标，nodeName 为空时返回全量节点指标
func GetNodeMetrics(ctx context.Context, Kube kubernetes.Interface, metrics metricsclient.Interface, nodeName string) ([]models.NodeMetricItem, error) {
	// 如果指定了节点名，只获取单个节点的指标
	if nodeName != "" {
		nm, err := metrics.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("metrics API error: %w (metrics-server installed?)", err)
		}

		node, err := Kube.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			global.Logger.Warn("get node capacity failed", zap.String("nodeName", nodeName), zap.Error(err))
			node = &corev1.Node{}
		}

		item := toMetricItem(nm, node)
		return []models.NodeMetricItem{item}, nil
	}

	// nodeName 为空时，获取全量节点指标
	nmList, err := metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("metrics API error: %w (metrics-server installed?)", err)
	}

	// 获取所有节点信息
	nodeList, err := Kube.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		global.Logger.Warn("get nodes failed", zap.Error(err))
		nodeList = &corev1.NodeList{}
	}

	// 构建节点 map 方便查找
	nodeMap := make(map[string]*corev1.Node)
	for i := range nodeList.Items {
		nodeMap[nodeList.Items[i].Name] = &nodeList.Items[i]
	}

	// 组装结果
	result := make([]models.NodeMetricItem, 0, len(nmList.Items))
	for i := range nmList.Items {
		nm := &nmList.Items[i]
		node := nodeMap[nm.Name]
		if node == nil {
			node = &corev1.Node{}
		}
		item := toMetricItem(nm, node)
		result = append(result, item)
	}

	return result, nil
}

func toMetricItem(nm *metricsv1beta1.NodeMetrics, node *corev1.Node) models.NodeMetricItem {
	// 获取时间戳和窗口时长（秒）
	ts := nm.Timestamp.Time
	windowSeconds := int64(nm.Window.Duration / time.Second)

	// 初始化CPU使用量（毫秒）和内存使用量（字节）为0
	cpuMilli := int64(0)
	memBytes := int64(0)
	// 检查CPU使用量是否存在，如果存在则获取其毫秒值
	if q := nm.Usage.Cpu(); q != nil {
		cpuMilli = q.MilliValue()
	}
	// 检查内存使用量是否存在，如果存在则获取其值
	if q := nm.Usage.Memory(); q != nil {
		memBytes = q.Value()
	}

	// 初始化容量相关的CPU和内存值为0
	capMilli, capMem := int64(0), int64(0)
	// 初始化分配相关的CPU和内存值为0
	allocMilli, allocMem := int64(0), int64(0)
	// 判断节点是否为空
	if node != nil {
		// 获取节点的CPU容量，如果存在则转换为毫秒值
		if q := node.Status.Capacity.Cpu(); q != nil {
			capMilli = q.MilliValue()
		}
		// 获取节点的内存容量，如果存在则获取其值
		if q := node.Status.Capacity.Memory(); q != nil {
			capMem = q.Value()
		}
		// 获取节点的可分配CPU资源，如果存在则转换为毫秒值
		if q := node.Status.Allocatable.Cpu(); q != nil {
			allocMilli = q.MilliValue()
		}
		// 获取节点的可分配内存资源，如果存在则获取其值
		if q := node.Status.Allocatable.Memory(); q != nil {
			allocMem = q.Value()
		}
	}

	/*
	 * 创建并返回一个NodeMetricItem结构体实例，该结构体包含了节点的各项指标数据
	 * 包括名称、时间戳、窗口时间以及CPU和内存的使用量和限制值
	 */
	m := models.NodeMetricItem{
		Name:          nm.Name,       // 节点名称
		Timestamp:     ts,            // 时间戳
		WindowSeconds: windowSeconds, // 窗口时间（秒）
		CPUUsageMilli: cpuMilli,      // CPU使用量（毫核）
		MemUsageBytes: memBytes,      // 内存使用量（字节）
		CPUCapMilli:   capMilli,      // CPU限制值（毫核）
		MemCapBytes:   capMem,        // 内存限制值（字节）
		CPUAllocMilli: allocMilli,    // CPU分配量（毫核）
		MemAllocBytes: allocMem,      // 内存分配量（字节）
	}
	// 计算CPU使用百分比
	// 首先确定分母，优先使用分配量，如果分配量为0则使用限制值
	denomCPU := allocMilli
	if denomCPU == 0 {
		denomCPU = capMilli
	}
	// 如果分母大于0，则计算CPU使用百分比
	if denomCPU > 0 {
		m.CPUUsagePercent = float64(cpuMilli) * 100 / float64(denomCPU)
	}
	// 计算内存使用百分比
	// 首先确定分母，优先使用分配量，如果分配量为0则使用限制值
	denomMem := allocMem
	if denomMem == 0 {
		denomMem = capMem
	}
	// 如果分母大于0，则计算内存使用百分比
	if denomMem > 0 {
		m.MemUsagePercent = float64(memBytes) * 100 / float64(denomMem)
	}
	// 返回填充好的NodeMetricItem结构体实例
	return m
}
