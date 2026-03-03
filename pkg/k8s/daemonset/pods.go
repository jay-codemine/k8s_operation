package daemonset

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetDaemonSetPods 获取 DaemonSet 管理的 Pod 列表
func GetDaemonSetPods(ctx context.Context, kube kubernetes.Interface, namespace, name string) ([]corev1.Pod, error) {
	// 1. 获取 DaemonSet
	ds, err := kube.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}

	// 2. 获取 selector
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析 selector 失败: %w", err)
	}

	// 3. 列出匹配的 Pods
	podList, err := kube.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 4. 过滤：只返回属于该 DaemonSet 的 Pod（通过 OwnerReference 判断）
	var result []corev1.Pod
	for _, pod := range podList.Items {
		for _, ref := range pod.OwnerReferences {
			if ref.Kind == "DaemonSet" && ref.Name == name {
				result = append(result, pod)
				break
			}
		}
	}

	return result, nil
}

// GetDaemonSetPodsBySelector 通过 selector 获取 Pod 列表（备用方法）
func GetDaemonSetPodsBySelector(ctx context.Context, kube kubernetes.Interface, namespace string, labelSelector map[string]string) ([]corev1.Pod, error) {
	selector := labels.SelectorFromSet(labelSelector)
	podList, err := kube.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}
