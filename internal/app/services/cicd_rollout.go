package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
)

// =============================================================================
// 统一滚动更新等待函数 — release flow 和 auto-deploy flow 共用
// 所有函数均为同步阻塞，等 rollout 全部完成或超时/失败后才返回
// =============================================================================

// RolloutResult Rollout 完成后的状态信息
type RolloutResult struct {
	Ready     int32 // 就绪 Pod 数
	Total     int32 // 期望 Pod 数
	Available int32 // 可用 Pod 数
}

// ==================== Deployment ====================

// WaitDeploymentRollout 等待 Deployment 滚动更新完成（5 条件 + Pod 故障检测）
// 超时默认 5 分钟，每 5 秒轮询一次
func WaitDeploymentRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration, logs *strings.Builder) (*RolloutResult, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		dp, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
		}

		replicas := int32(1)
		if dp.Spec.Replicas != nil {
			replicas = *dp.Spec.Replicas
		}

		logRollout(logs, "DEPLOY", replicas, dp.Status.Replicas,
			dp.Status.UpdatedReplicas, dp.Status.ReadyReplicas,
			dp.Status.AvailableReplicas, dp.Status.ObservedGeneration, dp.Generation)

		// 检查 Rollout 是否超过 ProgressDeadline
		for _, cond := range dp.Status.Conditions {
			if cond.Type == appv1.DeploymentProgressing && cond.Status == corev1.ConditionFalse {
				return nil, fmt.Errorf("rollout 失败: %s", cond.Message)
			}
			if cond.Reason == "ProgressDeadlineExceeded" {
				return nil, fmt.Errorf("rollout 进度超时: %s", cond.Message)
			}
		}

		// Pod 级别故障检测
		if dp.Spec.Selector != nil {
			if err := checkPodErrors(ctx, client, namespace, dp.Spec.Selector, logs); err != nil {
				return nil, err
			}
		}

		// 5 条件全部满足
		if dp.Status.ObservedGeneration >= dp.Generation &&
			dp.Status.UpdatedReplicas == replicas &&
			dp.Status.Replicas == dp.Status.UpdatedReplicas &&
			dp.Status.ReadyReplicas == replicas &&
			dp.Status.AvailableReplicas == replicas {
			logs.WriteString(fmt.Sprintf("[SUCCESS] Deployment 所有 %d 个副本已就绪（Ready=%d, Available=%d）\n",
				replicas, dp.Status.ReadyReplicas, dp.Status.AvailableReplicas))
			return &RolloutResult{
				Ready:     dp.Status.ReadyReplicas,
				Total:     replicas,
				Available: dp.Status.AvailableReplicas,
			}, nil
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("Deployment rollout 超时（%v），请检查 Pod 状态", timeout)
}

// ==================== StatefulSet ====================

// WaitStatefulSetRollout 等待 StatefulSet 滚动更新完成（5 条件 + Pod 故障检测）
// StatefulSet 按序号从高到低逐个更新
func WaitStatefulSetRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration, logs *strings.Builder) (*RolloutResult, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 StatefulSet 失败: %w", err)
		}

		replicas := int32(1)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		logRollout(logs, "STS", replicas, sts.Status.Replicas,
			sts.Status.UpdatedReplicas, sts.Status.ReadyReplicas,
			sts.Status.CurrentReplicas, sts.Status.ObservedGeneration, sts.Generation)

		// Pod 级别故障检测
		if sts.Spec.Selector != nil {
			if err := checkPodErrors(ctx, client, namespace, sts.Spec.Selector, logs); err != nil {
				return nil, err
			}
		}

		// 5 条件：
		// ① 控制器已处理最新配置
		// ② 所有 Pod 已更新到新版本
		// ③ 无旧版本 Pod 残留
		// ④ 所有 Pod 就绪
		// ⑤ CurrentRevision == UpdateRevision（确保版本切换完成）
		if sts.Status.ObservedGeneration >= sts.Generation &&
			sts.Status.UpdatedReplicas == replicas &&
			sts.Status.CurrentReplicas == sts.Status.UpdatedReplicas &&
			sts.Status.ReadyReplicas == replicas &&
			sts.Status.CurrentRevision == sts.Status.UpdateRevision {
			logs.WriteString(fmt.Sprintf("[SUCCESS] StatefulSet 所有 %d 个副本已就绪（Updated=%d, Ready=%d, Revision=%s）\n",
				replicas, sts.Status.UpdatedReplicas, sts.Status.ReadyReplicas, sts.Status.CurrentRevision))
			return &RolloutResult{
				Ready:     sts.Status.ReadyReplicas,
				Total:     replicas,
				Available: sts.Status.ReadyReplicas,
			}, nil
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("StatefulSet rollout 超时（%v），请检查 Pod 状态", timeout)
}

