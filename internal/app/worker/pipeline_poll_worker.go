package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/jenkins"
)

// PipelinePollWorker 流水线状态轮询 Worker
// 轮询 Jenkins 获取未终态且回调未收到的构建状态
type PipelinePollWorker struct {
	dao *dao.Dao
	svc *services.Services // 用于发送钉钉通知

	pollInterval time.Duration // 轮询间隔
	maxBuildTime int           // 最大构建时间(分钟)
	batchSize    int           // 每批处理数量
	workerCount  int           // 并行 worker 数量
	limiter      *rate.Limiter // 限流器（防止打爆 Jenkins）

	pollCount int64 // 轮询计数器（用于心跳日志）

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewPipelinePollWorker 创建轮询 Worker
func NewPipelinePollWorker() *PipelinePollWorker {
	// 从配置读取参数，设置默认值
	pollInterval := 10 * time.Second
	maxBuildTime := 30 // 分钟
	if global.JenkinsSetting != nil {
		if global.JenkinsSetting.PollInterval > 0 {
			pollInterval = time.Duration(global.JenkinsSetting.PollInterval) * time.Second
		}
		if global.JenkinsSetting.MaxBuildTime > 0 {
			maxBuildTime = global.JenkinsSetting.MaxBuildTime
		}
	}

	return &PipelinePollWorker{
		dao:          dao.NewDao(global.DB),
		svc:          services.NewBackgroundServices(),
		pollInterval: pollInterval,
		maxBuildTime: maxBuildTime,
		batchSize:    100,                                              // 扩大批次: 20 -> 100
		workerCount:  5,                                                // 5 个并行 worker
		limiter:      rate.NewLimiter(rate.Every(100*time.Millisecond), 10), // 10 QPS，突发 10
		stopCh:       make(chan struct{}),
	}
}

// Start 启动轮询 Worker
func (w *PipelinePollWorker) Start(ctx context.Context) {
	global.Logger.Info("[轮询Worker] 启动",
		zap.Duration("poll_interval", w.pollInterval),
		zap.Int("max_build_time_minutes", w.maxBuildTime),
	)

	w.wg.Add(1)
	go w.pollLoop(ctx)
}

// Stop 停止轮询 Worker
func (w *PipelinePollWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	global.Logger.Info("[轮询Worker] 已停止")
}

// pollLoop 轮询循环
func (w *PipelinePollWorker) pollLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollOnce(ctx)
		}
	}
}

// pollOnce 执行一次轮询（并行处理）
func (w *PipelinePollWorker) pollOnce(ctx context.Context) {
	w.pollCount++

	// 每 4 次轮询（约 1 分钟）输出一次心跳 INFO 日志，确认 worker 存活
	if w.pollCount%4 == 1 {
		global.Logger.Info("[轮询Worker] 心跳",
			zap.Int64("poll_count", w.pollCount),
		)
	}

	// 1. 先标记超时的记录
	w.markTimeoutRecords(ctx)

	// 1.5 修复孤儿 pipeline：status=running 但没有活跃 run 记录（历史遗留数据）
	w.fixOrphanedPipelines(ctx)

	// 2. 获取需要轮询的记录
	runs, err := w.dao.PipelineRunListPendingForPoll(ctx, w.maxBuildTime, w.batchSize)
	if err != nil {
		global.Logger.Error("[轮询Worker] 获取待轮询记录失败", zap.Error(err))
		return
	}

	if len(runs) == 0 {
		return
	}

	global.Logger.Info("[轮询Worker] 开始轮询",
		zap.Int("count", len(runs)),
	)

	// 3. 并行轮询：通过 channel 分发给多个 worker
	runCh := make(chan *models.CicdPipelineRun, len(runs))
	for _, run := range runs {
		runCh <- run
	}
	close(runCh)

	var pollWg sync.WaitGroup
	for i := 0; i < w.workerCount; i++ {
		pollWg.Add(1)
		go func() {
			defer pollWg.Done()
			for run := range runCh {
				// 限流
				if err := w.limiter.Wait(ctx); err != nil {
					return
				}
				w.safePollSingleRun(ctx, run)
			}
		}()
	}
	pollWg.Wait()
}

// safePollSingleRun 带 panic 恢复的 pollSingleRun 包装
func (w *PipelinePollWorker) safePollSingleRun(ctx context.Context, run *models.CicdPipelineRun) {
	defer func() {
		if r := recover(); r != nil {
			global.Logger.Error("[轮询Worker] pollSingleRun panic 恢复",
				zap.Int64("run_id", run.ID),
				zap.Any("panic", r),
			)
		}
	}()
	w.pollSingleRun(ctx, run)
}

