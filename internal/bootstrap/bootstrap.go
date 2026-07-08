package bootstrap

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/initialize"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/internal/app/worker"
	"k8soperation/pkg/k8s/crd"
	"k8soperation/pkg/openai"
)

var (
	// cicdWorker CICD 部署任务消费者
	cicdWorker *worker.CicdWorker
	// pollWorker 流水线状态轮询 Worker
	pollWorker *worker.PipelinePollWorker
	// gitOpsSyncWorker GitOps 同步状态轮询 Worker
	gitOpsSyncWorker *worker.GitOpsSyncWorker
	// auditCleanupWorker 审计日志清理 Worker
	auditCleanupWorker *worker.AuditCleanupWorker
	// alertEvalWorker 告警规则评估 Worker
	alertEvalWorker *worker.AlertEvalWorker
	// aiopsInspectionWorker AIOps 智能巡检 Worker
	aiopsInspectionWorker *worker.AIOpsInspectionWorker
)

func InitAll() error {
	// 初始化配置
	if err := initialize.SetupSetting(); err != nil {
		return err
	}
	// 初始校验规则
	if err := initialize.SetupValidator(); err != nil {
		return err
	}

	// 初始化日志
	if err := initialize.SetupLogger(); err != nil {
		return err
	}

	// 注入 AI 日志器到 openai 包（解耦依赖）
	if global.AILogger != nil {
		openai.AILog = global.AILogger
	}

	// 初始化数据库
	if err := initialize.SetupDB(); err != nil {
		global.Logger.Error("init db failed", zap.Error(err))
		return fmt.Errorf("init db failed: %w", err)
	}

	// 监控数据源引导：把 config.yaml 中的 PrometheusURL 一次性导入 DB（仅当 DB 中无任何 prometheus 数据源）
	// 并打印当前实际生效的数据源，方便运维确认
	BootstrapMonitorDatasource()

	// 初始化 Session（依赖 Redis）
	if err := initialize.SetupSession(); err != nil {
		global.Logger.Error("init session failed", zap.Error(err))
		return fmt.Errorf("init session failed: %w", err)
	}

	// 初始化 Redis 客户端
	// Redis 存储用户认证信息，属于核心依赖，初始化失败必须终止启动
	if err := initialize.SetupRedis(); err != nil {
		global.Logger.Error("init redis failed", zap.Error(err))
		return fmt.Errorf("init redis failed: %w", err)
	}
	// 初始化K8s（失败不阻塞启动，登录/RBAC/CICD 等功能仍可用）
	if err := initialize.SetupK8sBootstrap(); err != nil {
		global.Logger.Warn("K8s 集群初始化失败，集群管理功能暂不可用，其他功能正常", zap.Error(err))
	}

	// 初始化 AppConfig CRD 客户端（依赖 K8s，失败不阻塞）
	if err := crd.SetupAppConfigClient(); err != nil {
		global.Logger.Warn("AppConfig CRD 客户端初始化失败，CRD 功能暂不可用", zap.Error(err))
	}

	// 加载 swagger 接口文档
	initialize.LogDocsReady()

	// 初始化并启动 CICD Worker
	if err := StartCicdWorker(); err != nil {
		global.Logger.Warn("start cicd worker failed", zap.Error(err))
		// 不返回错误，Worker 启动失败不影响主服务
	}

	// 数据补全：同步历史流水线审批阶段到 cicd_approval 表
	SyncApprovalData()

	// 启动审计日志清理 Worker
	auditCleanupWorker = worker.NewAuditCleanupWorker()
	auditCleanupWorker.Start()

	// 启动告警规则评估 Worker（定期查询 Prometheus 评估规则，产生告警事件）
	alertEvalWorker = worker.NewAlertEvalWorker()
	alertEvalWorker.Start()

	// AIOps: 自动建表（如果不存在）
	if global.DB != nil {
		global.DB.AutoMigrate(&models.AIOpsAnalysisRecord{}, &models.AIOpsInspectionReport{})
	}

	// 启动 AIOps 智能巡检 Worker（每 6 小时自动巡检 + AI 分析）
	if global.AISetting != nil && global.AISetting.Enabled {
		aiopsInspectionWorker = worker.NewAIOpsInspectionWorker()
		aiopsInspectionWorker.Start()
	}

	return nil
}

