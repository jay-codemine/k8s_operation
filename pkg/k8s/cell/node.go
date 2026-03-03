package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

// NodeCell 用于封装 corev1.Node，使其实现 DataCell 接口
type NodeCell corev1.Node

// GetCreation 获取创建时间（用于排序）
func (n *NodeCell) GetCreation() time.Time {
	return n.CreationTimestamp.Time
}

// GetName 获取节点名称（用于模糊搜索或唯一标识）
func (n *NodeCell) GetName() string {
	return n.Name
}
