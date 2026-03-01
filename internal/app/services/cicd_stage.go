package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 阶段定义 ====================

// StageDefinition 阶段定义（用于配置流水线包含哪些阶段）
type StageDefinition struct {
	Order   int    `json:"order"`
	Type    string `json:"type"`    // checkout/build/test/push/approval/deploy
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// DefaultStageDefinitions 默认阶段定义
var DefaultStageDefinitions = []StageDefinition{
	{Order: 1, Type: models.StageTypeCheckout, Name: "代码检出", Enabled: true},
	{Order: 2, Type: models.StageTypeBuild, Name: "构建", Enabled: true},
	{Order: 3, Type: models.StageTypeTest, Name: "测试", Enabled: true},
	{Order: 4, Type: models.StageTypePush, Name: "推送镜像", Enabled: true},
	{Order: 5, Type: models.StageTypeApproval, Name: "人工审批", Enabled: false}, // 默认关闭
	{Order: 6, Type: models.StageTypeDeploy, Name: "部署", Enabled: false},     // 默认关闭
}

// ==================== 阶段执行服务 ====================

// CreateRunStages 为流水线运行创建阶段记录
func (s *Services) CreateRunStages(ctx context.Context, runID, pipelineID int64, pipeline *models.CicdPipeline) error {
	// 确定要创建的阶段
	stages := s.getStageDefinitionsForPipeline(pipeline)
	
	// 创建阶段记录
	stageRecords := make([]*models.CicdPipelineStage, 0, len(stages))
	for _, def := range stages {
		if !def.Enabled {
			continue
		}
		stage := &models.CicdPipelineStage{
			RunID:      runID,
			PipelineID: pipelineID,
			StageOrder: def.Order,
			StageType:  def.Type,
			StageName:  def.Name,
			Status:     models.StageStatusPending,
		}
		// 部署阶段预填充配置
		if def.Type == models.StageTypeDeploy && pipeline.AutoDeploy {
			stage.DeployClusterID = pipeline.TargetClusterID
			stage.DeployNamespace = pipeline.TargetNamespace
			stage.DeployWorkloadKind = pipeline.TargetWorkloadKind
			stage.DeployWorkloadName = pipeline.TargetWorkloadName
			stage.DeployContainer = pipeline.TargetContainer
		}
		stageRecords = append(stageRecords, stage)
	}

	return s.dao.StageCreateBatch(ctx, stageRecords)
}

// getStageDefinitionsForPipeline 获取流水线的阶段定义
func (s *Services) getStageDefinitionsForPipeline(pipeline *models.CicdPipeline) []StageDefinition {
	stages := make([]StageDefinition, len(DefaultStageDefinitions))
	copy(stages, DefaultStageDefinitions)
	
	// 根据流水线配置调整
	for i := range stages {
		switch stages[i].Type {
		case models.StageTypeApproval:
			stages[i].Enabled = pipeline.RequireApproval
		case models.StageTypeDeploy:
			stages[i].Enabled = pipeline.AutoDeploy
		}
	}
	
	return stages
}

// GetRunStages 获取运行记录的所有阶段（用于前端展示）
func (s *Services) GetRunStages(ctx context.Context, runID int64) ([]*models.StageDisplayInfo, error) {
	stages, err := s.dao.StageListByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	
	result := make([]*models.StageDisplayInfo, 0, len(stages))
	for _, stage := range stages {
		info := &models.StageDisplayInfo{
			ID:           stage.ID,
			Order:        stage.StageOrder,
			Type:         stage.StageType,
			Name:         stage.StageName,
			Status:       stage.Status,
			Duration:     s.formatStageDuration(stage.DurationSec),
			StartedAt:    stage.StartedAt,
			FinishedAt:   stage.FinishedAt,
			HasLogs:      stage.Logs != "",
			ErrorMsg:     stage.ErrorMessage,
			ErrorMessage: stage.ErrorMessage,
		}
		
		// 判断是否可操作
		if stage.StageType == models.StageTypeApproval {
			// 只有 waiting 状态才可操作
			if stage.Status == models.StageStatusWaiting {
				info.CanOperate = true
			}
			// 始终返回审批信息（方便前端展示审批人和审批时间）
			info.ApprovalInfo = &models.StageApprovalInfo{
				ApproverID:   stage.ApprovalUserID,
				Decision:     stage.ApprovalDecision,
				Comment:      stage.ApprovalComment,
				ApprovedAt:   stage.FinishedAt, // 使用完成时间作为审批时间
			}
		}
		
		// 部署阶段的特殊处理
		if stage.StageType == models.StageTypeDeploy {
			// 只有 pending 状态才可操作
			if stage.Status == models.StageStatusPending {
				info.CanOperate = true
			}
			// 始终返回部署信息
			info.DeployInfo = &models.StageDeployInfo{
				ClusterID:    stage.DeployClusterID,
				Namespace:    stage.DeployNamespace,
				WorkloadKind: stage.DeployWorkloadKind,
				WorkloadName: stage.DeployWorkloadName,
				Container:    stage.DeployContainer,
				Image:        stage.DeployImage,
				Replicas:     stage.DeployReplicas,
			}
			// 返回部署日志（包含 Rollout 进度）
			if stage.Logs != "" {
				info.Logs = stage.Logs
			}
		}
		
		result = append(result, info)
	}
	
	return result, nil
}

// formatStageDuration 格式化阶段时长
func (s *Services) formatStageDuration(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%d分%d秒", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%d时%d分", seconds/3600, (seconds%3600)/60)
}

