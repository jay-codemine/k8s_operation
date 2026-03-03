package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/k8s/deployment"
)

// CicdExecuteResult 执行结果
type CicdExecuteResult struct {
	Success   bool
	Message   string
	PrevImage string
}

// CicdTaskExecutor CICD 任务执行器
type CicdTaskExecutor struct {
	factory *ClusterClientFactory
}

// NewCicdTaskExecutor 创建执行器
func NewCicdTaskExecutor(factory *ClusterClientFactory) *CicdTaskExecutor {
	return &CicdTaskExecutor{factory: factory}
}

// Execute 执行部署任务
func (e *CicdTaskExecutor) Execute(ctx context.Context, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取 K8s 客户端
	cli, err := e.factory.GetClient(ctx, task.ClusterID)
	if err != nil {
		result.Message = fmt.Sprintf("获取K8s客户端失败: %v", err)
		global.Logger.Error("get k8s client failed",
			zap.Int64("cluster_id", task.ClusterID),
			zap.Error(err))
		return result
	}

	// 2. 根据 WorkloadKind 执行部署
	switch release.WorkloadKind {
	case "Deployment":
		return e.executeDeployment(ctx, cli.Kube, task, release)
	case "StatefulSet":
		return e.executeStatefulSet(ctx, cli.Kube, task, release)
	case "DaemonSet":
		return e.executeDaemonSet(ctx, cli.Kube, task, release)
	default:
		result.Message = fmt.Sprintf("不支持的工作负载类型: %s", release.WorkloadKind)
		return result
	}
}

// executeDeployment 执行 Deployment 部署
func (e *CicdTaskExecutor) executeDeployment(ctx context.Context, kube kubernetes.Interface, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取当前 Deployment
	dp, err := kube.AppsV1().Deployments(release.Namespace).Get(ctx, release.WorkloadName, metav1.GetOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("获取Deployment失败: %v", err)
		return result
	}

	// 2. 保存原镜像
	result.PrevImage = e.getContainerImage(dp.Spec.Template.Spec.Containers, release.ContainerName)

	// 3. Patch 更新镜像
	_, err = deployment.PatchDeploymentImage(ctx, kube, release.Namespace, release.WorkloadName, release.ContainerName, task.TargetImage)
	if err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	global.Logger.Info("patched deployment image",
		zap.String("namespace", release.Namespace),
		zap.String("deployment", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 4. 等待 Rollout 完成
	timeout := time.Duration(release.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	if err := e.waitDeploymentRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout); err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// waitDeploymentRollout 等待 Deployment Rollout 完成
func (e *CicdTaskExecutor) waitDeploymentRollout(ctx context.Context, kube kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("rollout timeout after %v", timeout)
		case <-ticker.C:
			dp, err := kube.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get deployment failed: %w", err)
			}

			// 检查是否完成
			if isDeploymentRolloutComplete(dp) {
				global.Logger.Info("deployment rollout complete",
					zap.String("namespace", namespace),
					zap.String("name", name))
				return nil
			}

			// 检查是否失败
			for _, cond := range dp.Status.Conditions {
				if cond.Type == appv1.DeploymentProgressing && cond.Status == corev1.ConditionFalse {
					return fmt.Errorf("rollout failed: %s", cond.Message)
				}
				if cond.Reason == "ProgressDeadlineExceeded" {
					return fmt.Errorf("rollout progress deadline exceeded: %s", cond.Message)
				}
			}
		}
	}
}

// isDeploymentRolloutComplete 检查 Deployment Rollout 是否完成
// 注意：不检查 Replicas == replicas，因为滚动更新期间旧 Pod 可能还在终止中
func isDeploymentRolloutComplete(dp *appv1.Deployment) bool {
	replicas := int32(1)
	if dp.Spec.Replicas != nil {
		replicas = *dp.Spec.Replicas
	}

	return dp.Status.ObservedGeneration >= dp.Generation &&
		dp.Status.UpdatedReplicas == replicas &&
		dp.Status.AvailableReplicas == replicas
}

// executeStatefulSet 执行 StatefulSet 部署
func (e *CicdTaskExecutor) executeStatefulSet(ctx context.Context, kube kubernetes.Interface, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取当前 StatefulSet
	sts, err := kube.AppsV1().StatefulSets(release.Namespace).Get(ctx, release.WorkloadName, metav1.GetOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("获取StatefulSet失败: %v", err)
		return result
	}

	// 2. 保存原镜像
	result.PrevImage = e.getContainerImage(sts.Spec.Template.Spec.Containers, release.ContainerName)

	// 3. Patch 更新镜像
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, release.ContainerName, task.TargetImage)
	_, err = kube.AppsV1().StatefulSets(release.Namespace).Patch(ctx, release.WorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	// 4. 等待 Rollout 完成
	timeout := time.Duration(release.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	if err := e.waitStatefulSetRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout); err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// waitStatefulSetRollout 等待 StatefulSet Rollout 完成
func (e *CicdTaskExecutor) waitStatefulSetRollout(ctx context.Context, kube kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("rollout timeout after %v", timeout)
		case <-ticker.C:
			sts, err := kube.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get statefulset failed: %w", err)
			}

			replicas := int32(1)
			if sts.Spec.Replicas != nil {
				replicas = *sts.Spec.Replicas
			}

			if sts.Status.UpdatedReplicas == replicas &&
				sts.Status.ReadyReplicas == replicas &&
				sts.Status.ObservedGeneration >= sts.Generation {
				return nil
			}
		}
	}
}

// executeDaemonSet 执行 DaemonSet 部署
func (e *CicdTaskExecutor) executeDaemonSet(ctx context.Context, kube kubernetes.Interface, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取当前 DaemonSet
	ds, err := kube.AppsV1().DaemonSets(release.Namespace).Get(ctx, release.WorkloadName, metav1.GetOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("获取DaemonSet失败: %v", err)
		return result
	}

	// 2. 保存原镜像
	result.PrevImage = e.getContainerImage(ds.Spec.Template.Spec.Containers, release.ContainerName)

	// 3. Patch 更新镜像
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, release.ContainerName, task.TargetImage)
	_, err = kube.AppsV1().DaemonSets(release.Namespace).Patch(ctx, release.WorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	// 4. 等待 Rollout 完成
	timeout := time.Duration(release.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	if err := e.waitDaemonSetRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout); err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// waitDaemonSetRollout 等待 DaemonSet Rollout 完成
func (e *CicdTaskExecutor) waitDaemonSetRollout(ctx context.Context, kube kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("rollout timeout after %v", timeout)
		case <-ticker.C:
			ds, err := kube.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get daemonset failed: %w", err)
			}

			if ds.Status.UpdatedNumberScheduled == ds.Status.DesiredNumberScheduled &&
				ds.Status.NumberReady == ds.Status.DesiredNumberScheduled &&
				ds.Status.ObservedGeneration >= ds.Generation {
				return nil
			}
		}
	}
}

// getContainerImage 获取容器镜像
func (e *CicdTaskExecutor) getContainerImage(containers []corev1.Container, containerName string) string {
	for _, c := range containers {
		if c.Name == containerName {
			return c.Image
		}
	}
	return ""
}