// ==================== DaemonSet ====================

// WaitDaemonSetRollout 等待 DaemonSet 滚动更新完成（4 条件 + Pod 故障检测）
func WaitDaemonSetRollout(ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration, logs *strings.Builder) (*RolloutResult, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
		}

		desired := ds.Status.DesiredNumberScheduled

		logRollout(logs, "DS", desired, ds.Status.CurrentNumberScheduled,
			ds.Status.UpdatedNumberScheduled, ds.Status.NumberReady,
			ds.Status.NumberAvailable, ds.Status.ObservedGeneration, ds.Generation)

		// Pod 级别故障检测
		if ds.Spec.Selector != nil {
			if err := checkPodErrors(ctx, client, namespace, ds.Spec.Selector, logs); err != nil {
				return nil, err
			}
		}

		// 4 条件：
		// ① 控制器已处理最新配置
		// ② 所有节点 Pod 已更新
		// ③ 所有 Pod 就绪
		// ④ 没有不可用的 Pod
		if ds.Status.ObservedGeneration >= ds.Generation &&
			ds.Status.UpdatedNumberScheduled == desired &&
			ds.Status.NumberReady == desired &&
			ds.Status.NumberUnavailable == 0 {
			logs.WriteString(fmt.Sprintf("[SUCCESS] DaemonSet 所有 %d 个 Pod 已就绪（Updated=%d, Ready=%d）\n",
				desired, ds.Status.UpdatedNumberScheduled, ds.Status.NumberReady))
			return &RolloutResult{
				Ready:     ds.Status.NumberReady,
				Total:     desired,
				Available: ds.Status.NumberAvailable,
			}, nil
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("DaemonSet rollout 超时（%v），请检查 Pod 状态", timeout)
}

// ==================== CronJob ====================

// ValidateCronJob 校验 CronJob 配置是否有效
// 检查是否有至少一个 Job 被成功创建过，防止镜像名称错误/配置异常
func ValidateCronJob(ctx context.Context, client kubernetes.Interface, namespace, name string) error {
	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 CronJob 失败: %w", err)
	}

	// 检查是否有活跃的 Job 或最近成功的 Job
	if len(cj.Status.Active) == 0 && cj.Status.LastScheduleTime == nil {
		global.Logger.Warn("CronJob 尚未被调度过，无法验证镜像是否有效",
			zap.String("namespace", namespace),
			zap.String("name", name))
		return nil // 不阻塞，CronJob 可能在下次调度时才触发
	}

	// 如果有活跃 Job，检查其 Pod 状态
	if len(cj.Status.Active) > 0 {
		for _, activeRef := range cj.Status.Active {
			job, err := client.BatchV1().Jobs(activeRef.Namespace).Get(ctx, activeRef.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			selector := metav1.FormatLabelSelector(job.Spec.Selector)
			pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
			if err != nil {
				continue
			}
			for _, pod := range pods.Items {
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil {
						reason := cs.State.Waiting.Reason
						if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
							return fmt.Errorf("CronJob 镜像验证失败: Pod %s 处于 %s", pod.Name, reason)
						}
					}
				}
			}
		}
	}

	return nil
}

