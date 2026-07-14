package cronjob

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CronJobJobItem 该 CronJob 触发的 Job 记录
type CronJobJobItem struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`      // Running / Succeeded / Failed
	Active     int32  `json:"active"`
	Succeeded  int32  `json:"succeeded"`
	Failed     int32  `json:"failed"`
	StartTime  string `json:"start_time"`
	Duration   string `json:"duration"`
}

// GetCronJobOwnedJobs 获取指定 CronJob 所创建的 Job 列表
func GetCronJobOwnedJobs(ctx context.Context, kube kubernetes.Interface, namespace, cronjobName string) ([]CronJobJobItem, error) {
	jobs, err := kube.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list Jobs failed: %w", err)
	}

	var result []CronJobJobItem
	for _, job := range jobs.Items {
		// 检查 ownerReferences 是否指向目标 CronJob
		isOwned := false
		for _, owner := range job.OwnerReferences {
			if owner.Kind == "CronJob" && owner.Name == cronjobName {
				isOwned = true
				break
			}
		}
		if !isOwned {
			continue
		}

		item := CronJobJobItem{
			Name:      job.Name,
			Namespace: job.Namespace,
			Active:    job.Status.Active,
			Succeeded: job.Status.Succeeded,
			Failed:    job.Status.Failed,
		}

		// 判断状态
		if job.Status.Active > 0 {
			item.Status = "Running"
		} else if job.Status.Failed > 0 {
			item.Status = "Failed"
		} else if job.Status.Succeeded > 0 {
			item.Status = "Succeeded"
		} else {
			item.Status = "Unknown"
		}

		if job.Status.StartTime != nil {
			item.StartTime = job.Status.StartTime.Time.Format("2006-01-02 15:04:05")
		}
		if job.Status.CompletionTime != nil && job.Status.StartTime != nil {
			d := job.Status.CompletionTime.Sub(job.Status.StartTime.Time)
			item.Duration = d.String()
		}

		result = append(result, item)
	}

	return result, nil
}
