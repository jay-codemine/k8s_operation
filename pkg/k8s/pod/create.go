package pod

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/requests"
)

func CreatePod(
	ctx context.Context,
	kube kubernetes.Interface,
	param *requests.KubePodCreateRequest,
) (*corev1.Pod, error) {

	containerName := param.ContainerName
	if containerName == "" {
		containerName = param.Name // 兜底：容器名 = Pod 名
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
			Labels:    param.Labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  containerName,
					Image: param.Image,
				},
			},
		},
	}

	return kube.CoreV1().
		Pods(param.Namespace).
		Create(ctx, pod, metav1.CreateOptions{})
}
