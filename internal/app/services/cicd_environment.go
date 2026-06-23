package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 环境管理 Service ====================

// EnvironmentList 获取环境列表
func (s *Services) EnvironmentList(ctx context.Context, param *requests.EnvironmentListRequest) ([]*models.EnvironmentListItem, int64, error) {
	return s.dao.EnvironmentList(ctx, param.Page, param.PageSize, param.Keyword)
}

// EnvironmentDetail 获取环境详情
func (s *Services) EnvironmentDetail(ctx context.Context, id int64) (*models.CicdEnvironment, error) {
	return s.dao.EnvironmentGetByID(ctx, id)
}

// EnvironmentCreate 创建环境
func (s *Services) EnvironmentCreate(ctx context.Context, param *requests.EnvironmentCreateRequest, userID int64) (int64, error) {
	// 检查环境名称是否已存在
	existing, err := s.dao.EnvironmentGetByName(ctx, param.Name)
	if err == nil && existing != nil && existing.ID > 0 {
		return 0, errors.New("环境名称已存在")
	}

	// 构造审批人员JSON（兼容旧接口）
	var approvalUsers models.JSONMap
	if len(param.ApprovalUserIDs) > 0 {
		approvalUsers = models.JSONMap{"user_ids": param.ApprovalUserIDs}
	}

	// 多级审批配置
	var approvalLevels models.ApprovalLevels
	if len(param.ApprovalLevels) > 0 {
		approvalLevels = models.ApprovalLevels(param.ApprovalLevels)
	}

	now := time.Now().Unix()
	env := &models.CicdEnvironment{
		Name:            param.Name,
		DisplayName:     param.DisplayName,
		Description:     param.Description,
		ClusterID:       param.ClusterID,
		Namespace:       param.Namespace,
		Color:           param.Color,
		SortOrder:       param.SortOrder,
		RequireApproval: param.RequireApproval,
		ApprovalUsers:   approvalUsers,
		ApprovalLevels:  approvalLevels,
		CreatedUserID:   userID,
		CreatedAt:       uint64(now),
		ModifiedAt:      uint64(now),
	}

	// 设置默认颜色
	if env.Color == "" {
		switch param.Name {
		case "dev":
			env.Color = "#52c41a" // 绿色
		case "staging":
			env.Color = "#faad14" // 橙色
		case "prod":
			env.Color = "#f5222d" // 红色
		default:
			env.Color = "#1890ff" // 蓝色
		}
	}

	return s.dao.EnvironmentCreate(ctx, env)
}

// EnvironmentUpdate 更新环境
func (s *Services) EnvironmentUpdate(ctx context.Context, param *requests.EnvironmentUpdateRequest) error {
	env, err := s.dao.EnvironmentGetByID(ctx, param.ID)
	if err != nil {
		return errors.New("环境不存在")
	}

	// 检查名称是否与其他环境冲突
	if param.Name != "" && param.Name != env.Name {
		existing, err := s.dao.EnvironmentGetByName(ctx, param.Name)
		if err == nil && existing != nil && existing.ID > 0 && existing.ID != param.ID {
			return errors.New("环境名称已存在")
		}
		env.Name = param.Name
	}

	if param.DisplayName != "" {
		env.DisplayName = param.DisplayName
	}
	if param.Description != "" {
		env.Description = param.Description
	}
	if param.ClusterID != nil {
		env.ClusterID = *param.ClusterID
	}
	if param.Namespace != "" {
		env.Namespace = param.Namespace
	}
	if param.Color != "" {
		env.Color = param.Color
	}
	if param.SortOrder != nil {
		env.SortOrder = *param.SortOrder
	}
	if param.RequireApproval != nil {
		env.RequireApproval = *param.RequireApproval
	}
	if len(param.ApprovalUserIDs) > 0 {
		approvalUsers := models.JSONMap{"user_ids": param.ApprovalUserIDs}
		env.ApprovalUsers = approvalUsers
	}
	// 更新多级审批配置（允许传空数组清除）
	if param.ApprovalLevels != nil {
		env.ApprovalLevels = models.ApprovalLevels(param.ApprovalLevels)
	}

	env.ModifiedAt = uint64(time.Now().Unix())

	return s.dao.EnvironmentUpdate(ctx, env)
}

