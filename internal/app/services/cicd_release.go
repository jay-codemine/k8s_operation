package services

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/builder"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/infra"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"

	"strings"
	"time"
)

func (s *Services) CicdReleaseCreate(
	ctx context.Context,
	req *requests.CicdReleaseCreateRequest,
	userID int64,
) (int64, error) {

	// 模板化发布：如果传入 pipeline_id，自动继承流水线的部署配置
	if req.PipelineID > 0 {
		pipeline, err := s.dao.PipelineGetByID(ctx, req.PipelineID)
		if err != nil {
			return 0, fmt.Errorf("关联流水线不存在(id=%d): %w", req.PipelineID, err)
		}
		// 自动填充缺失字段（用户显式传入的优先级更高）
		if req.AppName == "" {
			req.AppName = pipeline.Name
		}
		if req.Namespace == "" {
			req.Namespace = pipeline.TargetNamespace
		}
		if req.WorkloadKind == "" {
			req.WorkloadKind = pipeline.TargetWorkloadKind
		}
		if req.WorkloadName == "" {
			req.WorkloadName = pipeline.TargetWorkloadName
		}
		if req.ContainerName == "" {
			req.ContainerName = pipeline.TargetContainer
		}
		if req.ImageRepo == "" {
			// 尝试从流水线环境变量中获取 IMAGE_REPO
			for _, ev := range pipeline.EnvVars {
				if ev.Name == "IMAGE_REPO" {
					req.ImageRepo = ev.Value
					break
				}
			}
		}
		// 回退策略：从最近部署镜像解析仓库地址
		if req.ImageRepo == "" && pipeline.LastDeployImage != "" {
			img := pipeline.LastDeployImage
			// 去掉 digest 部分（如 repo@sha256:xxx）
			if atIdx := strings.Index(img, "@"); atIdx > 0 {
				img = img[:atIdx]
			}
			// 去掉 tag 部分（如 repo:v1.0.0）
			if colonIdx := strings.LastIndex(img, ":"); colonIdx > 0 && !strings.Contains(img[colonIdx:], "/") {
				img = img[:colonIdx]
			}
			if img != "" {
				req.ImageRepo = img
			}
		}
		if len(req.ClusterIDs) == 0 && pipeline.TargetClusterID > 0 {
			req.ClusterIDs = []int64{pipeline.TargetClusterID}
		}
		// 默认值补充
		if req.Namespace == "" {
			req.Namespace = "default"
		}
		if req.WorkloadKind == "" {
			req.WorkloadKind = "Deployment"
		}
		// 校验必要字段
		if req.WorkloadName == "" {
			return 0, fmt.Errorf("流水线未配置目标工作负载名称，请先编辑流水线补充部署配置")
		}
		if req.ContainerName == "" {
			req.ContainerName = req.WorkloadName // 默认与工作负载同名
		}
		if len(req.ClusterIDs) == 0 {
			return 0, fmt.Errorf("流水线未配置目标集群，请先编辑流水线补充部署配置")
		}
		if req.ImageRepo == "" {
			return 0, fmt.Errorf("流水线未配置镜像仓库(IMAGE_REPO)，且无最近部署记录，请先编辑流水线或补充 IMAGE_REPO 环境变量")
		}
	}

	// 幂等：request_id 不为空则复用已创建的 release
	reqID := strings.TrimSpace(req.RequestID)
	if reqID != "" {
		exist, err := s.dao.CicdReleaseGetByRequestID(ctx, reqID)
		if err == nil && exist != nil {
			return exist.ID, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	// 止血：防止空集群导致“无任务但 Queued”
	if len(req.ClusterIDs) == 0 {
		return 0, fmt.Errorf("clusterIDs is empty")
	}

	now := uint64(time.Now().Unix())

	imageRepo := strings.TrimSpace(req.ImageRepo)
	imageTag := strings.TrimSpace(req.ImageTag)
	imageDigest := strings.TrimSpace(req.ImageDigest)

	target := builder.BuildTargetImage(imageRepo, imageTag, imageDigest)
	rel := builder.BuildCicdRelease(req, userID, now, imageRepo, imageTag, imageDigest)

	var tasks []*models.CicdReleaseTask

	// 事务：Release + Tasks 原子写入
	if err := s.dao.WithTx(ctx, func(tx *dao.Dao) error {
		if err := tx.CicdReleaseCreate(ctx, rel); err != nil {
			return err
		}

		tasks = builder.BuildCicdReleaseTasks(rel.ID, req.ClusterIDs, target, now)
		return tx.CicdTasksCreate(ctx, tasks)
	}); err != nil {
		return 0, err
	}

	// ====== 审批判断：根据环境配置决定是否需要审批 ======
	// 优先按 namespace 查环境，找不到再按 name 查
	env, envErr := s.dao.EnvironmentGetByNamespace(ctx, req.Namespace)
	if envErr != nil {
		env, _ = s.dao.EnvironmentGetByName(ctx, req.Namespace)
	}

	// 判断是否需要审批：环境配置 RequireApproval=true 时才走审批
	needApproval := true // 默认需要审批（找不到环境配置时保守处理）
	if env != nil {
		needApproval = env.RequireApproval
	}

	// 不需要审批：直接入队部署
	if !needApproval {
		global.Logger.Info("[发布] 环境无需审批，直接入队部署",
			zap.Int64("release_id", rel.ID),
			zap.String("namespace", req.Namespace),
			zap.String("app_name", req.AppName),
		)
		_, enqErr := s.releaseEnqueue(ctx, rel.ID, tasks)
		if enqErr != nil {
			global.Logger.Error("[发布] 直接入队失败", zap.Error(enqErr))
			return 0, fmt.Errorf("发布入队失败: %w", enqErr)
		}
		return rel.ID, nil
	}

	// 需要审批：查询多级审批策略
	totalLevels := 1
	levelLabel := "发布审批"
	if env != nil && len(env.ApprovalLevels) > 0 {
		totalLevels = len(env.ApprovalLevels)
		levelLabel = env.ApprovalLevels[0].Label
	}

	// 构造完整镜像地址
	fullImage := imageRepo
	if imageTag != "" {
		fullImage = imageRepo + ":" + imageTag
	}

	// 创建第一级审批记录
	approval := &models.CicdApproval{
		PipelineID:    req.PipelineID,
		ReleaseID:     rel.ID,
		EnvName:       req.Namespace, // 用命名空间标识环境
		Status:        models.ApprovalStatusPending,
		Image:         fullImage,
		ImageDigest:   imageDigest,
		RequestUserID: userID,
		RequestReason: fmt.Sprintf("手动创建发布: %s → %s/%s", req.AppName, req.Namespace, req.WorkloadName),
		ExpireTime:    uint64(time.Now().Unix() + 86400*7), // 7天过期
		ApprovalLevel: 1,
		TotalLevels:   totalLevels,
		LevelLabel:    levelLabel,
		CreatedAt:     uint64(time.Now().Unix()),
		ModifiedAt:    uint64(time.Now().Unix()),
	}
	approvalID, err := s.dao.ApprovalCreate(ctx, approval)
	if err != nil {
		global.Logger.Error("[发布] 创建审批记录失败", zap.Error(err))
		// 审批失败不降级，直接报错阻断（不允许自动发布）
		_, _ = s.dao.CicdReleaseUpdateStatusCAS(ctx, rel.ID,
			[]string{models.CicdReleaseStatusPending},
			models.CicdReleaseStatusFailed,
			"创建审批记录失败，发布已阻断",
		)
		return 0, fmt.Errorf("创建审批记录失败: %w", err)
	}

	// 更新发布单状态为「等待审批」
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(
		ctx,
		rel.ID,
		[]string{models.CicdReleaseStatusPending},
		models.CicdReleaseStatusAwaitingApproval,
		fmt.Sprintf("等待审批(approval_id=%d, levels=%d)", approvalID, totalLevels),
	)

	global.Logger.Info("[发布] 已创建发布单，等待多级审批",
		zap.Int64("release_id", rel.ID),
		zap.Int64("approval_id", approvalID),
		zap.Int("total_levels", totalLevels),
		zap.String("app_name", req.AppName),
	)

	return rel.ID, nil
}

// releaseEnqueue 将发布单的 tasks 入队 Redis Stream 执行部署
func (s *Services) releaseEnqueue(ctx context.Context, releaseID int64, tasks []*models.CicdReleaseTask) (int64, error) {
	enqueued := 0
	for _, t := range tasks {
		if _, err := s.stream.XAdd(ctx, infra.CicdDeployStream, map[string]any{
			"task_id":    t.ID,
			"release_id": releaseID,
		}); err != nil {
			_, _ = s.dao.CicdReleaseUpdateStatusCAS(
				ctx,
				releaseID,
				[]string{"Pending", "Queued", models.CicdReleaseStatusAwaitingApproval},
				"Failed",
				fmt.Sprintf("enqueue failed after %d/%d", enqueued, len(tasks)),
			)
			_ = s.dao.CicdTasksFailByRelease(
				ctx,
				releaseID,
				fmt.Sprintf("enqueue failed after %d/%d", enqueued, len(tasks)),
			)
			return 0, err
		}
		enqueued++
	}

	// 全部入队成功：Release 标 Queued
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(
		ctx,
		releaseID,
		[]string{"Pending", models.CicdReleaseStatusAwaitingApproval},
		"Queued",
		"enqueued",
	)

	return releaseID, nil
}

// CicdReleaseDetail 获取发布单详情
func (s *Services) CicdReleaseDetail(ctx context.Context, releaseID int64) (*models.CicdRelease, []*models.CicdReleaseTask, error) {
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		return nil, nil, err
	}

	tasks, err := s.dao.CicdTasksByReleaseID(ctx, releaseID)
	if err != nil {
		return nil, nil, err
	}

	return rel, tasks, nil
}

// CicdReleaseList 发布单列表
func (s *Services) CicdReleaseList(ctx context.Context, req *requests.CicdReleaseListRequest) ([]*models.CicdRelease, int64, error) {
	return s.dao.CicdReleaseList(ctx, req.Keyword, req.AppName, req.Status, req.Page, req.PageSize)
}

// CicdReleaseStats 发布单统计
func (s *Services) CicdReleaseStats(ctx context.Context) (map[string]int64, error) {
	return s.dao.CicdReleaseStats(ctx)
}

// CicdReleaseHistory 应用发布历史查询（增强版）
func (s *Services) CicdReleaseHistory(ctx context.Context, appName, namespace, status string, startTime, endTime int64, page, pageSize int) ([]*models.CicdRelease, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.dao.CicdReleaseHistory(ctx, appName, namespace, status, startTime, endTime, page, pageSize)
}

// CicdReleaseStatsEnhanced 增强版发布统计
func (s *Services) CicdReleaseStatsEnhanced(ctx context.Context) (*dao.ReleaseStatsEnhanced, error) {
	return s.dao.CicdReleaseStatsEnhanced(ctx)
}

// CicdReleaseUpdate 编辑发布单（仅 Pending/Failed/Canceled 状态可编辑）
func (s *Services) CicdReleaseUpdate(ctx context.Context, req *requests.CicdReleaseUpdateRequest) error {
	// 先查询发布单是否存在
	rel, err := s.dao.CicdReleaseGetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("发布单不存在: %w", err)
	}

	// 检查状态：仅 Pending/AwaitingApproval/Failed/Canceled 可编辑
	if rel.Status != models.CicdReleaseStatusPending &&
		rel.Status != models.CicdReleaseStatusAwaitingApproval &&
		rel.Status != models.CicdReleaseStatusFailed &&
		rel.Status != models.CicdReleaseStatusCanceled {
		return fmt.Errorf("发布单当前状态为 %s，无法编辑", rel.Status)
	}

	updates := map[string]any{}
	if req.AppName != "" {
		updates["app_name"] = req.AppName
	}
	if req.Namespace != "" {
		updates["namespace"] = req.Namespace
	}
	if req.WorkloadKind != "" {
		updates["workload_kind"] = req.WorkloadKind
	}
	if req.WorkloadName != "" {
		updates["workload_name"] = req.WorkloadName
	}
	if req.ContainerName != "" {
		updates["container_name"] = req.ContainerName
	}
	if req.Strategy != "" {
		updates["strategy"] = req.Strategy
	}
	if req.TimeoutSec > 0 {
		updates["timeout_sec"] = req.TimeoutSec
	}
	if req.Concurrency > 0 {
		updates["concurrency"] = req.Concurrency
	}
	if req.ImageRepo != "" {
		updates["image_repo"] = req.ImageRepo
	}
	if req.ImageTag != "" {
		updates["image_tag"] = req.ImageTag
	}
	if req.Message != "" {
		updates["message"] = req.Message
	}

	if len(updates) == 0 {
		return nil // 无需更新
	}

	return s.dao.CicdReleaseUpdate(ctx, req.ID, updates)
}

