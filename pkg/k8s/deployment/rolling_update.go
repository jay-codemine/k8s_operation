package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// ==================== 滚动更新策略 ====================

// RollingUpdateConfig 滚动更新策略配置
type RollingUpdateConfig struct {
	MaxSurge                string `json:"max_surge"`                           // 最大超出副本数，如 "1" 或 "25%"
	MaxUnavailable          string `json:"max_unavailable"`                     // 最大不可用副本数，如 "0" 或 "25%"
	MinReadySeconds         int32  `json:"min_ready_seconds"`                   // Pod 就绪后最少等待秒数（0 表示立即可用）
	ProgressDeadlineSeconds *int32 `json:"progress_deadline_seconds,omitempty"` // 进度截止时间（秒），默认 600
	RevisionHistoryLimit    *int32 `json:"revision_history_limit,omitempty"`    // 历史版本保留数量
}

// UpdateRollingStrategy 更新 Deployment 滚动更新策略
func UpdateRollingStrategy(ctx context.Context, client kubernetes.Interface, namespace, name string, config *RollingUpdateConfig) (*appv1.Deployment, error) {
	// 获取当前 Deployment
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 构建更新策略
	deploy.Spec.Strategy.Type = appv1.RollingUpdateDeploymentStrategyType
	if deploy.Spec.Strategy.RollingUpdate == nil {
		deploy.Spec.Strategy.RollingUpdate = &appv1.RollingUpdateDeployment{}
	}

	// 解析 maxSurge
	if config.MaxSurge != "" {
		surge := intstr.Parse(config.MaxSurge)
		deploy.Spec.Strategy.RollingUpdate.MaxSurge = &surge
	}

	// 解析 maxUnavailable
	if config.MaxUnavailable != "" {
		unavail := intstr.Parse(config.MaxUnavailable)
		deploy.Spec.Strategy.RollingUpdate.MaxUnavailable = &unavail
	}

	// 更新 minReadySeconds
	deploy.Spec.MinReadySeconds = config.MinReadySeconds

	// 更新 progressDeadlineSeconds
	if config.ProgressDeadlineSeconds != nil {
		deploy.Spec.ProgressDeadlineSeconds = config.ProgressDeadlineSeconds
	}

	// 更新 revisionHistoryLimit
	if config.RevisionHistoryLimit != nil {
		deploy.Spec.RevisionHistoryLimit = config.RevisionHistoryLimit
	}

	// 添加审计注解
	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations["k8soperation/strategy-updated-at"] = time.Now().Format(time.RFC3339)

	// 执行更新
	updated, err := client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新滚动策略失败: %w", err)
	}

	global.Logger.Info("滚动更新策略已更新",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.String("max_surge", config.MaxSurge),
		zap.String("max_unavailable", config.MaxUnavailable),
	)

	return updated, nil
}

// ==================== 暂停/恢复 Rollout ====================

// PauseDeployment 暂停 Deployment 滚动更新
// 等同于 kubectl rollout pause deployment/<name>
func PauseDeployment(ctx context.Context, client kubernetes.Interface, namespace, name string) (*appv1.Deployment, error) {
	// 检查当前状态
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	if deploy.Spec.Paused {
		return deploy, fmt.Errorf("Deployment %s/%s 已经处于暂停状态", namespace, name)
	}

	// 使用 Patch 设置 paused=true
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"paused": true,
		},
	}
	patchData, _ := json.Marshal(patch)

	updated, err := client.AppsV1().Deployments(namespace).Patch(
		ctx, name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("暂停 Rollout 失败: %w", err)
	}

	global.Logger.Info("Deployment Rollout 已暂停",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	return updated, nil
}

// ResumeDeployment 恢复 Deployment 滚动更新
// 等同于 kubectl rollout resume deployment/<name>
func ResumeDeployment(ctx context.Context, client kubernetes.Interface, namespace, name string) (*appv1.Deployment, error) {
	// 检查当前状态
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	if !deploy.Spec.Paused {
		return deploy, fmt.Errorf("Deployment %s/%s 不在暂停状态", namespace, name)
	}

	// 使用 Patch 设置 paused=false
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"paused": false,
		},
	}
	patchData, _ := json.Marshal(patch)

	updated, err := client.AppsV1().Deployments(namespace).Patch(
		ctx, name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("恢复 Rollout 失败: %w", err)
	}

	global.Logger.Info("Deployment Rollout 已恢复",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	return updated, nil
}

// ==================== Rollout 状态查询 ====================