// ==================== Job ====================

// RecreateJobAndWait 删除旧 Job 并创建新 Job，等待完成
// Job 的 Pod 模板不可变，所以必须 delete + create
func RecreateJobAndWait(ctx context.Context, client kubernetes.Interface, namespace, name, containerName, targetImage string, timeout time.Duration, logs *strings.Builder) (*RolloutResult, error) {
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	// 1. 获取旧 Job，保存配置
	oldJob, err := client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Job 失败: %w", err)
	}

	// 2. 更新容器镜像
	containerFound := false
	for i, c := range oldJob.Spec.Template.Spec.Containers {
		if c.Name == containerName {
			oldJob.Spec.Template.Spec.Containers[i].Image = targetImage
			containerFound = true
			break
		}
	}
	if !containerFound {
		return nil, fmt.Errorf("容器 '%s' 不存在于 Job %s/%s", containerName, namespace, name)
	}

	// 3. 删除旧 Job（使用 Background 删除策略，让旧 Pod 被清理）
	deletePolicy := metav1.DeletePropagationBackground
	if err := client.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		return nil, fmt.Errorf("删除旧 Job 失败: %w", err)
	}
	logs.WriteString(fmt.Sprintf("[INFO] 已删除旧 Job %s/%s\n", namespace, name))

	// 4. 等待旧 Job 完全删除
	for i := 0; i < 30; i++ {
		_, err := client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			break // Job 已删除
		}
		time.Sleep(2 * time.Second)
	}

	// 5. 创建新 Job（清理旧 Job 的元数据）
	newJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      oldJob.Labels,
			Annotations: oldJob.Annotations,
		},
		Spec: oldJob.Spec,
	}
	// 清理不可变字段
	newJob.ResourceVersion = ""
	newJob.UID = ""
	newJob.Generation = 0
	newJob.Spec.Selector = nil // Job 的 selector 由控制器自动生成

	newJob, err = client.BatchV1().Jobs(namespace).Create(ctx, newJob, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建新 Job 失败: %w", err)
	}
	logs.WriteString(fmt.Sprintf("[INFO] 已创建新 Job %s/%s，等待完成...\n", namespace, name))

	// 6. 等待新 Job 完成
	return waitJobComplete(ctx, client, namespace, name, timeout, logs)
}

// waitJobComplete 等待 Job 成功完成
func waitJobComplete(ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration, logs *strings.Builder) (*RolloutResult, error) {
	interval := 5 * time.Second
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		job, err := client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Job 失败: %w", err)
		}

		// 获取关联 Pod 状态
		var runningPods, succeededPods, failedPods int32
		if job.Spec.Selector != nil {
			labelSelector := metav1.FormatLabelSelector(job.Spec.Selector)
			pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
			if err == nil {
				for _, pod := range pods.Items {
					switch pod.Status.Phase {
					case corev1.PodRunning:
						runningPods++
					case corev1.PodSucceeded:
						succeededPods++
					case corev1.PodFailed:
						failedPods++
					}
				}
			}
		}

		completions := int32(1)
		if job.Spec.Completions != nil {
			completions = *job.Spec.Completions
		}

		logs.WriteString(fmt.Sprintf("[JOB] 完成: %d/%d | 运行: %d | 成功: %d | 失败: %d | Active: %d\n",
			job.Status.Succeeded, completions, job.Status.Active, succeededPods, job.Status.Failed, job.Status.Active))

		// Pod 级别故障检测
		if job.Spec.Selector != nil {
			if err := checkPodErrors(ctx, client, namespace, job.Spec.Selector, logs); err != nil {
				return nil, err
			}
		}

		// 检查是否完成
		if job.Status.Succeeded >= completions {
			logs.WriteString(fmt.Sprintf("[SUCCESS] Job 已完成: %d/%d 成功\n", job.Status.Succeeded, completions))
			return &RolloutResult{
				Ready:     job.Status.Succeeded,
				Total:     completions,
				Available: job.Status.Succeeded,
			}, nil
		}

		// 检查是否失败（超过 backoffLimit）
		if job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit {
			return nil, fmt.Errorf("Job 失败次数超过限制: %d > %d", job.Status.Failed, *job.Spec.BackoffLimit)
		}

		// 检查 Pod 级别失败
		if failedPods > 0 && job.Status.Active == 0 && job.Status.Succeeded == 0 {
			return nil, fmt.Errorf("Job Pod 全部失败（%d 个失败），请检查容器日志", failedPods)
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("Job 执行超时（%v）", timeout)
}

