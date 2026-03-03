package statefulset

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

func CreateStatefulSet(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeStatefulSetCreateRequest) (*appv1.StatefulSet, error) {
	// 1) 构造 StatefulSet
	sts := BuildStatefulSetFromCreateReq(req)

	// 2) 创建 StatefulSet
	createdSts, err := Kube.AppsV1().
		StatefulSets(req.Namespace).
		Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("statefulset %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Warnf("create statefulset failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("statefulset %q created successfully", createdSts.Name)
	return createdSts, nil
}
