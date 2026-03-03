package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
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
