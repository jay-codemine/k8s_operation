package services

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// 聚合一套与某个集群绑定的客户端
type K8sClients struct {
	Config       *rest.Config             // 同源配置（给其它 client 复用）
	Kube         *kubernetes.Clientset    // Core/Apps/Batch… 客户端
	Metrics      *metricsclient.Clientset // metrics.k8s.io 客户端（可能为 nil）
	SupportsEvV1 bool                     // 是否支持 events.k8s.io/v1
}
