package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type NamespaceCell corev1.Namespace

func (n *NamespaceCell) GetCreation() time.Time {
	return n.CreationTimestamp.Time
}

func (n *NamespaceCell) GetName() string {
	return n.Name
}
