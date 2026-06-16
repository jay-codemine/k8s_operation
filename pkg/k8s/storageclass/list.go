package storageclass

import (
	"context"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetStorageClassList 列出 StorageClass（支持名称模糊过滤、排序、分页）
func GetStorageClassList(ctx context.Context, Kube kubernetes.Interface, name string, page, limit int) ([]storagev1.StorageClass, int, error) {
	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 拉取 SC 列表（集群级资源，无 namespace）
	list, err := Kube.StorageV1().
		StorageClasses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	// 名称过滤 + 排序 + 分页（与其他资源保持一致）
	selector := dataselect.NewStorageClassSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromStorageClassCells(data.GenericDataList), total, nil
}
