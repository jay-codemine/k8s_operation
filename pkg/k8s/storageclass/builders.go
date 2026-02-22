package storageclass

import (
	"fmt"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
)

// StorageClassListItem 列表项响应结构
type StorageClassListItem struct {
	Name              string            `json:"name"`
	Provisioner       string            `json:"provisioner"`
	ReclaimPolicy     string            `json:"reclaim_policy"`       // Delete/Retain
	VolumeBindingMode string            `json:"volume_binding_mode"` // Immediate/WaitForFirstConsumer
	AllowExpansion    bool              `json:"allow_expansion"`
	Parameters        map[string]string `json:"parameters"`
	MountOptions      []string          `json:"mount_options"`
	IsDefault         bool              `json:"is_default"`
	CreatedAt         string            `json:"created_at"`
}

// BuildStorageClassListResponse 将 StorageClass 列表转换为响应格式
func BuildStorageClassListResponse(storageClasses []storagev1.StorageClass) []StorageClassListItem {
	result := make([]StorageClassListItem, 0, len(storageClasses))

	for _, sc := range storageClasses {
		// 获取回收策略
		reclaimPolicy := "Delete" // 默认
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		// 获取绑定模式
		volumeBindingMode := "Immediate" // 默认
		if sc.VolumeBindingMode != nil {
			volumeBindingMode = string(*sc.VolumeBindingMode)
		}

		// 是否允许扩容
		allowExpansion := false
		if sc.AllowVolumeExpansion != nil {
			allowExpansion = *sc.AllowVolumeExpansion
		}

		// 判断是否为默认 StorageClass
		isDefault := false
		if sc.Annotations != nil {
			if v, ok := sc.Annotations["storageclass.kubernetes.io/is-default-class"]; ok && v == "true" {
				isDefault = true
			}
			// 兼容旧版注解
			if v, ok := sc.Annotations["storageclass.beta.kubernetes.io/is-default-class"]; ok && v == "true" {
				isDefault = true
			}
		}

		item := StorageClassListItem{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: volumeBindingMode,
			AllowExpansion:    allowExpansion,
			Parameters:        sc.Parameters,
			MountOptions:      sc.MountOptions,
			IsDefault:         isDefault,
			CreatedAt:         sc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}

		result = append(result, item)
	}

	return result
}

// BuildStorageClassResponse 用于统一构造创建 StorageClass 成功后的返回体
func BuildStorageClassResponse(sc *storagev1.StorageClass) gin.H {
	reclaimPolicy := "Delete"
	if sc.ReclaimPolicy != nil {
		reclaimPolicy = string(*sc.ReclaimPolicy)
	}

	volumeBindingMode := "Immediate"
	if sc.VolumeBindingMode != nil {
		volumeBindingMode = string(*sc.VolumeBindingMode)
	}

	allowExpansion := false
	if sc.AllowVolumeExpansion != nil {
		allowExpansion = *sc.AllowVolumeExpansion
	}

	return gin.H{
		"name":                 sc.Name,
		"provisioner":          sc.Provisioner,
		"reclaim_policy":       reclaimPolicy,
		"volume_binding_mode":  volumeBindingMode,
		"allow_expansion":      allowExpansion,
		"parameters":           sc.Parameters,
		"mount_options":        sc.MountOptions,
		"created_at":           sc.CreationTimestamp,
		"message":              "StorageClass 创建成功",
	}
}

// buildStorageClassFromReq 根据请求构造 StorageClass
func buildStorageClassFromReq(req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      nil,
			Annotations: nil,
		},
		Provisioner:          req.Provisioner,
		Parameters:           req.Parameters,           // map[string]string
		AllowVolumeExpansion: req.AllowVolumeExpansion, // *bool
		MountOptions:         req.MountOptions,         // []string
	}

	if req.ReclaimPolicy != "" {
		p, err := parseReclaimPolicy(req.ReclaimPolicy)
		if err != nil {
			return nil, err
		}
		sc.ReclaimPolicy = &p
	}

	if req.VolumeBindingMode != "" {
		m, err := parseVolumeBindingMode(req.VolumeBindingMode)
		if err != nil {
			return nil, err
		}
		sc.VolumeBindingMode = &m
	}

	return sc, nil
}

// parseReclaimPolicy 解析回收策略
func parseReclaimPolicy(s string) (corev1.PersistentVolumeReclaimPolicy, error) {
	switch s {
	case "Delete", "delete":
		return corev1.PersistentVolumeReclaimDelete, nil
	case "Retain", "retain":
		return corev1.PersistentVolumeReclaimRetain, nil
	default:
		return "", fmt.Errorf("不支持的 ReclaimPolicy: %q（仅支持 Delete/Retain）", s)
	}
}

// parseVolumeBindingMode 解析卷绑定模式
func parseVolumeBindingMode(s string) (storagev1.VolumeBindingMode, error) {
	switch s {
	case "Immediate", "immediate":
		return storagev1.VolumeBindingImmediate, nil
	case "WaitForFirstConsumer", "waitforfirstconsumer", "WaitForFirst":
		return storagev1.VolumeBindingWaitForFirstConsumer, nil
	default:
		return "", fmt.Errorf("不支持的 VolumeBindingMode: %q（仅支持 Immediate/WaitForFirstConsumer）", s)
	}
}