// EnvironmentDelete 删除环境
func (s *Services) EnvironmentDelete(ctx context.Context, id int64) error {
	return s.dao.EnvironmentDelete(ctx, id)
}

// ==================== 审批流程 Service ====================

// ApprovalList 获取审批列表
func (s *Services) ApprovalList(ctx context.Context, param *requests.ApprovalListRequest) ([]*models.ApprovalListItem, int64, error) {
	return s.dao.ApprovalList(ctx, param.Page, param.PageSize, param.Status, param.PipelineID)
}

// ApprovalListByUser 获取指定用户提交的审批列表
func (s *Services) ApprovalListByUser(ctx context.Context, userID int64, param *requests.ApprovalListRequest) ([]*models.ApprovalListItem, int64, error) {
	return s.dao.ApprovalListByUser(ctx, userID, param.Page, param.PageSize, param.Status, param.PipelineID)
}

// ApprovalStats 获取审批统计
func (s *Services) ApprovalStats(ctx context.Context) (map[string]int64, error) {
	return s.dao.ApprovalStats(ctx)
}

// ApprovalStatsByUser 获取指定用户的审批统计
func (s *Services) ApprovalStatsByUser(ctx context.Context, userID int64) (map[string]int64, error) {
	return s.dao.ApprovalStatsByUser(ctx, userID)
}

// ApprovalDetail 获取审批详情
func (s *Services) ApprovalDetail(ctx context.Context, id int64) (*models.CicdApproval, error) {
	return s.dao.ApprovalGetByID(ctx, id)
}

// ApprovalCreate 创建审批申请
func (s *Services) ApprovalCreate(ctx context.Context, param *requests.ApprovalCreateRequest, userID int64) (int64, error) {
	// 检查是否已有待审批记录
	existing, err := s.dao.ApprovalGetPendingByPipeline(ctx, param.PipelineID)
	if err == nil && existing != nil && existing.ID > 0 {
		return 0, errors.New("该流水线已有待审批的部署申请")
	}

	now := time.Now().Unix()
	approval := &models.CicdApproval{
		PipelineID:    param.PipelineID,
		PipelineRunID: param.PipelineRunID,
		EnvName:       param.EnvName,
		Image:         param.Image,
		ImageDigest:   param.ImageDigest,
		Status:        models.ApprovalStatusPending,
		RequestUserID: userID,
		RequestReason: param.RequestReason,
		ExpireTime:    uint64(now + 86400*7), // 7天过期
		CreatedAt:     uint64(now),
		ModifiedAt:    uint64(now),
	}

	return s.dao.ApprovalCreate(ctx, approval)
}

