package worker

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/infra"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
)

// CicdWorker CICD 部署任务消费者
type CicdWorker struct {
	stream   *infra.RedisStream
	dao      *dao.Dao
	executor *services.CicdTaskExecutor
	svc      *services.Services
	callback *CicdCallback

	consumerName string
	concurrency  int
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewCicdWorker 创建 Worker
func NewCicdWorker(rdb *redis.Client, factory *services.ClusterClientFactory) *CicdWorker {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "worker"
	}

	return &CicdWorker{
		stream:       infra.NewRedisStream(rdb),
		dao:          dao.NewDao(global.DB),
		executor:     services.NewCicdTaskExecutor(factory),
		svc:          services.NewServices(),
		callback:     NewCicdCallback(),
		consumerName: hostname + "-" + strconv.FormatInt(time.Now().Unix(), 10),
		concurrency:  3, // 并发处理任务数
		stopCh:       make(chan struct{}),
	}
}

// Start 启动 Worker
func (w *CicdWorker) Start(ctx context.Context) error {
	// 1. 创建消费者组（如果不存在）
	if err := w.stream.CreateGroup(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, "0"); err != nil {
		global.Logger.Error("create consumer group failed", zap.Error(err))
		return err
	}

	global.Logger.Info("cicd worker started",
		zap.String("consumer", w.consumerName),
		zap.Int("concurrency", w.concurrency))

	// 2. 先处理待处理的消息（已消费但未 ACK）
	go w.processPendingMessages(ctx)

	// 3. 启动消费协程
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.consumeLoop(ctx, i)
	}

	return nil
}

// Stop 停止 Worker
func (w *CicdWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	global.Logger.Info("cicd worker stopped")
}

// consumeLoop 消费循环
func (w *CicdWorker) consumeLoop(ctx context.Context, workerID int) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			w.consumeOnce(ctx, workerID)
		}
	}
}

// consumeOnce 消费一次
func (w *CicdWorker) consumeOnce(ctx context.Context, workerID int) {
	// 从 Redis Stream 读取消息
	streams, err := w.stream.XReadGroup(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, w.consumerName, 1, 5*time.Second)
	if err != nil {
		if err == redis.Nil {
			return // 没有新消息
		}
		global.Logger.Error("xreadgroup failed", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			w.processMessage(ctx, msg)
		}
	}
}

// processPendingMessages 处理待处理的消息
func (w *CicdWorker) processPendingMessages(ctx context.Context) {
	streams, err := w.stream.XReadGroupPending(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, w.consumerName, 100)
	if err != nil {
		global.Logger.Error("read pending messages failed", zap.Error(err))
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			w.processMessage(ctx, msg)
		}
	}
}

// processMessage 处理单条消息
func (w *CicdWorker) processMessage(ctx context.Context, msg redis.XMessage) {
	// 解析消息
	taskIDStr, ok := msg.Values["task_id"].(string)
	if !ok {
		global.Logger.Error("invalid task_id in message", zap.String("msg_id", msg.ID))
		w.ackMessage(ctx, msg.ID)
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		global.Logger.Error("parse task_id failed", zap.String("task_id", taskIDStr), zap.Error(err))
		w.ackMessage(ctx, msg.ID)
		return
	}

	releaseIDStr, _ := msg.Values["release_id"].(string)
	releaseID, _ := strconv.ParseInt(releaseIDStr, 10, 64)

	global.Logger.Info("processing task",
		zap.Int64("task_id", taskID),
		zap.Int64("release_id", releaseID),
		zap.String("msg_id", msg.ID))

	// 执行任务
	w.executeTask(ctx, taskID, releaseID)

	// ACK 消息
	w.ackMessage(ctx, msg.ID)
}

// executeTask 执行任务
func (w *CicdWorker) executeTask(ctx context.Context, taskID, releaseID int64) {
	// 1. 获取任务信息
	task, err := w.dao.CicdTaskGetByID(ctx, taskID)
	if err != nil {
		global.Logger.Error("get task failed", zap.Int64("task_id", taskID), zap.Error(err))
		return
	}

	// 检查任务状态（避免重复执行）
	if task.Status != models.CicdTaskStatusPending && task.Status != "Queued" {
		global.Logger.Info("task already processed", zap.Int64("task_id", taskID), zap.String("status", task.Status))
		return
	}

	// 2. 获取发布单信息
	release, err := w.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		global.Logger.Error("get release failed", zap.Int64("release_id", releaseID), zap.Error(err))
		w.markTaskFailed(ctx, taskID, releaseID, "获取发布单失败")
		return
	}

	// 检查发布单是否被取消
	if release.Status == models.CicdReleaseStatusCanceled {
		global.Logger.Info("release canceled, skip task", zap.Int64("task_id", taskID))
		w.markTaskFailed(ctx, taskID, releaseID, "发布单已取消")
		return
	}

	// 3. 标记任务开始执行
	if err := w.dao.CicdTaskMarkStarted(ctx, taskID); err != nil {
		global.Logger.Error("mark task started failed", zap.Int64("task_id", taskID), zap.Error(err))
	}

	// 更新 Release 状态为 Running（CAS，只有第一个任务开始时才更新）
	_, _ = w.dao.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{models.CicdReleaseStatusQueued},
		models.CicdReleaseStatusRunning,
		"deploying")

	// 4. 执行部署
	result := w.executor.Execute(ctx, task, release)

	// 5. 更新原镜像（用于回滚）
	if result.PrevImage != "" {
		_ = w.dao.CicdTaskUpdatePrevImage(ctx, taskID, result.PrevImage)
	}

	// 6. 更新任务状态
	if result.Success {
		w.markTaskSucceeded(ctx, taskID, releaseID, result.Message)
	} else {
		w.markTaskFailed(ctx, taskID, releaseID, result.Message)
	}
}

// markTaskSucceeded 标记任务成功
func (w *CicdWorker) markTaskSucceeded(ctx context.Context, taskID, releaseID int64, message string) {
	if err := w.dao.CicdTaskMarkFinished(ctx, taskID, models.CicdTaskStatusSucceeded, message); err != nil {
		global.Logger.Error("mark task succeeded failed", zap.Int64("task_id", taskID), zap.Error(err))
	}

	global.Logger.Info("task succeeded", zap.Int64("task_id", taskID))

	// 尝试完结发布单
	w.svc.TryFinalizeRelease(ctx, releaseID)

	// 发送回调通知
	w.callback.NotifyTaskComplete(ctx, taskID, releaseID, true, message)
}

// markTaskFailed 标记任务失败
func (w *CicdWorker) markTaskFailed(ctx context.Context, taskID, releaseID int64, message string) {
	if err := w.dao.CicdTaskMarkFinished(ctx, taskID, models.CicdTaskStatusFailed, message); err != nil {
		global.Logger.Error("mark task failed error", zap.Int64("task_id", taskID), zap.Error(err))
	}

	global.Logger.Error("task failed", zap.Int64("task_id", taskID), zap.String("message", message))

	// 尝试完结发布单
	w.svc.TryFinalizeRelease(ctx, releaseID)

	// 发送回调通知
	w.callback.NotifyTaskComplete(ctx, taskID, releaseID, false, message)
}

// ackMessage ACK 消息
func (w *CicdWorker) ackMessage(ctx context.Context, msgID string) {
	if _, err := w.stream.XAck(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, msgID); err != nil {
		global.Logger.Error("ack message failed", zap.String("msg_id", msgID), zap.Error(err))
	}
}
