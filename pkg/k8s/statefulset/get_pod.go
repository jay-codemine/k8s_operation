package statefulset

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

// GetPodByStatefulSet 获取某个 StatefulSet 对应的所有 Pod
func GetPodByStatefulSet(ctx context.Context, Kube kubernetes.Interface, namespace, stsName string) ([]corev1.Pod, error) {
	// 读取 StatefulSet
	sts, err := Kube.AppsV1().StatefulSets(namespace).
		Get(ctx, stsName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("statefulset %s/%s not found", namespace, stsName)
		}
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to get statefulset %s/%s", namespace, stsName)
		}
		return nil, fmt.Errorf("failed to get statefulset %s/%s: %v", namespace, stsName, err)
	}

	// 构造 LabelSelector
	if sts.Spec.Selector == nil {
		return nil, fmt.Errorf("statefulset %s/%s has no selector", namespace, stsName)
	}
	selector := metav1.FormatLabelSelector(sts.Spec.Selector)

	// 获取匹配的 Pod 列表
	podList, err := Kube.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("no permission to list pods in namespace %s: %v", namespace, err)
		}
		return nil, fmt.Errorf("failed to list pods for statefulset %s/%s: %v", namespace, stsName, err)
	}

	if len(podList.Items) == 0 {
		global.Logger.Infof("no pods found for statefulset %s/%s", namespace, stsName)
	}

	return podList.Items, nil
}
