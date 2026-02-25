package cronjob

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCronJobDetail(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*batchv1.CronJob, []batchv1.Job, error) {
	// 1. 获取 CronJob
	cj, err := Kube.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 2. 获取 CronJob 对应的 Jobs
	labelSelector := fmt.Sprintf("batch.kubernetes.io/cronjob-name=%s", name)
	jobList, err := Kube.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return cj, nil, err
	}

	return cj, jobList.Items, nil
}