// RolloutStatusInfo 滚动更新状态详情
type RolloutStatusInfo struct {
	// 基本信息
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// 滚动更新状态
	Status       string `json:"status"`        // Progressing / Complete / Paused / Failed / Waiting
	StatusReason string `json:"status_reason"`  // 状态原因描述
	Paused       bool   `json:"paused"`         // 是否暂停中
	IsRolling    bool   `json:"is_rolling"`     // 是否正在滚动更新中

	// 副本信息
	DesiredReplicas   int32 `json:"desired_replicas"`   // 期望副本数
	CurrentReplicas   int32 `json:"current_replicas"`   // 当前副本数
	UpdatedReplicas   int32 `json:"updated_replicas"`   // 已更新副本数
	ReadyReplicas     int32 `json:"ready_replicas"`     // 就绪副本数
	AvailableReplicas int32 `json:"available_replicas"` // 可用副本数
	UnavailableReplicas int32 `json:"unavailable_replicas"` // 不可用副本数

	// 进度信息
	Progress     int    `json:"progress"`      // 进度百分比 (0-100)
	Generation   int64  `json:"generation"`    // 当前配置版本
	ObservedGen  int64  `json:"observed_gen"`  // 已观察版本

	// 策略配置
	Strategy            string `json:"strategy"`              // 策略类型 (RollingUpdate/Recreate)
	MaxSurge            string `json:"max_surge"`             // 最大超出数
	MaxUnavailable      string `json:"max_unavailable"`       // 最大不可用数
	MinReadySeconds     int32  `json:"min_ready_seconds"`     // 最小就绪时间
	ProgressDeadline    int32  `json:"progress_deadline"`     // 进度截止时间
	RevisionHistoryLimit int32 `json:"revision_history_limit"` // 历史版本保留数

	// Conditions
	Conditions []RolloutCondition `json:"conditions"` // 当前 Conditions

	// 新旧 ReplicaSet
	NewReplicaSet string             `json:"new_replica_set,omitempty"` // 新 RS 名称
	OldReplicaSets []string          `json:"old_replica_sets,omitempty"` // 旧 RS 名称列表
	ReplicaSets   []ReplicaSetBrief  `json:"replica_sets"`              // RS 摘要列表
}

// RolloutCondition 简化的 Condition 信息
type RolloutCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	LastUpdate string `json:"last_update"`
}

// ReplicaSetBrief RS 摘要信息
type ReplicaSetBrief struct {
	Name            string `json:"name"`
	Revision        string `json:"revision"`
	Replicas        int32  `json:"replicas"`
	ReadyReplicas   int32  `json:"ready_replicas"`
	AvailableReplicas int32 `json:"available_replicas"`
	Image           string `json:"image"`
	IsCurrent       bool   `json:"is_current"` // 是否当前活跃 RS
	CreatedAt       string `json:"created_at"`
}

