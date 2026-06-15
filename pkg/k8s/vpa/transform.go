package vpa

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// VPAItem 前端友好结构
type VPAItem struct {
	Name              string                 `json:"name"`
	Namespace         string                 `json:"namespace"`
	TargetKind        string                 `json:"target_kind"`
	TargetName        string                 `json:"target_name"`
	UpdateMode        string                 `json:"update_mode"`
	MinAllowedCPU     string                 `json:"min_allowed_cpu,omitempty"`
	MinAllowedMem     string                 `json:"min_allowed_mem,omitempty"`
	MaxAllowedCPU     string                 `json:"max_allowed_cpu,omitempty"`
	MaxAllowedMem     string                 `json:"max_allowed_mem,omitempty"`
	ContainerName     string                 `json:"container_name,omitempty"`
	ControlledRes     []string               `json:"controlled_resources,omitempty"`
	Recommendation    map[string]interface{} `json:"recommendation,omitempty"` // 推荐资源
	Status            string                 `json:"status"`
	Labels            map[string]string      `json:"labels,omitempty"`
	CreationTimestamp string                 `json:"creation_timestamp"`
}

// BuildVPAItem 把 Unstructured VPA 转成前端友好结构
func BuildVPAItem(u *unstructured.Unstructured) VPAItem {
	item := VPAItem{
		Name:              u.GetName(),
		Namespace:         u.GetNamespace(),
		Labels:            u.GetLabels(),
		CreationTimestamp: u.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		Status:            "Pending",
	}

	// targetRef
	if tr, found, _ := unstructuredNestedMap(u.Object, "spec", "targetRef"); found {
		if v, ok := tr["kind"].(string); ok {
			item.TargetKind = v
		}
		if v, ok := tr["name"].(string); ok {
			item.TargetName = v
		}
	}

	// updatePolicy.updateMode
	if mode, found, _ := unstructuredNestedString(u.Object, "spec", "updatePolicy", "updateMode"); found {
		item.UpdateMode = mode
	} else {
		item.UpdateMode = "Auto"
	}

	// resourcePolicy.containerPolicies[0]
	if list, found, _ := unstructuredNestedSlice(u.Object, "spec", "resourcePolicy", "containerPolicies"); found && len(list) > 0 {
		if cp, ok := list[0].(map[string]interface{}); ok {
			if v, ok := cp["containerName"].(string); ok {
				item.ContainerName = v
			}
			if cr, ok := cp["controlledResources"].([]interface{}); ok {
				resList := make([]string, 0, len(cr))
				for _, r := range cr {
					if s, ok := r.(string); ok {
						resList = append(resList, s)
					}
				}
				item.ControlledRes = resList
			}
			if minA, ok := cp["minAllowed"].(map[string]interface{}); ok {
				if v, ok := minA["cpu"].(string); ok {
					item.MinAllowedCPU = v
				}
				if v, ok := minA["memory"].(string); ok {
					item.MinAllowedMem = v
				}
			}
			if maxA, ok := cp["maxAllowed"].(map[string]interface{}); ok {
				if v, ok := maxA["cpu"].(string); ok {
					item.MaxAllowedCPU = v
				}
				if v, ok := maxA["memory"].(string); ok {
					item.MaxAllowedMem = v
				}
			}
		}
	}

	// status.recommendation - 提取推荐数据
	if rec, found, _ := unstructuredNestedMap(u.Object, "status", "recommendation"); found {
		item.Recommendation = rec
		item.Status = "Active"
	}

	return item
}

// BuildVPAItemList 批量转换
func BuildVPAItemList(list []unstructured.Unstructured) []VPAItem {
	out := make([]VPAItem, 0, len(list))
	for i := range list {
		out = append(out, BuildVPAItem(&list[i]))
	}
	return out
}

// 内部帮助函数：手写以避免引入大量依赖
func unstructuredNestedMap(obj map[string]interface{}, fields ...string) (map[string]interface{}, bool, error) {
	cur := obj
	for i, f := range fields {
		v, ok := cur[f]
		if !ok {
			return nil, false, nil
		}
		if i == len(fields)-1 {
			m, ok := v.(map[string]interface{})
			return m, ok, nil
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return nil, false, nil
		}
		cur = next
	}
	return nil, false, nil
}

func unstructuredNestedString(obj map[string]interface{}, fields ...string) (string, bool, error) {
	cur := obj
	for i, f := range fields {
		v, ok := cur[f]
		if !ok {
			return "", false, nil
		}
		if i == len(fields)-1 {
			s, ok := v.(string)
			return s, ok, nil
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return "", false, nil
		}
		cur = next
	}
	return "", false, nil
}

func unstructuredNestedSlice(obj map[string]interface{}, fields ...string) ([]interface{}, bool, error) {
	cur := obj
	for i, f := range fields {
		v, ok := cur[f]
		if !ok {
			return nil, false, nil
		}
		if i == len(fields)-1 {
			s, ok := v.([]interface{})
			return s, ok, nil
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return nil, false, nil
		}
		cur = next
	}
	return nil, false, nil
}
