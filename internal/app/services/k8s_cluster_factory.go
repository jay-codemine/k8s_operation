package services

import (
	dm "k8soperation/internal/domain/k8s"
	"k8soperation/global"
)

// ClusterClientFactory 多集群客户端工厂（领域实现在 domain/k8s）
type ClusterClientFactory = dm.ClusterClientFactory

// NewClusterClientFactory 创建工厂（*Services 实现 ClusterClientProvider 接口）
func NewClusterClientFactory(s *Services) *ClusterClientFactory {
	return dm.NewClusterClientFactory(s, global.Logger)
}
