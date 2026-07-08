package pod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodLog(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace, container string,
	tail *int64,
	follow bool,
) (string, error) {

	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    follow,
	}

	// 只有用户显式传了 tail，才设置
	if tail != nil {
		opts.TailLines = tail
	}

	rc, err := kube.CoreV1().
		Pods(namespace).
		GetLogs(name, opts).
		Stream(ctx)
	if err != nil {
		// 检查是否是容器还在创建中
		if isContainerNotReady(err) {
			// 获取 Pod 状态，提供更详细的信息
			pod, getErr := kube.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			if getErr == nil {
				statusMsg := getPodContainerStatus(pod, container)
				if statusMsg != "" {
					return "", fmt.Errorf("容器还未就绪: %s", statusMsg)
				}
			}
			return "", fmt.Errorf("容器还未就绪，请稍后再试")
		}
		return "", fmt.Errorf("open log stream: %w", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", fmt.Errorf("read log: %w", err)
	}

	return buf.String(), nil
}

// isContainerNotReady 检查错误是否是因为容器未就绪
func isContainerNotReady(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "ContainerCreating") ||
		strings.Contains(errMsg, "PodInitializing") ||
		strings.Contains(errMsg, "waiting to start") ||
		strings.Contains(errMsg, "is waiting to start") ||
		apierrors.IsBadRequest(err)
}

// getPodContainerStatus 获取容器的详细状态信息
func getPodContainerStatus(pod *corev1.Pod, containerName string) string {
	if pod == nil {
		return ""
	}

	// 检查 Init 容器状态
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.Name == containerName {
			if cs.State.Waiting != nil {
				return fmt.Sprintf("Init 容器 %s，原因: %s", cs.State.Waiting.Reason, cs.State.Waiting.Message)
			}
		}
	}

	// 检查普通容器状态
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			if cs.State.Waiting != nil {
				return fmt.Sprintf("%s - %s", cs.State.Waiting.Reason, cs.State.Waiting.Message)
			}
			if cs.State.Terminated != nil {
				return fmt.Sprintf("容器已终止: %s", cs.State.Terminated.Reason)
			}
			if !cs.Ready {
				return "容器未就绪"
			}
		}
	}

	// 检查 Pod 阶段
	switch pod.Status.Phase {
	case corev1.PodPending:
		return "Pod 正在启动中"
	case corev1.PodFailed:
		return "Pod 已失败"
	case corev1.PodSucceeded:
		return "Pod 已完成"
	}

	return ""
}