// CicdReleaseDelete 删除发布单（软删除）
func (s *Services) CicdReleaseDelete(ctx context.Context, releaseID int64) error {
	// 查询发布单是否存在
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		return fmt.Errorf("发布单不存在: %w", err)
	}

	// 运行中的发布单不允许删除
	if rel.Status == models.CicdReleaseStatusRunning ||
		rel.Status == models.CicdReleaseStatusQueued {
		return fmt.Errorf("发布单当前状态为 %s，无法删除，请先取消", rel.Status)
	}

	return s.dao.CicdReleaseDelete(ctx, releaseID)
}

// CicdReleaseCancelResult 取消操作结果
type CicdReleaseCancelResult struct {
	Action          string `json:"action"`            // "canceled" 或 "rollback"
	RollbackReleaseID int64  `json:"rollback_release_id,omitempty"` // 回滚时返回新发布单ID
}

// CicdReleaseCancel 取消发布单（智能判断：已部署的触发回滚，未部署的直接取消）
func (s *Services) CicdReleaseCancel(ctx context.Context, releaseID int64, userID int64) (*CicdReleaseCancelResult, error) {
	// 1. 获取发布单
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		return nil, fmt.Errorf("获取发布单失败: %w", err)
	}

	// 2. 已经是终态，不能取消
	if rel.Status == models.CicdReleaseStatusCanceled ||
		rel.Status == models.CicdReleaseStatusRollback ||
		rel.Status == models.CicdReleaseStatusFailed {
		return nil, fmt.Errorf("发布单已经是终态: %s，无法取消", rel.Status)
	}

	// 3. 如果是 Succeeded 或 Running，触发回滚
	if rel.Status == models.CicdReleaseStatusSucceeded || rel.Status == models.CicdReleaseStatusRunning {
		rollbackID, err := s.CicdReleaseRollback(ctx, releaseID, userID)
		if err != nil {
			return nil, fmt.Errorf("触发回滚失败: %w", err)
		}
		return &CicdReleaseCancelResult{
			Action:          "rollback",
			RollbackReleaseID: rollbackID,
		}, nil
	}

	// 4. 其他状态（Pending/Queued/AwaitingApproval），直接取消
	ok, err := s.dao.CicdReleaseCancel(ctx, releaseID)
	if err != nil {
		return nil, fmt.Errorf("取消发布单失败: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("发布单状态已变化，无法取消")
	}

	// 5. 同时取消所有未完成的任务
	_ = s.dao.CicdTasksFailByRelease(ctx, releaseID, "release canceled")

	return &CicdReleaseCancelResult{
		Action: "canceled",
	}, nil
}

