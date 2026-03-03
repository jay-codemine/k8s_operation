package global

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	// =========================
	// 默认管理集群（系统启动初始化一次）
	// 仅用于：后台任务/系统巡检/无 clusterID 的兼容接口
	// =========================
	ManagementKubeClient       *kubernetes.Clientset
	ManagementSupportsEventsV1 bool
	ManagementMetricsClient    *metricsclient.Clientset
	ManagementKubeConfig       *rest.Config

	// =========================
	// 兼容旧代码（临时保留）
	// 多集群模式下：这些变量永远指向 Management 集群
	// 新业务代码禁止直接使用，统一走请求上下文里的 k8s_clients
	// =========================
	KubeClient       *kubernetes.Clientset
	SupportsEventsV1 bool
	MetricsClient    *metricsclient.Clientset
	KubeConfig       *rest.Config
)
