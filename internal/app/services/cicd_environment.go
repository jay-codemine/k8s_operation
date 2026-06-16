package services

import (
	"context"
	"errors"
	"time"

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

	// 构造审批人员JSON
	var approvalUsers models.JSONMap
	if len(param.ApprovalUserIDs) > 0 {
		approvalUsers = models.JSONMap{"user_ids": param.ApprovalUserIDs}
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

// ApprovalStats 获取审批统计
func (s *Services) ApprovalStats(ctx context.Context) (map[string]int64, error) {
	return s.dao.ApprovalStats(ctx)
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
			// 通过：更新阶段审批状态，并启动后续部署阶段
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
	}

	return nil
}

// ApprovalPendingList 获取待审批列表
func (s *Services) ApprovalPendingList(ctx context.Context, userID int64) ([]*models.ApprovalListItem, int64, error) {
	// TODO: 可以根据用户权限过滤
	return s.dao.ApprovalList(ctx, 1, 100, models.ApprovalStatusPending, 0)
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
func (s *Services) CheckAndCreateApproval(ctx context.Context, pipeline *models.CicdPipeline, image, digest string, userID int64) (bool, int64, error) {
	// 如果不需要审批，直接返回
	if !pipeline.RequireApproval {
		return false, 0, nil
	}

	// 检查环境是否需要审批
	if pipeline.DeployEnv == "prod" || pipeline.RequireApproval {
		// 创建审批记录
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
			CreatedAt:     uint64(now),
			ModifiedAt:    uint64(now),
		}

		id, err := s.dao.ApprovalCreate(ctx, approval)
		if err != nil {
			return true, 0, err
		}

		return true, id, nil
	}

	return false, 0, nil
}
