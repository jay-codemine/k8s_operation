// DeletePod 删除Pod
package pod

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeletePod 统一封装 Pod 删除（包含默认值/force策略）
// graceSeconds: nil 表示未传；force=true 表示强制删除（grace=0 + 后台级联）
func DeletePod(ctx context.Context, kube kubernetes.Interface, namespace,
	name string, graceSeconds *int64, force bool,
) error {
	opts := metav1.DeleteOptions{}

	// 默认 grace
	if graceSeconds != nil {
		opts.GracePeriodSeconds = graceSeconds
	} else {
		def := int64(30)
		opts.GracePeriodSeconds = &def
	}

	// force 策略
	if force {
		zero := int64(0)
		opts.GracePeriodSeconds = &zero
		policy := metav1.DeletePropagationBackground
		opts.PropagationPolicy = &policy
	}

	return kube.CoreV1().Pods(namespace).Delete(ctx, name, opts)
}
