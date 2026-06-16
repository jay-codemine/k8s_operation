package pod

import (
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func GetPodLogStream(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace, container string,
	tail *int64,
) (io.ReadCloser, error) {

	opts := &corev1.PodLogOptions{
		Container:  container,
		Follow:     true,
		Timestamps: global.PodLogSetting.Timestamps,
		Previous:   global.PodLogSetting.Previous,
	}

	// 核心：不再“强行补默认值”
	if tail != nil {
		// 可选：安全上限
		t := *tail
		if max := global.PodLogSetting.TailMax; max > 0 && t > max {
			t = max
		}
		opts.TailLines = &t
	}

	return kube.CoreV1().
		Pods(namespace).
		GetLogs(name, opts).
		Stream(ctx)
}
