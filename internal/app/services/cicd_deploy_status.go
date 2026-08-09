package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 部署实时状态查询 ====================
// 说明：CICD 流水线详情页（非 /c/:clusterId 路由）无法提供 X-Cluster-ID，
// 因此新增此端点，由服务端根据部署阶段记录/流水线配置解析目标集群，
// 查询真实的工作负载 Rollout 状态与 Pod 列表；同时对「卡在 running 的
// 孤立部署阶段」执行状态修正（reconcile），解决后端重启导致的“发布中”不消失问题。

// DeployStatusPod 部署 Pod 状态（前端卡片展示用）
type DeployStatusPod struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Phase           string `json:"phase"`            // Running/Pending/Succeeded/Failed
	Status          string `json:"status"`           // 展示状态：Running/ImagePullBackOff/CrashLoopBackOff...
	Ready           bool   `json:"ready"`            // 是否所有容器就绪
	ReadyContainers int    `json:"ready_containers"` // 就绪容器数
	TotalContainers int    `json:"total_containers"` // 容器总数
	RestartCount    int32  `json:"restart_count"`    // 重启次数
	NodeName        string `json:"node_name"`
	PodIP           string `json:"pod_ip"`
	Image           string `json:"image"`
	Reason          string `json:"reason,omitempty"`  // 异常原因
	Message         string `json:"message,omitempty"` // 异常详情
	CreatedAt       int64  `json:"created_at"`        // 创建时间(Unix)
}

// DeployStatusWorkload 工作负载 Rollout 概览
type DeployStatusWorkload struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	ClusterID int64  `json:"cluster_id"`
	Image     string `json:"image"`
	Desired   int32  `json:"desired"`
	Ready     int32  `json:"ready"`
	Updated   int32  `json:"updated"`
	Available int32  `json:"available"`
}

// DeployStatusResult 部署状态查询结果
type DeployStatusResult struct {
	StageID     int64                `json:"stage_id"`
	StageStatus string               `json:"stage_status"` // 数据库中的阶段状态（可能已被 reconcile 修正）
	RealStatus  string               `json:"real_status"`  // 真实部署状态：success/failed/deploying
	Reconciled  bool                 `json:"reconciled"`   // 是否发生了状态修正
	Message     string               `json:"message"`
	Workload    DeployStatusWorkload `json:"workload"`
	Pods        []DeployStatusPod    `json:"pods"`
}

// workloadRollout 工作负载副本数快照
type workloadRollout struct {
	desired   int32
	ready     int32
	updated   int32
	available int32
}

