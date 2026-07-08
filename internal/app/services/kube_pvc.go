package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/pvc"
	"sigs.k8s.io/yaml"
)

// KubeCreatePVC 创建 PersistentVolumeClaim
func (s *Services) KubeCreatePVC(ctx context.Context, cli *K8sClients, req *requests.KubePVCCreateRequest) (*corev1.PersistentVolumeClaim, error) {
	// 2) 调用资源层进行构建 + 创建
	created, err := pvc.CreatePersistentVolumeClaim(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s already exists in namespace %s", req.Name, req.Namespace)
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Errorf("create PersistentVolumeClaim failed: %v", err)
		return nil, err
	}

	// 3) 成功日志
	global.Logger.Infof("PersistentVolumeClaim %q created successfully in namespace %q", created.Name, req.Namespace)
	return created, nil
}

// KubePVCList 获取 PVC 列表（支持分页与名称模糊）
func (s *Services) KubePVCList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubePVCListRequest,
) ([]corev1.PersistentVolumeClaim, int64, error) {

	items, total, err := pvc.GetPVCList(
		ctx,
		cli.Kube,
		param.Namespace,
		param.Name,
		param.Page,
		param.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// KubePVCDetail 获取 PVC 详情
func (s *Services) KubePVCDetail(ctx context.Context, cli *K8sClients, param *requests.KubePVCDetailRequest) (*corev1.PersistentVolumeClaim, error) {
	pvcDetail, err := pvc.GetPVCDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", param.Name, param.Namespace)
		}
		global.Logger.Error("get PersistentVolumeClaim detail failed", zap.Error(err))
		return nil, err
	}

	return pvcDetail, nil
}