// ApprovalAction 审批操作
func (s *Services) ApprovalAction(ctx context.Context, param *requests.ApprovalActionRequest, userID int64) error {
	approval, err := s.dao.ApprovalGetByID(ctx, param.ID)
	if err != nil {
		return errors.New("审批记录不存在")
	}

	if approval.Status != models.ApprovalStatusPending {
		return errors.New("该审批已处理，无法重复操作")
	}

	// 权限分离：审批人不能审批自己提交的申请
	if approval.RequestUserID == userID {
		return errors.New("不能审批自己提交的申请，请联系其他审批人处理")
	}

	// 检查是否过期
	if approval.ExpireTime > 0 && uint64(time.Now().Unix()) > approval.ExpireTime {
		// 更新状态为已过期
		_ = s.dao.ApprovalUpdateStatus(ctx, param.ID, models.ApprovalStatusExpired, 0, "")
		return errors.New("该审批申请已过期")
	}

	var status string
	if param.Action == "approve" {
		status = models.ApprovalStatusApproved
	} else {
		status = models.ApprovalStatusRejected
	}

	err = s.dao.ApprovalUpdateStatus(ctx, param.ID, status, userID, param.Reason)
	if err != nil {
		return err
	}

	// 同步更新关联的流水线审批阶段（cicd_pipeline_stage）
	if approval.StageID > 0 {
		if status == models.ApprovalStatusApproved {
			// ====== 多级审批级联逻辑 ======
			if approval.ApprovalLevel < approval.TotalLevels {
				// 还有下一级，创建下一级审批
				nextID, nextErr := s.CreateNextLevelApproval(ctx, approval)
				if nextErr == nil && nextID > 0 {
					// 发送下一级审批飞书通知
					nextApproval, _ := s.dao.ApprovalGetByID(ctx, nextID)
					if nextApproval != nil {
						pipeline, _ := s.dao.PipelineGetByID(ctx, nextApproval.PipelineID)
						run, _ := s.dao.PipelineRunGetByID(ctx, nextApproval.PipelineRunID)
						if pipeline != nil && run != nil {
							s.NotifyFeishuApproval(ctx, pipeline, run, nextApproval)
						}
					}
				}
				return nil // 不触发部署，等待下一级审批
			}

			// 最后一级通过：更新阶段审批状态，并自动触发部署
			_ = s.dao.StageUpdateApproval(ctx, approval.StageID, userID, "approved", param.Reason)

			// 获取阶段信息以启动部署阶段
			stage, stageErr := s.dao.StageGetByID(ctx, approval.StageID)
			if stageErr == nil && stage != nil {
				// 检查是否有部署阶段需要启动
				deployStage, dErr := s.dao.StageGetByRunIDAndType(ctx, stage.RunID, models.StageTypeDeploy)
				if dErr == nil && deployStage != nil {
					run, _ := s.dao.PipelineRunGetByID(ctx, stage.RunID)
					if run != nil && run.ImageURL != "" {
						_ = s.dao.StageUpdate(ctx, deployStage.ID, map[string]interface{}{
							"status":       models.StageStatusPending,
							"deploy_image": run.ImageURL,
						})

						// ====== 审批通过后自动触发部署（核心修复） ======
						global.Logger.Info("[审批] 审批通过，自动触发部署",
							zap.Int64("approval_id", param.ID),
							zap.Int64("deploy_stage_id", deployStage.ID),
							zap.String("image", run.ImageURL),
						)
						go func() {
							deployReq := &requests.StageDeployRequest{
								StageID: deployStage.ID,
							}
							if err := s.ExecuteDeployStage(context.Background(), deployReq, userID); err != nil {
								global.Logger.Error("[审批] 自动触发部署失败",
									zap.Int64("deploy_stage_id", deployStage.ID),
									zap.Error(err),
								)
							}
						}()
					}
				}
			}
		} else {
			// 拒绝：更新阶段审批状态，标记后续阶段为跳过，更新流水线状态为失败
			_ = s.dao.StageUpdateApproval(ctx, approval.StageID, userID, "rejected", param.Reason)

			stage, stageErr := s.dao.StageGetByID(ctx, approval.StageID)
			if stageErr == nil && stage != nil {
				// 将后续阶段标记为跳过
				stages, _ := s.dao.StageListByRunID(ctx, stage.RunID)
				for _, stg := range stages {
					if stg.StageOrder > stage.StageOrder && stg.Status == models.StageStatusPending {
						_ = s.dao.StageUpdateStatus(ctx, stg.ID, models.StageStatusSkipped)
					}
				}

				// 更新流水线运行状态为失败
				_ = s.dao.PipelineRunUpdateStatus(ctx, stage.RunID, models.PipelineRunStatusFailed)
				run, _ := s.dao.PipelineRunGetByID(ctx, stage.RunID)
				if run != nil {
					_ = s.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed)
				}
			}
		}
	} else if approval.ReleaseID > 0 {
		// ====== 手动发布单多级审批逻辑 ======
		if status == models.ApprovalStatusApproved {
			// 多级审批级联：当前级未到最后一级，创建下一级审批
			if approval.ApprovalLevel < approval.TotalLevels {
				nextID, nextErr := s.CreateNextLevelApproval(ctx, approval)
				if nextErr == nil && nextID > 0 {
					global.Logger.Info("[审批] 发布单当前级审批通过，已创建下一级审批",
						zap.Int64("release_id", approval.ReleaseID),
						zap.Int("current_level", approval.ApprovalLevel),
						zap.Int("total_levels", approval.TotalLevels),
						zap.Int64("next_approval_id", nextID),
					)
				} else if nextErr != nil {
					global.Logger.Error("[审批] 创建下一级审批失败",
						zap.Int64("release_id", approval.ReleaseID),
						zap.Error(nextErr),
					)
					return fmt.Errorf("创建下一级审批失败: %w", nextErr)
				}
				return nil // 不触发部署，等待下一级审批
			}

			// 最后一级审批通过：将发布单任务入队执行部署
			tasks, tErr := s.dao.CicdTasksByReleaseID(ctx, approval.ReleaseID)
			if tErr != nil || len(tasks) == 0 {
				global.Logger.Error("[审批] 获取发布单任务失败",
					zap.Int64("release_id", approval.ReleaseID),
					zap.Error(tErr),
				)
				return fmt.Errorf("发布单任务不存在(release_id=%d)", approval.ReleaseID)
			}
			_, enqErr := s.releaseEnqueue(ctx, approval.ReleaseID, tasks)
			if enqErr != nil {
				global.Logger.Error("[审批] 发布单入队失败",
					zap.Int64("release_id", approval.ReleaseID),
					zap.Error(enqErr),
				)
				return fmt.Errorf("全部审批通过但入队失败: %w", enqErr)
			}
			global.Logger.Info("[审批] 发布单全部审批通过，已入队部署",
				zap.Int64("release_id", approval.ReleaseID),
				zap.Int64("approval_id", param.ID),
				zap.Int("total_levels", approval.TotalLevels),
			)
		} else {
			// 审批拒绝：将发布单标记为取消
			reason := "审批被拒绝"
			if param.Reason != "" {
				reason = "审批拒绝: " + param.Reason
			}
			_, _ = s.dao.CicdReleaseUpdateStatusCAS(
				ctx,
				approval.ReleaseID,
				[]string{models.CicdReleaseStatusAwaitingApproval, models.CicdReleaseStatusPending},
				models.CicdReleaseStatusCanceled,
				reason,
			)
			global.Logger.Info("[审批] 发布单审批拒绝，已取消",
				zap.Int64("release_id", approval.ReleaseID),
				zap.Int64("approval_id", param.ID),
			)
		}
	}

	return nil
}

