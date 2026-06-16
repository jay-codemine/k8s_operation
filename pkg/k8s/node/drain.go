package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// DrainNode cordon 节点并驱逐其上可驱逐 Pod
func DrainNode(ctx context.Context, Kube kubernetes.Interface, nodeName string) error {
	// 1) 先 cordon 节点
	if err := CordonNode(ctx, Kube, nodeName, true); err != nil {
		return fmt.Errorf("cordon node %s failed: %w", nodeName, err)
	}

	// 2) 列出该节点上的所有 Pod（所有 namespace）
	podList, err := Kube.CoreV1().Pods("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("spec.nodeName", nodeName).String(),
		})
	if err != nil {
		return fmt.Errorf("list pods.js on node %s failed: %w", nodeName, err)
	}

	// 3) 遍历 Pod，过滤掉 DaemonSet / Mirror Pod，然后逐个驱逐
	for _, pod := range podList.Items {
		// 跳过已经完成/终止的 Pod
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}

		// 跳过 DaemonSet Pod
		if isDaemonSetPod(&pod) {
			continue
		}

		// 跳过 Mirror Pod（静态 Pod）
		if isMirrorPod(&pod) {
			continue
		}

		// 这里可以按需：跳过 kube-system 某些关键组件，自己看情况：
		if pod.Namespace == "kube-system" {
			continue
		}

		// 4) 调用 Eviction API 驱逐 Pod
		graceSeconds := int64(30) // 你也可以做成入参
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &graceSeconds,
			},
		}

		err := Kube.PolicyV1().
			Evictions(eviction.Namespace).
			Evict(ctx, eviction)
		if err != nil {
			// 如果是 NotFound 就忽略
			if apierrors.IsNotFound(err) {
				global.Logger.Warn("pod already gone when evict",
					zap.String("node", nodeName),
					zap.String("pod", pod.Name),
					zap.String("ns", pod.Namespace),
				)
				continue
			}
			global.Logger.Error("evict pod failed",
				zap.String("node", nodeName),
				zap.String("pod", pod.Name),
				zap.String("ns", pod.Namespace),
				zap.Error(err),
			)
			return fmt.Errorf("evict pod %s/%s failed: %w", pod.Namespace, pod.Name, err)
		}
	}

	return nil
}

// 是否是 DaemonSet 管理的 Pod
func isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" && owner.Controller != nil && *owner.Controller {
			return true
		}
	}
	return false
}

// 是否是 Mirror Pod（静态 Pod）
func isMirrorPod(pod *corev1.Pod) bool {
	if _, ok := pod.Annotations["kubernetes.io/config.mirror"]; ok {
		return true
	}
	return false
}
