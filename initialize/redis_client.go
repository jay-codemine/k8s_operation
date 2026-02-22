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
	if global.CacheSetting.Address == "" {
		return fmt.Errorf("redis address is empty")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     global.CacheSetting.Address,  // "host:port"
		Username: global.CacheSetting.Username, // ACL 用，不需要可空
		Password: global.CacheSetting.Password,
		DB:       0,

		PoolSize:     global.CacheSetting.MaxConnect, // 连接池大小
		MinIdleConns: 2,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	global.RedisCli = rdb
	return nil
}
