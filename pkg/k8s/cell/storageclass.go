package cell

import (
	storagev1 "k8s.io/api/storage/v1"
	"time"
)

type StorageClassCell storagev1.StorageClass

func (s *StorageClassCell) GetCreation() time.Time {
	return s.CreationTimestamp.Time
}

func (s *StorageClassCell) GetName() string {
	return s.Name
}
