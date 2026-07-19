package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 环境目标管理（cicd_pipeline_target）====================

// PipelineTargetList 获取某条流水线的全部环境部署目标
func (s *Services) PipelineTargetList(ctx context.Context, pipelineID int64) ([]*models.CicdPipelineTargetView, error) {
	return s.dao.PipelineTargetListByPipeline(ctx, pipelineID)
}

// PipelineTargetSave 全量保存某条流水线的环境部署目标（存在则更新，缺失则新增，多余则软删除）
func (s *Services) PipelineTargetSave(ctx context.Context, req *requests.PipelineTargetSaveRequest, userID int64) error {
	// 校验流水线存在
	if _, err := s.dao.PipelineGetByID(ctx, req.PipelineID); err != nil {
		return fmt.Errorf("关联流水线不存在(id=%d): %w", req.PipelineID, err)
	}

	// 现有目标：用于计算需要软删除的环境
	existing, err := s.dao.PipelineTargetListByPipeline(ctx, req.PipelineID)
	if err != nil {
		return fmt.Errorf("查询现有环境目标失败: %w", err)
	}

	// 提交的环境集合
	submitted := make(map[string]struct{}, len(req.Targets))
	for _, item := range req.Targets {
		env := strings.TrimSpace(item.Env)
		if env == "" {
			return fmt.Errorf("环境标识(env)不能为空")
		}
		if item.ClusterID <= 0 {
			return fmt.Errorf("环境[%s]未选择目标集群", env)
		}
		if strings.TrimSpace(item.WorkloadName) == "" {
			return fmt.Errorf("环境[%s]未填写目标工作负载名称", env)
		}
		submitted[env] = struct{}{}

		workloadKind := strings.TrimSpace(item.WorkloadKind)
		if workloadKind == "" {
			workloadKind = "Deployment"
		}
		namespace := strings.TrimSpace(item.Namespace)
		if namespace == "" {
			namespace = "default"
		}
		container := strings.TrimSpace(item.Container)
		if container == "" {
			container = item.WorkloadName
		}

		t := &models.CicdPipelineTarget{
			PipelineID:      req.PipelineID,
			Env:             env,
			ClusterID:       item.ClusterID,
			Namespace:       namespace,
			WorkloadKind:    workloadKind,
			WorkloadName:    item.WorkloadName,
			Container:       container,
			AutoDeploy:      item.AutoDeploy,
			RequireApproval: item.RequireApproval,
			PromoteFrom:     strings.TrimSpace(item.PromoteFrom),
			SortOrder:       item.SortOrder,
			CreatedUserID:   userID,
		}
		if err := s.dao.PipelineTargetUpsert(ctx, t); err != nil {
			return fmt.Errorf("保存环境[%s]目标失败: %w", env, err)
		}
	}

	// 软删除已不在提交列表中的环境目标
	for _, old := range existing {
		if _, ok := submitted[old.Env]; !ok {
			_ = s.dao.PipelineTargetDelete(ctx, old.ID)
		}
	}

	return nil
}

// PipelineTargetDelete 删除单个环境目标
func (s *Services) PipelineTargetDelete(ctx context.Context, id int64) error {
	return s.dao.PipelineTargetDelete(ctx, id)
}

// ==================== 镜像晋级 ====================

