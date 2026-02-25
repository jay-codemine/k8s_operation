package pod

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

// PodListItem Pod 列表项响应结构
type PodListItem struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`        // Running/Pending/Failed/Succeeded/Unknown/CrashLoopBackOff/ImagePullBackOff等
	StatusReason string `json:"status_reason"` // 状态原因
	Node         string `json:"node"`
	PodIP        string `json:"pod_ip"`
	HostIP       string `json:"host_ip"`
	Image        string `json:"image"`        // 第一个容器的镜像
	RestartCount int32  `json:"restart_count"` // 第一个容器的重启次数
	CreatedAt    string `json:"created_at"`
	Containers   []string `json:"containers"`   // 容器名称列表（用于前端日志选择）
	Conditions   []corev1.PodCondition       `json:"conditions"`         // 原始 conditions
	ContainerStatuses []corev1.ContainerStatus `json:"container_statuses"` // 容器状态列表
}

// BuildPodListResponse 将 Pod 列表转换为响应格式
func BuildPodListResponse(pods []corev1.Pod) []PodListItem {
	result := make([]PodListItem, 0, len(pods))

	for _, pod := range pods {
		// 提取容器信息
		var image string
		var restartCount int32

		if len(pod.Spec.Containers) > 0 {
			image = pod.Spec.Containers[0].Image
		}

		if len(pod.Status.ContainerStatuses) > 0 {
			restartCount = pod.Status.ContainerStatuses[0].RestartCount
			// 如果 spec 中没有镜像，从 status 获取
			if image == "" {
				image = pod.Status.ContainerStatuses[0].Image
			}
		}

		// 提取容器名称列表
		containers := make([]string, 0, len(pod.Spec.Containers))
		for _, c := range pod.Spec.Containers {
			containers = append(containers, c.Name)
		}

		// 获取状态
		status, reason := getPodStatus(&pod)

		item := PodListItem{
			Name:              pod.Name,
			Namespace:         pod.Namespace,
			Status:            status,
			StatusReason:      reason,
			Node:              pod.Spec.NodeName,
			PodIP:             pod.Status.PodIP,
			HostIP:            pod.Status.HostIP,
			Image:             image,
			RestartCount:      restartCount,
			CreatedAt:         pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Containers:        containers,
			Conditions:        pod.Status.Conditions,
			ContainerStatuses: pod.Status.ContainerStatuses,
		}

		result = append(result, item)
	}

	return result
}

// getPodStatus 根据 Pod 的 status 判断状态
func getPodStatus(pod *corev1.Pod) (status, reason string) {
	// 1. 首先检查 Pod Phase
	phase := pod.Status.Phase

	// 2. 检查容器状态 - 这是最重要的部分
	for _, cs := range pod.Status.ContainerStatuses {
		// 2.1 容器正在等待
		if cs.State.Waiting != nil {
			waiting := cs.State.Waiting
			reason = waiting.Reason

			// 常见的等待原因
			switch waiting.Reason {
			case "ContainerCreating", "PodInitializing":
				status = "Pending"
				reason = fmt.Sprintf("容器创建中: %s", waiting.Message)
				return
			case "CrashLoopBackOff":
				status = "CrashLoopBackOff"
				reason = fmt.Sprintf("容器崩溃重启: %s", waiting.Message)
				return
			case "ImagePullBackOff", "ErrImagePull":
				status = "ImagePullBackOff"
				reason = fmt.Sprintf("镜像拉取失败: %s", waiting.Message)
				return
			case "CreateContainerConfigError":
				status = "Failed"
				reason = fmt.Sprintf("容器配置错误: %s", waiting.Message)
				return
			case "InvalidImageName":
				status = "Failed"
				reason = fmt.Sprintf("无效的镜像名称: %s", waiting.Message)
				return
			default:
				status = "Pending"
				reason = fmt.Sprintf("%s: %s", waiting.Reason, waiting.Message)
				return
			}
		}

		// 2.2 容器正在终止
		if cs.State.Terminated != nil {
			terminated := cs.State.Terminated
			if terminated.ExitCode != 0 {
				status = "Failed"
				reason = fmt.Sprintf("容器异常退出(退出码:%d): %s", terminated.ExitCode, terminated.Reason)
				return
			}
			// 正常退出的情况继续检查其他容器
		}

		// 2.3 容器未就绪但正在运行
		if !cs.Ready && cs.State.Running != nil {
			status = "Running"
			reason = "容器运行中但未就绪"
			// 继续检查其他容器
		}
	}

	// 3. 检查 Init 容器状态
	for _, ics := range pod.Status.InitContainerStatuses {
		if ics.State.Waiting != nil {
			waiting := ics.State.Waiting
			status = "Pending"
			reason = fmt.Sprintf("Init 容器: %s - %s", waiting.Reason, waiting.Message)
			return
		}
		if ics.State.Terminated != nil && ics.State.Terminated.ExitCode != 0 {
			status = "Failed"
			reason = fmt.Sprintf("Init 容器失败(退出码:%d)", ics.State.Terminated.ExitCode)
			return
		}
	}

	// 4. 检查 Pod Conditions
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
			status = "Pending"
			reason = fmt.Sprintf("调度失败: %s", cond.Message)
			return
		}
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionFalse && phase == corev1.PodRunning {
			status = "NotReady"
			reason = fmt.Sprintf("未就绪: %s", cond.Message)
			return
		}
	}

	// 5. 根据 Phase 返回状态
	switch phase {
	case corev1.PodPending:
		status = "Pending"
		reason = "Pod 等待调度或容器启动中"
		// 检查是否有更具体的原因
		if pod.Status.Reason != "" {
			reason = pod.Status.Reason + ": " + pod.Status.Message
		}

	case corev1.PodRunning:
		// 所有容器都在运行且就绪
		allReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			status = "Running"
			reason = "所有容器运行正常"
		} else {
			status = "Running"
			reason = "部分容器未就绪"
		}

	case corev1.PodSucceeded:
		status = "Succeeded"
		reason = "Pod 已成功完成"

	case corev1.PodFailed:
		status = "Failed"
		reason = "Pod 执行失败"
		if pod.Status.Reason != "" {
			reason = pod.Status.Reason + ": " + pod.Status.Message
		}

	case corev1.PodUnknown:
		status = "Unknown"
		reason = "无法获取 Pod 状态"

	default:
		status = "Unknown"
		reason = fmt.Sprintf("未知阶段: %s", phase)
	}

	return
}
