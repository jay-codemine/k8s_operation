package pv

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/requests"
)

func ReclaimPersistentVolume(ctx context.Context, Kube kubernetes.Interface, req *requests.KubePVReclaimRequest) (*corev1.PersistentVolume, error) {
	patchBytes, err := BuildReclaimPolicyPatch(req.ReclaimPolicy)
	if err != nil {
		return nil, err
	}

	updated, err := Kube.CoreV1().
		PersistentVolumes().
		Patch(ctx, req.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("patch reclaim policy failed: %w", err)
	}
	return updated, nil
}

// ExpandPersistentVolume 扩容 PersistentVolume
// 注意：
// 1. 只能扩大不能缩小
// 2. 需要 StorageClass 支持 allowVolumeExpansion: true
// 3. 底层存储驱动必须支持扩容（如 AWS EBS、GCE PD、Ceph 等）
// 4. 静态 PV（HostPath/NFS）需要手动调整底层存储后再修改 PV
func ExpandPersistentVolume(ctx context.Context, Kube kubernetes.Interface, req *requests.KubePVExpandRequest) (*corev1.PersistentVolume, error) {
	// 1. 获取当前 PV
	currentPV, err := Kube.CoreV1().PersistentVolumes().Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get PV failed: %w", err)
	}

	// 2. 验证状态：只有 Bound 或 Available 状态的 PV 才能扩容
	if currentPV.Status.Phase != corev1.VolumeBound && currentPV.Status.Phase != corev1.VolumeAvailable {
		return nil, fmt.Errorf("PV 必须处于 Bound 或 Available 状态，当前状态: %s", currentPV.Status.Phase)
	}

	// 3. 构建扩容 Patch
	patchBytes, err := BuildCapacityPatch(req.NewCapacity)
	if err != nil {
		return nil, err
	}

	// 4. 应用 Patch
	updated, err := Kube.CoreV1().
		PersistentVolumes().
		Patch(ctx, req.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("patch capacity failed: %w", err)
	}

	return updated, nil
}
