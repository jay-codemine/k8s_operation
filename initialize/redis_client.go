package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"k8soperation/global"
)

func SetupRedis() error {
	if global.CacheSetting == nil {
		return fmt.Errorf("CacheSetting is nil")
	}

	// 判断是否为 Cluster 模式（Addresses 非空则使用 Cluster）
	if len(global.CacheSetting.Addresses) > 0 {
		// Redis Cluster 模式
		clusterCli := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    global.CacheSetting.Addresses,
			Password: global.CacheSetting.Password,
			Username: global.CacheSetting.Username,

			PoolSize:     global.CacheSetting.MaxConnect,
			MinIdleConns: 2,

			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if pingErr := clusterCli.Ping(ctx).Err(); pingErr != nil {
			return fmt.Errorf("redis cluster ping failed: %w", pingErr)
		}

		// Cluster 模式直接使用 ClusterClient，支持自动 MOVED/ASK 重定向
		global.RedisCli = clusterCli
		fmt.Printf("[Redis] Cluster mode enabled, nodes: %v\n", global.CacheSetting.Addresses)
	} else {
		// 单节点模式
		if global.CacheSetting.Address == "" {
			return fmt.Errorf("redis address is empty")
		}

		rdb := redis.NewClient(&redis.Options{
			Addr:     global.CacheSetting.Address,
			Username: global.CacheSetting.Username,
			Password: global.CacheSetting.Password,
			DB:       0,

			PoolSize:     global.CacheSetting.MaxConnect,
			MinIdleConns: 2,

			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if pingErr := rdb.Ping(ctx).Err(); pingErr != nil {
			return fmt.Errorf("redis ping failed: %w", pingErr)
		}

		global.RedisCli = rdb
	}

	return nil
}
