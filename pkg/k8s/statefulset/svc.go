package statefulset

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

func CreateServiceFromStatefulSet(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeStatefulSetCreateRequest,
) (*corev1.Service, error) {

	// 1) 构造 Headless Service
	svc := BuildServiceFromStatefulSetReq(req) // name 里已做了回退

	// 2) 创建
	createdSvc, err := Kube.CoreV1().
		Services(req.Namespace).
		Create(ctx, svc, metav1.CreateOptions{FieldManager: "k8soperation"})
	if err != nil {
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("forbidden to create service %s/%s: %w", req.Namespace, svc.Name, err)
		}
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("service %s/%s already exists: %w", req.Namespace, svc.Name, err)
		}
		global.Logger.Errorf("create headless service %s/%s failed: %v", req.Namespace, svc.Name, err)
		return nil, fmt.Errorf("failed to create service %s/%s: %w", req.Namespace, svc.Name, err)
	}

	global.Logger.Infof("headless service %s/%s created successfully", createdSvc.Namespace, createdSvc.Name)
	return createdSvc, nil
}
