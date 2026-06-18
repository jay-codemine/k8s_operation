package global

import "github.com/redis/go-redis/v9"

var (
	// Redis: go-redis 客户端，用于 Stream/MQ、缓存、分布式锁等
	// 支持单节点 (*redis.Client) 和集群 (*redis.ClusterClient)
	RedisCli redis.UniversalClient
)
