package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/pv"
)

// KubeCreatePV 创建 PersistentVolume
func (s *Services) KubeCreatePV(ctx context.Context, cli *K8sClients, req *requests.KubePVCreateRequest) (*corev1.PersistentVolume, error) {
	// 调用资源层进行构建 + 创建
	created, err := pv.CreatePersistentVolume(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("PersistentVolume %s already exists", req.Name)
			return nil, fmt.Errorf("PersistentVolume %q already exists", req.Name)
		}
		global.Logger.Errorf("create PersistentVolume failed: %v", err)
		return nil, err
	}

	// 3) 成功日志
	global.Logger.Infof("PersistentVolume %q created successfully", created.Name)
	return created, nil
}
func (s *Services) KubePVList(ctx context.Context, cli *K8sClients, param *requests.KubePVListRequest) ([]corev1.PersistentVolume, int, error) {
	items, total, err := pv.GetPVList(ctx, cli.Kube, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list PV failed: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Services) KubePVDetail(ctx context.Context, cli *K8sClients, param *requests.KubePVDetailRequest) (*corev1.PersistentVolume, error) {
	pvDetail, err := pv.GetPVDetail(ctx, cli.Kube, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolume %s not found", param.Name)
			return nil, fmt.Errorf("PersistentVolume %q not found", param.Name)
		}
		global.Logger.Error("get PersistentVolume detail failed", zap.Error(err))
		return nil, err
	}

	return pvDetail, nil
}

// KubePVDetailEnhanced 获取增强的 PV 详情（包含关联 PVC 信息和事件）
func (s *Services) KubePVDetailEnhanced(ctx context.Context, cli *K8sClients, param *requests.KubePVDetailRequest) (*requests.PVDetailEnhanced, error) {
	// 1. 获取 PV 详情
	pvObj, err := pv.GetPVDetail(ctx, cli.Kube, param.Name)
	if err != nil {
		return nil, err
	}

	// 2. 构建增强响应
	result := &requests.PVDetailEnhanced{
		Name:             pvObj.Name,
		UID:              string(pvObj.UID),
		CreatedAt:        pvObj.CreationTimestamp.Unix(),
		Labels:           pvObj.Labels,
		Annotations:      pvObj.Annotations,
		Phase:            string(pvObj.Status.Phase),
		StorageClassName: pvObj.Spec.StorageClassName,
		ReclaimPolicy:    string(pvObj.Spec.PersistentVolumeReclaimPolicy),
	}

	// 3. 设置状态信息
	switch pvObj.Status.Phase {
	case corev1.VolumeAvailable:
		result.StatusMessage = "可用，尚未绑定到 PVC"
		result.StatusColor = "success"
	case corev1.VolumeBound:
		result.StatusMessage = "已绑定到 PVC"
		result.StatusColor = "success"
	case corev1.VolumeReleased:
		result.StatusMessage = "已释放，PVC 已删除"
		result.StatusColor = "warning"
	case corev1.VolumeFailed:
		result.StatusMessage = "回收失败"
		result.StatusColor = "error"
	default:
		result.StatusMessage = string(pvObj.Status.Phase)
		result.StatusColor = "default"
	}

	// 4. 设置存储信息
	if cap := pvObj.Spec.Capacity[corev1.ResourceStorage]; !cap.IsZero() {
		result.Capacity = cap.String()
	}

	// 访问模式
	for _, mode := range pvObj.Spec.AccessModes {
		result.AccessModes = append(result.AccessModes, string(mode))
	}

	// 卷模式
	if pvObj.Spec.VolumeMode != nil {
		result.VolumeMode = string(*pvObj.Spec.VolumeMode)
	} else {
		result.VolumeMode = "Filesystem"
	}

	// 5. 卷类型和源
	result.VolumeType, result.VolumeSource = getPVVolumeTypeAndSource(pvObj)

	// 6. 节点亲和性
	if pvObj.Spec.NodeAffinity != nil && pvObj.Spec.NodeAffinity.Required != nil {
		result.NodeAffinity = formatNodeAffinity(pvObj.Spec.NodeAffinity.Required)
	}

	// 7. 获取关联的 PVC 信息
	if pvObj.Spec.ClaimRef != nil {
		pvcObj, err := cli.Kube.CoreV1().PersistentVolumeClaims(pvObj.Spec.ClaimRef.Namespace).Get(ctx, pvObj.Spec.ClaimRef.Name, metav1.GetOptions{})
		if err == nil {
			result.BoundPVC = &requests.BoundPVCInfo{
				Name:      pvcObj.Name,
				Namespace: pvcObj.Namespace,
				Status:    string(pvcObj.Status.Phase),
				CreatedAt: pvcObj.CreationTimestamp.Unix(),
			}
			
			// 请求容量
			if req := pvcObj.Spec.Resources.Requests[corev1.ResourceStorage]; !req.IsZero() {
				result.BoundPVC.RequestStorage = req.String()
			}
			
			// 访问模式
			for _, mode := range pvcObj.Spec.AccessModes {
				result.BoundPVC.AccessModes = append(result.BoundPVC.AccessModes, string(mode))
			}
			
			// 存储类
			if pvcObj.Spec.StorageClassName != nil {
				result.BoundPVC.StorageClassName = *pvcObj.Spec.StorageClassName
			}
		} else {
			// PVC 不存在，但有引用（可能已删除）
			result.BoundPVC = &requests.BoundPVCInfo{
				Name:      pvObj.Spec.ClaimRef.Name,
				Namespace: pvObj.Spec.ClaimRef.Namespace,
				Status:    "不存在（可能已删除）",
			}
		}
	}

	// 8. 获取最近事件（PV 事件在默认命名空间）
	events, _ := cli.Kube.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=PersistentVolume", param.Name),
		Limit:         5,
	})
	if events != nil {
		for _, ev := range events.Items {
			result.RecentEvents = append(result.RecentEvents, requests.StorageEvent{
				Type:      ev.Type,
				Reason:    ev.Reason,
				Message:   ev.Message,
				Count:     ev.Count,
				FirstSeen: ev.FirstTimestamp.Unix(),
				LastSeen:  ev.LastTimestamp.Unix(),
			})
		}
	}

	// 9. 其他状态信息
	result.Reason = pvObj.Status.Reason
	result.Message = pvObj.Status.Message

	return result, nil
}