// Sync() 会做两件事：
// 调用底层 WriteSyncer 的 Sync()（例如 os.File.Sync()）；
// 把缓冲日志强制写到文件。
func FlushLoggers() {
	// 系统日志落盘
	_ = global.Logger.Sync()
	if global.BizLogger != nil {
		// 业务日志落盘
		_ = global.BizLogger.Sync()
	}
	if global.AILogger != nil {
		_ = global.AILogger.Sync()
	}
}

// StartCicdWorker 启动 CICD Worker
func StartCicdWorker() error {
	if global.RedisCli == nil {
		global.Logger.Warn("redis client not initialized, cicd worker will not start")
		return nil
	}

	// 创建集群客户端工厂
	svc := services.NewBackgroundServices()
	factory := services.NewClusterClientFactory(svc)

	// 创建并启动 Worker
	cicdWorker = worker.NewCicdWorker(global.RedisCli, factory)
	if err := cicdWorker.Start(context.Background()); err != nil {
		return err
	}

	// 启动流水线状态轮询 Worker（兼容回调失败的傅底机制）
	if global.JenkinsSetting != nil && global.JenkinsSetting.URL != "" {
		pollWorker = worker.NewPipelinePollWorker()
		pollWorker.Start(context.Background())
	}

	// 启动 GitOps 同步状态轮询 Worker（如果 GitOps 已配置）
	if global.GitOpsSetting != nil && global.GitOpsSetting.ArgoCDURL != "" {
		gitOpsSyncWorker = worker.NewGitOpsSyncWorker()
		gitOpsSyncWorker.Start(context.Background())
	}

	return nil
}

// StopCicdWorker 停止 CICD Worker
func StopCicdWorker() {
	if cicdWorker != nil {
		cicdWorker.Stop()
	}
	if pollWorker != nil {
		pollWorker.Stop()
	}
	if gitOpsSyncWorker != nil {
		gitOpsSyncWorker.Stop()
	}
	if auditCleanupWorker != nil {
		auditCleanupWorker.Stop()
	}
	if alertEvalWorker != nil {
		alertEvalWorker.Stop()
	}
	if aiopsInspectionWorker != nil {
		aiopsInspectionWorker.Stop()
	}
}

// SyncApprovalData 启动时补全历史审批数据
// 扫描 cicd_pipeline_stage 中的审批阶段，对没有对应 cicd_approval 记录的自动补建
func SyncApprovalData() {
	if global.DB == nil {
		return
	}

	ctx := context.Background()
	d := dao.NewDao(global.DB)

	// 查询所有审批类型的阶段记录
	stages, err := d.StageListApprovalAll(ctx)
	if err != nil {
		global.Logger.Warn("审批数据补全: 查询审批阶段失败", zap.Error(err))
		return
	}

	if len(stages) == 0 {
		// 无审批阶段，静默跳过
		return
	}

	var synced int
	for _, stage := range stages {
		// 检查是否已有对应的 cicd_approval 记录
		exists, err := d.ApprovalExistsByStageID(ctx, stage.ID)
		if err != nil {
			continue
		}
		if exists {
			continue
		}

		// 根据阶段状态确定审批记录的状态
		approvalStatus := models.ApprovalStatusPending
		switch stage.ApprovalDecision {
		case "approved":
			approvalStatus = models.ApprovalStatusApproved
		case "rejected":
			approvalStatus = models.ApprovalStatusRejected
		default:
			// 如果阶段已经完成但没有审批决策，可能是旧数据
			if stage.Status == models.StageStatusSuccess {
				approvalStatus = models.ApprovalStatusApproved
			} else if stage.Status == models.StageStatusFailed {
				approvalStatus = models.ApprovalStatusRejected
			} else if stage.Status == models.StageStatusWaiting {
				approvalStatus = models.ApprovalStatusPending
			} else if stage.Status == models.StageStatusSkipped || stage.Status == models.StageStatusAborted {
				approvalStatus = models.ApprovalStatusExpired
			}
		}

		// 获取运行记录信息
		var imageURL string
		var triggerUserID int64
		run, runErr := d.PipelineRunGetByID(ctx, stage.RunID)
		if runErr == nil && run != nil {
			imageURL = run.ImageURL
			triggerUserID = run.TriggerUserID
		}

		approval := &models.CicdApproval{
			PipelineID:    stage.PipelineID,
			PipelineRunID: stage.RunID,
			StageID:       stage.ID,
			Status:        approvalStatus,
			Image:         imageURL,
			RequestUserID: triggerUserID,
			RequestReason: "流水线构建完成，等待人工审批",
		}

		// 如果已审批，填充审批人和时间
		if approvalStatus != models.ApprovalStatusPending {
			approval.ApproveUserID = stage.ApprovalUserID
			approval.ApproveReason = stage.ApprovalComment
			if stage.FinishedAt > 0 {
				approval.ApproveTime = uint64(stage.FinishedAt)
			} else {
				approval.ApproveTime = uint64(time.Now().Unix())
			}
		}

		_, createErr := d.ApprovalCreate(ctx, approval)
		if createErr != nil {
			global.Logger.Warn("审批数据补全: 创建审批记录失败",
				zap.Int64("stage_id", stage.ID),
				zap.Error(createErr),
			)
			continue
		}
		synced++
	}

	global.Logger.Info("审批数据补全完成",
		zap.Int("total_stages", len(stages)),
		zap.Int("synced", synced),
	)
}

