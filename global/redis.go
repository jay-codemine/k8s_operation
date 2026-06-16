package global

import "github.com/redis/go-redis/v9"

var (
	// Redis: go-redis 客户端，用于 Stream/MQ、缓存、分布式锁等
	RedisCli *redis.Client
)
