package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// CreateStatefulSetWithService：创建 StatefulSet，并按需创建（或复用）Service；失败则回滚删除 STS
func CreateStatefulSetWithService(ctx context.Context, kube kubernetes.Interface, req *requests.KubeStatefulSetCreateRequest,
) (*appv1.StatefulSet, *corev1.Service, error) {

	// 1) 创建 StatefulSet（建议你的 CreateStatefulSet 也支持 kube 传入）
	sts, err := CreateStatefulSet(ctx, kube, req)
	if err != nil {
		return nil, nil, fmt.Errorf("create statefulset failed: %w", err)
	}
	global.Logger.Infof("statefulset %s/%s created successfully", req.Namespace, req.Name)

	// 2) 不需要创建 Service，直接返回
	if !req.IsCreateService {
		return sts, nil, nil
	}

	// 3) 创建 Service（同样建议支持 kube 传入）
	svcObj, err := CreateServiceFromStatefulSet(ctx, kube, req)
	if err != nil {
		// 3.1 Service 已存在：复用
		if apierrors.IsAlreadyExists(err) {
			exist, gerr := kube.CoreV1().
				Services(req.Namespace).
				Get(ctx, req.ServiceName, metav1.GetOptions{})
			if gerr == nil {
				global.Logger.Infof("service %s/%s already exists, reuse it", req.Namespace, req.ServiceName)
				return sts, exist, nil
			}
		}

		// 3.2 真失败：回滚删除 StatefulSet
		pol := metav1.DeletePropagationForeground
		_ = kube.AppsV1().
			StatefulSets(req.Namespace).
			Delete(ctx, sts.Name, metav1.DeleteOptions{PropagationPolicy: &pol})
		global.Logger.Errorf("rollback delete statefulset %s/%s after service failed: %v", req.Namespace, req.Name, err)

		return nil, nil, fmt.Errorf("create service failed: %w", err)
	}

	global.Logger.Infof("service %s/%s created successfully", req.Namespace, req.ServiceName)
	return sts, svcObj, nil
}
