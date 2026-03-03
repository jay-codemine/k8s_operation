package cell

import (
	batchv1 "k8s.io/api/batch/v1"
	"time"
)

type JobCell batchv1.Job

func (d *JobCell) GetCreation() time.Time {
	return d.CreationTimestamp.Time
}

func (d *JobCell) GetName() string {
	return d.Name
}
