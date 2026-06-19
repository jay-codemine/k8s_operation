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

// PatchStatefulSetResources 修改 StatefulSet 容器资源配置（requests/limits）
func PatchStatefulSetResources(ctx context.Context, Kube kubernetes.Interface, namespace, name, containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) (*appv1.StatefulSet, error) {
	patchData, err := patchbuilder.BuildResourcesPatch(containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit)
	if err != nil {
		return nil, err
	}

	sts, err := Kube.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 StatefulSet 资源配置失败: %w", err)
	}
	return sts, nil
}
