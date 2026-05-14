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

	global.Logger.Info("cicd worker started successfully")

	// 启动流水线状态轮询 Worker（兼容回调失败的儹底机制）
	if global.JenkinsSetting != nil && global.JenkinsSetting.URL != "" {
		pollWorker = worker.NewPipelinePollWorker()
		pollWorker.Start(context.Background())
		global.Logger.Info("pipeline poll worker started successfully")
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
		global.Logger.Info("审批数据补全: 无审批阶段，跳过")
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