// GetDeployStatus 查询部署阶段的真实状态与 Pod 列表，并修正卡住的阶段
func (s *Services) GetDeployStatus(ctx context.Context, stageID int64) (*DeployStatusResult, error) {
	stage, err := s.cicdSvc().StageGetByID(ctx, stageID)
	if err != nil {
		return nil, errors.New("阶段不存在")
	}
	if stage.StageType != models.StageTypeDeploy {
		return nil, errors.New("该阶段不是部署阶段")
	}

	run, err := s.cicdSvc().PipelineRunGetByID(ctx, stage.RunID)
	if err != nil {
		return nil, errors.New("运行记录不存在")
	}
	pipeline, _ := s.cicdSvc().PipelineGetByID(ctx, run.PipelineID)

	// 解析部署目标：优先阶段记录，回退流水线最新配置
	clusterID := stage.DeployClusterID
	namespace := stage.DeployNamespace
	workloadKind := stage.DeployWorkloadKind
	workloadName := stage.DeployWorkloadName
	container := stage.DeployContainer
	if pipeline != nil {
		if clusterID == 0 {
			clusterID = pipeline.TargetClusterID
		}
		if namespace == "" {
			namespace = pipeline.TargetNamespace
		}
		if workloadKind == "" {
			workloadKind = pipeline.TargetWorkloadKind
		}
		if workloadName == "" {
			workloadName = pipeline.TargetWorkloadName
		}
		if container == "" {
			container = pipeline.TargetContainer
		}
	}
	if workloadKind == "" {
		workloadKind = "Deployment"
	}

	result := &DeployStatusResult{
		StageID:     stageID,
		StageStatus: stage.Status,
		Workload: DeployStatusWorkload{
			Kind:      workloadKind,
			Namespace: namespace,
			Name:      workloadName,
			ClusterID: clusterID,
			Image:     stage.DeployImage,
		},
		Pods: []DeployStatusPod{},
	}

	if clusterID == 0 || namespace == "" || workloadName == "" {
		result.RealStatus = stage.Status
		result.Message = "部署目标参数不完整，无法查询实时状态"
		return result, nil
	}

	// 服务端解析集群客户端（免 X-Cluster-ID）
	client, err := s.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: uint32(clusterID)})
	if err != nil {
		result.RealStatus = stage.Status
		result.Message = fmt.Sprintf("初始化集群客户端失败: %v", err)
		return result, nil
	}

	// 查询工作负载 Rollout 状态
	selector, wl, err := s.getWorkloadRollout(ctx, client.Kube, workloadKind, namespace, workloadName)
	if err != nil {
		result.RealStatus = stage.Status
		result.Message = fmt.Sprintf("获取工作负载状态失败: %v", err)
		return result, nil
	}
	result.Workload.Desired = wl.desired
	result.Workload.Ready = wl.ready
	result.Workload.Updated = wl.updated
	result.Workload.Available = wl.available

	// 收集 Pod 列表并计算真实状态
	pods, realStatus, msg := s.collectDeployPods(ctx, client.Kube, namespace, selector, container, wl)
	result.Pods = pods
	result.RealStatus = realStatus
	result.Message = msg

	// reconcile：
	// ① DB 阶段仍处于运行中（后端重启丢 goroutine），但真实状态已达终态 → 回写终态；
	// ② DB 阶段已置失败，但集群实际全部就绪（rollout 等待期陈旧条件误判 /
	//    超时后新副本才就绪）→ 仅当就绪 Pod 运行的是目标镜像时纠正为成功，
	//    镜像校验用于避免「部署失败后回滚到旧版且旧版已就绪」被误纠正为成功。
	needReconcile := false
	switch {
	case (stage.Status == models.StageStatusRunning || stage.Status == "deploying") &&
		(realStatus == models.StageStatusSuccess || realStatus == models.StageStatusFailed):
		needReconcile = true
	case stage.Status == models.StageStatusFailed && realStatus == models.StageStatusSuccess &&
		s.deployImageMatches(stage.DeployImage, pods):
		needReconcile = true
	}
	if needReconcile {
		s.reconcileStuckDeployStage(ctx, stage, run, realStatus, msg)
		result.Reconciled = true
		result.StageStatus = realStatus
	}

	return result, nil
}

// deployImageMatches 判断当前就绪 Pod 是否运行目标镜像。
// 用于 failed → success 的纠正场景：只有至少一个就绪 Pod 跑的是目标镜像，
// 才确认部署真正生效，避免回滚到旧版本也被当成新部署成功。
func (s *Services) deployImageMatches(targetImage string, pods []DeployStatusPod) bool {
	if targetImage == "" {
		return false // 无目标镜像信息，保守不纠正
	}
	for _, p := range pods {
		if p.Ready && p.Image == targetImage {
			return true
		}
	}
	return false
}

// getWorkloadRollout 获取工作负载的选择器与副本数快照
func (s *Services) getWorkloadRollout(ctx context.Context, client kubernetes.Interface, kind, namespace, name string) (*metav1.LabelSelector, *workloadRollout, error) {
	wl := &workloadRollout{}
	switch kind {
	case "Deployment", "":
		d, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		if d.Spec.Replicas != nil {
			wl.desired = *d.Spec.Replicas
		}
		wl.ready = d.Status.ReadyReplicas
		wl.updated = d.Status.UpdatedReplicas
		wl.available = d.Status.AvailableReplicas
		return d.Spec.Selector, wl, nil
	case "StatefulSet":
		ss, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		if ss.Spec.Replicas != nil {
			wl.desired = *ss.Spec.Replicas
		}
		wl.ready = ss.Status.ReadyReplicas
		wl.updated = ss.Status.UpdatedReplicas
		wl.available = ss.Status.CurrentReplicas
		return ss.Spec.Selector, wl, nil
	case "DaemonSet":
		ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		wl.desired = ds.Status.DesiredNumberScheduled
		wl.ready = ds.Status.NumberReady
		wl.updated = ds.Status.UpdatedNumberScheduled
		wl.available = ds.Status.NumberAvailable
		return ds.Spec.Selector, wl, nil
	default:
		return nil, nil, fmt.Errorf("不支持的工作负载类型: %s", kind)
	}
}