// GetStageLogs 获取阶段日志
func (s *Services) GetStageLogs(ctx context.Context, stageID int64) (string, error) {
	return s.dao.StageGetLogs(ctx, stageID)
}

// ==================== 阶段状态更新 ====================

// StageCallback 处理 Jenkins 阶段回调（实时更新阶段状态）
func (s *Services) StageCallback(ctx context.Context, req *requests.StageCallbackRequest) error {
	// 1. 查找流水线和运行记录
	var runID int64
	
	if req.PipelineID > 0 {
		// 根据 pipeline_id + build_number 查找
		run, err := s.dao.PipelineRunGetByBuildNumber(ctx, req.PipelineID, req.BuildNumber)
		if err == nil && run != nil {
			runID = run.ID
		}
	}
	
	if runID == 0 && req.JobName != "" {
		// 根据 job_name + build_number 查找
		pipeline, err := s.dao.PipelineGetByJenkinsJob(ctx, req.JobName)
		if err == nil && pipeline != nil {
			run, err := s.dao.PipelineRunGetByBuildNumber(ctx, pipeline.ID, req.BuildNumber)
			if err == nil && run != nil {
				runID = run.ID
			}
		}
	}
	
	if runID == 0 {
		global.Logger.Warn("[阶段回调] 未找到对应的运行记录",
			zap.String("job_name", req.JobName),
			zap.Int("build_number", req.BuildNumber),
			zap.Int64("pipeline_id", req.PipelineID),
		)
		return nil // 不报错，避影响 Jenkins 构建
	}
	
	// 2. 查找对应的阶段
	stage, err := s.dao.StageGetByRunIDAndType(ctx, runID, req.StageType)
	if err != nil || stage == nil {
		global.Logger.Debug("[阶段回调] 未找到对应阶段",
			zap.Int64("run_id", runID),
			zap.String("stage_type", req.StageType),
		)
		return nil
	}
	
	// 3. 更新阶段状态
	updates := map[string]interface{}{
		"status": req.Status,
	}
	
	now := time.Now().Unix()
	switch req.Status {
	case models.StageStatusRunning:
		if stage.StartedAt == 0 {
			updates["started_at"] = now
		}
	case models.StageStatusSuccess, models.StageStatusFailed:
		updates["finished_at"] = now
		if stage.StartedAt > 0 {
			updates["duration_sec"] = int(now - int64(stage.StartedAt))
		}
	}
	
	if err := s.dao.StageUpdate(ctx, stage.ID, updates); err != nil {
		return err
	}
	
	global.Logger.Info("[阶段回调] 更新阶段状态",
		zap.Int64("stage_id", stage.ID),
		zap.String("stage_type", req.StageType),
		zap.String("status", req.Status),
	)
	
	return nil
}

