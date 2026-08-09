package adapter

import (
	"context"

	"k8soperation/internal/domain/k8s"
	"k8soperation/internal/domain/rbac"
)

// ClusterListerAdapter 将 k8s.ClusterService 适配为 rbac.ClusterLister 接口
type ClusterListerAdapter struct {
	svc *k8s.ClusterService
}

func NewClusterLister(svc *k8s.ClusterService) rbac.ClusterLister {
	return &ClusterListerAdapter{svc: svc}
}

func (a *ClusterListerAdapter) ListAllClusters(ctx context.Context) ([]rbac.ClusterInfo, error) {
	clusters, _, err := a.svc.List(ctx, "", 1, 1000)
	if err != nil {
		return nil, err
	}
	result := make([]rbac.ClusterInfo, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, rbac.ClusterInfo{
			ID:          c.ID,
			ClusterName: c.ClusterName,
			Status:      c.Status,
		})
	}
	return result, nil
}

func (a *ClusterListerAdapter) ListUserClusters(ctx context.Context, userID int64) ([]rbac.ClusterInfo, error) {
	// 委托给 ClusterService（内部已有权限过滤逻辑）
	clusters, _, err := a.svc.List(ctx, "", 1, 1000)
	if err != nil {
		return nil, err
	}
	_ = userID // RbacService.GetUserAccessibleClusters 自行处理权限逻辑
	result := make([]rbac.ClusterInfo, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, rbac.ClusterInfo{
			ID:          c.ID,
			ClusterName: c.ClusterName,
			Status:      c.Status,
		})
	}
	return result, nil
}
