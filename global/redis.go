package global

import "github.com/redis/go-redis/v9"

var (
	// Redis: go-redis 客户端，用于 Stream/MQ、缓存、分布式锁等
	// 支持单节点 (*redis.Client) 和集群 (*redis.ClusterClient)
	RedisCli redis.UniversalClient
)

// OnlineUsersKey Redis ZSET key：member=用户ID，score=最近活跃时间(unix秒)
// auth 中间件写入，平台健康页用于统计"当前在线用户数"
const OnlineUsersKey = "platform:online_users"
