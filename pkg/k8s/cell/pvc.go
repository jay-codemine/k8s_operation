package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type PersistentVolumeClaimCell corev1.PersistentVolumeClaim

func (pvc *PersistentVolumeClaimCell) GetCreation() time.Time {
	return pvc.CreationTimestamp.Time
}

func (pvc *PersistentVolumeClaimCell) GetName() string {
	return pvc.Name
}
