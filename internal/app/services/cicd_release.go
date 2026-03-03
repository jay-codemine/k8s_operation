package services

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
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

	// 入队：逐个 task 写入 Redis Stream
	enqueued := 0
	for _, t := range tasks {
		if _, err := s.stream.XAdd(ctx, infra.CicdDeployStream, map[string]any{
			"task_id":    t.ID,
			"release_id": rel.ID,
		}); err != nil {

			// 1) Release 标 Failed（CAS，避免并发覆盖终态）
			_, _ = s.dao.CicdReleaseUpdateStatusCAS(
				ctx,
				rel.ID,
				[]string{"Pending", "Queued"},
				"Failed",
				fmt.Sprintf("enqueue failed after %d/%d", enqueued, len(tasks)),
			)

			// 2) 止血：把仍未执行的任务（Pending/Queued）批量标 Failed，避免悬挂
			_ = s.dao.CicdTasksFailByRelease(
				ctx,
				rel.ID,
				fmt.Sprintf("enqueue failed after %d/%d", enqueued, len(tasks)),
			)

			return 0, err
		}

		enqueued++
	}

	// 全部入队成功：Release 标 Queued（CAS）
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(
		ctx,
		rel.ID,
		[]string{"Pending"},
		"Queued",
		"enqueued",
	)

	return rel.ID, nil
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

	// 4. 其他状态（Pending/Queued），直接取消
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
