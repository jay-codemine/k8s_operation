package pvc

import (
	"context"
	"fmt"
	"k8soperation/internal/app/requests"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResizePVC 扩容 PVC：仅允许修改 spec.resources.requests.storage
func ResizePVC(ctx context.Context, kube kubernetes.Interface, req *requests.KubePVCResizeRequest,
) (*corev1.PersistentVolumeClaim, error) {

	c, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// 1) 读取当前 PVC
	curr, err := kube.CoreV1().
		PersistentVolumeClaims(req.Namespace).
		Get(c, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", req.Name, req.Namespace)
		}
		return nil, err
	}

	// 2) 解析并比较容量（新值必须更大）
	currReq := curr.Spec.Resources.Requests[corev1.ResourceStorage]
	newQty, err := resource.ParseQuantity(req.Storage)
	if err != nil {
		return nil, fmt.Errorf("invalid storage quantity %q: %w", req.Storage, err)
	}
	if newQty.Cmp(currReq) <= 0 {
		return nil, fmt.Errorf("new storage %q must be greater than current %q", newQty.String(), currReq.String())
	}

	// 3) 校验 StorageClass 是否允许扩容（若有 SC）
	if curr.Spec.StorageClassName != nil && *curr.Spec.StorageClassName != "" {
		scName := *curr.Spec.StorageClassName
		sc, err := kube.StorageV1().
			StorageClasses().
			Get(c, scName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("get StorageClass %q failed: %w", scName, err)
		}
		if sc.AllowVolumeExpansion == nil || !*sc.AllowVolumeExpansion {
			return nil, fmt.Errorf("StorageClass %q does not allow volume expansion", scName)
		}
	}

	// 4) 调用资源层 Patch（你已实现）
	updated, err := PatchPVC(c, kube, req) // 👈 建议 PatchPVC 也改成传 kube（见下）
	if err != nil {
		return nil, err
	}
	return updated, nil
}
