package cell

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type PersistentVolumeCell corev1.PersistentVolume

func (pv *PersistentVolumeCell) GetCreation() time.Time {
	return pv.CreationTimestamp.Time
}

func (pv *PersistentVolumeCell) GetName() string {
	return pv.Name
}