// UpdateStageFromJenkins 从 Jenkins 更新阶段状态
func (s *Services) UpdateStageFromJenkins(ctx context.Context, runID int64, jenkinsStages []PipelineStageInfo) error {
	dbStages, err := s.dao.StageListByRunID(ctx, runID)
	if err != nil {
		return err
	}

	// 构建映射：Jenkins 阶段名 -> DB 阶段
	jenkinsMap := make(map[string]PipelineStageInfo)
	for _, js := range jenkinsStages {
		jenkinsMap[js.Name] = js
	}

	// 更新匹配的阶段
	for _, dbStage := range dbStages {
		// 只更新 Jenkins 相关阶段
		if dbStage.StageType == models.StageTypeApproval || dbStage.StageType == models.StageTypeDeploy {
			continue
		}

		// 根据阶段类型匹配 Jenkins 阶段
		var jenkinsStage PipelineStageInfo
		var found bool
		switch dbStage.StageType {
		case models.StageTypeCheckout:
			jenkinsStage, found = jenkinsMap["Checkout"]
			if !found {
				jenkinsStage, found = jenkinsMap["代码检出"]
			}
		case models.StageTypeBuild:
			jenkinsStage, found = jenkinsMap["Build"]
			if !found {
				jenkinsStage, found = jenkinsMap["构建"]
			}
		case models.StageTypeTest:
			jenkinsStage, found = jenkinsMap["Test"]
			if !found {
				jenkinsStage, found = jenkinsMap["测试"]
			}
		case models.StageTypePush:
			jenkinsStage, found = jenkinsMap["Push"]
			if !found {
				jenkinsStage, found = jenkinsMap["推送镜像"]
			}
			if !found {
				jenkinsStage, found = jenkinsMap["Deploy"] // 兼容旧的 Deploy 阶段名
			}
		}

		if found {
			updates := map[string]interface{}{
				"status":           jenkinsStage.Status,
				"jenkins_stage_id": jenkinsStage.ID,
			}
			if jenkinsStage.Status == "running" {
				updates["started_at"] = time.Now().Unix()
			}
			if jenkinsStage.Status == "success" || jenkinsStage.Status == "failed" {
				updates["finished_at"] = time.Now().Unix()
			}
			_ = s.dao.StageUpdate(ctx, dbStage.ID, updates)
		}
	}

	return nil
}

// UpdateBuildStagesComplete 构建完成后更新阶段状态
func (s *Services) UpdateBuildStagesComplete(ctx context.Context, runID int64, status string, imageURL, imageDigest string) error {
	dbStages, err := s.dao.StageListByRunID(ctx, runID)
	if err != nil {
		return err
	}

	// 更新构建相关阶段为完成状态
	for _, stage := range dbStages {
		if stage.StageType == models.StageTypeApproval || stage.StageType == models.StageTypeDeploy {
			continue
		}
		if stage.Status == models.StageStatusPending || stage.Status == models.StageStatusRunning {
			finalStatus := models.StageStatusSuccess
			if status == models.PipelineRunStatusFailed {
				finalStatus = models.StageStatusFailed
			}
			_ = s.dao.StageUpdateStatus(ctx, stage.ID, finalStatus)
		}
	}

	// 如果构建成功，更新推送镜像阶段的镜像信息
	if status == models.PipelineRunStatusSuccess && imageURL != "" {
		pushStage, err := s.dao.StageGetByRunIDAndType(ctx, runID, models.StageTypePush)
		if err == nil && pushStage != nil {
			_ = s.dao.StageUpdate(ctx, pushStage.ID, map[string]interface{}{
				"deploy_image": imageURL,
			})
		}
	}

	// 如果需要审批，将审批阶段设为等待状态
	approvalStage, err := s.dao.StageGetByRunIDAndType(ctx, runID, models.StageTypeApproval)
	if err == nil && approvalStage != nil && status == models.PipelineRunStatusSuccess {
		_ = s.dao.StageUpdateStatus(ctx, approvalStage.ID, models.StageStatusWaiting)
	}

	// 如果不需要审批但需要部署，将部署阶段设为待执行
	deployStage, err := s.dao.StageGetByRunIDAndType(ctx, runID, models.StageTypeDeploy)
	if err == nil && deployStage != nil && status == models.PipelineRunStatusSuccess {
		// 检查是否有审批阶段
		if approvalStage == nil {
			// 无审批阶段，部署阶段可以开始
			_ = s.dao.StageUpdate(ctx, deployStage.ID, map[string]interface{}{
				"status":       models.StageStatusPending,
				"deploy_image": imageURL,
			})
		}
	}

	return nil
}