// ApprovalPendingList 获取待审批列表
func (s *Services) ApprovalPendingList(ctx context.Context, userID int64) ([]*models.ApprovalListItem, int64, error) {
	// TODO: 可以根据用户权限过滤
	return s.dao.ApprovalList(ctx, 1, 100, models.ApprovalStatusPending, 0)
}

// ApprovalBatchAction 批量审批操作
func (s *Services) ApprovalBatchAction(ctx context.Context, ids []int64, action string, reason string, userID int64) (int, []string) {
	success := 0
	var failures []string

	for _, id := range ids {
		param := &requests.ApprovalActionRequest{
			ID:     id,
			Action: action,
			Reason: reason,
		}
		if err := s.ApprovalAction(ctx, param, userID); err != nil {
			failures = append(failures, fmt.Sprintf("#%d: %s", id, err.Error()))
		} else {
			success++
		}
	}

	return success, failures
}

// ApprovalUpdate 更新审批记录
func (s *Services) ApprovalUpdate(ctx context.Context, param *requests.ApprovalUpdateRequest) error {
	approval, err := s.dao.ApprovalGetByID(ctx, param.ID)
	if err != nil {
		return errors.New("审批记录不存在")
	}

	// 只有待审批状态的记录可以编辑
	if approval.Status != models.ApprovalStatusPending {
		return errors.New("该审批已处理，无法编辑")
	}

	updates := make(map[string]interface{})
	if param.EnvName != "" {
		updates["env_name"] = param.EnvName
	}
	if param.Image != "" {
		updates["image"] = param.Image
	}
	if param.ImageDigest != "" {
		updates["image_digest"] = param.ImageDigest
	}
	if param.RequestReason != "" {
		updates["request_reason"] = param.RequestReason
	}

	if len(updates) == 0 {
		return errors.New("无有效的更新字段")
	}

	return s.dao.ApprovalUpdate(ctx, param.ID, updates)
}

