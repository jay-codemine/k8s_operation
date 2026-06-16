package pod

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodMetrics 表示 Pod 的资源使用情况
type PodMetrics struct {
	PodName        string              `json:"pod_name"`
	Namespace      string              `json:"namespace"`
	Containers     []ContainerMetrics  `json:"containers"`
	TotalCPU       string              `json:"total_cpu"`        // 总 CPU 使用（如 "125m"）
	TotalMemory    string              `json:"total_memory"`     // 总内存使用（如 "256Mi"）
}

// ContainerMetrics 表示单个容器的资源使用情况
type ContainerMetrics struct {
	Name   string `json:"name"`
	CPU    string `json:"cpu"`     // CPU 使用（如 "50m"）
	Memory string `json:"memory"`  // 内存使用（如 "128Mi"）
}

// GetPodMetrics 获取单个 Pod 的资源使用情况
func GetPodMetrics(ctx context.Context, metricsClient *metricsclient.Clientset, namespace, podName string) (*PodMetrics, error) {
	if metricsClient == nil {
		return nil, fmt.Errorf("metrics-server 未安装或不可用")
	}

	// 获取 Pod metrics
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod metrics 失败: %w", err)
	}

	result := &PodMetrics{
		PodName:    podMetrics.Name,
		Namespace:  podMetrics.Namespace,
		Containers: make([]ContainerMetrics, 0, len(podMetrics.Containers)),
	}

	var totalCPU, totalMemory int64

	// 遍历容器指标
	for _, container := range podMetrics.Containers {
		cpu := container.Usage.Cpu().MilliValue()      // CPU 使用（毫核）
		memory := container.Usage.Memory().Value()      // 内存使用（字节）

		totalCPU += cpu
		totalMemory += memory

		result.Containers = append(result.Containers, ContainerMetrics{
			Name:   container.Name,
			CPU:    formatCPU(cpu),
			Memory: formatMemory(memory),
		})
	}

	result.TotalCPU = formatCPU(totalCPU)
	result.TotalMemory = formatMemory(totalMemory)

	return result, nil
}

// GetPodsMetrics 批量获取多个 Pod 的资源使用情况
func GetPodsMetrics(ctx context.Context, metricsClient *metricsclient.Clientset, namespace string) (map[string]*PodMetrics, error) {
	if metricsClient == nil {
		return nil, fmt.Errorf("metrics-server 未安装或不可用")
	}

	// 获取命名空间下所有 Pod 的 metrics
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod metrics 列表失败: %w", err)
	}

	result := make(map[string]*PodMetrics, len(podMetricsList.Items))

	for _, podMetrics := range podMetricsList.Items {
		pm := &PodMetrics{
			PodName:    podMetrics.Name,
			Namespace:  podMetrics.Namespace,
			Containers: make([]ContainerMetrics, 0, len(podMetrics.Containers)),
		}

		var totalCPU, totalMemory int64

		for _, container := range podMetrics.Containers {
			cpu := container.Usage.Cpu().MilliValue()
			memory := container.Usage.Memory().Value()

			totalCPU += cpu
			totalMemory += memory

			pm.Containers = append(pm.Containers, ContainerMetrics{
				Name:   container.Name,
				CPU:    formatCPU(cpu),
				Memory: formatMemory(memory),
			})
		}

		pm.TotalCPU = formatCPU(totalCPU)
		pm.TotalMemory = formatMemory(totalMemory)

		result[podMetrics.Name] = pm
	}

	return result, nil
}

// formatCPU 格式化 CPU 使用量（毫核转为可读格式）
func formatCPU(milliCores int64) string {
	if milliCores >= 1000 {
		return fmt.Sprintf("%.2f", float64(milliCores)/1000.0)
	}
	return fmt.Sprintf("%dm", milliCores)
}

// formatMemory 格式化内存使用量（字节转为可读格式）
func formatMemory(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2fGi", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2fMi", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2fKi", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