// pollSingleRun 轮询单个运行记录
func (w *PipelinePollWorker) pollSingleRun(ctx context.Context, run *models.CicdPipelineRun) {
	// 没有构建号：检查是否卡住太久（超过 5 分钟没获取到 buildNumber，说明触发失败了）
	if run.BuildNumber == 0 {
		createdAt := time.Unix(int64(run.CreatedAt), 0)
		if time.Since(createdAt) > 5*time.Minute {
			global.Logger.Warn("[轮询Worker] 运行记录无构建号且超过 5 分钟，标记为失败",
				zap.Int64("run_id", run.ID),
				zap.Int64("pipeline_id", run.PipelineID),
			)
			run.ErrorMessage = "Jenkins 构建触发失败（未获取到构建号）"
			_ = w.dao.PipelineRunUpdateCallback(ctx, run.ID, models.PipelineRunStatusFailed, "", "", run.ErrorMessage, 0)
			_ = w.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed)
			// 发送钉钉构建失败通知
			w.notifyBuildFailed(ctx, run)
		}
		return
	}

	// 获取流水线信息
	pipeline, err := w.dao.PipelineGetByID(ctx, run.PipelineID)
	if err != nil {
		global.Logger.Warn("[轮询Worker] 获取流水线失败",
			zap.Int64("run_id", run.ID),
			zap.Int64("pipeline_id", run.PipelineID),
			zap.Error(err),
		)
		return
	}

	// 创建 Jenkins 客户端
	client := w.getJenkinsClient(pipeline.JenkinsURL)
	if client == nil {
		return
	}

	// 使用 10 秒超时 context，避免 Jenkins 慢响应阻塞整个轮询周期
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 获取构建信息
	buildInfo, err := client.GetBuildInfo(pollCtx, pipeline.JenkinsJob, run.BuildNumber)
	if err != nil {
		global.Logger.Warn("[轮询Worker] 获取构建信息失败",
			zap.Int64("run_id", run.ID),
			zap.String("job", pipeline.JenkinsJob),
			zap.Int("build_number", run.BuildNumber),
			zap.Error(err),
		)
		return
	}

	// 如果构建还在进行中，记录日志并跳过
	if buildInfo.Building {
		global.Logger.Debug("[轮询Worker] 构建仍在进行中",
			zap.Int64("run_id", run.ID),
			zap.String("job", pipeline.JenkinsJob),
			zap.Int("build_number", run.BuildNumber),
		)
		return
	}

	// 构建已完成，更新状态
	runStatus := jenkins.BuildStatusToRunStatus(false, buildInfo.Result)
	duration := int(buildInfo.Duration / 1000) // 毫秒转秒

	global.Logger.Info("[轮询Worker] 检测到构建完成",
		zap.Int64("run_id", run.ID),
		zap.String("job", pipeline.JenkinsJob),
		zap.Int("build_number", run.BuildNumber),
		zap.String("status", runStatus),
		zap.String("jenkins_result", buildInfo.Result),
		zap.Int("duration", duration),
	)

	// 更新运行记录（标记回调已收到，因为是轮询到的）
	if err := w.dao.PipelineRunUpdateCallback(ctx, run.ID, runStatus, "", "", "", duration); err != nil {
		global.Logger.Error("[轮询Worker] 更新运行记录失败",
			zap.Int64("run_id", run.ID),
			zap.Error(err),
		)
		return
	}

	// 更新流水线状态
	if err := w.dao.PipelineUpdateRunComplete(ctx, pipeline.ID, runStatus); err != nil {
		global.Logger.Warn("[轮询Worker] 更新流水线状态失败",
			zap.Int64("pipeline_id", pipeline.ID),
			zap.Error(err),
		)
	}

	// 发送钉钉通知（构建失败/中止时）
	if runStatus == models.PipelineRunStatusFailed || runStatus == models.PipelineRunStatusAborted {
		run.DurationSec = duration
		w.notifyBuildFailed(ctx, run)
	}

	// 确保关键操作日志即时落盘
	_ = global.Logger.Sync()
}

// markTimeoutRecords 标记超时的记录，同时同步更新 pipeline 表状态
func (w *PipelinePollWorker) markTimeoutRecords(ctx context.Context) {
	// 先查出即将被标记超时的运行记录（用于同步更新 pipeline 表）
	timedOutRuns, err := w.dao.PipelineRunListTimedOut(ctx, w.maxBuildTime)
	if err != nil {
		global.Logger.Error("[轮询Worker] 查询超时记录失败", zap.Error(err))
		return
	}

	if len(timedOutRuns) == 0 {
		return
	}

	global.Logger.Info("[轮询Worker] 发现超时记录，准备标记",
		zap.Int("count", len(timedOutRuns)),
	)

	affected, err := w.dao.PipelineRunMarkTimeout(ctx, w.maxBuildTime)
	if err != nil {
		global.Logger.Error("[轮询Worker] 标记超时记录失败", zap.Error(err))
		return
	}
	if affected > 0 {
		global.Logger.Info("[轮询Worker] 已标记超时记录",
			zap.Int64("count", affected),
		)

		// 同步更新 pipeline 表状态为 idle + failed，并发送钉钉通知
		for _, run := range timedOutRuns {
			global.Logger.Info("[轮询Worker] 处理超时记录",
				zap.Int64("run_id", run.ID),
				zap.Int64("pipeline_id", run.PipelineID),
				zap.Int("build_number", run.BuildNumber),
			)
			if err := w.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed); err != nil {
				global.Logger.Warn("[轮询Worker] 同步更新流水线状态失败",
					zap.Int64("pipeline_id", run.PipelineID),
					zap.Int64("run_id", run.ID),
					zap.Error(err),
				)
			}
			// 发送钉钉构建超时通知
			run.ErrorMessage = fmt.Sprintf("构建超时（超过 %d 分钟未完成）", w.maxBuildTime)
			w.notifyBuildFailed(ctx, run)
		}

		// 确保通知日志即时落盘
		_ = global.Logger.Sync()
	}
}

