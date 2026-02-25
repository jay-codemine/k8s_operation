package deployment

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// CreateServiceFromDeploymentReq 根据 Deployment 创建 Service
// CreateDeployment 只负责创建 Deployment，不管 Service
func CreateDeployment(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeDeploymentCreateRequest) (*appv1.Deployment, error) {
	// 1) 构造 Deployment
	dp := BuildDeploymentFromCreateReq(req)

	// 2) 创建 Deployment
	createdDp, err := Kube.AppsV1().
		Deployments(req.Namespace).
		Create(ctx, dp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("deployment %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Warnf("create deployment failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("deployment %q created successfully", createdDp.Name)
	return createdDp, nil
}
