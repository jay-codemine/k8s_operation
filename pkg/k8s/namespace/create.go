package namespace

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

func CreateNamespace(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeNamespaceCreateRequest) (*corev1.Namespace, error) {
	// 1) 兜底默认值
	cpu := req.QuotaCPU
	if cpu == "" {
		cpu = "4"
	}
	mem := req.QuotaMemory
	if mem == "" {
		mem = "8Gi"
	}
	pods := req.QuotaPods
	if pods == "" {
		pods = "110"
	}

	// 2) 构造 Namespace 对象（使用构造函数）
	ns := BuildNamespaceFromReq(req)

	// 3) 创建 Namespace
	created, err := Kube.CoreV1().
		Namespaces().
		Create(ctx, ns, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("namespace %q already exists", req.Name)
			exist, getErr := Kube.CoreV1().
				Namespaces().
				Get(ctx, req.Name, metav1.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}
			return exist, nil
		}
		return nil, err
	}

	global.Logger.Infof("namespace %q created", created.Name)

	// 4) 创建 ResourceQuota（使用构造函数）
	quota := BuildQuotaForNamespace(req.Name, cpu, mem, pods)
	if _, err := Kube.CoreV1().
		ResourceQuotas(req.Name).
		Create(ctx, quota, metav1.CreateOptions{}); err != nil {

		global.Logger.Errorf("quota create failed: %v", err)
		return created, fmt.Errorf("namespace created but quota failed: %w", err)
	}

	// 5) 创建 LimitRange（使用构造函数）
	limit := BuildLimitRangeForNamespace(req.Name)
	if _, err := Kube.CoreV1().
		LimitRanges(req.Name).
		Create(ctx, limit, metav1.CreateOptions{}); err != nil {

		global.Logger.Errorf("limitrange create failed: %v", err)
		return created, fmt.Errorf("namespace created but limitrange failed: %w", err)
	}

	return created, nil
}
