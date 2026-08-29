package deployment

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// =============================================================================
// 金丝雀部署 —— 双 Deployment + 副本比流量拆分（纯 K8s 原生，无外部依赖）
//
// 原理：
//   - stable Deployment 标签：{app: xxx, version: stable}
//   - canary Deployment 标签：{app: xxx, version: canary}
//   - Service selector 仅含 {app: xxx}，同时选中 stable 和 canary Pod
//   - 流量比 = canaryReplicas / (stableReplicas + canaryReplicas)
// =============================================================================

const (
	LabelVersion     = "version"
	VersionStable    = "stable"
	VersionCanary    = "canary"
	CanaryNameSuffix = "-canary"
)

// CanaryStatusInfo 金丝雀状态信息
type CanaryStatusInfo struct {
	StableName            string `json:"stable_name"`
	CanaryName            string `json:"canary_name"`
	StableReadyReplicas   int32  `json:"stable_ready_replicas"`
	StableDesiredReplicas int32  `json:"stable_desired_replicas"`
	StableImage           string `json:"stable_image"`
	CanaryReadyReplicas   int32  `json:"canary_ready_replicas"`
	CanaryDesiredReplicas int32  `json:"canary_desired_replicas"`
	CanaryImage           string `json:"canary_image"`
	CanaryReplicas        int32  `json:"canary_replicas"`
	TrafficSplitPercent   int32  `json:"traffic_split_percent"`
	CanaryExists          bool   `json:"canary_exists"` // canary Deployment 是否存在
	Phase                 string `json:"phase"`         // none/pending/running/monitoring/promoting/promoted/rolling_back/rolled_back
}

const (
	CanaryPhaseNone        = "none"
	CanaryPhasePending     = "pending"
	CanaryPhaseRunning     = "running"
	CanaryPhaseMonitoring  = "monitoring"
	CanaryPhasePromoting   = "promoting"
	CanaryPhasePromoted    = "promoted"
	CanaryPhaseRollingBack = "rolling_back"
	CanaryPhaseRolledBack  = "rolled_back"
)

// ensureCanaryServiceSelector 确保关联 Service 的 selector 不含 version 标签。
//
// 金丝雀流量拆分的前提是 Service selector 只含 {app: xxx} 这类稳定标签，
// 同时命中 stable 与 canary 两个 Deployment 的 Pod。若 selector 里带了 version
// 条件，Service 只会命中其中一个版本，流量根本进不去金丝雀。
//
// 该函数列出命名空间下所有 Service，找到命中 stable Pod 的 Service，并移除其
// selector 中的 version key。返回命中 stable 的 Service 名称列表。
func ensureCanaryServiceSelector(ctx context.Context, kube kubernetes.Interface, namespace string, stable *appv1.Deployment) []string {
	podLabels := stable.Spec.Template.Labels
	if len(podLabels) == 0 {
		global.Logger.Warn("[Canary] stable Deployment 无 Pod 标签，跳过 Service selector 校验")
		return nil
	}

	svcList, err := kube.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		global.Logger.Warn("[Canary] 列出 Service 失败，跳过 selector 校验", zap.Error(err))
		return nil
	}

	var matched []string
	for i := range svcList.Items {
		svc := &svcList.Items[i]
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		if !labels.SelectorFromSet(svc.Spec.Selector).Matches(labels.Set(podLabels)) {
			continue
		}
		matched = append(matched, svc.Name)
		if _, hasVersion := svc.Spec.Selector[LabelVersion]; !hasVersion {
			continue
		}

		// JSON Merge Patch 中 null 表示删除该 key，移除 version 让 Service 同时选中两个版本
		patch := []byte(`{"spec":{"selector":{"version":null}}}`)
		if _, perr := kube.CoreV1().Services(namespace).Patch(ctx, svc.Name, types.MergePatchType, patch, metav1.PatchOptions{}); perr != nil {
			global.Logger.Warn("[Canary] 移除 Service selector version 失败",
				zap.String("service", svc.Name), zap.Error(perr))
			continue
		}
		global.Logger.Info("[Canary] 已移除 Service selector 的 version 标签，启用流量切分",
			zap.String("service", svc.Name))
	}
	return matched
}

