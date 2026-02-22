package cell

import (
	appv1 "k8s.io/api/apps/v1"
	"time"
)

type StatefulSetCell appv1.StatefulSet

func (s *StatefulSetCell) GetCreation() time.Time {
	return s.CreationTimestamp.Time
}

func (s *StatefulSetCell) GetName() string {
	return s.Name
}
