package rbac

import "context"

// ——— 跨域接口（依赖倒置：rbac 域定义契约，外部实现）———

// ClusterLister 集群列表查询器（由 k8s 域实现）
type ClusterLister interface {
	ListAllClusters(ctx context.Context) ([]ClusterInfo, error)
	ListUserClusters(ctx context.Context, userID int64) ([]ClusterInfo, error)
}

// ClusterInfo 跨域传递用的集群摘要（避免直接依赖 k8s.Cluster）
type ClusterInfo struct {
	ID          uint32
	ClusterName string
	Status      uint8
}

// UserLookup 用户查询器（由 user 域实现）
type UserLookup interface {
	FindUsername(ctx context.Context, userID int64) (string, error)
	ListUsersByRole(ctx context.Context, roleID int64) ([]UserInfo, error)
}

// UserInfo 跨域传递用的用户摘要
type UserInfo struct {
	ID       int64
	Username string
}

// ——— 当前实现方式 ———
// 由于 k8s、user 域尚未暴露这些接口，RbacService 暂时通过 s.db 直接 JOIN 查询
// kube_cluster / user 表。待各域提供 ClusterLister / UserLookup 实现后，替换 s.db 调用。