// CicdReleaseRollback 回滚发布单（将已部署的工作负载回滚到上一个版本）
func (s *Services) CicdReleaseRollback(ctx context.Context, releaseID int64, userID int64) (int64, error) {
	// 1. 获取原发布单
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		return 0, fmt.Errorf("获取发布单失败: %w", err)
	}

	// 2. 检查发布单状态（只有成功或运行中的发布单才能回滚）
	if rel.Status != models.CicdReleaseStatusSucceeded && rel.Status != models.CicdReleaseStatusRunning {
		return 0, fmt.Errorf("发布单状态不支持回滚: %s，仅支持 Succeeded/Running 状态", rel.Status)
	}

	// 3. 获取已执行成功的任务（有 PrevImage 的任务）
	tasks, err := s.dao.CicdTasksByReleaseID(ctx, releaseID)
	if err != nil {
		return 0, fmt.Errorf("获取任务列表失败: %w", err)
	}

	// 4. 筛选有 PrevImage 的任务（说明已经执行过）
	var rollbackTasks []*models.CicdReleaseTask
	for _, t := range tasks {
		if t.PrevImage != "" && t.Status == models.CicdTaskStatusSucceeded {
			rollbackTasks = append(rollbackTasks, t)
		}
	}

	if len(rollbackTasks) == 0 {
		return 0, fmt.Errorf("没有可回滚的任务，发布单可能未执行或已失败")
	}

	// 5. 提取第一个任务的 PrevImage 作为回滚目标
	// 注：所有任务的 PrevImage 应该相同（同一次发布）
	rollbackImage := rollbackTasks[0].PrevImage

	// 6. 提取集群 ID
	clusterIDs := make([]int64, 0, len(rollbackTasks))
	for _, t := range rollbackTasks {
		clusterIDs = append(clusterIDs, t.ClusterID)
	}

	// 7. 解析回滚镜像的 repo 和 tag
	imageRepo, imageTag := parseImage(rollbackImage)

	// 8. 创建回滚发布单
	rollbackReq := &requests.CicdReleaseCreateRequest{
		AppName:       rel.AppName + "-rollback",
		Namespace:     rel.Namespace,
		WorkloadKind:  rel.WorkloadKind,
		WorkloadName:  rel.WorkloadName,
		ContainerName: rel.ContainerName,
		Strategy:      rel.Strategy,
		TimeoutSec:    rel.TimeoutSec,
		Concurrency:   rel.Concurrency,
		ImageRepo:     imageRepo,
		ImageTag:      imageTag,
		ClusterIDs:    clusterIDs,
	}

	newID, err := s.CicdReleaseCreate(ctx, rollbackReq, userID)
	if err != nil {
		return 0, fmt.Errorf("创建回滚发布单失败: %w", err)
	}

	// 9. 标记原发布单为已回滚状态
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(
		ctx,
		releaseID,
		[]string{models.CicdReleaseStatusSucceeded, models.CicdReleaseStatusRunning},
		models.CicdReleaseStatusRollback,
		fmt.Sprintf("rolled back to release %d", newID),
	)

	return newID, nil
}