// ==================== 审批阶段操作 ====================

// ApproveStage 审批通过阶段
func (s *Services) ApproveStage(ctx context.Context, stageID int64, userID int64, comment string) error {
	stage, err := s.dao.StageGetByID(ctx, stageID)
	if err != nil {
		return errors.New("阶段不存在")
	}

	if stage.StageType != models.StageTypeApproval {
		return errors.New("该阶段不是审批阶段")
	}

	if stage.Status != models.StageStatusWaiting {
		return errors.New("该阶段当前不处于等待审批状态")
	}

	// 更新审批信息
	if err := s.dao.StageUpdateApproval(ctx, stageID, userID, "approved", comment); err != nil {
		return err
	}

	global.Logger.Info("[流水线] 阶段审批通过",
		zap.Int64("stage_id", stageID),
		zap.Int64("user_id", userID),
	)

	// 检查是否有部署阶段需要启动
	deployStage, err := s.dao.StageGetByRunIDAndType(ctx, stage.RunID, models.StageTypeDeploy)
	if err == nil && deployStage != nil {
		// 获取构建产物镜像
		run, _ := s.dao.PipelineRunGetByID(ctx, stage.RunID)
		if run != nil && run.ImageURL != "" {
			_ = s.dao.StageUpdate(ctx, deployStage.ID, map[string]interface{}{
				"status":       models.StageStatusPending,
				"deploy_image": run.ImageURL,
			})
		}
	}

	return nil
}

// RejectStage 审批拒绝阶段
func (s *Services) RejectStage(ctx context.Context, stageID int64, userID int64, reason string) error {
	stage, err := s.dao.StageGetByID(ctx, stageID)
	if err != nil {
		return errors.New("阶段不存在")
	}

	if stage.StageType != models.StageTypeApproval {
		return errors.New("该阶段不是审批阶段")
	}

	if stage.Status != models.StageStatusWaiting {
		return errors.New("该阶段当前不处于等待审批状态")
	}

	// 更新审批信息
	if err := s.dao.StageUpdateApproval(ctx, stageID, userID, "rejected", reason); err != nil {
		return err
	}

	global.Logger.Info("[流水线] 阶段审批拒绝",
		zap.Int64("stage_id", stageID),
		zap.Int64("user_id", userID),
		zap.String("reason", reason),
	)

	// 将后续阶段标记为跳过
	stages, _ := s.dao.StageListByRunID(ctx, stage.RunID)
	for _, stg := range stages {
		if stg.StageOrder > stage.StageOrder && stg.Status == models.StageStatusPending {
			_ = s.dao.StageUpdateStatus(ctx, stg.ID, models.StageStatusSkipped)
		}
	}

	// 更新流水线运行状态为失败
	_ = s.dao.PipelineRunUpdateStatus(ctx, stage.RunID, models.PipelineRunStatusFailed)
	
	// 获取流水线ID并更新状态
	run, _ := s.dao.PipelineRunGetByID(ctx, stage.RunID)
	if run != nil {
		_ = s.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed)
	}

	return nil
}

