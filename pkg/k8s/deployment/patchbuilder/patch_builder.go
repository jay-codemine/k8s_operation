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

// BuildResourcesPatch 构造修改容器资源限制的 patch（StrategicMergePatch）
// 用于快速调整 CPU/Memory requests/limits，应对 OOM 等紧急场景
func BuildResourcesPatch(containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) ([]byte, error) {
	if containerName == "" {
		return nil, fmt.Errorf("containerName must not be empty")
	}
	if memoryLimit == "" {
		return nil, fmt.Errorf("memoryLimit must not be empty (required to prevent OOM)")
	}

	resources := map[string]interface{}{}
	requests := map[string]string{}
	limits := map[string]string{}

	if cpuRequest != "" {
		requests["cpu"] = cpuRequest
	}
	if memoryRequest != "" {
		requests["memory"] = memoryRequest
	}
	if cpuLimit != "" {
		limits["cpu"] = cpuLimit
	}
	if memoryLimit != "" {
		limits["memory"] = memoryLimit
	}

	if len(requests) > 0 {
		resources["requests"] = requests
	}
	if len(limits) > 0 {
		resources["limits"] = limits
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":      containerName,
							"resources": resources,
						},
					},
				},
			},
		},
	}
	return json.Marshal(patch)
}