// PipelinePromote 镜像晋级：将已构建的不可变镜像发布到目标环境。
//
// 核心复用：解析出「目标环境部署目标 + 待晋级镜像」后，组装成发布单请求交给
// CicdReleaseCreate —— 它已内置按环境的多级审批、多集群下发、回滚等能力，
// 因此晋级本身不重新构建镜像，实现 build once, promote everywhere。
func (s *Services) PipelinePromote(ctx context.Context, req *requests.PipelinePromoteRequest, userID int64) (int64, error) {
	// 1. 校验流水线
	pipeline, err := s.dao.PipelineGetByID(ctx, req.PipelineID)
	if err != nil {
		return 0, fmt.Errorf("关联流水线不存在(id=%d): %w", req.PipelineID, err)
	}

	targetEnv := strings.TrimSpace(req.TargetEnv)

	// 2. 解析目标环境部署目标（优先用 cicd_pipeline_target；回退到流水线自带单目标）
	appName := pipeline.Name
	var clusterID int64
	var namespace, workloadKind, workloadName, container string
	var sourceEnv string

	target, tErr := s.dao.PipelineTargetGetByPipelineAndEnv(ctx, req.PipelineID, targetEnv)
	if tErr == nil && target != nil {
		clusterID = target.ClusterID
		namespace = target.Namespace
		workloadKind = target.WorkloadKind
		workloadName = target.WorkloadName
		container = target.Container
		sourceEnv = target.PromoteFrom
	} else if pipeline.DeployEnv == targetEnv && pipeline.TargetClusterID > 0 {
		// 回退：流水线自带的单一部署目标恰好是该环境
		clusterID = pipeline.TargetClusterID
		namespace = pipeline.TargetNamespace
		workloadKind = pipeline.TargetWorkloadKind
		workloadName = pipeline.TargetWorkloadName
		container = pipeline.TargetContainer
	} else {
		return 0, fmt.Errorf("环境[%s]未配置部署目标，请先在「环境目标配置」中绑定集群/命名空间/工作负载", targetEnv)
	}

	if clusterID <= 0 {
		return 0, fmt.Errorf("环境[%s]未配置目标集群", targetEnv)
	}
	if workloadName == "" {
		return 0, fmt.Errorf("环境[%s]未配置目标工作负载", targetEnv)
	}
	if workloadKind == "" {
		workloadKind = "Deployment"
	}
	if namespace == "" {
		namespace = "default"
	}
	if container == "" {
		container = workloadName
	}

	// 3. 解析待晋级的不可变镜像（优先级：显式镜像 > source_release_id > source_run_id）
	imageRepo := strings.TrimSpace(req.ImageRepo)
	imageTag := strings.TrimSpace(req.ImageTag)
	imageDigest := strings.TrimSpace(req.ImageDigest)
	sourceRunID := req.SourceRunID

	switch {
	case imageRepo != "":
		// 使用显式传入的镜像
	case req.SourceReleaseID > 0:
		// 从已有发布单晋级（跨环境晋级：dev -> staging -> prod）
		srcRel, rErr := s.dao.CicdReleaseGetByID(ctx, req.SourceReleaseID)
		if rErr != nil {
			return 0, fmt.Errorf("来源发布单不存在(id=%d): %w", req.SourceReleaseID, rErr)
		}
		imageRepo = srcRel.ImageRepo
		imageTag = srcRel.ImageTag
		if srcRel.ImageDigest != nil {
			imageDigest = *srcRel.ImageDigest
		}
		if srcRel.Env != "" {
			sourceEnv = srcRel.Env
		}
		if srcRel.SourceRunID > 0 {
			sourceRunID = srcRel.SourceRunID
		}
	case req.SourceRunID > 0:
		// 从流水线运行记录（构建产物）晋级
		run, runErr := s.dao.PipelineRunGetByID(ctx, req.SourceRunID)
		if runErr != nil {
			return 0, fmt.Errorf("来源运行记录不存在(id=%d): %w", req.SourceRunID, runErr)
		}
		if run.ImageURL == "" {
			return 0, fmt.Errorf("运行记录 #%d 无构建产物镜像，无法晋级（请确认构建/推送阶段已成功）", req.SourceRunID)
		}
		imageRepo, imageTag = splitImageRepoTag(run.ImageURL)
		if run.ImageDigest != "" {
			imageDigest = run.ImageDigest
		}
	default:
		// 未显式指定来源：回退到「最新一次已构建产物」，对应前端「留空使用最新一次构建」
		run, runErr := s.dao.PipelineRunGetLatestBuilt(ctx, req.PipelineID)
		if runErr != nil {
			return 0, fmt.Errorf("流水线暂无可晋级的构建产物，请先成功执行一次构建后再晋级")
		}
		if run.ImageURL == "" {
			return 0, fmt.Errorf("最新构建（#%d）无产物镜像，无法晋级（请确认构建/推送阶段已成功）", run.ID)
		}
		imageRepo, imageTag = splitImageRepoTag(run.ImageURL)
		if run.ImageDigest != "" {
			imageDigest = run.ImageDigest
		}
		sourceRunID = run.ID
	}

	if imageRepo == "" {
		return 0, fmt.Errorf("无法解析镜像仓库地址，晋级已阻断")
	}
	if imageTag == "" && imageDigest == "" {
		return 0, fmt.Errorf("镜像 tag 与 digest 均为空，无法晋级")
	}

	// 4. 组装发布单请求，复用 CicdReleaseCreate（内含按环境审批 + 多集群下发 + 入队部署）
	message := strings.TrimSpace(req.Reason)
	if message == "" {
		message = fmt.Sprintf("镜像晋级: %s → %s", firstNonEmpty(sourceEnv, "build"), targetEnv)
	}

	releaseReq := &requests.CicdReleaseCreateRequest{
		PipelineID:    req.PipelineID,
		AppName:       appName,
		Namespace:     namespace,
		WorkloadKind:  workloadKind,
		WorkloadName:  workloadName,
		ContainerName: container,
		Strategy:      "rolling",
		TimeoutSec:    300,
		Concurrency:   3,
		ImageRepo:     imageRepo,
		ImageTag:      imageTag,
		ImageDigest:   imageDigest,
		ClusterIDs:    []int64{clusterID},
		Message:       message,
		Env:           targetEnv,
		SourceEnv:     sourceEnv,
		SourceRunID:   sourceRunID,
	}

	return s.CicdReleaseCreate(ctx, releaseReq, userID)
}

