package deployment

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/deployment/patchbuilder"
)

// PatchDeployment 修改Deployment
func PatchDeployment(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*appv1.Deployment, error) {
	dep, err := Kube.AppsV1().
		Deployments(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return dep, nil
}

// PatchDeploymentReplicas 修改副本数
func PatchDeploymentReplicas(ctx context.Context, Kube kubernetes.Interface, namespace, name string, replicas int32) (*appv1.Deployment, error) {
	patchReplicas, err := patchbuilder.BuildReplicasPatch(replicas)
	if err != nil {
		return nil, err
	}
	return PatchDeployment(ctx, Kube, namespace, name, patchReplicas)
}

// PatchDeploymentImage 修改容器镜像
func PatchDeploymentImage(ctx context.Context, Kube kubernetes.Interface, namespace, name, containerName, image string) (*appv1.Deployment, error) {
	patchImage, err := patchbuilder.BuildImagePatch(containerName, image)
	if err != nil {
		return nil, err
	}
	return PatchDeployment(ctx, Kube, namespace, name, patchImage)
}

// PatchDeploymentResources 修改容器资源配置（requests/limits）
func PatchDeploymentResources(ctx context.Context, Kube kubernetes.Interface, namespace, name, containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) (*appv1.Deployment, error) {
	patchData, err := patchbuilder.BuildResourcesPatch(containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit)
	if err != nil {
		return nil, err
	}
	return PatchDeployment(ctx, Kube, namespace, name, patchData)
}