func (s *Services) KubePVDelete(ctx context.Context, cli *K8sClients, param *requests.KubePVDeleteRequest) error {
	if err := pv.DeletePersistentVolume(ctx, cli.Kube, param.Name); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolume %s not found", param.Name)
			return nil
		}
		global.Logger.Errorf("delete PersistentVolume %s failed: %v", param.Name, err)
		return err
	}

	global.Logger.Infof("PersistentVolume %s deleted successfully", param.Name)
	return nil
}

// 修改回收策略
func (s *Services) KubePVReclaim(ctx context.Context, cli *K8sClients, req *requests.KubePVReclaimRequest) (*corev1.PersistentVolume, error) {
	return pv.ReclaimPersistentVolume(ctx, cli.Kube, req)
}

// KubePVExpand PV 扩容
func (s *Services) KubePVExpand(ctx context.Context, cli *K8sClients, req *requests.KubePVExpandRequest) (*corev1.PersistentVolume, error) {
	// 获取当前 PV 信息用于验证
	currentPV, err := pv.GetPVDetail(ctx, cli.Kube, req.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolume %s not found", req.Name)
			return nil, fmt.Errorf("PersistentVolume %q not found", req.Name)
		}
		global.Logger.Errorf("get PersistentVolume failed: %v", err)
		return nil, err
	}

	// 验证新容量大于旧容量
	currentCapacity := currentPV.Spec.Capacity[corev1.ResourceStorage]
	newCapacity, err := resource.ParseQuantity(req.NewCapacity)
	if err != nil {
		return nil, fmt.Errorf("无效的容量格式: %w", err)
	}

	if newCapacity.Cmp(currentCapacity) <= 0 {
		return nil, fmt.Errorf("PV 只能扩大不能缩小，当前容量: %s，新容量: %s",
			currentCapacity.String(), req.NewCapacity)
	}

	// 执行扩容
	updated, err := pv.ExpandPersistentVolume(ctx, cli.Kube, req)
	if err != nil {
		global.Logger.Errorf("expand PersistentVolume %s failed: %v", req.Name, err)
		return nil, err
	}

	global.Logger.Infof("PersistentVolume %s expanded from %s to %s successfully",
		req.Name, currentCapacity.String(), req.NewCapacity)
	return updated, nil
}

// KubePVGetYaml 获取 PV 的 YAML 配置
func (s *Services) KubePVGetYaml(ctx context.Context, cli *K8sClients, name string) (string, error) {
	yamlStr, err := pv.GetYaml(ctx, cli.Kube, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolume %s not found", name)
			return "", fmt.Errorf("PersistentVolume %q not found", name)
		}
		global.Logger.Errorf("get PersistentVolume YAML failed: %v", err)
		return "", err
	}
	return yamlStr, nil
}

// KubePVApplyYaml 应用 PV YAML 配置
func (s *Services) KubePVApplyYaml(ctx context.Context, cli *K8sClients, name, yamlContent string) (*corev1.PersistentVolume, error) {
	updated, err := pv.ApplyYaml(ctx, cli.Kube, name, yamlContent)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolume %s not found", name)
			return nil, fmt.Errorf("PersistentVolume %q not found", name)
		}
		global.Logger.Errorf("apply PersistentVolume YAML failed: %v", err)
		return nil, err
	}
	global.Logger.Infof("PersistentVolume %s YAML applied successfully", name)
	return updated, nil
}

// KubePVCreateFromYaml 从 YAML 创建 PV
func (s *Services) KubePVCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.PersistentVolume, error) {
	// 预处理：提取第一个有效 YAML 文档（跳过纯注释文档）
	yamlContent = common.ExtractFirstValidDocument(yamlContent)

	created, err := pv.CreateFromYaml(ctx, cli.Kube, yamlContent)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("PersistentVolume already exists")
			return nil, fmt.Errorf("PersistentVolume already exists")
		}
		global.Logger.Errorf("create PersistentVolume from YAML failed: %v", err)
		return nil, err
	}
	global.Logger.Infof("PersistentVolume %s created from YAML successfully", created.Name)
	return created, nil
}