// ApprovalDelete 删除审批记录
func (s *Services) ApprovalDelete(ctx context.Context, id int64) error {
	approval, err := s.dao.ApprovalGetByID(ctx, id)
	if err != nil {
		return errors.New("审批记录不存在")
	}

	// 已通过的审批不允许删除（保护已执行的审批流程）
	if approval.Status == models.ApprovalStatusApproved {
		return errors.New("已通过的审批不允许删除")
	}

	return s.dao.ApprovalDelete(ctx, id)
}

// CheckAndCreateApproval 检查是否需要审批，如果需要则创建审批记录
// 支持多级审批：先创建第一级审批，通过后级联触发下一级
func (s *Services) CheckAndCreateApproval(ctx context.Context, pipeline *models.CicdPipeline, image, digest string, userID int64) (bool, int64, error) {
	// 判断是否需要审批：pipeline 标志 OR 环境配置（双重保险，防止前端漏设导致绕过审批）
	needApproval := pipeline.RequireApproval

	// 查询环境配置，获取审批策略
	env, _ := s.dao.EnvironmentGetByName(ctx, pipeline.DeployEnv)

	// 如果 pipeline 未开启审批，但环境配置要求审批，则强制开启
	if !needApproval && env != nil && env.RequireApproval {
		needApproval = true
	}

	if !needApproval {
		return false, 0, nil
	}

	totalLevels := 1
	levelLabel := "审批"
	if env != nil && len(env.ApprovalLevels) > 0 {
		totalLevels = len(env.ApprovalLevels)
		if len(env.ApprovalLevels) > 0 {
			levelLabel = env.ApprovalLevels[0].Label
		}
	}

	// 创建第一级审批记录
	now := time.Now().Unix()
	approval := &models.CicdApproval{
		PipelineID:    pipeline.ID,
		EnvName:       pipeline.DeployEnv,
		Image:         image,
		ImageDigest:   digest,
		Status:        models.ApprovalStatusPending,
		RequestUserID: userID,
		RequestReason: "构建成功，申请部署到" + pipeline.DeployEnv + "环境",
		ExpireTime:    uint64(now + 86400*7),
		ApprovalLevel: 1,
		TotalLevels:   totalLevels,
		LevelLabel:    levelLabel,
		CreatedAt:     uint64(now),
		ModifiedAt:    uint64(now),
	}

	id, err := s.dao.ApprovalCreate(ctx, approval)
	if err != nil {
		return true, 0, err
	}

	return true, id, nil
}

// CreateNextLevelApproval 创建下一级审批记录（当前级通过后调用）
func (s *Services) CreateNextLevelApproval(ctx context.Context, prevApproval *models.CicdApproval) (int64, error) {
	nextLevel := prevApproval.ApprovalLevel + 1
	if nextLevel > prevApproval.TotalLevels {
		return 0, nil // 已是最后一级，无需创建
	}

	// 查询环境配置获取下一级信息
	env, _ := s.dao.EnvironmentGetByName(ctx, prevApproval.EnvName)
	levelLabel := "审批"
	if env != nil && len(env.ApprovalLevels) >= nextLevel {
		levelLabel = env.ApprovalLevels[nextLevel-1].Label
	}

	now := time.Now().Unix()
	approval := &models.CicdApproval{
		PipelineID:    prevApproval.PipelineID,
		PipelineRunID: prevApproval.PipelineRunID,
		StageID:       prevApproval.StageID,
		ReleaseID:     prevApproval.ReleaseID,
		EnvName:       prevApproval.EnvName,
		Image:         prevApproval.Image,
		ImageDigest:   prevApproval.ImageDigest,
		Status:        models.ApprovalStatusPending,
		RequestUserID: prevApproval.RequestUserID,
		RequestReason: prevApproval.RequestReason,
		ExpireTime:    uint64(now + 86400*7),
		ApprovalLevel: nextLevel,
		TotalLevels:   prevApproval.TotalLevels,
		LevelLabel:    levelLabel,
		CreatedAt:     uint64(now),
		ModifiedAt:    uint64(now),
	}

	return s.dao.ApprovalCreate(ctx, approval)
}