// ==================== 部署阶段操作 ====================

// ExecuteDeployStage 执行部署阶段
func (s *Services) ExecuteDeployStage(ctx context.Context, req *requests.StageDeployRequest, userID int64) error {
	stage, err := s.dao.StageGetByID(ctx, req.StageID)
	if err != nil {
		return errors.New("阶段不存在")
	}

	if stage.StageType != models.StageTypeDeploy {
		return errors.New("该阶段不是部署阶段")
	}

	// 允许 pending 和 failed 状态的阶段执行部署（支持重试）
	if stage.Status != models.StageStatusPending && stage.Status != models.StageStatusFailed {
		return errors.New("该阶段当前不可执行部署")
	}

	// 获取流水线运行记录
	run, err := s.dao.PipelineRunGetByID(ctx, stage.RunID)
	if err != nil {
		return errors.New("运行记录不存在")
	}

	// 确定部署参数
	clusterID := stage.DeployClusterID
	namespace := stage.DeployNamespace
	workloadKind := stage.DeployWorkloadKind
	workloadName := stage.DeployWorkloadName
	container := stage.DeployContainer
	image := stage.DeployImage

	// 请求参数可覆盖默认配置
	if req.ClusterID > 0 {
		clusterID = req.ClusterID
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	if req.WorkloadKind != "" {
		workloadKind = req.WorkloadKind
	}
	if req.WorkloadName != "" {
		workloadName = req.WorkloadName
	}
	if req.Container != "" {
		container = req.Container
	}
	if req.Image != "" {
		image = req.Image
	} else if run.ImageURL != "" {
		image = run.ImageURL
	}

	if clusterID == 0 || namespace == "" || workloadName == "" || image == "" {
		return errors.New("部署参数不完整")
	}

	// 更新阶段为执行中
	_ = s.dao.StageUpdateStatus(ctx, stage.ID, models.StageStatusRunning)
	_ = s.dao.StageUpdateDeploy(ctx, stage.ID, clusterID, namespace, workloadKind, workloadName, container, image, 0)

	global.Logger.Info("[流水线] 开始执行部署阶段",
		zap.Int64("stage_id", stage.ID),
		zap.Int64("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("workload", workloadName),
		zap.String("image", image),
	)

	// 异步执行部署
	go s.executeDeployAsync(context.Background(), stage.ID, run, clusterID, namespace, workloadKind, workloadName, container, image)

	return nil
}

// executeDeployAsync 异步执行部署
func (s *Services) executeDeployAsync(ctx context.Context, stageID int64, run *models.CicdPipelineRun, clusterID int64, namespace, workloadKind, workloadName, container, image string) {
	startTime := time.Now()
	var logs strings.Builder
	logs.WriteString(fmt.Sprintf("[%s] 开始部署\n", startTime.Format("2006-01-02 15:04:05")))
	logs.WriteString(fmt.Sprintf("目标集群: %d\n", clusterID))
	logs.WriteString(fmt.Sprintf("命名空间: %s\n", namespace))
	logs.WriteString(fmt.Sprintf("工作负载: %s/%s\n", workloadKind, workloadName))
	logs.WriteString(fmt.Sprintf("容器: %s\n", container))
	logs.WriteString(fmt.Sprintf("镜像: %s\n\n", image))

	// 初始化 K8s 客户端
	client, err := s.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: uint32(clusterID)})
	if err != nil {
		errMsg := fmt.Sprintf("初始化集群客户端失败: %v", err)
		logs.WriteString(fmt.Sprintf("[ERROR] %s\n", errMsg))
		s.finishDeployStage(ctx, stageID, run, models.StageStatusFailed, errMsg, logs.String(), startTime)
		return
	}

	logs.WriteString("[INFO] 集群客户端初始化成功\n")

	// 执行镜像更新
	switch workloadKind {
	case "Deployment", "":
		err = s.updateDeploymentImage(ctx, client.Kube, namespace, workloadName, container, image, &logs)
	case "StatefulSet":
		err = s.updateStatefulSetImage(ctx, client.Kube, namespace, workloadName, container, image, &logs)
	case "DaemonSet":
		err = s.updateDaemonSetImage(ctx, client.Kube, namespace, workloadName, container, image, &logs)
	default:
		err = fmt.Errorf("不支持的工作负载类型: %s", workloadKind)
	}

	if err != nil {
		errMsg := fmt.Sprintf("更新镜像失败: %v", err)
		logs.WriteString(fmt.Sprintf("[ERROR] %s\n", errMsg))
		s.finishDeployStage(ctx, stageID, run, models.StageStatusFailed, errMsg, logs.String(), startTime)
		return
	}

	logs.WriteString(fmt.Sprintf("\n[%s] 部署完成\n", time.Now().Format("2006-01-02 15:04:05")))
	s.finishDeployStage(ctx, stageID, run, models.StageStatusSuccess, "", logs.String(), startTime)
}

