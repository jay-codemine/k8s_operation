package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type PodCell corev1.Pod

func (p PodCell) GetCreation() time.Time {
	return p.CreationTimestamp.Time
}

func (p PodCell) GetName() string {
	return p.Name
}
