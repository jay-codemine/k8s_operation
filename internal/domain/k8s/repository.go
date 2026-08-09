package k8s

import "context"

// ClusterRepository 集群仓储接口
type ClusterRepository interface {
	Save(ctx context.Context, cluster *Cluster) error
	Update(ctx context.Context, id uint32, values map[string]interface{}) error
	SoftDelete(ctx context.Context, id uint32) error
	BatchDelete(ctx context.Context, ids []uint32) (int64, error)
	FindByID(ctx context.Context, id uint32) (*Cluster, error)
	FindByName(ctx context.Context, name string) (*Cluster, error)
	Query(ctx context.Context, name string, page, limit int) ([]*Cluster, int64, error)
	UpdateHealth(ctx context.Context, id uint32, status uint8, lastErr string) error
}
