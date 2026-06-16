package cell

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type ServiceCell v1.Service

func (s *ServiceCell) GetCreation() time.Time {
	return s.CreationTimestamp.Time
}

func (s *ServiceCell) GetName() string {
	return s.Name
}