// BootstrapMonitorDatasource 启动引导：把 config.yaml 中的 Monitoring.PrometheusURL 一次性导入 DB，后续完全由【数据源管理】维护。
// 设计原则：
//   - config.yaml 仅作为首次启动的引导值（bootstrap），不是运行期以上为准的来源
//   - 所有运行期修改（增/删/改默认）由前端【数据源管理】写入 DB，MonitoringService 实时从 DB 读取
//   - 启动后明确打印“实际生效”的数据源地址（避免运维误以为“前端改了没生效”）
func BootstrapMonitorDatasource() {
	if global.DB == nil {
		return
	}
	ctx := context.Background()

	// 1) 在 DB 中查是否已有 prometheus 数据源
	var count int64
	if err := global.DB.WithContext(ctx).Model(&models.MonitorDatasource{}).
		Where("type = ? AND is_del = 0", "prometheus").Count(&count).Error; err != nil {
		global.Logger.Warn("[Monitoring Bootstrap] 查询 DB 数据源失败，跳过引导", zap.Error(err))
		return
	}

	// 2) 若 DB 中一条没有且 config.yaml 配了 URL，自动 INSERT 作为默认数据源（首次启动引导）
	if count == 0 && global.MonitoringSetting != nil && global.MonitoringSetting.PrometheusURL != "" {
		ds := &models.MonitorDatasource{
			Name:           "Prometheus-默认",
			Type:           "prometheus",
			URL:            global.MonitoringSetting.PrometheusURL,
			Description:    "由 config.yaml 首次启动自动引导，后续请在【数据源管理】中维护",
			AccessMode:     "proxy",
			AuthType:       "none",
			IsDefault:      true,
			Enabled:        true,
			Timeout:        30,
			ScrapeInterval: 15,
			Status:         "unknown",
		}
		if err := global.DB.WithContext(ctx).Create(ds).Error; err != nil {
			global.Logger.Warn("[Monitoring Bootstrap] 引导写入默认数据源失败", zap.Error(err))
		} else {
			global.Logger.Info("[Monitoring Bootstrap] 首次启动：已将 config.yaml 中的 PrometheusURL 引导至 DB",
				zap.Int64("id", ds.ID), zap.String("url", ds.URL))
		}
	}

	// 3) 打印当前实际生效的数据源（优先级：DB is_default=1 > DB 任一 enabled=1 > config.yaml staticURL）
	var active models.MonitorDatasource
	found := false
	if err := global.DB.WithContext(ctx).
		Where("type IN (?,?,?) AND is_default = 1 AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").First(&active).Error; err == nil {
		found = true
	} else if err := global.DB.WithContext(ctx).
		Where("type IN (?,?,?) AND enabled = 1 AND is_del = 0",
			"prometheus", "victoriametrics", "thanos").Order("id DESC").First(&active).Error; err == nil {
		found = true
	}
	if found {
		global.Logger.Info("[Monitoring] 实际生效数据源（来自 DB）",
			zap.Int64("id", active.ID),
			zap.String("name", active.Name),
			zap.String("type", active.Type),
			zap.String("url", active.URL),
			zap.Bool("is_default", active.IsDefault))
	} else if global.MonitoringSetting != nil && global.MonitoringSetting.PrometheusURL != "" {
		global.Logger.Warn("[Monitoring] DB 中无可用数据源，将使用 config.yaml 中的兜底地址",
			zap.String("url", global.MonitoringSetting.PrometheusURL))
	} else {
		global.Logger.Warn("[Monitoring] DB 和 config.yaml 都未配置 Prometheus 数据源，监控功能不可用")
	}
}