// updateDeploymentImage 更新 Deployment 镜像
func (s *Services) updateDeploymentImage(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string, logs *strings.Builder) error {
	logs.WriteString(fmt.Sprintf("[INFO] 正在更新 Deployment %s/%s 的镜像...\n", namespace, name))
	
	// 1. Patch 更新镜像
	patchData := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	
	_, err := client.AppsV1().Deployments(namespace).Patch(
		ctx,
		name,
		types.StrategicMergePatchType,
		[]byte(patchData),
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("镜像更新失败: %v", err)
	}

	logs.WriteString(fmt.Sprintf("[INFO] 容器 %s 的新镜像: %s\n", container, image))
	logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")

	// 2. 等待 Rollout 完成（健康检查）
	err = s.waitDeploymentRollout(ctx, client, namespace, name, logs)
	if err != nil {
		return err
	}

	logs.WriteString("[INFO] Deployment Rollout 完成\n")
	return nil
}

// waitDeploymentRollout 等待 Deployment Rollout 完成
func (s *Services) waitDeploymentRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, logs *strings.Builder) error {
	timeout := 5 * time.Minute
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	logs.WriteString(fmt.Sprintf("[INFO] Rollout 超时时间: %v\n", timeout))

	for time.Now().Before(endTime) {
		dp, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取 Deployment 失败: %v", err)
		}

		// 获取期望副本数
		replicas := int32(1)
		if dp.Spec.Replicas != nil {
			replicas = *dp.Spec.Replicas
		}

		// 记录当前状态
		logs.WriteString(fmt.Sprintf("[ROLLOUT] 副本: %d/%d | 更新: %d | 就绪: %d | 可用: %d\n",
			dp.Status.ReadyReplicas, replicas,
			dp.Status.UpdatedReplicas,
			dp.Status.ReadyReplicas,
			dp.Status.AvailableReplicas))

		// 检查 Rollout 是否失败
		for _, cond := range dp.Status.Conditions {
			if cond.Type == "Progressing" {
				if cond.Reason == "ProgressDeadlineExceeded" {
					return fmt.Errorf("Rollout 超时: %s", cond.Message)
				}
			}
			if cond.Type == "Available" && cond.Status == "False" {
				logs.WriteString(fmt.Sprintf("[WARN] Deployment 不可用: %s\n", cond.Message))
			}
		}

		// 检查 Pod 状态（捕获 ImagePullBackOff 等错误）
		podErr := s.checkDeploymentPodStatus(ctx, client, namespace, name, dp.Spec.Selector, logs)
		if podErr != nil {
			return podErr
		}

		// Rollout 完成的条件：所有副本都已更新、就绪、可用
		if dp.Status.UpdatedReplicas == replicas &&
			dp.Status.Replicas == replicas &&
			dp.Status.AvailableReplicas == replicas &&
			dp.Status.ObservedGeneration >= dp.Generation {
			// 最终确认：所有 Pod 已就绪并可对外提供服务
			logs.WriteString(fmt.Sprintf("[SUCCESS] 所有 %d 个副本已就绪，服务可用\n", replicas))
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("Rollout 超时（%v），副本未就绪", timeout)
}

// checkDeploymentPodStatus 检查 Pod 状态，捕获 ImagePullBackOff 等错误
func (s *Services) checkDeploymentPodStatus(ctx context.Context, client kubernetes.Interface, namespace, name string, selector *metav1.LabelSelector, logs *strings.Builder) error {
	if selector == nil {
		return nil
	}

	// 构建 label selector
	labelSelector := metav1.FormatLabelSelector(selector)
	
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil // 获取 Pod 失败不需要成为致命错误
	}

	for _, pod := range pods.Items {
		// 检查容器状态
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				msg := cs.State.Waiting.Message
				
				// 检测镜像拉取失败
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					errMsg := fmt.Sprintf("镜像拉取失败 [%s]: %s", reason, msg)
					logs.WriteString(fmt.Sprintf("[ERROR] Pod %s: %s\n", pod.Name, errMsg))
					return fmt.Errorf(errMsg)
				}
				
				// 检测 CrashLoopBackOff
				if reason == "CrashLoopBackOff" {
					errMsg := fmt.Sprintf("容器崩溃重启 [%s]: %s", reason, msg)
					logs.WriteString(fmt.Sprintf("[ERROR] Pod %s: %s\n", pod.Name, errMsg))
					return fmt.Errorf(errMsg)
				}
			}
		}
	}

	return nil
}

