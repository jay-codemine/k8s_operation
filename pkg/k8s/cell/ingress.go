package cell

import (
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

// IngressCell 将 Ingress 封装为通用 DataCell
type IngressCell networkingv1.Ingress

// GetCreation 获取创建时间，用于排序
func (i *IngressCell) GetCreation() time.Time {
	return i.CreationTimestamp.Time
}

// GetName 获取资源名称，用于过滤
func (i *IngressCell) GetName() string {
	return i.Name
}
