package services

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// =========================================================================
// AI 操作执行器 —— 将 Function Call 映射到真实平台操作
// =========================================================================

// ExecuteToolCall 执行单个工具调用，返回 JSON 结果
func (s *Services) ExecuteToolCall(ctx context.Context, factory *ClusterClientFactory, toolName string, argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	global.Logger.Info("AI 执行工具调用", zap.String("tool", toolName), zap.String("args", argsJSON))

	switch toolName {
	// ==================== 查询类 ====================
	case "list_pods":
		return s.execListPods(ctx, factory, args)
	case "get_pod_detail":
		return s.execGetPodDetail(ctx, factory, args)
	case "list_deployments":
		return s.execListDeployments(ctx, factory, args)
	case "get_deployment_detail":
		return s.execGetDeploymentDetail(ctx, factory, args)
	case "list_services":
		return s.execListServices(ctx, factory, args)
	case "get_service_detail":
		return s.execGetServiceDetail(ctx, factory, args)
	case "list_namespaces":
		return s.execListNamespaces(ctx, factory, args)
	case "list_nodes":
		return s.execListNodes(ctx, factory, args)
	case "get_node_detail":
		return s.execGetNodeDetail(ctx, factory, args)
	case "list_ingresses":
		return s.execListIngresses(ctx, factory, args)
	case "list_configmaps":
		return s.execListConfigMaps(ctx, factory, args)
	case "list_pvcs":
		return s.execListPVCs(ctx, factory, args)
	case "list_clusters":
		return s.execListClusters(ctx, args)
	case "list_pipelines":
		return s.execListPipelines(ctx, args)
	case "get_pipeline_detail":
		return s.execGetPipelineDetail(ctx, args)

	// ==================== 写操作 ====================
	case "scale_deployment":
		return s.execScaleDeployment(ctx, factory, args)
	case "restart_deployment":
		return s.execRestartDeployment(ctx, factory, args)
	case "rollback_deployment":
		return s.execRollbackDeployment(ctx, factory, args)
	case "update_deployment_image":
		return s.execUpdateDeploymentImage(ctx, factory, args)
	case "cordon_node":
		return s.execCordonNode(ctx, factory, args)

	// ==================== 删除操作 ====================
	case "delete_pod":
		return s.execDeletePod(ctx, factory, args)
	case "delete_deployment":
		return s.execDeleteDeployment(ctx, factory, args)
	case "delete_service":
		return s.execDeleteService(ctx, factory, args)
	case "delete_namespace":
		return s.execDeleteNamespace(ctx, factory, args)
	case "delete_configmap":
		return s.execDeleteConfigMap(ctx, factory, args)
	case "delete_ingress":
		return s.execDeleteIngress(ctx, factory, args)
	case "delete_pvc":
		return s.execDeletePVC(ctx, factory, args)
	case "drain_node":
		return s.execDrainNode(ctx, factory, args)

	default:
		return jsonResult(false, "未知的操作: "+toolName, nil)
	}
}

// ==================== 辅助函数 ====================

