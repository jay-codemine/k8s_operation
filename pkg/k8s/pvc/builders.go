package pvc

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8soperation/internal/app/requests"
)

func BuildPVCFromReq(req *requests.KubePVCCreateRequest) *corev1.PersistentVolumeClaim {
	var selector *metav1.LabelSelector
	if len(req.SelectorMatchLabels) > 0 {
		selector = &metav1.LabelSelector{MatchLabels: req.SelectorMatchLabels}
	}
	
	// 处理访问模式：优先使用 AccessModes，其次使用 AccessMode
	accessModes := req.AccessModes
	if len(accessModes) == 0 && req.AccessMode != "" {
		// 如果没有 AccessModes 但有 AccessMode，将其转换为 AccessModes
		accessModes = []corev1.PersistentVolumeAccessMode{
			corev1.PersistentVolumeAccessMode(req.AccessMode),
		}
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			// 这里直接用请求里的类型：[]corev1.PersistentVolumeAccessMode
			AccessModes: accessModes,

			// 新类型：VolumeResourceRequirements（不是 ResourceRequirements）
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParse(req.Storage), // e.g. "10Gi"
				},
			},

			Selector: selector,
		},
	}

	// 可选字段
	if req.StorageClassName != "" {
		pvc.Spec.StorageClassName = &req.StorageClassName
	}
	if req.VolumeMode != nil {
		pvc.Spec.VolumeMode = req.VolumeMode // Filesystem / Block
	}

	return pvc
}

func mustParse(q string) resource.Quantity {
	v, _ := resource.ParseQuantity(q)
	return v
}
