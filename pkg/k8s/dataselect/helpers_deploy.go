package dataselect

import (
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"

	"k8soperation/pkg/k8s/cell"
)

// -------------------------------
// 通用：把切片元素安全地转为指针并映射为 DataCell
// -------------------------------
func ToPointerCells[T any](items []T, convert func(*T) DataCell) []DataCell {
	cells := make([]DataCell, 0, len(items))
	for i := range items {
		cells = append(cells, convert(&items[i]))
	}
	return cells
}

// -------------------------------
// Pod
// -------------------------------
func ToPodCells(items []corev1.Pod) []DataCell {
	return ToPointerCells(items, func(p *corev1.Pod) DataCell {
		return (*cell.PodCell)(p)
	})
}

func FromPodCells(cells []DataCell) []corev1.Pod {
	out := make([]corev1.Pod, 0, len(cells))
	for _, c := range cells {
		if pc, ok := c.(*cell.PodCell); ok {
			out = append(out, corev1.Pod(*pc))
		}
	}
	return out
}

func NewPodSelector(items []corev1.Pod, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToPodCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// -------------------------------
// Deployment
// -------------------------------

func ToDeploymentCells(items []appv1.Deployment) []DataCell {
	return ToPointerCells(items, func(d *appv1.Deployment) DataCell {
		return (*cell.DeploymentCell)(d)
	})
}

func FromDeploymentCells(cells []DataCell) []appv1.Deployment {
	out := make([]appv1.Deployment, 0, len(cells))
	for _, c := range cells {
		if dc, ok := c.(*cell.DeploymentCell); ok {
			out = append(out, appv1.Deployment(*dc))
		}
	}
	return out
}

func NewDeploymentSelector(items []appv1.Deployment, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToDeploymentCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// -------------------------------
// StatefulSet
// -------------------------------

func ToStatefulSetCells(items []appv1.StatefulSet) []DataCell {
	return ToPointerCells(items, func(s *appv1.StatefulSet) DataCell {
		return (*cell.StatefulSetCell)(s)
	})
}

func FromStatefulSetCells(cells []DataCell) []appv1.StatefulSet {
	out := make([]appv1.StatefulSet, 0, len(cells))
	for _, c := range cells {
		if sc, ok := c.(*cell.StatefulSetCell); ok {
			out = append(out, appv1.StatefulSet(*sc))
		}
	}
	return out
}

func NewStatefulSetSelector(items []appv1.StatefulSet, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToStatefulSetCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// -------------------------------
// DaemonSet
// -------------------------------

func ToDaemonSetCells(items []appv1.DaemonSet) []DataCell {
	return ToPointerCells(items, func(ds *appv1.DaemonSet) DataCell {
		return (*cell.DaemonSetCell)(ds)
	})
}

func FromDaemonSetCells(cells []DataCell) []appv1.DaemonSet {
	out := make([]appv1.DaemonSet, 0, len(cells))
	for _, c := range cells {
		if dsc, ok := c.(*cell.DaemonSetCell); ok {
			out = append(out, appv1.DaemonSet(*dsc))
		}
	}
	return out
}

func NewDaemonSetSelector(items []appv1.DaemonSet, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToDaemonSetCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// -------------------------------
// Service
// -------------------------------

func ToServiceCells(items []v1.Service) []DataCell {
	return ToPointerCells(items, func(svc *v1.Service) DataCell {
		return (*cell.ServiceCell)(svc)
	})
}

func FromServiceCells(cells []DataCell) []v1.Service {
	out := make([]v1.Service, 0, len(cells))
	for _, c := range cells {
		if sc, ok := c.(*cell.ServiceCell); ok {
			out = append(out, v1.Service(*sc))
		}
	}
	return out
}

// 统一命名为 Selector；保留一个别名兼容旧方法名
func NewServiceSelector(items []v1.Service, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToServiceCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// 兼容：旧名字（如果不需要兼容，可删除）
func NewServiceCell(items []v1.Service, name string, page, limit int) DataSelector {
	return NewServiceSelector(items, name, page, limit)
}

// -------------------------------
// Job
// -------------------------------

func ToJobCells(items []batchv1.Job) []DataCell {
	return ToPointerCells(items, func(j *batchv1.Job) DataCell {
		return (*cell.JobCell)(j)
	})
}

func FromJobCells(cells []DataCell) []batchv1.Job {
	out := make([]batchv1.Job, 0, len(cells))
	for _, c := range cells {
		if jc, ok := c.(*cell.JobCell); ok {
			out = append(out, batchv1.Job(*jc))
		}
	}
	return out
}

func NewJobSelector(items []batchv1.Job, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToJobCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},                 // 按名称过滤（配合你的 Filter 实现）
			Paginate: &PaginateQuery{Limit: limit, Page: page}, // 分页
		},
	}
}

// -------------------------------
// CronJob
// -------------------------------

func ToCronJobCells(items []batchv1.CronJob) []DataCell {
	return ToPointerCells(items, func(cj *batchv1.CronJob) DataCell {
		return (*cell.CronJobCell)(cj)
	})
}

func FromCronJobCells(cells []DataCell) []batchv1.CronJob {
	out := make([]batchv1.CronJob, 0, len(cells))
	for _, c := range cells {
		if cjc, ok := c.(*cell.CronJobCell); ok {
			out = append(out, batchv1.CronJob(*cjc))
		}
	}
	return out
}

func NewCronJobSelector(items []batchv1.CronJob, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToCronJobCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToIngressCells 将 Ingress 列表转换为通用 DataCell 列表
func ToIngressCells(items []networkingv1.Ingress) []DataCell {
	return ToPointerCells(items, func(ing *networkingv1.Ingress) DataCell {
		return (*cell.IngressCell)(ing)
	})
}

// FromIngressCells 将通用 DataCell 列表还原为 Ingress 列表
func FromIngressCells(cells []DataCell) []networkingv1.Ingress {
	out := make([]networkingv1.Ingress, 0, len(cells))
	for _, c := range cells {
		if ic, ok := c.(*cell.IngressCell); ok {
			out = append(out, networkingv1.Ingress(*ic))
		}
	}
	return out
}

// NewIngressSelector 构造 Ingress 的 DataSelector
// 支持名称过滤 + 分页
func NewIngressSelector(items []networkingv1.Ingress, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToIngressCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToSecretCells 将 Secret 列表转换为通用 DataCell 列表
func ToSecretCells(items []corev1.Secret) []DataCell {
	return ToPointerCells(items, func(sec *corev1.Secret) DataCell {
		return (*cell.SecretCell)(sec)
	})
}

// FromSecretCells 将通用 DataCell 列表还原为 Secret 列表
func FromSecretCells(cells []DataCell) []corev1.Secret {
	out := make([]corev1.Secret, 0, len(cells))
	for _, c := range cells {
		if sc, ok := c.(*cell.SecretCell); ok {
			out = append(out, corev1.Secret(*sc))
		}
	}
	return out
}

// NewSecretSelector 构造 Secret 的 DataSelector（名称过滤 + 分页）
func NewSecretSelector(items []corev1.Secret, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToSecretCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToConfigMapCells 将 ConfigMap 列表转换为通用 DataCell 列表
func ToConfigMapCells(items []corev1.ConfigMap) []DataCell {
	return ToPointerCells(items, func(cm *corev1.ConfigMap) DataCell {
		return (*cell.ConfigMapCell)(cm)
	})
}

// FromConfigMapCells 将通用 DataCell 列表还原为 ConfigMap 列表
func FromConfigMapCells(cells []DataCell) []corev1.ConfigMap {
	out := make([]corev1.ConfigMap, 0, len(cells))
	for _, c := range cells {
		if cmc, ok := c.(*cell.ConfigMapCell); ok {
			out = append(out, corev1.ConfigMap(*cmc))
		}
	}
	return out
}

// NewConfigMapSelector 构造 ConfigMap 的 DataSelector（名称过滤 + 分页）
func NewConfigMapSelector(items []corev1.ConfigMap, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToConfigMapCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToStorageClassCells 将 StorageClass 列表转换为通用 DataCell 列表
func ToStorageClassCells(items []storagev1.StorageClass) []DataCell {
	return ToPointerCells(items, func(sc *storagev1.StorageClass) DataCell {
		return (*cell.StorageClassCell)(sc)
	})
}

// FromStorageClassCells 将通用 DataCell 列表还原为 StorageClass 列表
func FromStorageClassCells(cells []DataCell) []storagev1.StorageClass {
	out := make([]storagev1.StorageClass, 0, len(cells))
	for _, c := range cells {
		if scc, ok := c.(*cell.StorageClassCell); ok {
			out = append(out, storagev1.StorageClass(*scc))
		}
	}
	return out
}

// NewStorageClassSelector 构造 StorageClass 的 DataSelector（名称过滤 + 分页）
func NewStorageClassSelector(items []storagev1.StorageClass, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToStorageClassCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToPersistentVolumeCells 将 PersistentVolume 列表转换为通用 DataCell 列表
func ToPersistentVolumeCells(items []corev1.PersistentVolume) []DataCell {
	return ToPointerCells(items, func(pv *corev1.PersistentVolume) DataCell {
		return (*cell.PersistentVolumeCell)(pv)
	})
}

// FromPersistentVolumeCells 将通用 DataCell 列表还原为 PersistentVolume 列表
func FromPersistentVolumeCells(cells []DataCell) []corev1.PersistentVolume {
	out := make([]corev1.PersistentVolume, 0, len(cells))
	for _, c := range cells {
		if pvc, ok := c.(*cell.PersistentVolumeCell); ok {
			out = append(out, corev1.PersistentVolume(*pvc))
		}
	}
	return out
}

// NewPersistentVolumeSelector 构造 PersistentVolume 的 DataSelector（名称过滤 + 分页）
func NewPersistentVolumeSelector(items []corev1.PersistentVolume, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToPersistentVolumeCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToPersistentVolumeClaimCells 将 PersistentVolumeClaim 列表转换为通用 DataCell 列表
func ToPersistentVolumeClaimCells(items []corev1.PersistentVolumeClaim) []DataCell {
	return ToPointerCells(items, func(pvc *corev1.PersistentVolumeClaim) DataCell {
		return (*cell.PersistentVolumeClaimCell)(pvc)
	})
}

// FromPersistentVolumeClaimCells 将通用 DataCell 列表还原为 PersistentVolumeClaim 列表
func FromPersistentVolumeClaimCells(cells []DataCell) []corev1.PersistentVolumeClaim {
	out := make([]corev1.PersistentVolumeClaim, 0, len(cells))
	for _, c := range cells {
		if pvc, ok := c.(*cell.PersistentVolumeClaimCell); ok {
			out = append(out, corev1.PersistentVolumeClaim(*pvc))
		}
	}
	return out
}

// NewPersistentVolumeClaimSelector 构造 PVC 的 DataSelector（名称过滤 + 分页）
func NewPersistentVolumeClaimSelector(items []corev1.PersistentVolumeClaim, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToPersistentVolumeClaimCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToNodeCells 将 Node 列表转换为通用 DataCell 列表
func ToNodeCells(items []corev1.Node) []DataCell {
	return ToPointerCells(items, func(n *corev1.Node) DataCell {
		return (*cell.NodeCell)(n)
	})
}

// FromNodeCells 将通用 DataCell 列表还原为 Node 列表
func FromNodeCells(cells []DataCell) []corev1.Node {
	out := make([]corev1.Node, 0, len(cells))
	for _, c := range cells {
		if n, ok := c.(*cell.NodeCell); ok {
			out = append(out, corev1.Node(*n))
		}
	}
	return out
}

// NewNodeSelector 构造 Node 的 DataSelector（名称过滤 + 分页）
func NewNodeSelector(items []corev1.Node, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToNodeCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToNamespaceCells 将 Namespace 列表转换为通用 DataCell 列表
func ToNamespaceCells(items []corev1.Namespace) []DataCell {
	return ToPointerCells(items, func(n *corev1.Namespace) DataCell {
		return (*cell.NamespaceCell)(n)
	})
}

// FromNamespaceCells 将通用 DataCell 列表还原为 Namespace 列表
func FromNamespaceCells(cells []DataCell) []corev1.Namespace {
	out := make([]corev1.Namespace, 0, len(cells))
	for _, c := range cells {
		if n, ok := c.(*cell.NamespaceCell); ok {
			out = append(out, corev1.Namespace(*n))
		}
	}
	return out
}

// NewNamespaceSelector 构造 Namespace 的 DataSelector（名称过滤 + 分页）
// Namespace 本身也是 cluster-scoped 资源，没有 namespace 字段。
func NewNamespaceSelector(items []corev1.Namespace, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToNamespaceCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}

// ToAppConfigCells 将 AppConfig 列表转换为通用 DataCell 列表
func ToAppConfigCells(items []appv1alpha1.AppConfig) []DataCell {
	return ToPointerCells(items, func(a *appv1alpha1.AppConfig) DataCell {
		return (*cell.AppConfigCell)(a)
	})
}

// FromAppConfigCells 将通用 DataCell 列表还原为 AppConfig 列表
func FromAppConfigCells(cells []DataCell) []appv1alpha1.AppConfig {
	out := make([]appv1alpha1.AppConfig, 0, len(cells))
	for _, c := range cells {
		if a, ok := c.(*cell.AppConfigCell); ok {
			out = append(out, appv1alpha1.AppConfig(*a))
		}
	}
	return out
}

// NewAppConfigSelector 构造 AppConfig 的 DataSelector（名称过滤 + 分页）
//
// 注意：AppConfig 是 namespace-scoped，namespace 一般在 List 的时候通过 client.InNamespace(ns) 先筛过了，
// 这里主要做 name 模糊匹配 + 分页。
func NewAppConfigSelector(items []appv1alpha1.AppConfig, name string, page, limit int) DataSelector {
	return DataSelector{
		GenericDataList: ToAppConfigCells(items),
		DataSelect: &DataSelectQuery{
			Filter:   &FilterQuery{Name: name},
			Paginate: &PaginateQuery{Limit: limit, Page: page},
		},
	}
}
