package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

// SecretCell 将 Secret 封装为通用 DataCell
type SecretCell corev1.Secret

// GetCreation 获取创建时间，用于排序
func (s *SecretCell) GetCreation() time.Time {
	return s.CreationTimestamp.Time
}

// GetName 获取资源名称，用于过滤
func (s *SecretCell) GetName() string {
	return s.Name
}