// fixOrphanedPipelines 修复孤儿 pipeline：pipeline.status="running" 但没有活跃 run 记录
// 场景：旧代码标记了 run 为 failed 但未同步更新 pipeline 表，导致 pipeline 永久卡在 running
func (w *PipelinePollWorker) fixOrphanedPipelines(ctx context.Context) {
	orphaned, err := w.dao.PipelineListStuckRunning(ctx)
	if err != nil {
		global.Logger.Error("[轮询Worker] 查询孤儿 pipeline 失败", zap.Error(err))
		return
	}

	for _, p := range orphaned {
		// 查找最新的 run 记录来确定实际状态
		latestRun, err := w.dao.PipelineRunGetLatest(ctx, p.ID)
		if err != nil {
			// 没有 run 记录，直接重置为 idle
			_ = w.dao.PipelineUpdateRunComplete(ctx, p.ID, models.PipelineRunStatusFailed)
			global.Logger.Info("[轮询Worker] 修复孤儿 pipeline（无 run 记录）",
				zap.Int64("pipeline_id", p.ID),
				zap.String("pipeline_name", p.Name),
			)
			continue
		}

		// run 已经是终态，同步 pipeline 状态
		if latestRun.Status == models.PipelineRunStatusSuccess ||
			latestRun.Status == models.PipelineRunStatusFailed ||
			latestRun.Status == models.PipelineRunStatusAborted {
			_ = w.dao.PipelineUpdateRunComplete(ctx, p.ID, latestRun.Status)
			global.Logger.Info("[轮询Worker] 修复孤儿 pipeline（run 已终态）",
				zap.Int64("pipeline_id", p.ID),
				zap.String("pipeline_name", p.Name),
				zap.String("actual_status", latestRun.Status),
			)
			continue
		}

		// run 还在 pending/running，但已经超时太久（超过 maxBuildTime），强制标记失败
		createdAt := time.Unix(int64(latestRun.CreatedAt), 0)
		if time.Since(createdAt) > time.Duration(w.maxBuildTime)*time.Minute {
			_ = w.dao.PipelineRunUpdateCallback(ctx, latestRun.ID, models.PipelineRunStatusFailed, "", "", "构建超时（孤儿修复）", 0)
			_ = w.dao.PipelineUpdateRunComplete(ctx, p.ID, models.PipelineRunStatusFailed)
			global.Logger.Info("[轮询Worker] 修复孤儿 pipeline（run 超时）",
				zap.Int64("pipeline_id", p.ID),
				zap.String("pipeline_name", p.Name),
				zap.Int64("run_id", latestRun.ID),
			)
		}
	}
}

// notifyBuildFailed 发送钉钉构建失败通知（同步调用，确保通知可靠送达）
func (w *PipelinePollWorker) notifyBuildFailed(ctx context.Context, run *models.CicdPipelineRun) {
	pipeline, err := w.dao.PipelineGetByID(ctx, run.PipelineID)
	if err != nil {
		global.Logger.Warn("[轮询Worker] 发送通知前获取流水线失败",
			zap.Int64("run_id", run.ID),
			zap.Int64("pipeline_id", run.PipelineID),
			zap.Error(err),
		)
		return
	}

	global.Logger.Info("[轮询Worker] 准备发送构建失败钉钉通知",
		zap.Int64("run_id", run.ID),
		zap.String("pipeline_name", pipeline.Name),
		zap.Int("build_number", run.BuildNumber),
	)

	// 同步调用通知，不走 goroutine，确保通知结果可追踪
	w.svc.NotifyBuildResultSync(ctx, pipeline, run, false)
}

// getJenkinsClient 获取 Jenkins 客户端（全局缓存单例，复用连接池）
func (w *PipelinePollWorker) getJenkinsClient(pipelineJenkinsURL string) *jenkins.Client {
	if global.JenkinsSetting == nil {
		return nil
	}

	jenkinsURL := pipelineJenkinsURL
	if jenkinsURL == "" {
		jenkinsURL = global.JenkinsSetting.URL
	}
	if jenkinsURL == "" {
		return nil
	}

	return jenkins.GetOrCreateClient(
		jenkinsURL,
		global.JenkinsSetting.Username,
		global.JenkinsSetting.APIToken,
	)
}