// GetRolloutStatus 获取 Deployment 滚动更新状态
func GetRolloutStatus(ctx context.Context, client kubernetes.Interface, namespace, name string) (*RolloutStatusInfo, error) {
	// 1. 获取 Deployment
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 2. 获取关联 ReplicaSets
	selector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("获取 ReplicaSet 列表失败: %w", err)
	}

	// 3. 构建基本信息
	desired := int32(1)
	if deploy.Spec.Replicas != nil {
		desired = *deploy.Spec.Replicas
	}

	info := &RolloutStatusInfo{
		Name:                name,
		Namespace:           namespace,
		Paused:              deploy.Spec.Paused,
		DesiredReplicas:     desired,
		CurrentReplicas:     deploy.Status.Replicas,
		UpdatedReplicas:     deploy.Status.UpdatedReplicas,
		ReadyReplicas:       deploy.Status.ReadyReplicas,
		AvailableReplicas:   deploy.Status.AvailableReplicas,
		UnavailableReplicas: deploy.Status.UnavailableReplicas,
		Generation:          deploy.Generation,
		ObservedGen:         deploy.Status.ObservedGeneration,
	}

	// 4. 计算进度
	if desired > 0 {
		info.Progress = int(float64(deploy.Status.AvailableReplicas) / float64(desired) * 100)
		if info.Progress > 100 {
			info.Progress = 100
		}
	}

	// 5. 策略信息
	info.Strategy = string(deploy.Spec.Strategy.Type)
	if deploy.Spec.Strategy.RollingUpdate != nil {
		ru := deploy.Spec.Strategy.RollingUpdate
		if ru.MaxSurge != nil {
			info.MaxSurge = ru.MaxSurge.String()
		}
		if ru.MaxUnavailable != nil {
			info.MaxUnavailable = ru.MaxUnavailable.String()
		}
	}
	info.MinReadySeconds = deploy.Spec.MinReadySeconds
	if deploy.Spec.ProgressDeadlineSeconds != nil {
		info.ProgressDeadline = *deploy.Spec.ProgressDeadlineSeconds
	}
	if deploy.Spec.RevisionHistoryLimit != nil {
		info.RevisionHistoryLimit = *deploy.Spec.RevisionHistoryLimit
	}

	// 6. 计算滚动状态
	info.Status, info.StatusReason, info.IsRolling = calcRolloutStatus(deploy)

	// 7. Conditions
	info.Conditions = make([]RolloutCondition, 0, len(deploy.Status.Conditions))
	for _, cond := range deploy.Status.Conditions {
		info.Conditions = append(info.Conditions, RolloutCondition{
			Type:       string(cond.Type),
			Status:     string(cond.Status),
			Reason:     cond.Reason,
			Message:    cond.Message,
			LastUpdate: cond.LastUpdateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 8. ReplicaSet 信息
	info.ReplicaSets = make([]ReplicaSetBrief, 0, len(rsList.Items))
	for _, rs := range rsList.Items {
		// 验证 OwnerReference
		isOwned := false
		for _, owner := range rs.OwnerReferences {
			if owner.UID == deploy.UID {
				isOwned = true
				break
			}
		}
		if !isOwned {
			continue
		}

		revision := ""
		if rev, ok := rs.Annotations["deployment.kubernetes.io/revision"]; ok {
			revision = rev
		}

		image := ""
		if len(rs.Spec.Template.Spec.Containers) > 0 {
			image = rs.Spec.Template.Spec.Containers[0].Image
		}

		replicas := int32(0)
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}

		isCurrent := replicas > 0
		brief := ReplicaSetBrief{
			Name:              rs.Name,
			Revision:          revision,
			Replicas:          replicas,
			ReadyReplicas:     rs.Status.ReadyReplicas,
			AvailableReplicas: rs.Status.AvailableReplicas,
			Image:             image,
			IsCurrent:         isCurrent,
			CreatedAt:         rs.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		info.ReplicaSets = append(info.ReplicaSets, brief)

		if isCurrent && replicas == desired {
			info.NewReplicaSet = rs.Name
		} else if replicas > 0 && rs.Name != info.NewReplicaSet {
			info.OldReplicaSets = append(info.OldReplicaSets, rs.Name)
		}
	}

	return info, nil
}

// calcRolloutStatus 计算 Rollout 状态
func calcRolloutStatus(deploy *appv1.Deployment) (status, reason string, isRolling bool) {
	desired := int32(1)
	if deploy.Spec.Replicas != nil {
		desired = *deploy.Spec.Replicas
	}

	// 暂停中
	if deploy.Spec.Paused {
		return "Paused", "Rollout 已暂停", false
	}

	// 检查 Conditions
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appv1.DeploymentProgressing {
			if cond.Reason == "ProgressDeadlineExceeded" {
				return "Failed", fmt.Sprintf("Rollout 超时: %s", cond.Message), false
			}
		}
		if cond.Type == appv1.DeploymentAvailable && cond.Status == "False" {
			return "Failed", fmt.Sprintf("不可用: %s", cond.Message), false
		}
	}

	// 正在滚动更新
	if deploy.Status.UpdatedReplicas < desired {
		return "Progressing",
			fmt.Sprintf("正在滚动更新: %d/%d 已更新", deploy.Status.UpdatedReplicas, desired),
			true
	}

	// 更新完成但还没全部就绪
	if deploy.Status.AvailableReplicas < desired {
		return "Progressing",
			fmt.Sprintf("等待就绪: %d/%d 已可用", deploy.Status.AvailableReplicas, desired),
			true
	}

	// 旧 Pod 还在终止中
	if deploy.Status.Replicas > desired {
		return "Progressing",
			fmt.Sprintf("清理旧 Pod: 当前 %d > 期望 %d", deploy.Status.Replicas, desired),
			true
	}

	// 控制器还没处理最新配置
	if deploy.Status.ObservedGeneration < deploy.Generation {
		return "Waiting", "等待控制器处理最新配置", true
	}

	// 一切正常
	if deploy.Status.UpdatedReplicas == desired &&
		deploy.Status.AvailableReplicas == desired &&
		deploy.Status.ReadyReplicas == desired {
		return "Complete", fmt.Sprintf("所有 %d 个副本已就绪", desired), false
	}

	return "Unknown", "状态未知", false
}
