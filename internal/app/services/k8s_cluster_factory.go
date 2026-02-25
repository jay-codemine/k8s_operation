package services

import (
	"context"
	"fmt"
	"k8soperation/internal/app/requests"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type cachedClients struct {
	clients   *K8sClients
	version   int64
	createdAt time.Time
	expiresAt time.Time
}

type ClusterClientFactory struct {
	s *Services

	mu sync.RWMutex
	m  map[uint32]*cachedClients
	g  singleflight.Group

	baseTTL     time.Duration
	jitterRange time.Duration
}

func NewClusterClientFactory(s *Services) *ClusterClientFactory {
	return &ClusterClientFactory{
		s:           s,
		m:           make(map[uint32]*cachedClients),
		baseTTL:     30 * time.Minute,
		jitterRange: 3 * time.Minute,
	}
}

func (f *ClusterClientFactory) randJitter() time.Duration {
	if f.jitterRange <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(f.jitterRange)))
}

func (f *ClusterClientFactory) Get(ctx context.Context, clusterID uint32) (*K8sClients, error) {
	// 1) 先查 DB 拿版本
	cluster, err := f.s.dao.KubeClusterGetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	ver := int64(cluster.ModifiedAt)

	// 2) 缓存命中（版本一致 + 未过期）
	now := time.Now()
	f.mu.RLock()
	if c, ok := f.m[clusterID]; ok &&
		c.version == ver &&
		now.Before(c.expiresAt) {

		cli := c.clients
		f.mu.RUnlock()

		// 建议加 debug log：cache hit
		fmt.Sprintf("cluster client cache hit", "cluster_id", clusterID, "ver", ver, "client", fmt.Sprintf("%p", cli))
		return cli, nil
	}
	f.mu.RUnlock()

	// 3) singleflight：key 必须带版本，避免更新后仍复用旧初始化结果
	key := fmt.Sprintf("%d:%d", clusterID, ver)

	v, err, _ := f.g.Do(key, func() (any, error) {
		// 在 singleflight 临界区内重新读取 DB，
		// 以数据库的最新状态作为最终一致性来源，
		// 避免并发更新 kubeconfig 时产生缓存与实际配置不一致的问题
		latest, e := f.s.dao.KubeClusterGetByID(ctx, clusterID)
		if e != nil {
			return nil, e
		}
		latestVer := int64(latest.ModifiedAt)

		// latestVer 表示当前数据库中集群记录的最新版本（如 modified_at）
		// useVer 表示本次初始化并写入缓存的 client 所绑定的版本快照
		useVer := latestVer

		cli, e := f.s.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: clusterID})
		if e != nil {
			// 初始化失败：保险起见驱逐（防止坏对象残留）
			f.Invalidate(clusterID)
			return nil, e
		}

		exp := time.Now().Add(f.baseTTL + f.randJitter())

		f.mu.Lock()
		f.m[clusterID] = &cachedClients{
			clients:   cli,
			version:   useVer,
			createdAt: time.Now(),
			expiresAt: exp,
		}
		f.mu.Unlock()

		return cli, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*K8sClients), nil
}

func (f *ClusterClientFactory) Invalidate(clusterID uint32) {
	f.mu.Lock()
	delete(f.m, clusterID)
	f.mu.Unlock()
}

// GetClient 获取集群客户端（支持 int64 类型的 clusterID）
func (f *ClusterClientFactory) GetClient(ctx context.Context, clusterID int64) (*K8sClients, error) {
	return f.Get(ctx, uint32(clusterID))
}
