package daemonset

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

// CreateServiceFromDaemonSet 创建与 DaemonSet 对应的 Service（可选）
func CreateServiceFromDaemonSet(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeDaemonSetCreateRequest) (*corev1.Service, error) {
	// 1) 构造 Service 对象
	svc := BuildServiceFromDaemonSetReq(req)

	// 2) 创建 Service
	createdSvc, err := Kube.CoreV1().
		Services(req.Namespace).
		Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("service %q already exists in namespace %q", svc.Name, svc.Namespace)
		}
		global.Logger.Errorf("create service failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("service %q created successfully", createdSvc.Name)
	return createdSvc, nil
}
