package node

import (
	"context"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventItem 事件简化结构
type EventItem struct {
	Type           string      `json:"type"`            // Normal/Warning
	Reason         string      `json:"reason"`          // 事件原因
	Message        string      `json:"message"`         // 事件消息
	Count          int32       `json:"count"`           // 发生次数
	FirstTimestamp metav1.Time `json:"first_timestamp"` // 首次发生时间
	LastTimestamp  metav1.Time `json:"last_timestamp"`  // 最后发生时间
	Source         string      `json:"source"`          // 事件来源组件
}

// GetNodeEvents 获取节点相关事件
func GetNodeEvents(ctx context.Context, kube kubernetes.Interface, nodeName string, limit int) ([]EventItem, error) {
	// 查询与节点相关的事件
	// 事件在 default 命名空间，involvedObject.name = nodeName, involvedObject.kind = Node
	eventList, err := kube.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.kind=Node,involvedObject.name=" + nodeName,
	})
	if err != nil {
		return nil, err
	}

	// 按最后发生时间倒序排序
	events := eventList.Items
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastTimestamp.After(events[j].LastTimestamp.Time)
	})

	// 限制数量
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	// 转换为简化结构
	result := make([]EventItem, 0, len(events))
	for _, e := range events {
		source := e.Source.Component
		if e.Source.Host != "" {
			source += "/" + e.Source.Host
		}

		result = append(result, EventItem{
			Type:           e.Type,
			Reason:         e.Reason,
			Message:        e.Message,
			Count:          e.Count,
			FirstTimestamp: e.FirstTimestamp,
			LastTimestamp:  e.LastTimestamp,
			Source:         source,
		})
	}

	return result, nil
}

// GetNodeConditions 获取节点条件状态
func GetNodeConditions(nodeObj *corev1.Node) []map[string]interface{} {
	conditions := make([]map[string]interface{}, 0, len(nodeObj.Status.Conditions))
	for _, c := range nodeObj.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":               c.Type,
			"status":             c.Status,
			"reason":             c.Reason,
			"message":            c.Message,
			"last_heartbeat":     c.LastHeartbeatTime,
			"last_transition":    c.LastTransitionTime,
		})
	}
	return conditions
}

// GetNodeSystemInfo 获取节点系统信息
func GetNodeSystemInfo(nodeObj *corev1.Node) map[string]string {
	info := nodeObj.Status.NodeInfo
	return map[string]string{
		"machine_id":               info.MachineID,
		"system_uuid":              info.SystemUUID,
		"boot_id":                  info.BootID,
		"kernel_version":           info.KernelVersion,
		"os_image":                 info.OSImage,
		"container_runtime":        info.ContainerRuntimeVersion,
		"kubelet_version":          info.KubeletVersion,
		"kube_proxy_version":       info.KubeProxyVersion,
		"operating_system":         info.OperatingSystem,
		"architecture":             info.Architecture,
	}
}
