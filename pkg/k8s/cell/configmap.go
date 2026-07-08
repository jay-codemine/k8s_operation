package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type ConfigMapCell corev1.ConfigMap

func (c *ConfigMapCell) GetCreation() time.Time {
	return c.CreationTimestamp.Time
}

func (c *ConfigMapCell) GetName() string {
	return c.Name
}