// CreateCanaryDeployment 克隆 stable Deployment 创建金丝雀版本
func CreateCanaryDeployment(ctx context.Context, kube kubernetes.Interface, namespace, stableName, newImage, containerName string, canaryReplicas int32) (*appv1.Deployment, error) {
	stable, err := GetDeploymentDetail(ctx, kube, stableName, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取 stable Deployment 失败: %w", err)
	}

	// ===== 给 stable 补打 version 标签（不触发 Pod 重启） =====
	needStableLabel := stable.Spec.Template.Labels[LabelVersion] != VersionStable
	if needStableLabel {
		global.Logger.Info("[Canary] 给 stable Deployment 补打 version: stable 标签",
			zap.String("deployment", stableName),
		)
		if stable.Labels == nil {
			stable.Labels = make(map[string]string)
		}
		stable.Labels[LabelVersion] = VersionStable
		if stable.Spec.Template.Labels == nil {
			stable.Spec.Template.Labels = make(map[string]string)
		}
		stable.Spec.Template.Labels[LabelVersion] = VersionStable

		updated, updateErr := kube.AppsV1().Deployments(namespace).Update(ctx, stable, metav1.UpdateOptions{})
		if updateErr != nil {
			return nil, fmt.Errorf("补打 stable 标签失败: %w", updateErr)
		}
		stable = updated
	}

	// 确保关联 Service 的 selector 不含 version，使流量能同时命中 stable 与 canary
	matchedServices := ensureCanaryServiceSelector(ctx, kube, namespace, stable)
	if len(matchedServices) > 0 {
		global.Logger.Info("[Canary] 已关联的 Service", zap.Strings("services", matchedServices))
	}

	// 检查是否已有 canary 在运行
	canaryName := stableName + CanaryNameSuffix
	existing, _ := GetDeploymentDetail(ctx, kube, canaryName, namespace)
	if existing != nil {
		canaryVersion := existing.Spec.Template.Labels[LabelVersion]
		if canaryVersion == VersionCanary && existing.Status.ReadyReplicas > 0 {
			return nil, fmt.Errorf("金丝雀 Deployment %s 已在运行中（就绪副本: %d），请先 promote 或 rollback", canaryName, existing.Status.ReadyReplicas)
		}
		global.Logger.Warn("[Canary] 发现残留金丝雀 Deployment，将删除重建",
			zap.String("deployment", canaryName),
		)
		_ = DeleteDeployment(ctx, kube, canaryName, namespace)
	}

	// ===== 克隆 stable 创建 canary Deployment =====
	canary := stable.DeepCopy()
	canary.Name = canaryName
	canary.ResourceVersion = ""
	canary.UID = ""
	canary.Generation = 0
	canary.CreationTimestamp = metav1.Time{}
	canary.Status = appv1.DeploymentStatus{}

	// 标签
	if canary.Labels == nil {
		canary.Labels = make(map[string]string)
	}
	canary.Labels[LabelVersion] = VersionCanary
	if canary.Spec.Template.Labels == nil {
		canary.Spec.Template.Labels = make(map[string]string)
	}
	canary.Spec.Template.Labels[LabelVersion] = VersionCanary

	// 副本数
	canary.Spec.Replicas = &canaryReplicas

	// 替换镜像
	if containerName == "" && len(canary.Spec.Template.Spec.Containers) > 0 {
		containerName = canary.Spec.Template.Spec.Containers[0].Name
	}
	found := false
	for i := range canary.Spec.Template.Spec.Containers {
		if canary.Spec.Template.Spec.Containers[i].Name == containerName {
			canary.Spec.Template.Spec.Containers[i].Image = newImage
			found = true
			break
		}
	}
	if !found && containerName != "" {
		return nil, fmt.Errorf("未找到容器 %s", containerName)
	}

	created, err := kube.AppsV1().Deployments(namespace).Create(ctx, canary, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 canary Deployment 失败: %w", err)
	}

	global.Logger.Info("[Canary] 金丝雀 Deployment 创建成功",
		zap.String("deployment", canaryName),
		zap.String("image", newImage),
		zap.Int32("replicas", canaryReplicas),
	)

	return created, nil
}

