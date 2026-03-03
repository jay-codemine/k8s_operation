package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/node"
)

// services/node_service.go
func (s *Services) KubeNodeList(ctx context.Context, cli *K8sClients, param *requests.KubeNodeListRequest) ([]corev1.Node, int, error) {
	items, total, err := node.GetNodeList(ctx, cli.Kube, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list Node failed: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

// KubeNodeDetail 获取 Node 详情
func (s *Services) KubeNodeDetail(ctx context.Context, cli *K8sClients, param *requests.KubeNodeDetailRequest) (*corev1.Node, error) {
	nodeObj, err := node.GetNodeDetail(ctx, cli.Kube, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found", param.Name)
			return nil, fmt.Errorf("Node %q not found", param.Name)
		}
		global.Logger.Error("get Node detail failed", zap.Error(err))
		return nil, err
	}
	return nodeObj, nil
}

// KubeNodePods 列出指定 Node 上的 Pod
func (s *Services) KubeNodePods(ctx context.Context, cli *K8sClients, param *requests.KubeNodePodsRequest) ([]corev1.Pod, error) {
	pods, err := node.ListPodsByNode(ctx, cli.Kube, param.Name)
	if err != nil {
		global.Logger.Error("list Pods by node failed",
			zap.String("nodeName", param.Name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list Pods on node %q: %v", param.Name, err)
	}

	return pods, nil
}

// KubeNodeMetricsList 获取 Node 指标列表（支持单节点或全量）
func (s *Services) KubeNodeMetricsList(ctx context.Context, cli *K8sClients, param *requests.KubeNodeMetricsRequest) ([]models.NodeMetricItem, int, error) {
	global.Logger.Info("use k8s_clients in service",
		zap.String("apiserver", cli.Config.Host),
		zap.Bool("has_metrics", cli.Metrics != nil),
	)

	if cli.Metrics == nil {
		return nil, 0, fmt.Errorf("metrics client is nil (metrics-server not available for this cluster)")
	}
	items, err := node.GetNodeMetrics(ctx, cli.Kube, cli.Metrics, param.Name)

	if err != nil {
		global.Logger.Errorf("list Node metrics failed: %v", err)
		return nil, 0, err
	}
	return items, len(items), nil
}

// KubeNodeCordon 标记 Node 是否可调度
func (s *Services) KubeNodeCordon(ctx context.Context, cli *K8sClients, param *requests.KubeNodeCordonRequest) error {
	err := node.CordonNode(ctx, cli.Kube, param.NodeName, param.Unschedulable)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found", param.NodeName)
			return fmt.Errorf("Node %q not found", param.NodeName)
		}
		global.Logger.Error("cordon Node failed",
			zap.String("name", param.NodeName),
			zap.Bool("unscheduled", param.Unschedulable),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// KubeNodeDrain 节点 drain：cordon + 驱逐 Pod
func (s *Services) KubeNodeDrain(ctx context.Context, cli *K8sClients, param *requests.KubeNodeDrainRequest) error {
	if err := node.DrainNode(ctx, cli.Kube, param.NodeName); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found when drain", param.NodeName)
			return fmt.Errorf("Node %q not found", param.NodeName)
		}
		global.Logger.Error("drain Node failed",
			zap.String("name", param.NodeName),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// KubeNodePatchLabels 修改节点标签
func (s *Services) KubeNodePatchLabels(ctx context.Context, cli *K8sClients, param *requests.KubeNodeLabelPatchRequest) error {
	err := node.PatchLabels(ctx, cli.Kube, param.Name, param.Add, param.Remove)
	if err != nil {
		global.Logger.Error("patch Node labels failed",
			zap.String("name", param.Name),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// KubeNodePatchTaints 修改节点污点
func (s *Services) KubeNodePatchTaints(ctx context.Context, cli *K8sClients, param *requests.KubeNodeTaintPatchRequest) error {
	err := node.PatchTaints(ctx, cli.Kube, param.Name, param.Add, param.RemoveKeys)
	if err != nil {
		global.Logger.Error("patch Node taints failed",
			zap.String("name", param.Name),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// KubeNodeEvents 获取节点事件
func (s *Services) KubeNodeEvents(ctx context.Context, cli *K8sClients, param *requests.KubeNodeEventsRequest) ([]node.EventItem, error) {
	events, err := node.GetNodeEvents(ctx, cli.Kube, param.Name, param.Limit)
	if err != nil {
		global.Logger.Error("get Node events failed",
			zap.String("name", param.Name),
			zap.Error(err),
		)
		return nil, err
	}
	return events, nil
}
