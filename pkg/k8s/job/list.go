package job

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

func GetJobList(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace string,
	page, limit int,
	labelSelector string, // 新增：标签选择器
) ([]batchv1.Job, int64, error) {
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

	list, err := kube.BatchV1().Jobs(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, 0, err
	}

	selector := dataselect.NewJobSelector(list.Items, name, page, limit)
	selector.Filter().Sort()

	total := int64(selector.TotalCount())

	data := selector.Paginate()

	return dataselect.FromJobCells(data.GenericDataList), total, nil
}
