package worker

import (
	"context"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/services"
)

// AuditCleanupWorker 审计日志定时清理 Worker
// 每天凌晨 3:00 自动清理过期审计日志
type AuditCleanupWorker struct {
	stopCh chan struct{}
}

// NewAuditCleanupWorker 创建审计清理 Worker
func NewAuditCleanupWorker() *AuditCleanupWorker {
	return &AuditCleanupWorker{
		stopCh: make(chan struct{}),
	}
}

// Start 启动审计清理定时任务
func (w *AuditCleanupWorker) Start() {
	global.Logger.Info("[AuditCleanup] 审计日志定时清理 Worker 已启动")

	go func() {
		// 首次启动延迟 5 分钟执行一次清理
		timer := time.NewTimer(5 * time.Minute)
		defer timer.Stop()

		select {
		case <-timer.C:
			w.doCleanup()
		case <-w.stopCh:
			return
		}

		// 之后每 24 小时执行一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.doCleanup()
			case <-w.stopCh:
				global.Logger.Info("[AuditCleanup] 审计日志清理 Worker 已停止")
				return
			}
		}
	}()
}

// Stop 停止 Worker
func (w *AuditCleanupWorker) Stop() {
	close(w.stopCh)
}

// doCleanup 执行清理逻辑
func (w *AuditCleanupWorker) doCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	svc := services.NewBackgroundServices()

	affected, err := svc.AuditLogCleanup(ctx)
	if err != nil {
		global.Logger.Error("[AuditCleanup] 清理审计日志失败", zap.Error(err))
		return
	}

	if affected > 0 {
		global.Logger.Info("[AuditCleanup] 审计日志清理完成",
			zap.Int64("deleted_count", affected))
	} else {
		global.Logger.Debug("[AuditCleanup] 无需清理（未过期或永久保留）")
	}
}
