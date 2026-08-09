package k8s

import "context"

// ClusterClientProvider 为 ClusterClientFactory 提供集群数据和客户端初始化能力。
// 由外层（services 包）实现，注入到 domain 层，避免 domain 直接依赖 DAO/DB。
type ClusterClientProvider interface {
	// GetCluster 获取集群信息（至少包含 ModifiedAt，用于缓存版本校验）
	GetCluster(ctx context.Context, clusterID uint32) (*Cluster, error)

	// BuildClientsForCluster 为指定集群构建 K8s 客户端集
	BuildClientsForCluster(ctx context.Context, clusterID uint32) (*K8sClients, error)
}
