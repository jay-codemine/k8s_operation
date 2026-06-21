// Package cache 提供基于 Redis 的 Jenkins API 结果缓存
// 目的：减少对 Jenkins 的高频实时轮询，解决 PipelineStatus/PipelineStages
// 接口 P99 延迟高达 800ms+ 的问题。
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8soperation/global"
)

const (
	// running 状态缓存时间短（构建状态变化快）
	ttlRunning = 5 * time.Second
	// 已完成状态缓存时间长（结果不再变化）
	ttlFinished = 60 * time.Second
	// stages 缓存时间
	ttlStagesRunning  = 3 * time.Second
	ttlStagesFinished = 60 * time.Second
)

// buildInfoKey 生成 BuildInfo 的 Redis 缓存 key
func buildInfoKey(pipelineID int64, buildNumber int) string {
	return fmt.Sprintf("cache:jenkins:build:%d:%d", pipelineID, buildNumber)
}

// stagesKey 生成 Stages 的 Redis 缓存 key
func stagesKey(pipelineID int64, buildNumber int) string {
	return fmt.Sprintf("cache:jenkins:stages:%d:%d", pipelineID, buildNumber)
}

// GetBuildInfo 从 Redis 获取缓存的 BuildInfo（未命中返回 nil, nil）
func GetBuildInfo(ctx context.Context, pipelineID int64, buildNumber int, dest interface{}) error {
	if global.RedisCli == nil {
		return nil
	}
	key := buildInfoKey(pipelineID, buildNumber)
	data, err := global.RedisCli.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil 表示 key 不存在，不是真正的错误
		return nil
	}
	return json.Unmarshal(data, dest)
}

// SetBuildInfo 将 BuildInfo 存入 Redis
// isBuilding=true 时用短 TTL（5s），否则用长 TTL（60s）
func SetBuildInfo(ctx context.Context, pipelineID int64, buildNumber int, v interface{}, isBuilding bool) {
	if global.RedisCli == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	ttl := ttlFinished
	if isBuilding {
		ttl = ttlRunning
	}
	key := buildInfoKey(pipelineID, buildNumber)
	// 忽略缓存写入错误，不影响主流程
	_ = global.RedisCli.Set(ctx, key, data, ttl).Err()
}

// GetStages 从 Redis 获取缓存的阶段数据
func GetStages(ctx context.Context, pipelineID int64, buildNumber int, dest interface{}) error {
	if global.RedisCli == nil {
		return nil
	}
	key := stagesKey(pipelineID, buildNumber)
	data, err := global.RedisCli.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	return json.Unmarshal(data, dest)
}

// SetStages 将阶段数据存入 Redis
func SetStages(ctx context.Context, pipelineID int64, buildNumber int, v interface{}, isRunning bool) {
	if global.RedisCli == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	ttl := ttlStagesFinished
	if isRunning {
		ttl = ttlStagesRunning
	}
	key := stagesKey(pipelineID, buildNumber)
	_ = global.RedisCli.Set(ctx, key, data, ttl).Err()
}

// InvalidatePipeline 流水线状态变更时主动清除缓存
// 在回调、停止等改变构建状态的操作后调用
func InvalidatePipeline(ctx context.Context, pipelineID int64, buildNumber int) {
	if global.RedisCli == nil || buildNumber <= 0 {
		return
	}
	_ = global.RedisCli.Del(ctx,
		buildInfoKey(pipelineID, buildNumber),
		stagesKey(pipelineID, buildNumber),
	).Err()
}
