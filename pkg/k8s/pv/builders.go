package pv

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
)

// BuildPersistentVolumeFromReq 构建 PersistentVolume 对象
func BuildPersistentVolumeFromReq(req *requests.KubePVCreateRequest) *corev1.PersistentVolume {
	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(req.Capacity),
			},
			AccessModes: req.AccessModes,
		},
	}

	// 可选字段
	if req.ReclaimPolicy != "" {
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimPolicy(req.ReclaimPolicy)
	}
	if req.StorageClassName != "" {
		pv.Spec.StorageClassName = req.StorageClassName
	}
	if req.VolumeMode != nil {
		pv.Spec.VolumeMode = req.VolumeMode
	}

	// 存储后端：HostPath / NFS（二选一）
	switch {
	case req.HostPath != "":
		pv.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: req.HostPath,
			},
		}
	case req.NFS != nil:
		pv.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
			NFS: &corev1.NFSVolumeSource{
				Server:   req.NFS.Server,
				Path:     req.NFS.Path,
				ReadOnly: req.NFS.ReadOnly,
			},
		}
	}

	return pv
}

// BuildReclaimPolicyPatch 构造修改 ReclaimPolicy 的 Strategic Merge Patch
func BuildReclaimPolicyPatch(policy string) ([]byte, error) {
	if policy == "" {
		return nil, fmt.Errorf("reclaimPolicy 不能为空")
	}

	patchObj := map[string]interface{}{
		"spec": map[string]interface{}{
			"persistentVolumeReclaimPolicy": policy,
		},
	}
	return json.Marshal(patchObj)
}

// BuildCapacityPatch 构造修改 Capacity 的 Strategic Merge Patch
func BuildCapacityPatch(capacity string) ([]byte, error) {
	if capacity == "" {
		return nil, fmt.Errorf("capacity 不能为空")
	}

	// 验证容量格式
	_, err := resource.ParseQuantity(capacity)
	if err != nil {
		return nil, fmt.Errorf("无效的容量格式 %s: %w", capacity, err)
	}

	patchObj := map[string]interface{}{
		"spec": map[string]interface{}{
			"capacity": map[string]interface{}{
				"storage": capacity,
			},
		},
	}
	return json.Marshal(patchObj)
}