// collectDeployPods 列出关联 Pod 并计算真实部署状态
func (s *Services) collectDeployPods(ctx context.Context, client kubernetes.Interface, namespace string, selector *metav1.LabelSelector, container string, wl *workloadRollout) ([]DeployStatusPod, string, string) {
	pods := []DeployStatusPod{}
	if selector == nil {
		return pods, "deploying", "无法获取工作负载选择器"
	}

	labelSelector := metav1.FormatLabelSelector(selector)
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return pods, "deploying", fmt.Sprintf("获取 Pod 列表失败: %v", err)
	}

	hasFatalError := false
	fatalReason := ""
	for _, pod := range podList.Items {
		p := DeployStatusPod{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Phase:           string(pod.Status.Phase),
			NodeName:        pod.Spec.NodeName,
			PodIP:           pod.Status.PodIP,
			CreatedAt:       pod.CreationTimestamp.Unix(),
			TotalContainers: len(pod.Spec.Containers),
		}

		// 镜像：优先取目标容器
		for _, c := range pod.Spec.Containers {
			if p.Image == "" {
				p.Image = c.Image
			}
			if c.Name == container {
				p.Image = c.Image
				break
			}
		}

		var restart int32
		readyCount := 0
		podStatus := string(pod.Status.Phase)
		for _, cs := range pod.Status.ContainerStatuses {
			restart += cs.RestartCount
			if cs.Ready {
				readyCount++
			}
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				switch reason {
				case "ImagePullBackOff", "ErrImagePull", "CrashLoopBackOff", "CreateContainerError", "CreateContainerConfigError", "InvalidImageName":
					podStatus = reason
					p.Reason = reason
					p.Message = cs.State.Waiting.Message
					if reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CrashLoopBackOff" || reason == "InvalidImageName" {
						hasFatalError = true
						fatalReason = reason
					}
				}
			}
			if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" && cs.State.Terminated.Reason != "Completed" {
				podStatus = cs.State.Terminated.Reason
				p.Reason = cs.State.Terminated.Reason
				p.Message = cs.State.Terminated.Message
			}
		}
		p.RestartCount = restart
		p.ReadyContainers = readyCount
		p.Ready = p.TotalContainers > 0 && readyCount == p.TotalContainers
		p.Status = podStatus
		pods = append(pods, p)
	}

	// 计算真实部署状态
	realStatus := "deploying"
	msg := "部署进行中，等待所有副本就绪"
	switch {
	case hasFatalError:
		realStatus = models.StageStatusFailed
		msg = fmt.Sprintf("检测到 Pod 异常（%s），部署失败", fatalReason)
	case wl.desired > 0 && wl.ready == wl.desired && wl.updated == wl.desired && wl.available == wl.desired:
		realStatus = models.StageStatusSuccess
		msg = "所有副本已就绪，部署成功"
	}
	return pods, realStatus, msg
}

// reconcileStuckDeployStage 修正卡在运行中的部署阶段（后端重启导致 goroutine 丢失场景）
func (s *Services) reconcileStuckDeployStage(ctx context.Context, stage *models.CicdPipelineStage, run *models.CicdPipelineRun, realStatus, msg string) {
	now := time.Now().Unix()
	duration := 0
	if stage.StartedAt > 0 {
		duration = int(now - int64(stage.StartedAt))
	}

	updates := map[string]interface{}{
		"status":       realStatus,
		"finished_at":  now,
		"duration_sec": duration,
	}
	if realStatus == models.StageStatusFailed {
		updates["error_message"] = msg
	}
	_ = s.cicdSvc().StageUpdate(ctx, stage.ID, updates)

	runStatus := models.PipelineRunStatusSuccess
	if realStatus == models.StageStatusFailed {
		runStatus = models.PipelineRunStatusFailed
	}
	_ = s.cicdSvc().PipelineRunUpdateStatus(ctx, run.ID, runStatus)
	_ = s.cicdSvc().PipelineUpdateRunComplete(ctx, run.PipelineID, runStatus)

	if realStatus == models.StageStatusSuccess {
		_ = s.cicdSvc().PipelineUpdateDeployInfo(ctx, run.PipelineID, stage.DeployImage, "", uint64(now), "success", "")
	} else {
		_ = s.cicdSvc().PipelineRunUpdateError(ctx, run.ID, models.PipelineRunStatusFailed, msg)
	}

	global.Logger.Info("[流水线] 部署阶段状态已修正",
		zap.Int64("stage_id", stage.ID),
		zap.String("old_status", stage.Status),
		zap.String("new_status", realStatus),
		zap.Int("duration", duration),
	)
}
