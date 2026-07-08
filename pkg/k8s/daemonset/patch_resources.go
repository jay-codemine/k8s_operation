package daemonset

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/deployment/patchbuilder"
)

// PatchDaemonSetResources 修改 DaemonSet 容器资源配置（requests/limits）
func PatchDaemonSetResources(ctx context.Context, Kube kubernetes.Interface, namespace, name, containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) (*appv1.DaemonSet, error) {
	patchData, err := patchbuilder.BuildResourcesPatch(containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit)
	if err != nil {
		return nil, err
	}
	return PatchDaemonSet(ctx, Kube, namespace, name, patchData)
}
