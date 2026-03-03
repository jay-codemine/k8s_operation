package patchbuilder

import (
	"encoding/json"
	"fmt"
)

// 构造修改 replicas 的 patch
func BuildReplicasPatch(replicas int32) ([]byte, error) {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": replicas,
		},
	}
	return json.Marshal(patch)
}

// 构造修改镜像的 patch（指定容器名）
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