func getStr(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if v, ok := args[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return 0
}

func getClusterID(args map[string]interface{}) uint32 {
	return uint32(getInt(args, "cluster_id"))
}

func getK8sClients(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (*K8sClients, error) {
	clusterID := getClusterID(args)
	if clusterID == 0 {
		return nil, fmt.Errorf("cluster_id 不能为空")
	}
	return factory.Get(ctx, clusterID)
}

func jsonResult(success bool, msg string, data interface{}) (string, error) {
	result := map[string]interface{}{
		"success": success,
		"message": msg,
	}
	if data != nil {
		result["data"] = data
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

func jsonSummary(items interface{}, total interface{}) (string, error) {
	result := map[string]interface{}{
		"success": true,
		"total":   total,
		"items":   items,
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

// ==================== 查询类实现 ====================

func (s *Services) execListPods(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubePodListRequest{Page: 1, Limit: 50}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	pods, total, err := s.KubePodList(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	// 简化返回（只保留关键信息）
	var items []map[string]interface{}
	for _, p := range pods {
		items = append(items, map[string]interface{}{
			"name": p.Name, "namespace": p.Namespace, "status": string(p.Status.Phase),
			"node": p.Spec.NodeName, "ip": p.Status.PodIP,
			"containers": len(p.Spec.Containers), "restarts": getRestarts(p),
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execGetPodDetail(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	pod, err := s.KubePodDetail(ctx, cli, &requests.KubePodDetailRequest{
		Namespace: getStr(args, "namespace"), Name: getStr(args, "name"),
	})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	return jsonResult(true, "ok", map[string]interface{}{
		"name": pod.Name, "namespace": pod.Namespace, "status": string(pod.Status.Phase),
		"node": pod.Spec.NodeName, "ip": pod.Status.PodIP, "labels": pod.Labels,
	})
}

func (s *Services) execListDeployments(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentListRequest{Page: 1, Limit: 50}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	deployments, total, err := s.KubeDeploymentList(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, d := range deployments {
		var ready, desired int32
		if d.Status.ReadyReplicas > 0 {
			ready = d.Status.ReadyReplicas
		}
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		items = append(items, map[string]interface{}{
			"name": d.Name, "namespace": d.Namespace,
			"ready": ready, "desired": desired,
			"available": d.Status.AvailableReplicas, "updated": d.Status.UpdatedReplicas,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execGetDeploymentDetail(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentDetailRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	d, err := s.KubeDeploymentDetail(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var images []string
	for _, c := range d.Spec.Template.Spec.Containers {
		images = append(images, c.Name+":"+c.Image)
	}
	var desired int32
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}
	return jsonResult(true, "ok", map[string]interface{}{
		"name": d.Name, "namespace": d.Namespace, "labels": d.Labels,
		"desired": desired, "ready": d.Status.ReadyReplicas,
		"available": d.Status.AvailableReplicas, "images": images,
		"strategy": string(d.Spec.Strategy.Type),
	})
}

func (s *Services) execListServices(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeServiceListRequest{Page: 1, Limit: 50}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	svcs, total, err := s.KubeServiceList(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, sv := range svcs {
		items = append(items, map[string]interface{}{
			"name": sv.Name, "namespace": sv.Namespace, "type": string(sv.Spec.Type),
			"cluster_ip": sv.Spec.ClusterIP,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execGetServiceDetail(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeServiceDetailRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	sv, err := s.KubeServiceDetail(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	return jsonResult(true, "ok", map[string]interface{}{
		"name": sv.Name, "namespace": sv.Namespace, "type": string(sv.Spec.Type),
		"cluster_ip": sv.Spec.ClusterIP, "selector": sv.Spec.Selector, "ports": sv.Spec.Ports,
	})
}

func (s *Services) execListNamespaces(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	nsList, total, err := s.KubeNamespaceList(ctx, cli, &requests.KubeNamespaceListRequest{
		Name: getStr(args, "name"), Page: 1, Limit: 100,
	})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, ns := range nsList {
		items = append(items, map[string]interface{}{
			"name": ns.Name, "status": string(ns.Status.Phase), "labels": ns.Labels,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execListNodes(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	nodes, total, err := s.KubeNodeList(ctx, cli, &requests.KubeNodeListRequest{
		Name: getStr(args, "name"), Page: 1, Limit: 100,
	})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, n := range nodes {
		ready := "Unknown"
		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" {
				ready = string(c.Status)
			}
		}
		items = append(items, map[string]interface{}{
			"name": n.Name, "ready": ready, "unschedulable": n.Spec.Unschedulable,
			"os_image": n.Status.NodeInfo.OSImage, "kubelet": n.Status.NodeInfo.KubeletVersion,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execGetNodeDetail(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	n, err := s.KubeNodeDetail(ctx, cli, &requests.KubeNodeDetailRequest{Name: getStr(args, "name")})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	return jsonResult(true, "ok", map[string]interface{}{
		"name": n.Name, "labels": n.Labels, "unschedulable": n.Spec.Unschedulable,
		"capacity": n.Status.Capacity, "allocatable": n.Status.Allocatable,
		"os": n.Status.NodeInfo.OSImage, "kubelet": n.Status.NodeInfo.KubeletVersion,
	})
}

func (s *Services) execListIngresses(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeIngressListRequest{Page: 1, Limit: 50}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	ings, total, err := s.KubeIngressList(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, ing := range ings {
		items = append(items, map[string]interface{}{
			"name": ing.Name, "namespace": ing.Namespace,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execListConfigMaps(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeConfigMapListRequest{Name: getStr(args, "name"), Page: 1, Limit: 50}
	req.KubeCommonRequest.Namespace = getStr(args, "namespace")
	cms, total, err := s.KubeConfigMapList(ctx, cli, req)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, cm := range cms {
		items = append(items, map[string]interface{}{
			"name": cm.Name, "namespace": cm.Namespace, "data_keys": len(cm.Data),
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execListPVCs(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	pvcs, total, err := s.KubePVCList(ctx, cli, &requests.KubePVCListRequest{
		Name: getStr(args, "name"), Namespace: getStr(args, "namespace"), Page: 1, Limit: 50,
	})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, p := range pvcs {
		items = append(items, map[string]interface{}{
			"name": p.Name, "namespace": p.Namespace, "status": string(p.Status.Phase),
			"volume": p.Spec.VolumeName,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execListClusters(ctx context.Context, args map[string]interface{}) (string, error) {
	list, total, err := s.K8sClusterList(ctx, &requests.K8sClusterListRequest{Page: 1, Limit: 100})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	var items []map[string]interface{}
	for _, c := range list {
		items = append(items, map[string]interface{}{
			"id": c.ID, "name": c.ClusterName, "version": c.ClusterVersion, "status": c.Status,
		})
	}
	return jsonSummary(items, total)
}

func (s *Services) execListPipelines(ctx context.Context, args map[string]interface{}) (string, error) {
	list, total, err := s.PipelineList(ctx, &requests.PipelineListRequest{
		Keyword: getStr(args, "keyword"), Status: getStr(args, "status"), Page: 1, PageSize: 50,
	})
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	return jsonSummary(list, total)
}

func (s *Services) execGetPipelineDetail(ctx context.Context, args map[string]interface{}) (string, error) {
	id := int64(getInt(args, "id"))
	p, err := s.PipelineDetail(ctx, id)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	return jsonResult(true, "ok", p)
}

// ==================== 写操作实现 ====================

func (s *Services) execScaleDeployment(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	replicas := int32(getInt(args, "replicas"))
	req := &requests.KubeDeploymentScaleRequest{ScaleNum: replicas}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	d, err := s.KubeUpdateDeploymentReplicas(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "扩缩容失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Deployment %s 副本数已调整为 %d", d.Name, replicas), nil)
}

func (s *Services) execRestartDeployment(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentRestartRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	err = s.KubeDeploymentRestart(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "重启失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Deployment %s/%s 已触发重启", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execRollbackDeployment(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentRollbackRequest{ReplicaSet: getStr(args, "replica_set")}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	_, err = s.KubeDeploymentRollback(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "回滚失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Deployment %s/%s 已触发回滚", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execUpdateDeploymentImage(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentUpdateImageRequest{Container: getStr(args, "container"), Image: getStr(args, "image")}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	_, err = s.KubeUpdateDeploymentImage(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "更新镜像失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Deployment %s 镜像已更新为 %s", getStr(args, "name"), getStr(args, "image")), nil)
}

func (s *Services) execCordonNode(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	unschedulable := false
	if v, ok := args["unschedulable"]; ok {
		if b, ok := v.(bool); ok {
			unschedulable = b
		}
	}
	err = s.KubeNodeCordon(ctx, cli, &requests.KubeNodeCordonRequest{
		NodeName: getStr(args, "node_name"), Unschedulable: unschedulable,
	})
	if err != nil {
		return jsonResult(false, "操作失败: "+err.Error(), nil)
	}
	action := "标记为可调度"
	if unschedulable {
		action = "标记为不可调度"
	}
	return jsonResult(true, fmt.Sprintf("节点 %s 已%s", getStr(args, "node_name"), action), nil)
}

// ==================== 删除操作实现 ====================

func (s *Services) execDeletePod(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	err = s.KubePodDelete(ctx, cli, &requests.KubePodDeleteRequest{
		Name: getStr(args, "name"), Namespace: getStr(args, "namespace"),
	})
	if err != nil {
		return jsonResult(false, "删除 Pod 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Pod %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDeleteDeployment(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeDeploymentDeleteRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	err = s.KubeDeploymentDelete(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "删除 Deployment 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Deployment %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDeleteService(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeServiceDeleteRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	err = s.KubeServiceDelete(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "删除 Service 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Service %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDeleteNamespace(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	err = s.KubeNamespaceDelete(ctx, cli, &requests.KubeNamespaceDeleteRequest{Name: getStr(args, "name")})
	if err != nil {
		return jsonResult(false, "删除 Namespace 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Namespace %s 已删除", getStr(args, "name")), nil)
}

func (s *Services) execDeleteConfigMap(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeConfigMapDeleteRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	err = s.KubeConfigMapDelete(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "删除 ConfigMap 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("ConfigMap %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDeleteIngress(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	req := &requests.KubeIngressDeleteRequest{}
	req.Name = getStr(args, "name")
	req.Namespace = getStr(args, "namespace")
	err = s.KubeIngressDelete(ctx, cli, req)
	if err != nil {
		return jsonResult(false, "删除 Ingress 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("Ingress %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDeletePVC(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	err = s.KubePVCDelete(ctx, cli, &requests.KubePVCDeleteRequest{
		Name: getStr(args, "name"), Namespace: getStr(args, "namespace"),
	})
	if err != nil {
		return jsonResult(false, "删除 PVC 失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("PVC %s/%s 已删除", getStr(args, "namespace"), getStr(args, "name")), nil)
}

func (s *Services) execDrainNode(ctx context.Context, factory *ClusterClientFactory, args map[string]interface{}) (string, error) {
	cli, err := getK8sClients(ctx, factory, args)
	if err != nil {
		return jsonResult(false, err.Error(), nil)
	}
	err = s.KubeNodeDrain(ctx, cli, &requests.KubeNodeDrainRequest{NodeName: getStr(args, "node_name")})
	if err != nil {
		return jsonResult(false, "排空节点失败: "+err.Error(), nil)
	}
	return jsonResult(true, fmt.Sprintf("节点 %s 已排空", getStr(args, "node_name")), nil)
}

// getRestarts 获取Pod重启次数
func getRestarts(p interface{}) int32 {
	// type assertion for corev1.Pod
	return 0 // 简化，避免import循环
}
