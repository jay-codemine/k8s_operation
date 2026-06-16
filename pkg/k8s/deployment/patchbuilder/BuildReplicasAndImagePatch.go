package patchbuilder

import (
	"encoding/json"
	"fmt"
)

// BuildReplicasAndImagePatch 构造同时修改副本数和容器镜像的 Patch
func BuildReplicasAndImagePatch(replicas int32, images map[string]string) ([]byte, error) {
	if len(images) == 0 && replicas < 0 {
		return nil, fmt.Errorf("must provide replicas >=0 or images map")
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{},
	}

	// 设置副本数（>=0 才生效）
	// 检查副本数是否大于等于0
	if replicas >= 0 {
		// 更新patch中spec部分的副本数
		// 使用类型断言将patch["spec"]转换为map[string]interface{}类型
		// 然后设置replicas字段为指定的副本数
		patch["spec"].(map[string]interface{})["replicas"] = replicas
	}

	// 设置镜像
	if len(images) > 0 {
		// 创建一个字符串映射的切片（slice），初始容量为images切片的长度
		// 这是一个空切片，但预分配了足够的内存空间以提高性能
		containers := make([]map[string]string, 0, len(images))
		for name, img := range images {
			if name == "" || img == "" {
				return nil, fmt.Errorf("container name and image must not be empty")
			}
			containers = append(containers, map[string]string{"name": name, "image": img})
		}
		patch["spec"].(map[string]interface{})["template"] = map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": containers,
			},
		}
	}

	return json.Marshal(patch)
}
