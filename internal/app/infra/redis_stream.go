package infra

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// CicdDeployStream CICD 发布任务消息队列
	CicdDeployStream = "cicd:deploy:stream"
	// CicdDeployGroup 消费者组名称
	CicdDeployGroup = "cicd-deploy-workers"
)

type RedisStream struct {
	rdb *redis.Client
}

func NewRedisStream(rdb *redis.Client) *RedisStream {
	return &RedisStream{rdb: rdb}
}

// XAdd 往 Redis Stream 写入一条消息，返回 message_id
func (s *RedisStream) XAdd(ctx context.Context, stream string, values map[string]any) (string, error) {
	return s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
		MaxLen: 10000, // 限制 stream 的大小，超过 10000 条消息就删除最旧的消息
	}).Result()
}

// CreateGroup 创建消费者组（如果不存在）
// startID: "$" 表示只消费新消息，"0" 表示从头开始消费
func (s *RedisStream) CreateGroup(ctx context.Context, stream, group, startID string) error {
	// MKSTREAM 开启：如果 stream 不存在则自动创建
	err := s.rdb.XGroupCreateMkStream(ctx, stream, group, startID).Err()
	if err != nil {
		// 如果组已存在，忽略错误
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil
		}
		return err
	}
	return nil
}

// XReadGroup 从消费者组读取消息
// consumer: 消费者名称（通常使用 hostname 或 worker-id）
// count: 一次最多读取条数
// block: 阻塞等待时间，0 表示永久阻塞
func (s *RedisStream) XReadGroup(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	return s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"}, // ">" 表示只读取新消息
		Count:    count,
		Block:    block,
		NoAck:    false, // 需要手动 ACK
	}).Result()
}

// XReadGroupPending 读取待处理的消息（已消费但未 ACK）
func (s *RedisStream) XReadGroupPending(ctx context.Context, stream, group, consumer string, count int64) ([]redis.XStream, error) {
	return s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, "0"}, // "0" 表示读取待处理消息
		Count:    count,
		Block:    0, // 不阻塞
		NoAck:    false,
	}).Result()
}

// XAck 确认消息已处理
func (s *RedisStream) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	return s.rdb.XAck(ctx, stream, group, ids...).Result()
}

// XPending 获取待处理消息信息
func (s *RedisStream) XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	return s.rdb.XPending(ctx, stream, group).Result()
}

// XClaim 认领超时消息（用于处理挂掉的 worker 未 ACK 的消息）
func (s *RedisStream) XClaim(ctx context.Context, stream, group, consumer string, minIdleTime time.Duration, ids ...string) ([]redis.XMessage, error) {
	return s.rdb.XClaim(ctx, &redis.XClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdleTime,
		Messages: ids,
	}).Result()
}

// XLen 获取 Stream 长度
func (s *RedisStream) XLen(ctx context.Context, stream string) (int64, error) {
	return s.rdb.XLen(ctx, stream).Result()
}
