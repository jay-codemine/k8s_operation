package namespace

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

func CreateNamespace(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeNamespaceCreateRequest) (*corev1.Namespace, error) {
	ns := BuildNamespaceFromReq(req)

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
	return created, nil
}
