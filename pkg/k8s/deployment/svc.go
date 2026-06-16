package deployment

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

// CreateServiceFromDeploymentReq 根据 Deployment 创建 Service
func CreateServiceFromDeployment(ctx context.Context, kube kubernetes.Interface, req *requests.KubeDeploymentCreateRequest) (*corev1.Service, error) {
	svc := BuildServiceFromDeploymentReq(req)

	createdSvc, err := kube.CoreV1().
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