// updateStatefulSetImage 更新 StatefulSet 镜像
func (s *Services) updateStatefulSetImage(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string, logs *strings.Builder) error {
	logs.WriteString(fmt.Sprintf("[INFO] 正在更新 StatefulSet %s/%s 的镜像...\n", namespace, name))
	
	patchData := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	
	_, err := client.AppsV1().StatefulSets(namespace).Patch(
		ctx,
		name,
		types.StrategicMergePatchType,
		[]byte(patchData),
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("镜像更新失败: %v", err)
	}

	logs.WriteString(fmt.Sprintf("[INFO] 容器 %s 的新镜像: %s\n", container, image))
	logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")

	// 等待 Rollout 完成
	err = s.waitStatefulSetRollout(ctx, client, namespace, name, logs)
	if err != nil {
		return err
	}

	logs.WriteString("[INFO] StatefulSet Rollout 完成\n")
	return nil
}

// waitStatefulSetRollout 等待 StatefulSet Rollout 完成
func (s *Services) waitStatefulSetRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, logs *strings.Builder) error {
	timeout := 5 * time.Minute
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		ss, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取 StatefulSet 失败: %v", err)
		}

		replicas := int32(1)
		if ss.Spec.Replicas != nil {
			replicas = *ss.Spec.Replicas
		}

		logs.WriteString(fmt.Sprintf("[ROLLOUT] 副本: %d/%d | 更新: %d | 就绪: %d\n",
			ss.Status.ReadyReplicas, replicas,
			ss.Status.UpdatedReplicas,
			ss.Status.ReadyReplicas))

		// 检查 Pod 状态
		podErr := s.checkDeploymentPodStatus(ctx, client, namespace, name, ss.Spec.Selector, logs)
		if podErr != nil {
			return podErr
		}

		if ss.Status.UpdatedReplicas == replicas &&
			ss.Status.ReadyReplicas == replicas &&
			ss.Status.ObservedGeneration >= ss.Generation {
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("StatefulSet Rollout 超时（%v）", timeout)
}