// PromoteCanaryToStable 晋升金丝雀为稳定版本
func PromoteCanaryToStable(ctx context.Context, kube kubernetes.Interface, namespace, stableName, containerName string) (*appv1.Deployment, error) {
	canaryName := stableName + CanaryNameSuffix

	canary, err := GetDeploymentDetail(ctx, kube, canaryName, namespace)
	if err != nil {
		return nil, fmt.Errorf("canary Deployment 不存在: %w", err)
	}

	if containerName == "" && len(canary.Spec.Template.Spec.Containers) > 0 {
		containerName = canary.Spec.Template.Spec.Containers[0].Name
	}

	newImage := ""
	for _, c := range canary.Spec.Template.Spec.Containers {
		if c.Name == containerName {
			newImage = c.Image
			break
		}
	}
	if newImage == "" {
		return nil, fmt.Errorf("未找到容器 %s 的镜像", containerName)
	}

	global.Logger.Info("[Canary] 晋升金丝雀为稳定版",
		zap.String("stable", stableName),
		zap.String("canary", canaryName),
		zap.String("image", newImage),
	)

	_, err = PatchDeploymentImage(ctx, kube, namespace, stableName, containerName, newImage)
	if err != nil {
		return nil, fmt.Errorf("更新 stable 镜像失败: %w", err)
	}

	// 晋升完成后删除 canary Deployment，避免残留 0 副本垃圾对象
	if err := DeleteDeployment(ctx, kube, canaryName, namespace); err != nil {
		global.Logger.Warn("[Canary] 删除 canary Deployment 失败", zap.Error(err))
	}

	global.Logger.Info("[Canary] 晋升完成", zap.String("stable", stableName))
	return GetDeploymentDetail(ctx, kube, stableName, namespace)
}

// RollbackCanary 回滚金丝雀：删除 canary Deployment，stable 不受影响
func RollbackCanary(ctx context.Context, kube kubernetes.Interface, namespace, stableName string) error {
	canaryName := stableName + CanaryNameSuffix

	global.Logger.Info("[Canary] 回滚金丝雀", zap.String("canary", canaryName))

	err := DeleteDeployment(ctx, kube, canaryName, namespace)
	if err != nil {
		return fmt.Errorf("删除 canary Deployment 失败: %w", err)
	}

	global.Logger.Info("[Canary] 回滚完成，stable 不受影响")
	return nil
}

// GetCanaryStatus 获取金丝雀部署状态
func GetCanaryStatus(ctx context.Context, kube kubernetes.Interface, namespace, stableName string) (*CanaryStatusInfo, error) {
	info := &CanaryStatusInfo{
		StableName: stableName,
		CanaryName: stableName + CanaryNameSuffix,
		Phase:      CanaryPhaseNone,
	}

	stable, err := GetDeploymentDetail(ctx, kube, stableName, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取 stable Deployment 失败: %w", err)
	}
	info.StableReadyReplicas = stable.Status.ReadyReplicas
	info.StableDesiredReplicas = 1
	if stable.Spec.Replicas != nil {
		info.StableDesiredReplicas = *stable.Spec.Replicas
	}
	if len(stable.Spec.Template.Spec.Containers) > 0 {
		info.StableImage = stable.Spec.Template.Spec.Containers[0].Image
	}

	canary, err := GetDeploymentDetail(ctx, kube, info.CanaryName, namespace)
	if err != nil {
		return info, nil
	}
	info.CanaryExists = true

	info.CanaryReadyReplicas = canary.Status.ReadyReplicas
	info.CanaryDesiredReplicas = 1
	if canary.Spec.Replicas != nil {
		info.CanaryDesiredReplicas = *canary.Spec.Replicas
		info.CanaryReplicas = *canary.Spec.Replicas
	}
	if len(canary.Spec.Template.Spec.Containers) > 0 {
		info.CanaryImage = canary.Spec.Template.Spec.Containers[0].Image
	}

	// 流量比
	total := info.StableReadyReplicas + info.CanaryReadyReplicas
	if total > 0 {
		info.TrafficSplitPercent = info.CanaryReadyReplicas * 100 / total
	}

	// 判断阶段
	canaryVersion := canary.Spec.Template.Labels[LabelVersion]
	if canaryVersion != VersionCanary {
		info.Phase = CanaryPhaseNone
	} else if info.CanaryDesiredReplicas == 0 {
		info.Phase = CanaryPhasePromoted
	} else if info.CanaryReadyReplicas == 0 {
		info.Phase = CanaryPhasePending
	} else if info.CanaryReadyReplicas < info.CanaryDesiredReplicas {
		info.Phase = CanaryPhaseRunning
	} else {
		info.Phase = CanaryPhaseMonitoring
	}

	return info, nil
}

// SetCanaryReplicas 手动调整金丝雀副本数（动态改变流量比例）
func SetCanaryReplicas(ctx context.Context, kube kubernetes.Interface, namespace, stableName string, replicas int32) (*appv1.Deployment, error) {
	canaryName := stableName + CanaryNameSuffix

	global.Logger.Info("[Canary] 调整金丝雀副本数",
		zap.String("canary", canaryName),
		zap.Int32("replicas", replicas),
	)

	return PatchDeploymentReplicas(ctx, kube, namespace, canaryName, replicas)
}
