package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/deployment/patchbuilder"
)

func PatchImageStatefulSet(ctx context.Context, Kube kubernetes.Interface, namespace, name, container, image string) (*appv1.StatefulSet, error) {
	patchBytes, err := patchbuilder.BuildImagePatch(container, image)
	if err != nil {
		return nil, err
	}

	sts, err := Kube.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 StatefulSet 镜像失败: %w", err)
	}
	return sts, nil
}
