package job

import (
	"context"
	"encoding/json"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// PatchJob 修改 Job
func PatchJob(ctx context.Context, Kube kubernetes.Interface, namespace, name string, patch []byte) (*batchv1.Job, error) {
	jobObj, err := Kube.BatchV1().
		Jobs(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return jobObj, nil
}

// PatchJobImage 修改容器镜像
func PatchJobImage(ctx context.Context, Kube kubernetes.Interface, namespace, name, containerName, image string) (*batchv1.Job, error) {
	patchImage, err := BuildImagePatch(containerName, image)
	if err != nil {
		return nil, err
	}
	return PatchJob(ctx, Kube, namespace, name, patchImage)
}

// BuildImagePatch 构造修改镜像的 patch（指定容器名）
func BuildImagePatch(containerName, image string) ([]byte, error) {
	if containerName == "" || image == "" {
		return nil, fmt.Errorf("containerName and image must not be empty")
	}
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]string{
						{
							"name":  containerName,
							"image": image,
						},
					},
				},
			},
		},
	}
	return json.Marshal(patch)
}
