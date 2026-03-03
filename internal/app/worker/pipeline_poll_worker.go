package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/jenkins"
)

// PipelinePollWorker 流水线状态轮询 Worker
// 轮询 Jenkins 获取未终态且回调未收到的构建状态
type PipelinePollWorker struct {
	dao *dao.Dao

	pollInterval time.Duration // 轮询间隔
	maxBuildTime int           // 最大构建时间(分钟)
	batchSize    int           // 每批处理数量
	limiter      *rate.Limiter // 限流器（防止打爆 Jenkins）

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewPipelinePollWorker 创建轮询 Worker
func NewPipelinePollWorker() *PipelinePollWorker {
	// 从配置读取参数，设置默认值
	pollInterval := 15 * time.Second
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
		pollInterval: pollInterval,
		maxBuildTime: maxBuildTime,
		batchSize:    20,
		limiter:      rate.NewLimiter(rate.Every(500*time.Millisecond), 5), // 每秒最多2次请求，突发5次
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

// pollOnce 执行一次轮询
func (w *PipelinePollWorker) pollOnce(ctx context.Context) {
	// 1. 先标记超时的记录
	w.markTimeoutRecords(ctx)

	// 2. 获取需要轮询的记录
	runs, err := w.dao.PipelineRunListPendingForPoll(ctx, w.maxBuildTime, w.batchSize)
	if err != nil {
		global.Logger.Error("[轮询Worker] 获取待轮询记录失败", zap.Error(err))
		return
	}

	if len(runs) == 0 {
		return
	}

	global.Logger.Debug("[轮询Worker] 开始轮询",
		zap.Int("count", len(runs)),
	)

	// 3. 逐个轮询状态
	for _, run := range runs {
		// 限流
		if err := w.limiter.Wait(ctx); err != nil {
			global.Logger.Warn("[轮询Worker] 限流等待被中断", zap.Error(err))
			return
		}

		w.pollSingleRun(ctx, run)
	}
}

// pollSingleRun 轮询单个运行记录
func (w *PipelinePollWorker) pollSingleRun(ctx context.Context, run *models.CicdPipelineRun) {
	// 没有构建号，跳过
	if run.BuildNumber == 0 {
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

	// 获取构建信息
	buildInfo, err := client.GetBuildInfo(ctx, pipeline.JenkinsJob, run.BuildNumber)
	if err != nil {
		global.Logger.Warn("[轮询Worker] 获取构建信息失败",
			zap.Int64("run_id", run.ID),
			zap.String("job", pipeline.JenkinsJob),
			zap.Int("build_number", run.BuildNumber),
			zap.Error(err),
		)
		return
	}

	// 如果构建还在进行中，跳过
	if buildInfo.Building {
		return
	}

	// 构建已完成，更新状态
	runStatus := jenkins.BuildStatusToRunStatus(false, buildInfo.Result)
	duration := int(buildInfo.Duration / 1000) // 毫秒转秒

	global.Logger.Info("[轮询Worker] 检测到构建完成",
		zap.Int64("run_id", run.ID),
		zap.String("status", runStatus),
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
}

// markTimeoutRecords 标记超时的记录
func (w *PipelinePollWorker) markTimeoutRecords(ctx context.Context) {
	affected, err := w.dao.PipelineRunMarkTimeout(ctx, w.maxBuildTime)
	if err != nil {
		global.Logger.Error("[轮询Worker] 标记超时记录失败", zap.Error(err))
		return
	}
	if affected > 0 {
		global.Logger.Info("[轮询Worker] 已标记超时记录",
			zap.Int64("count", affected),
		)
	}
}

// getJenkinsClient 获取 Jenkins 客户端
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

	return jenkins.NewClient(
		jenkinsURL,
		global.JenkinsSetting.Username,
		global.JenkinsSetting.APIToken,
	)
}