// parseImage 解析镜像地址为 repo 和 tag
func parseImage(image string) (repo, tag string) {
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		return image[:idx], image[idx+1:]
	}
	return image, "latest"
}

// CicdReleaseRetry 重试发布单（创建新的发布单）
func (s *Services) CicdReleaseRetry(ctx context.Context, releaseID int64, userID int64) (int64, error) {
	// 获取原发布单
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		return 0, err
	}

	// 获取原任务的集群ID
	tasks, err := s.dao.CicdTasksByReleaseID(ctx, releaseID)
	if err != nil {
		return 0, err
	}

	clusterIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		clusterIDs = append(clusterIDs, t.ClusterID)
	}

	// 回退策略：如果原发布单没有关联任务（早期同步的数据），通过 BuildID 回查流水线配置获取集群ID
	if len(clusterIDs) == 0 && rel.BuildID > 0 {
		run, runErr := s.dao.PipelineRunGetByID(ctx, rel.BuildID)
		if runErr == nil && run != nil {
			pipeline, pErr := s.dao.PipelineGetByID(ctx, run.PipelineID)
			if pErr == nil && pipeline != nil && pipeline.TargetClusterID > 0 {
				clusterIDs = []int64{pipeline.TargetClusterID}
			}
		}
	}

	// 构建新的发布请求
	newReq := &requests.CicdReleaseCreateRequest{
		AppName:       rel.AppName,
		Namespace:     rel.Namespace,
		WorkloadKind:  rel.WorkloadKind,
		WorkloadName:  rel.WorkloadName,
		ContainerName: rel.ContainerName,
		Strategy:      rel.Strategy,
		TimeoutSec:    rel.TimeoutSec,
		Concurrency:   rel.Concurrency,
		ImageRepo:     rel.ImageRepo,
		ImageTag:      rel.ImageTag,
		ClusterIDs:    clusterIDs,
	}
	if rel.ImageDigest != nil {
		newReq.ImageDigest = *rel.ImageDigest
	}

	return s.CicdReleaseCreate(ctx, newReq, userID)
}