// updateDaemonSetImage 更新 DaemonSet 镜像
func (s *Services) updateDaemonSetImage(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string, logs *strings.Builder) error {
	logs.WriteString(fmt.Sprintf("[INFO] 正在更新 DaemonSet %s/%s 的镜像...\n", namespace, name))
	
	patchData := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	
	_, err := client.AppsV1().DaemonSets(namespace).Patch(
		ctx,
		name,
		types.StrategicMergePatchType,
		[]byte(patchData),
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("镜像更新失败: %v", err)
	}

	logs.WriteString(fmt.Sprintf("[INFO] 容器 %s 的新镜像: %s\n", container, image))
	logs.WriteString("[INFO] 镜像更新已提交，等待 Rollout 完成...\n")

	// 等待 Rollout 完成
	err = s.waitDaemonSetRollout(ctx, client, namespace, name, logs)
	if err != nil {
		return err
	}

	logs.WriteString("[INFO] DaemonSet Rollout 完成\n")
	return nil
}

// waitDaemonSetRollout 等待 DaemonSet Rollout 完成
func (s *Services) waitDaemonSetRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, logs *strings.Builder) error {
	timeout := 5 * time.Minute
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取 DaemonSet 失败: %v", err)
		}

		logs.WriteString(fmt.Sprintf("[ROLLOUT] 期望: %d | 更新: %d | 就绪: %d | 可用: %d\n",
			ds.Status.DesiredNumberScheduled,
			ds.Status.UpdatedNumberScheduled,
			ds.Status.NumberReady,
			ds.Status.NumberAvailable))

		// 检查 Pod 状态
		podErr := s.checkDeploymentPodStatus(ctx, client, namespace, name, ds.Spec.Selector, logs)
		if podErr != nil {
			return podErr
		}

		if ds.Status.UpdatedNumberScheduled == ds.Status.DesiredNumberScheduled &&
			ds.Status.NumberReady == ds.Status.DesiredNumberScheduled &&
			ds.Status.ObservedGeneration >= ds.Generation {
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("DaemonSet Rollout 超时（%v）", timeout)
}

// finishDeployStage 完成部署阶段
func (s *Services) finishDeployStage(ctx context.Context, stageID int64, run *models.CicdPipelineRun, status, errMsg, logs string, startTime time.Time) {
	duration := int(time.Since(startTime).Seconds())
	
	// 更新阶段状态
	updates := map[string]interface{}{
		"status":       status,
		"finished_at":  time.Now().Unix(),
		"duration_sec": duration,
		"logs":         logs,
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	_ = s.dao.StageUpdate(ctx, stageID, updates)

	// 更新流水线运行状态
	runStatus := models.PipelineRunStatusSuccess
	if status == models.StageStatusFailed {
		runStatus = models.PipelineRunStatusFailed
	}
	_ = s.dao.PipelineRunUpdateStatus(ctx, run.ID, runStatus)
	_ = s.dao.PipelineUpdateRunComplete(ctx, run.PipelineID, runStatus)

	// 如果部署失败，更新运行记录的错误信息
	if status == models.StageStatusFailed && errMsg != "" {
		_ = s.dao.PipelineRunUpdateError(ctx, run.ID, models.PipelineRunStatusFailed, errMsg)
	}

	// 获取阶段和流水线信息用于通知
	stage, _ := s.dao.StageGetByID(ctx, stageID)
	pipeline, _ := s.dao.PipelineGetByID(ctx, run.PipelineID)

	// 如果成功，更新流水线部署信息
	if status == models.StageStatusSuccess && stage != nil {
		_ = s.dao.PipelineUpdateDeployInfo(ctx, run.PipelineID, stage.DeployImage, "", uint64(time.Now().Unix()), "success")
	}

	// 发送钉钉通知（异步）
	if pipeline != nil && stage != nil {
		s.NotifyDeployResult(ctx, pipeline, stage, status == models.StageStatusSuccess, errMsg)
	}

	global.Logger.Info("[流水线] 部署阶段完成",
		zap.Int64("stage_id", stageID),
		zap.String("status", status),
		zap.Int("duration", duration),
	)
}