// ==================== 通用工具函数 ====================

// logRollout 统一输出 rollout 进度日志
func logRollout(logs *strings.Builder, kind string, desired, current, updated, ready, available int32, observedGen, gen int64) {
	if logs == nil {
		return
	}
	logs.WriteString(fmt.Sprintf("[ROLLOUT-%s] 期望: %d | 当前: %d | 更新: %d | 就绪: %d | 可用: %d | Gen: %d/%d\n",
		kind, desired, current, updated, ready, available, observedGen, gen))
}

// checkPodErrors 检查关联 Pod 是否有致命错误（ImagePullBackOff / CrashLoopBackOff）
func checkPodErrors(ctx context.Context, client kubernetes.Interface, namespace string, selector *metav1.LabelSelector, logs *strings.Builder) error {
	if selector == nil {
		return nil
	}

	labelSelector := metav1.FormatLabelSelector(selector)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil // 获取 Pod 失败不阻塞
	}

	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				msg := cs.State.Waiting.Message

				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					errMsg := fmt.Sprintf("镜像拉取失败 [%s]: %s", reason, msg)
					if logs != nil {
						logs.WriteString(fmt.Sprintf("[ERROR] Pod %s: %s\n", pod.Name, errMsg))
					}
					return errors.New(errMsg)
				}
				if reason == "CrashLoopBackOff" {
					errMsg := fmt.Sprintf("容器崩溃重启 [%s]: %s", reason, msg)
					if logs != nil {
						logs.WriteString(fmt.Sprintf("[ERROR] Pod %s: %s\n", pod.Name, errMsg))
					}
					return errors.New(errMsg)
				}
			}
		}
	}

	return nil
}

// PatchWorkloadImage 通用的镜像 Patch 更新
func PatchWorkloadImage(ctx context.Context, client kubernetes.Interface, patchFn func(context.Context, kubernetes.Interface, string, string, string, string) error,
	namespace, name, container, image string, logs *strings.Builder) error {

	if logs != nil {
		logs.WriteString(fmt.Sprintf("[INFO] 正在更新 %s/%s 的镜像: %s -> %s\n", namespace, name, container, image))
	}

	if err := patchFn(ctx, client, namespace, name, container, image); err != nil {
		return fmt.Errorf("更新镜像失败: %w", err)
	}

	if logs != nil {
		logs.WriteString(fmt.Sprintf("[INFO] 镜像更新已提交\n"))
	}

	return nil
}

// PatchDeploymentImageFn Deployment 镜像 patch 函数签名
func PatchDeploymentImageFn(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	_, err := client.AppsV1().Deployments(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	return err
}

// PatchStatefulSetImageFn StatefulSet 镜像 patch 函数签名
func PatchStatefulSetImageFn(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	_, err := client.AppsV1().StatefulSets(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	return err
}

// PatchDaemonSetImageFn DaemonSet 镜像 patch 函数签名
func PatchDaemonSetImageFn(ctx context.Context, client kubernetes.Interface, namespace, name, container, image string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`, container, image)
	_, err := client.AppsV1().DaemonSets(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	return err
}