// CicdTasksByRelease 获取发布单下的任务列表
func (s *Services) CicdTasksByRelease(ctx context.Context, releaseID int64) ([]*models.CicdReleaseTask, error) {
	return s.dao.CicdTasksByReleaseID(ctx, releaseID)
}

// CicdBuildCallback 处理 Jenkins 构建回调
// 当 Jenkins 构建完成后，会调用此接口通知后端，后端根据构建结果决定是否触发发布
func (s *Services) CicdBuildCallback(ctx context.Context, req *requests.CicdBuildCallbackRequest) error {
	// 1. 根据 build_id 查找关联的发布单
	rel, err := s.dao.CicdReleaseGetByBuildID(ctx, req.BuildID)
	if err != nil {
		// 如果没有关联的发布单，说明是独立的 CI 构建，记录日志即可
		return fmt.Errorf("未找到关联的发布单: build_id=%d", req.BuildID)
	}

	// 2. 检查构建状态
	if req.Status != "SUCCESS" {
		// 构建失败，更新发布单状态
		_, _ = s.dao.CicdReleaseUpdateStatusCAS(
			ctx,
			rel.ID,
			[]string{models.CicdReleaseStatusPending, models.CicdReleaseStatusQueued},
			models.CicdReleaseStatusFailed,
			fmt.Sprintf("Jenkins构建失败: %s", req.Message),
		)
		return fmt.Errorf("构建失败: %s", req.Message)
	}

	// 3. 更新镜像信息（如果回调中包含新镜像信息）
	if req.ImageRepo != "" && req.ImageTag != "" {
		if err := s.dao.CicdReleaseUpdateImage(ctx, rel.ID, req.ImageRepo, req.ImageTag, req.ImageDigest); err != nil {
			return fmt.Errorf("更新镜像信息失败: %w", err)
		}
	}

	// 4. 获取任务列表并更新目标镜像
	tasks, err := s.dao.CicdTasksByReleaseID(ctx, rel.ID)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	// 5. 构建新的目标镜像
	targetImage := builder.BuildTargetImage(req.ImageRepo, req.ImageTag, req.ImageDigest)
	if targetImage == "" {
		targetImage = builder.BuildTargetImage(rel.ImageRepo, rel.ImageTag, "")
	}

	// 6. 更新所有任务的目标镜像并入队
	for _, task := range tasks {
		if task.Status != models.CicdTaskStatusPending {
			continue
		}

		// 更新目标镜像
		if err := s.dao.CicdTaskUpdateTargetImage(ctx, task.ID, targetImage); err != nil {
			continue
		}

		// 入队执行
		if _, err := s.stream.XAdd(ctx, infra.CicdDeployStream, map[string]any{
			"task_id":    task.ID,
			"release_id": rel.ID,
		}); err != nil {
			return fmt.Errorf("任务入队失败: task_id=%d, err=%w", task.ID, err)
		}
	}

	// 7. 更新发布单状态为 Queued
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(
		ctx,
		rel.ID,
		[]string{models.CicdReleaseStatusPending},
		models.CicdReleaseStatusQueued,
		"Jenkins构建成功，任务已入队",
	)

	return nil
}

