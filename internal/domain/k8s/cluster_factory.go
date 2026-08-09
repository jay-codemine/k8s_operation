package k8s

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"k8soperation/pkg/logger"
)

// 连接超时配置
const (
	DefaultConnectTimeout = 15 * time.Second
	MaxConnectTimeout     = 30 * time.Second
	DefaultFailureTTL     = 20 * time.Second
)

type cachedClients struct {
	clients   *K8sClients
	version   int64
	createdAt time.Time
	expiresAt time.Time
}

type failureRecord struct {
	at      time.Time
	version int64
}

// ClusterClientFactory 多集群客户端工厂
type ClusterClientFactory struct {
	provider ClusterClientProvider

	mu sync.RWMutex
	m  map[uint32]*cachedClients
	g  singleflight.Group

	failures map[uint32]failureRecord

	baseTTL        time.Duration
	jitterRange    time.Duration
	connectTimeout time.Duration
	failureTTL     time.Duration

	logger *logger.Logger
}

// NewClusterClientFactory 创建工厂
func NewClusterClientFactory(provider ClusterClientProvider, logger *logger.Logger) *ClusterClientFactory {
	return &ClusterClientFactory{
		provider:       provider,
		m:              make(map[uint32]*cachedClients),
		failures:       make(map[uint32]failureRecord),
		baseTTL:        30 * time.Minute,
		jitterRange:    3 * time.Minute,
		connectTimeout: DefaultConnectTimeout,
		failureTTL:     DefaultFailureTTL,
		logger:         logger,
	}
}

// SetConnectTimeout 设置连接超时
func (f *ClusterClientFactory) SetConnectTimeout(timeout time.Duration) {
	if timeout > 0 && timeout <= MaxConnectTimeout {
		f.connectTimeout = timeout
	}
}

// SetFailureTTL 设置失败负缓存时长
func (f *ClusterClientFactory) SetFailureTTL(ttl time.Duration) {
	f.mu.Lock()
	f.failureTTL = ttl
	f.mu.Unlock()
}

// ResetFailure 清除失败标记
func (f *ClusterClientFactory) ResetFailure(clusterID uint32) {
	f.clearFailure(clusterID)
}

func (f *ClusterClientFactory) markFailure(clusterID uint32, version int64) {
	f.mu.Lock()
	if f.failures == nil {
		f.failures = make(map[uint32]failureRecord)
	}
	f.failures[clusterID] = failureRecord{at: time.Now(), version: version}
	f.mu.Unlock()
}

func (f *ClusterClientFactory) clearFailure(clusterID uint32) {
	f.mu.Lock()
	delete(f.failures, clusterID)
	f.mu.Unlock()
}

func (f *ClusterClientFactory) inFailureWindow(clusterID uint32, version int64) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.failureTTL <= 0 {
		return false
	}
	rec, ok := f.failures[clusterID]
	if !ok || rec.version != version {
		return false
	}
	return time.Since(rec.at) < f.failureTTL
}

func (f *ClusterClientFactory) randJitter() time.Duration {
	if f.jitterRange <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(f.jitterRange)))
}

// Get 获取集群客户端
func (f *ClusterClientFactory) Get(ctx context.Context, clusterID uint32) (*K8sClients, error) {
	cluster, err := f.provider.GetCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	ver := int64(cluster.ModifiedAt)

	// 缓存命中
	now := time.Now()
	f.mu.RLock()
	if c, ok := f.m[clusterID]; ok &&
		c.version == ver &&
		now.Before(c.expiresAt) {
		cli := c.clients
		f.mu.RUnlock()
		return cli, nil
	}
	f.mu.RUnlock()

	// 失败负缓存
	if f.inFailureWindow(clusterID, ver) {
		return nil, fmt.Errorf("cluster %d recently unreachable, skipped within failure window", clusterID)
	}

	// 设置连接超时
	connectCtx, cancel := context.WithTimeout(ctx, f.connectTimeout)
	defer cancel()

	// singleflight 防止雷群效应
	key := fmt.Sprintf("%d:%d", clusterID, ver)

	type result struct {
		clients *K8sClients
		err     error
	}
	resultCh := make(chan result, 1)

	go func() {
		v, err, _ := f.g.Do(key, func() (any, error) {
			latest, e := f.provider.GetCluster(connectCtx, clusterID)
			if e != nil {
				return nil, e
			}
			latestVer := int64(latest.ModifiedAt)

			cli, e := f.provider.BuildClientsForCluster(connectCtx, clusterID)
			if e != nil {
				f.Invalidate(clusterID)
				f.markFailure(clusterID, latestVer)
				return nil, e
			}

			exp := time.Now().Add(f.baseTTL + f.randJitter())

			f.mu.Lock()
			f.m[clusterID] = &cachedClients{
				clients:   cli,
				version:   latestVer,
				createdAt: time.Now(),
				expiresAt: exp,
			}
			delete(f.failures, clusterID)
			f.mu.Unlock()

			return cli, nil
		})

		if err != nil {
			resultCh <- result{nil, err}
			return
		}
		resultCh <- result{v.(*K8sClients), nil}
	}()

	select {
	case <-connectCtx.Done():
		f.markFailure(clusterID, ver)
		f.logger.Warn("集群连接超时，跳过该集群",
			zap.Uint32("cluster_id", clusterID),
			zap.Duration("timeout", f.connectTimeout))
		return nil, fmt.Errorf("cluster %d connection timeout after %v", clusterID, f.connectTimeout)
	case r := <-resultCh:
		return r.clients, r.err
	}
}

// Invalidate 驱逐缓存
func (f *ClusterClientFactory) Invalidate(clusterID uint32) {
	f.mu.Lock()
	delete(f.m, clusterID)
	f.mu.Unlock()
}

// GetClient 获取集群客户端（int64 类型 clusterID）
func (f *ClusterClientFactory) GetClient(ctx context.Context, clusterID int64) (*K8sClients, error) {
	return f.Get(ctx, uint32(clusterID))
}
