package services

import (
	"context"
	"fmt"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// =============================================================================
// ClusterClientFactory - 多集群客户端工厂（大厂设计模式）
// 
// 设计原则：
// 1. 快速失败：单个集群连接超时不影响其他集群
// 2. 缓存优先：命中缓存直接返回，避免重复连接
// 3. 并发安全：singleflight 防止雷群效应
// 4. 优雅降级：连接失败时记录状态，不阻塞业务
// =============================================================================

// 连接超时配置（参考 Rancher/KubeSphere 设计）
const (
	// DefaultConnectTimeout 默认连接超时时间
	// 生产环境建议 5-10 秒，避免单个集群拖慢整体
	DefaultConnectTimeout = 5 * time.Second

	// MaxConnectTimeout 最大连接超时时间
	MaxConnectTimeout = 30 * time.Second
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

	baseTTL        time.Duration
	jitterRange    time.Duration
	connectTimeout time.Duration // 连接超时时间
}

func NewClusterClientFactory(s *Services) *ClusterClientFactory {
	return &ClusterClientFactory{
		s:              s,
		m:              make(map[uint32]*cachedClients),
		baseTTL:        30 * time.Minute,
		jitterRange:    3 * time.Minute,
		connectTimeout: DefaultConnectTimeout,
	}
}

// SetConnectTimeout 设置连接超时时间
func (f *ClusterClientFactory) SetConnectTimeout(timeout time.Duration) {
	if timeout > 0 && timeout <= MaxConnectTimeout {
		f.connectTimeout = timeout
	}
}

func (f *ClusterClientFactory) randJitter() time.Duration {
	if f.jitterRange <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(f.jitterRange)))
}

// Get 获取集群客户端（带超时控制）
// 
// 设计要点（参考 Rancher/KubeSphere）：
// 1. 缓存优先：命中直接返回，延迟 < 1ms
// 2. 超时快速失败：连接超时立即返回错误，不阻塞其他集群
// 3. singleflight：并发请求合并，防止雷群效应
func (f *ClusterClientFactory) Get(ctx context.Context, clusterID uint32) (*K8sClients, error) {
	// 1) 先查 DB 拿版本
	cluster, err := f.s.dao.KubeClusterGetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	ver := int64(cluster.ModifiedAt)

	// 2) 缓存命中（版本一致 + 未过期）→ 快速返回
	now := time.Now()
	f.mu.RLock()
	if c, ok := f.m[clusterID]; ok &&
		c.version == ver &&
		now.Before(c.expiresAt) {

		cli := c.clients
		f.mu.RUnlock()

		// cache hit - 延迟 < 1ms
		return cli, nil
	}
	f.mu.RUnlock()

	// 3) 缓存未命中：需要初始化连接
	// 设置连接超时，避免单个集群阻塞整体
	connectCtx, cancel := context.WithTimeout(ctx, f.connectTimeout)
	defer cancel()

	// 4) singleflight：key 必须带版本，避免更新后仍复用旧初始化结果
	key := fmt.Sprintf("%d:%d", clusterID, ver)

	// 使用 channel 实现超时控制（singleflight 不直接支持 context）
	type result struct {
		clients *K8sClients
		err     error
	}
	resultCh := make(chan result, 1)

	go func() {
		v, err, _ := f.g.Do(key, func() (any, error) {
			// 在 singleflight 临界区内重新读取 DB，
			// 以数据库的最新状态作为最终一致性来源
			latest, e := f.s.dao.KubeClusterGetByID(connectCtx, clusterID)
			if e != nil {
				return nil, e
			}
			latestVer := int64(latest.ModifiedAt)
			useVer := latestVer

			// 初始化集群客户端（带超时）
			cli, e := f.s.K8sClusterInit(connectCtx, &requests.K8sClusterInitRequest{ID: clusterID})
			if e != nil {
				// 初始化失败：驱逐缓存
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
			resultCh <- result{nil, err}
			return
		}
		resultCh <- result{v.(*K8sClients), nil}
	}()

	// 5) 等待结果或超时
	select {
	case <-connectCtx.Done():
		// 连接超时：记录日志，快速失败
		global.Logger.Warn("集群连接超时，跳过该集群",
			zap.Uint32("cluster_id", clusterID),
			zap.Duration("timeout", f.connectTimeout))
		return nil, fmt.Errorf("cluster %d connection timeout after %v", clusterID, f.connectTimeout)
	case r := <-resultCh:
		return r.clients, r.err
	}
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