// TryFinalizeRelease 尝试完结发布单（公开给 Worker 调用）
func (s *Services) TryFinalizeRelease(ctx context.Context, releaseID int64) {
	s.tryFinalizeRelease(ctx, releaseID)
}

// ========== 批量操作 ==========

// BatchRetryResult 批量发布结果
type BatchRetryResult struct {
	ID       int64  `json:"id"`
	NewID    int64  `json:"new_id,omitempty"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// CicdReleaseBatchRetry 批量重新发布（根据最近一次发布记录发布）
func (s *Services) CicdReleaseBatchRetry(ctx context.Context, ids []int64, userID int64) ([]BatchRetryResult, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("发布单ID列表不能为空")
	}

	results := make([]BatchRetryResult, 0, len(ids))
	for _, id := range ids {
		newID, err := s.CicdReleaseRetry(ctx, id, userID)
		if err != nil {
			results = append(results, BatchRetryResult{
				ID:      id,
				Success: false,
				Message: err.Error(),
			})
		} else {
			results = append(results, BatchRetryResult{
				ID:      id,
				NewID:   newID,
				Success: true,
				Message: "发布成功",
			})
		}
	}
	return results, nil
}

// BatchRollbackResult 批量回滚结果
type BatchRollbackResult struct {
	ID       int64  `json:"id"`
	NewID    int64  `json:"new_id,omitempty"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// CicdReleaseBatchRollback 批量回滚（根据最近一次发布记录回滚到上一个版本）
func (s *Services) CicdReleaseBatchRollback(ctx context.Context, ids []int64, userID int64) ([]BatchRollbackResult, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("发布单ID列表不能为空")
	}

	results := make([]BatchRollbackResult, 0, len(ids))
	for _, id := range ids {
		newID, err := s.CicdReleaseRollback(ctx, id, userID)
		if err != nil {
			results = append(results, BatchRollbackResult{
				ID:      id,
				Success: false,
				Message: err.Error(),
			})
		} else {
			results = append(results, BatchRollbackResult{
				ID:      id,
				NewID:   newID,
				Success: true,
				Message: "回滚成功",
			})
		}
	}
	return results, nil
}