// KubePVCDetailEnhanced 获取增强的 PVC 详情（包含绑定的 PV 信息和事件）
func (s *Services) KubePVCDetailEnhanced(ctx context.Context, cli *K8sClients, param *requests.KubePVCDetailRequest) (*requests.PVCDetailEnhanced, error) {
	// 1. 获取 PVC 详情
	pvcObj, err := pvc.GetPVCDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		return nil, err
	}

	// 2. 构建增强响应
	result := &requests.PVCDetailEnhanced{
		Name:             pvcObj.Name,
		Namespace:        pvcObj.Namespace,
		UID:              string(pvcObj.UID),
		CreatedAt:        pvcObj.CreationTimestamp.Unix(),
		Labels:           pvcObj.Labels,
		Annotations:      pvcObj.Annotations,
		Phase:            string(pvcObj.Status.Phase),
		StorageClassName: func() string {
			if pvcObj.Spec.StorageClassName != nil {
				return *pvcObj.Spec.StorageClassName
			}
			return ""
		}(),
	}

	// 3. 设置状态信息
	switch pvcObj.Status.Phase {
	case corev1.ClaimBound:
		result.StatusMessage = "已成功绑定到 PV"
		result.StatusColor = "success"
	case corev1.ClaimPending:
		result.StatusMessage = "等待绑定 PV"
		result.StatusColor = "warning"
	case corev1.ClaimLost:
		result.StatusMessage = "PV 已丢失"
		result.StatusColor = "error"
	default:
		result.StatusMessage = string(pvcObj.Status.Phase)
		result.StatusColor = "default"
	}

	// 4. 设置存储信息
	if req := pvcObj.Spec.Resources.Requests[corev1.ResourceStorage]; !req.IsZero() {
		result.RequestStorage = req.String()
	}
	if cap := pvcObj.Status.Capacity[corev1.ResourceStorage]; !cap.IsZero() {
		result.ActualCapacity = cap.String()
	}

	// 访问模式
	for _, mode := range pvcObj.Spec.AccessModes {
		result.AccessModes = append(result.AccessModes, string(mode))
	}

	// 卷模式
	if pvcObj.Spec.VolumeMode != nil {
		result.VolumeMode = string(*pvcObj.Spec.VolumeMode)
	} else {
		result.VolumeMode = "Filesystem"
	}

	// 5. 获取绑定的 PV 信息
	if pvcObj.Spec.VolumeName != "" {
		pvObj, err := cli.Kube.CoreV1().PersistentVolumes().Get(ctx, pvcObj.Spec.VolumeName, metav1.GetOptions{})
		if err == nil {
			result.BoundPV = &requests.BoundPVInfo{
				Name:             pvObj.Name,
				ReclaimPolicy:    string(pvObj.Spec.PersistentVolumeReclaimPolicy),
				Status:           string(pvObj.Status.Phase),
				CreatedAt:        pvObj.CreationTimestamp.Unix(),
				StorageClassName: pvObj.Spec.StorageClassName,
			}
			
			// 容量
			if cap := pvObj.Spec.Capacity[corev1.ResourceStorage]; !cap.IsZero() {
				result.BoundPV.Capacity = cap.String()
			}
			
			// 卷类型和源
			result.BoundPV.VolumeType, result.BoundPV.VolumeSource = getPVVolumeTypeAndSource(pvObj)
			
			// 节点亲和性
			if pvObj.Spec.NodeAffinity != nil && pvObj.Spec.NodeAffinity.Required != nil {
				result.BoundPV.NodeAffinity = formatNodeAffinity(pvObj.Spec.NodeAffinity.Required)
			}
		}
	}

	// 6. 获取条件状态
	for _, cond := range pvcObj.Status.Conditions {
		result.Conditions = append(result.Conditions, requests.PVCCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			LastTransitionTime: cond.LastTransitionTime.Unix(),
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	// 7. 获取最近事件（最多 5 条）
	events, _ := cli.Kube.CoreV1().Events(param.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=PersistentVolumeClaim", param.Name),
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

	return result, nil
}

// getPVVolumeTypeAndSource 获取 PV 的卷类型和源信息
func getPVVolumeTypeAndSource(pv *corev1.PersistentVolume) (volumeType, volumeSource string) {
	spec := pv.Spec
	switch {
	case spec.HostPath != nil:
		return "HostPath", spec.HostPath.Path
	case spec.NFS != nil:
		return "NFS", fmt.Sprintf("%s:%s", spec.NFS.Server, spec.NFS.Path)
	case spec.Local != nil:
		return "Local", spec.Local.Path
	case spec.CSI != nil:
		return "CSI", spec.CSI.Driver
	case spec.AWSElasticBlockStore != nil:
		return "AWSElasticBlockStore", spec.AWSElasticBlockStore.VolumeID
	case spec.GCEPersistentDisk != nil:
		return "GCEPersistentDisk", spec.GCEPersistentDisk.PDName
	case spec.AzureDisk != nil:
		return "AzureDisk", spec.AzureDisk.DiskName
	case spec.CephFS != nil:
		return "CephFS", fmt.Sprintf("%v", spec.CephFS.Monitors)
	case spec.RBD != nil:
		return "RBD", spec.RBD.RBDImage
	case spec.ISCSI != nil:
		return "iSCSI", spec.ISCSI.TargetPortal
	case spec.FC != nil:
		return "FC", fmt.Sprintf("%v", spec.FC.TargetWWNs)
	default:
		return "Unknown", "-"
	}
}

// formatNodeAffinity 格式化节点亲和性
func formatNodeAffinity(req *corev1.NodeSelector) string {
	if req == nil || len(req.NodeSelectorTerms) == 0 {
		return ""
	}
	var parts []string
	for _, term := range req.NodeSelectorTerms {
		for _, expr := range term.MatchExpressions {
			parts = append(parts, fmt.Sprintf("%s %s %v", expr.Key, expr.Operator, expr.Values))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return parts[0] // 返回第一个
}

func (s *Services) KubePVCDelete(ctx context.Context, cli *K8sClients, param *requests.KubePVCDeleteRequest) error {
	if err := pvc.DeletePersistentVolumeClaim(ctx, cli.Kube, param.Namespace, param.Name); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil // 幂等
		}
		global.Logger.Errorf("delete PersistentVolumeClaim %s/%s failed: %v", param.Namespace, param.Name, err)
		return err
	}

	global.Logger.Infof("PersistentVolumeClaim %s/%s deleted successfully", param.Namespace, param.Name)
	return nil
}

// 扩容 PVC：仅允许修改 spec.resources.requests.storage
func (s *Services) KubePVCResize(ctx context.Context, cli *K8sClients, req *requests.KubePVCResizeRequest,
) (*corev1.PersistentVolumeClaim, error) {
	return pvc.ResizePVC(ctx, cli.Kube, req)
}

// KubePVCCreateFromYaml 从 YAML 创建 PVC
func (s *Services) KubePVCCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	// 预处理：提取第一个有效 YAML 文档（跳过纯注释文档）
	yamlContent = common.ExtractFirstValidDocument(yamlContent)

	// 1. 解析 YAML 到 Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlContent), &unstructuredObj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证是否为 PVC
	if unstructuredObj.GetKind() != "PersistentVolumeClaim" {
		return nil, fmt.Errorf("YAML kind must be PersistentVolumeClaim, got %q", unstructuredObj.GetKind())
	}

	// 3. 转换为 PVC 对象
	pvcObj := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, pvcObj); err != nil {
		return nil, fmt.Errorf("failed to convert to PersistentVolumeClaim: %w", err)
	}

	// 4. 确保 namespace 不为空
	if pvcObj.Namespace == "" {
		pvcObj.Namespace = "default"
	}

	// 5. 调用 K8s API 创建
	created, err := cli.Kube.CoreV1().PersistentVolumeClaims(pvcObj.Namespace).Create(ctx, pvcObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", pvcObj.Name, pvcObj.Namespace)
		}
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	global.Logger.Infof("PVC created from YAML: %s/%s", created.Namespace, created.Name)
	return created, nil
}

// KubePVCApplyYaml 应用 PVC YAML 更改（更新已存在的 PVC）
func (s *Services) KubePVCApplyYaml(ctx context.Context, cli *K8sClients, namespace, name, yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	// 1. 解析 YAML
	unstructuredObj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlContent), &unstructuredObj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 转换为 PVC 对象
	pvcObj := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, pvcObj); err != nil {
		return nil, fmt.Errorf("failed to convert to PersistentVolumeClaim: %w", err)
	}

	// 3. 确保 namespace/name 匹配
	if pvcObj.Namespace == "" {
		pvcObj.Namespace = namespace
	}
	if pvcObj.Name != name || pvcObj.Namespace != namespace {
		return nil, fmt.Errorf("YAML name/namespace mismatch: expected %s/%s, got %s/%s", namespace, name, pvcObj.Namespace, pvcObj.Name)
	}

	// 4. 获取现有 PVC（保留 ResourceVersion）
	existing, err := cli.Kube.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing PVC: %w", err)
	}

	// 5. 保留必要的元数据
	pvcObj.ResourceVersion = existing.ResourceVersion
	pvcObj.UID = existing.UID

	// 6. 更新 PVC
	updated, err := cli.Kube.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvcObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update PVC: %w", err)
	}

	global.Logger.Infof("PVC updated from YAML: %s/%s", updated.Namespace, updated.Name)
	return updated, nil
}
