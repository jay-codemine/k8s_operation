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
func (s *Services) ApprovalList(ctx context.Context, param *requests.ApprovalListRequest) ([]*models.CicdApproval, int64, error) {
	return s.dao.ApprovalList(ctx, param.Page, param.PageSize, param.Status)
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

	// 如果是通过审批，触发实际部署
	if status == models.ApprovalStatusApproved {
		// 获取流水线信息
		pipeline, err := s.dao.PipelineGetByID(ctx, approval.PipelineID)
		if err == nil && pipeline != nil && pipeline.AutoDeploy {
			// 通过 PipelineCallback 处理部署（已有完整的部署逻辑）
			// 这里可以记录日志或触发异步部署
			_ = pipeline // 审批通过，后续由用户手动触发部署
		}
	}

	return nil
}

// ApprovalPendingList 获取待审批列表
func (s *Services) ApprovalPendingList(ctx context.Context, userID int64) ([]*models.CicdApproval, int64, error) {
	// TODO: 可以根据用户权限过滤
	return s.dao.ApprovalList(ctx, 1, 100, models.ApprovalStatusPending)
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