// BatchCancelResult 批量取消结果
type BatchCancelResult struct {
	ID      int64  `json:"id"`
	Action  string `json:"action"` // "canceled" 或 "rollback"
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CicdReleaseBatchCancel 批量取消发布单（智能判断：已部署的触发回滚，未部署的直接取消）
func (s *Services) CicdReleaseBatchCancel(ctx context.Context, ids []int64, userID int64) ([]BatchCancelResult, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("发布单ID列表不能为空")
	}

	results := make([]BatchCancelResult, 0, len(ids))
	for _, id := range ids {
		cancelResult, err := s.CicdReleaseCancel(ctx, id, userID)
		if err != nil {
			results = append(results, BatchCancelResult{
				ID:      id,
				Success: false,
				Message: err.Error(),
			})
		} else {
			results = append(results, BatchCancelResult{
				ID:      id,
				Action:  cancelResult.Action,
				Success: true,
				Message: "取消成功",
			})
		}
	}
	return results, nil
}

// CicdReleaseSyncFromPipeline 将未同步的流水线运行记录同步到发布管理
func (s *Services) CicdReleaseSyncFromPipeline(ctx context.Context) (int, error) {
	// 获取未同步的已完成运行记录（最多50条）
	runs, err := s.dao.PipelineRunListCompletedUnsynced(ctx, 50)
	if err != nil {
		return 0, fmt.Errorf("查询未同步运行记录失败: %w", err)
	}

	if len(runs) == 0 {
		return 0, nil
	}

	synced := 0
	for _, run := range runs {
		// 获取对应的流水线信息
		pipeline, err := s.dao.PipelineGetByID(ctx, run.PipelineID)
		if err != nil {
			continue
		}

		// 转换状态
		releaseStatus := models.CicdReleaseStatusSucceeded
		if run.Status == models.PipelineRunStatusFailed {
			releaseStatus = models.CicdReleaseStatusFailed
		}

		// 解析镜像
		imageRepo := run.ImageURL
		imageTag := ""
		if run.ImageURL != "" {
			if idx := strings.LastIndex(run.ImageURL, ":"); idx > 0 && !strings.Contains(run.ImageURL[idx:], "/") {
				imageRepo = run.ImageURL[:idx]
				imageTag = run.ImageURL[idx+1:]
			}
		}

		workloadKind := pipeline.TargetWorkloadKind
		if workloadKind == "" {
			workloadKind = "Deployment"
		}
		namespace := pipeline.TargetNamespace
		if namespace == "" {
			namespace = "default"
		}

		// 使用运行记录的创建时间作为发布时间
		createdAt := run.CreatedAt
		if createdAt == 0 {
			createdAt = uint64(time.Now().Unix())
		}

		release := &models.CicdRelease{
			AppName:       pipeline.Name,
			Namespace:     namespace,
			WorkloadKind:  workloadKind,
			WorkloadName:  pipeline.TargetWorkloadName,
			ContainerName: pipeline.TargetContainer,
			Strategy:      "rolling",
			TimeoutSec:    300,
			Status:        releaseStatus,
			Message:       fmt.Sprintf("流水线同步: %s #%d", pipeline.Name, run.BuildNumber),
			CreatedUserID: run.TriggerUserID,
			RequestID:     fmt.Sprintf("pipeline-sync-%d", run.ID),
			BuildID:       run.ID,
			ImageRepo:     imageRepo,
			ImageTag:      imageTag,
			CreatedAt:     createdAt,
			ModifiedAt:    uint64(time.Now().Unix()),
		}
		if run.ImageDigest != "" {
			digest := run.ImageDigest
			release.ImageDigest = &digest
		}

		if err := s.dao.CicdReleaseCreate(ctx, release); err == nil {
			synced++
		}
	}

	return synced, nil
}
