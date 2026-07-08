package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/models"
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
	case "CronJob":
		return e.executeCronJob(ctx, cli.Kube, task, release)
	case "Job":
		return e.executeJob(ctx, cli.Kube, task, release)
	default:
		result.Message = fmt.Sprintf("不支持的工作负载类型: %s", release.WorkloadKind)
		return result
	}
}

// ==================== Deployment ====================

// executeDeployment 执行 Deployment 部署（统一使用 WaitDeploymentRollout）
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
	if err := PatchDeploymentImageFn(ctx, kube, release.Namespace, release.WorkloadName, release.ContainerName, task.TargetImage); err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	global.Logger.Info("patched deployment image",
		zap.String("namespace", release.Namespace),
		zap.String("deployment", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 4. 等待 Rollout 完成（统一逻辑：5 条件 + Pod 故障检测）
	timeout := time.Duration(release.TimeoutSec) * time.Second
	var logs strings.Builder
	_, err = WaitDeploymentRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout, &logs)
	if err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v\n%s", err, logs.String())
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// ==================== StatefulSet ====================

// executeStatefulSet 执行 StatefulSet 部署（统一使用 WaitStatefulSetRollout）
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
	if err := PatchStatefulSetImageFn(ctx, kube, release.Namespace, release.WorkloadName, release.ContainerName, task.TargetImage); err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	global.Logger.Info("patched statefulset image",
		zap.String("namespace", release.Namespace),
		zap.String("statefulset", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 4. 等待 Rollout 完成（统一逻辑：5 条件 + Pod 故障检测）
	timeout := time.Duration(release.TimeoutSec) * time.Second
	var logs strings.Builder
	_, err = WaitStatefulSetRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout, &logs)
	if err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v\n%s", err, logs.String())
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// ==================== DaemonSet ====================

// executeDaemonSet 执行 DaemonSet 部署（统一使用 WaitDaemonSetRollout）
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
	if err := PatchDaemonSetImageFn(ctx, kube, release.Namespace, release.WorkloadName, release.ContainerName, task.TargetImage); err != nil {
		result.Message = fmt.Sprintf("更新镜像失败: %v", err)
		return result
	}

	global.Logger.Info("patched daemonset image",
		zap.String("namespace", release.Namespace),
		zap.String("daemonset", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 4. 等待 Rollout 完成（统一逻辑：4 条件 + Pod 故障检测）
	timeout := time.Duration(release.TimeoutSec) * time.Second
	var logs strings.Builder
	_, err = WaitDaemonSetRollout(ctx, kube, release.Namespace, release.WorkloadName, timeout, &logs)
	if err != nil {
		result.Message = fmt.Sprintf("等待Rollout完成失败: %v\n%s", err, logs.String())
		return result
	}

	result.Success = true
	result.Message = "部署成功"
	return result
}

// ==================== CronJob ====================

// executeCronJob 执行 CronJob 镜像更新
// CronJob 不等待 Rollout（模板更新后下次调度生效），但会做快速校验
func (e *CicdTaskExecutor) executeCronJob(ctx context.Context, kube kubernetes.Interface, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取当前 CronJob
	cj, err := kube.BatchV1().CronJobs(release.Namespace).Get(ctx, release.WorkloadName, metav1.GetOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("获取CronJob失败: %v", err)
		return result
	}

	// 2. 保存原镜像
	result.PrevImage = e.getContainerImage(cj.Spec.JobTemplate.Spec.Template.Spec.Containers, release.ContainerName)

	// 3. Patch 更新镜像（CronJob 的容器在 spec.jobTemplate.spec.template.spec.containers 中）
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"jobTemplate": map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{"name": release.ContainerName, "image": task.TargetImage},
							},
						},
					},
				},
			},
		},
	}
	patchBytes, _ := json.Marshal(patchData)
	_, err = kube.BatchV1().CronJobs(release.Namespace).Patch(ctx, release.WorkloadName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		result.Message = fmt.Sprintf("更新CronJob镜像失败: %v", err)
		return result
	}

	global.Logger.Info("patched cronjob image",
		zap.String("namespace", release.Namespace),
		zap.String("cronjob", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 4. 快速校验：检查是否有活跃 Job 存在 Pod 故障
	if err := ValidateCronJob(ctx, kube, release.Namespace, release.WorkloadName); err != nil {
		result.Message = fmt.Sprintf("CronJob 镜像验证失败: %v（模板已更新，下次调度使用新镜像）", err)
		return result
	}

	result.Success = true
	result.Message = "CronJob 镜像更新成功，下次调度将使用新镜像"
	return result
}

// ==================== Job ====================

// executeJob 执行 Job 镜像更新
// Job 的 Pod 模板不可变，采用 delete + create 方式重建，并等待执行完成
func (e *CicdTaskExecutor) executeJob(ctx context.Context, kube kubernetes.Interface, task *models.CicdReleaseTask, release *models.CicdRelease) *CicdExecuteResult {
	result := &CicdExecuteResult{}

	// 1. 获取当前 Job，验证存在
	oldJob, err := kube.BatchV1().Jobs(release.Namespace).Get(ctx, release.WorkloadName, metav1.GetOptions{})
	if err != nil {
		// 如果 Job 不存在，尝试直接创建
		result.Message = fmt.Sprintf("获取Job失败: %v", err)
		return result
	}

	// 2. 保存原镜像
	result.PrevImage = e.getContainerImage(oldJob.Spec.Template.Spec.Containers, release.ContainerName)

	global.Logger.Info("recreating job with new image",
		zap.String("namespace", release.Namespace),
		zap.String("job", release.WorkloadName),
		zap.String("container", release.ContainerName),
		zap.String("prev_image", result.PrevImage),
		zap.String("target_image", task.TargetImage))

	// 3. 删除旧 Job + 创建新 Job + 等待完成
	timeout := time.Duration(release.TimeoutSec) * time.Second
	var logs strings.Builder
	_, err = RecreateJobAndWait(ctx, kube, release.Namespace, release.WorkloadName, release.ContainerName, task.TargetImage, timeout, &logs)
	if err != nil {
		result.Message = fmt.Sprintf("Job 执行失败: %v\n%s", err, logs.String())
		return result
	}

	result.Success = true
	result.Message = "Job 执行成功"
	return result
}

// ==================== 工具函数 ====================

// getContainerImage 获取容器镜像
func (e *CicdTaskExecutor) getContainerImage(containers []corev1.Container, containerName string) string {
	for _, c := range containers {
		if c.Name == containerName {
			return c.Image
		}
	}
	return ""
}
