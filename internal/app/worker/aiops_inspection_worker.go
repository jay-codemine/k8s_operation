package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/services"
)

// AIOpsInspectionWorker 智能巡检 Worker
// 定时执行平台巡检，生成 AI 健康报告
type AIOpsInspectionWorker struct {
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	svc      *services.Services
}

// NewAIOpsInspectionWorker 创建巡检 Worker
// 默认每 6 小时执行一次全量巡检
func NewAIOpsInspectionWorker() *AIOpsInspectionWorker {
	return &AIOpsInspectionWorker{
		interval: 6 * time.Hour,
		stopCh:   make(chan struct{}),
		svc:      services.NewBackgroundServices(),
	}
}

// Start 启动
func (w *AIOpsInspectionWorker) Start() {
	global.Logger.Info("[AIOps-InspectionWorker] 启动智能巡检引擎",
		zap.Duration("interval", w.interval))
	w.wg.Add(1)
	go w.loop()
}

// Stop 停止
func (w *AIOpsInspectionWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	global.Logger.Info("[AIOps-InspectionWorker] 已停止")
}

func (w *AIOpsInspectionWorker) loop() {
	defer w.wg.Done()

	// 启动 60 秒后执行首次巡检（等待其他组件就绪）
	select {
	case <-time.After(60 * time.Second):
	case <-w.stopCh:
		return
	}
	w.runOnce()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.runOnce()
		}
	}
}

func (w *AIOpsInspectionWorker) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	global.Logger.Info("[AIOps-InspectionWorker] 开始定时巡检")

	// AI 未启用时仍执行基础巡检（数据采集 + 健康评分 + 落报告），仅 AI 分析字段留空；
	// executeInspection 内部已对 AI 不可用做降级处理（getAIOpsClient 返回 error 时跳过分析），不会报错
	_, err := w.svc.AIOpsSvc().RunInspection(ctx, 0) // triggerBy=0 表示系统定时
	if err != nil {
		global.Logger.Warn("[AIOps-InspectionWorker] 巡检执行失败", zap.Error(err))
		return
	}

	global.Logger.Info("[AIOps-InspectionWorker] 定时巡检已提交")
}
