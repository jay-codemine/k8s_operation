package cell

import (
	batchv1 "k8s.io/api/batch/v1"
	"time"
)

type CronJobCell batchv1.CronJob

func (d *CronJobCell) GetCreation() time.Time {
	return d.CreationTimestamp.Time
}

func (d *CronJobCell) GetName() string {
	return d.Name
}
