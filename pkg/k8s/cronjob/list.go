package cronjob

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// CronJobListResult 包含 CronJob 列表和关联的 Job 列表
type CronJobListResult struct {
	CronJobs []batchv1.CronJob
	Jobs     []batchv1.Job
	Total    int64
}

func GetCronJobList(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace string,
	page, limit int,
	labelSelector string, // 新增：标签选择器
) (*CronJobListResult, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 构建 ListOptions
	listOpts := metav1.ListOptions{}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}

	// 获取 CronJob 列表
	list, err := kube.BatchV1().CronJobs(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	// 获取同命名空间下的 Job 列表（用于统计）
	jobList, err := kube.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	selector := dataselect.NewCronJobSelector(list.Items, name, page, limit)
	selector.Filter().Sort()

	total := int64(selector.TotalCount())

	pageData := selector.Paginate()

	return &CronJobListResult{
		CronJobs: dataselect.FromCronJobCells(pageData.GenericDataList),
		Jobs:     jobList.Items,
		Total:    total,
	}, nil
}