// PipelinePromotionChain 构建晋级链视图：各环境的部署目标 + 当前部署的镜像/发布单
func (s *Services) PipelinePromotionChain(ctx context.Context, pipelineID int64) ([]*models.PromotionEnvNode, error) {
	// 校验流水线
	if _, err := s.dao.PipelineGetByID(ctx, pipelineID); err != nil {
		return nil, fmt.Errorf("关联流水线不存在(id=%d): %w", pipelineID, err)
	}

	targets, err := s.dao.PipelineTargetListByPipeline(ctx, pipelineID)
	if err != nil {
		return nil, err
	}

	nodes := make([]*models.PromotionEnvNode, 0, len(targets))
	for _, t := range targets {
		node := &models.PromotionEnvNode{
			Env:             t.Env,
			ClusterID:       t.ClusterID,
			ClusterName:     t.ClusterName,
			Namespace:       t.Namespace,
			WorkloadKind:    t.WorkloadKind,
			WorkloadName:    t.WorkloadName,
			Container:       t.Container,
			AutoDeploy:      t.AutoDeploy,
			RequireApproval: t.RequireApproval,
			PromoteFrom:     t.PromoteFrom,
			SortOrder:       t.SortOrder,
			Configured:      true,
		}

		// 该环境最新一条发布单
		rel, relErr := s.dao.CicdReleaseLatestByPipelineEnv(ctx, pipelineID, t.Env)
		if relErr == nil && rel != nil {
			node.CurrentReleaseID = rel.ID
			node.CurrentImageRepo = rel.ImageRepo
			node.CurrentImageTag = rel.ImageTag
			if rel.ImageDigest != nil {
				node.CurrentImageDigest = *rel.ImageDigest
			}
			node.CurrentReleaseStatus = rel.Status
			node.CurrentDeployTime = rel.ModifiedAt
		} else if relErr != nil && !errors.Is(relErr, gorm.ErrRecordNotFound) {
			return nil, relErr
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// splitImageRepoTag 将 registry/repo:tag 拆分为 repo 与 tag（忽略 digest 部分）
func splitImageRepoTag(image string) (repo, tag string) {
	img := image
	// 去掉 digest 部分（repo@sha256:xxx）
	if atIdx := strings.Index(img, "@"); atIdx > 0 {
		img = img[:atIdx]
	}
	// 最后一个冒号且其后不含 '/' 才是 tag（避免把 host:port 误判）
	if colonIdx := strings.LastIndex(img, ":"); colonIdx > 0 && !strings.Contains(img[colonIdx:], "/") {
		return img[:colonIdx], img[colonIdx+1:]
	}
	return img, ""
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
