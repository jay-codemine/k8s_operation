package pv

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetPVList 获取 PV 列表（名称过滤 + 排序 + 分页）
func GetPVList(ctx context.Context, Kube kubernetes.Interface, name string, page, limit int) ([]corev1.PersistentVolume, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 集群级，无 namespace
	list, err := Kube.CoreV1().
		PersistentVolumes().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		// PV 是集群级，不会 NotFound namespace；保留兜底
		if apierrors.IsNotFound(err) {
			return nil, 0, fmt.Errorf("cluster PersistentVolumes not found")
		}
		return nil, 0, err
	}

	selector := dataselect.NewPersistentVolumeSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromPersistentVolumeCells(data.GenericDataList), total, nil
}
