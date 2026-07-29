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
	DefaultConnectTimeout = 15 * time.Second

	// MaxConnectTimeout 最大连接超时时间
	MaxConnectTimeout = 30 * time.Second

	// DefaultFailureTTL 连接失败负缓存时长
	// 作用：某集群连接失败（超时/不可达）后，在该时长内对同一集群的请求直接快速失败，
	// 避免聚合类接口（如平台健康）在一次请求中对同一离线集群反复等待完整连接超时。
	// 集群配置变更或手动 Invalidate 时会立即清除失败标记。
	DefaultFailureTTL = 20 * time.Second
)

type cachedClients struct {
	clients   *K8sClients
	version   int64
	createdAt time.Time
	expiresAt time.Time
}

// failureRecord 连接失败记录（负缓存）
// version 为失败时集群的配置版本（modified_at），集群配置变更后版本不再匹配，
// 负缓存自然失效，无需外部手动清理。
type failureRecord struct {
	at      time.Time
	version int64
}

type ClusterClientFactory struct {
	s *Services

	mu sync.RWMutex
	m  map[uint32]*cachedClients
	g  singleflight.Group

	// failures 记录连接失败的集群（负缓存），键为 clusterID
	failures map[uint32]failureRecord

	baseTTL        time.Duration
	jitterRange    time.Duration
	connectTimeout time.Duration // 连接超时时间
	failureTTL     time.Duration // 失败负缓存时长
}

func NewClusterClientFactory(s *Services) *ClusterClientFactory {
	return &ClusterClientFactory{
		s:              s,
		m:              make(map[uint32]*cachedClients),
		failures:       make(map[uint32]failureRecord),
		baseTTL:        30 * time.Minute,
		jitterRange:    3 * time.Minute,
		connectTimeout: DefaultConnectTimeout,
		failureTTL:     DefaultFailureTTL,
	}
}

// SetConnectTimeout 设置连接超时时间
func (f *ClusterClientFactory) SetConnectTimeout(timeout time.Duration) {
	if timeout > 0 && timeout <= MaxConnectTimeout {
		f.connectTimeout = timeout
	}
}

// SetFailureTTL 设置连接失败负缓存时长（<=0 表示关闭负缓存）
func (f *ClusterClientFactory) SetFailureTTL(ttl time.Duration) {
	f.mu.Lock()
	f.failureTTL = ttl
	f.mu.Unlock()
}

// ResetFailure 清除集群失败标记（用于集群配置变更、用户主动重试/测试连接等需要绕过负缓存的场景）
func (f *ClusterClientFactory) ResetFailure(clusterID uint32) {
	f.clearFailure(clusterID)
}

// markFailure 记录集群连接失败（带配置版本）
func (f *ClusterClientFactory) markFailure(clusterID uint32, version int64) {
	f.mu.Lock()
	if f.failures == nil {
		f.failures = make(map[uint32]failureRecord)
	}
	f.failures[clusterID] = failureRecord{at: time.Now(), version: version}
	f.mu.Unlock()
}

// clearFailure 清除集群失败标记
func (f *ClusterClientFactory) clearFailure(clusterID uint32) {
	f.mu.Lock()
	delete(f.failures, clusterID)
	f.mu.Unlock()
}

// inFailureWindow 判断集群是否处于失败负缓存窗口内（仅当配置版本一致时生效）
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

	// 2.5) 失败负缓存：该集群（同一配置版本）近期连接失败过，直接快速失败，避免重复等待完整超时
	if f.inFailureWindow(clusterID, ver) {
		return nil, fmt.Errorf("cluster %d recently unreachable, skipped within failure window", clusterID)
	}

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
				// 初始化失败：驱逐缓存并记录失败（负缓存）
				f.Invalidate(clusterID)
				f.markFailure(clusterID, useVer)
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
			delete(f.failures, clusterID) // 连接成功：清除失败标记
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
		// 连接超时：记录失败（负缓存）并快速失败
		f.markFailure(clusterID, ver)
		global.Logger.Warn("集群连接超时，跳过该集群",
			zap.Uint32("cluster_id", clusterID),
			zap.Duration("timeout", f.connectTimeout))
		return nil, fmt.Errorf("cluster %d connection timeout after %v", clusterID, f.connectTimeout)
	case r := <-resultCh:
		return r.clients, r.err
	}
}

// Invalidate 驱逐指定集群的客户端缓存
// 注意：不清除失败标记，以保留负缓存的快速失败能力；
// 需要立即重试时请配合调用 ResetFailure。
func (f *ClusterClientFactory) Invalidate(clusterID uint32) {
	f.mu.Lock()
	delete(f.m, clusterID)
	f.mu.Unlock()
}

// GetClient 获取集群客户端（支持 int64 类型的 clusterID）
func (f *ClusterClientFactory) GetClient(ctx context.Context, clusterID int64) (*K8sClients, error) {
	return f.Get(ctx, uint32(clusterID))
}
