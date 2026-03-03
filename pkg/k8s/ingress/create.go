package ingress

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress 创建 Ingress
func CreateIngress(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeIngressCreateRequest) (*networkingv1.Ingress, error) {
	ing := BuildIngressFromReq(req)

	created, err := Kube.NetworkingV1().
		Ingresses(req.Namespace).
		Create(ctx, ing, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("ingress %q already exists in namespace %q", ing.Name, ing.Namespace)
		}
		global.Logger.Errorf("create ingress failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("ingress %q created successfully", created.Name)
	return created, nil
}
