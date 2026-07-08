package services

import (
	"k8soperation/pkg/openai"
)

// =========================================================================
// AI Function Calling 工具定义
// GPT 会根据用户意图自动选择调用哪个工具
// =========================================================================

// RiskLevel 操作风险等级
const (
	RiskRead     = "read"     // 只读，无需审批
	RiskWrite    = "write"    // 写操作，需确认
	RiskDanger   = "danger"   // 高危，需管理员审批
	RiskCritical = "critical" // 极高危，需多人审批
)

// ToolMeta 工具元数据（风险等级、是否需要审批）
type ToolMeta struct {
	Name          string
	RiskLevel     string
	NeedApproval  bool   // 是否强制需要审批
	Description   string // 中文描述，用于审批展示
}

// toolRegistry 工具元数据注册表
var toolRegistry = map[string]ToolMeta{
	// ===== 查询类（只读，无需审批）=====
	"list_pods":            {Name: "list_pods", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Pod 列表"},
	"get_pod_detail":       {Name: "get_pod_detail", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Pod 详情"},
	"get_pod_logs":         {Name: "get_pod_logs", RiskLevel: RiskRead, NeedApproval: false, Description: "查看 Pod 日志"},
	"list_deployments":     {Name: "list_deployments", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Deployment 列表"},
	"get_deployment_detail":{Name: "get_deployment_detail", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Deployment 详情"},
	"list_services":        {Name: "list_services", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Service 列表"},
	"get_service_detail":   {Name: "get_service_detail", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Service 详情"},
	"list_namespaces":      {Name: "list_namespaces", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Namespace 列表"},
	"list_nodes":           {Name: "list_nodes", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Node 列表"},
	"get_node_detail":      {Name: "get_node_detail", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Node 详情"},
	"list_ingresses":       {Name: "list_ingresses", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Ingress 列表"},
	"list_configmaps":      {Name: "list_configmaps", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 ConfigMap 列表"},
	"list_secrets":         {Name: "list_secrets", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Secret 列表"},
	"list_pvcs":            {Name: "list_pvcs", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 PVC 列表"},
	"list_clusters":        {Name: "list_clusters", RiskLevel: RiskRead, NeedApproval: false, Description: "查询集群列表"},
	"list_pipelines":       {Name: "list_pipelines", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 CI/CD 流水线列表"},
	"get_pipeline_detail":  {Name: "get_pipeline_detail", RiskLevel: RiskRead, NeedApproval: false, Description: "查询流水线详情"},
	"list_statefulsets":    {Name: "list_statefulsets", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 StatefulSet 列表"},
	"list_daemonsets":      {Name: "list_daemonsets", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 DaemonSet 列表"},
	"list_jobs":            {Name: "list_jobs", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 Job 列表"},
	"list_cronjobs":        {Name: "list_cronjobs", RiskLevel: RiskRead, NeedApproval: false, Description: "查询 CronJob 列表"},
	"get_node_metrics":     {Name: "get_node_metrics", RiskLevel: RiskRead, NeedApproval: false, Description: "查询节点指标"},
	"get_events":           {Name: "get_events", RiskLevel: RiskRead, NeedApproval: false, Description: "查询事件列表"},

	// ===== 写操作（需人工确认）=====
	"scale_deployment":     {Name: "scale_deployment", RiskLevel: RiskWrite, NeedApproval: true, Description: "Deployment 扩缩容"},
	"restart_deployment":   {Name: "restart_deployment", RiskLevel: RiskWrite, NeedApproval: true, Description: "重启 Deployment"},
	"rollback_deployment":  {Name: "rollback_deployment", RiskLevel: RiskWrite, NeedApproval: true, Description: "回滚 Deployment"},
	"update_deployment_image":{Name: "update_deployment_image", RiskLevel: RiskWrite, NeedApproval: true, Description: "更新 Deployment 镜像"},
	"create_deployment":    {Name: "create_deployment", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 Deployment"},
	"create_service":       {Name: "create_service", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 Service"},
	"create_namespace":     {Name: "create_namespace", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 Namespace"},
	"create_configmap":     {Name: "create_configmap", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 ConfigMap"},
	"create_ingress":       {Name: "create_ingress", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 Ingress"},
	"create_pvc":           {Name: "create_pvc", RiskLevel: RiskWrite, NeedApproval: true, Description: "创建 PVC"},
	"trigger_pipeline":     {Name: "trigger_pipeline", RiskLevel: RiskWrite, NeedApproval: true, Description: "触发 CI/CD 流水线"},
	"cordon_node":          {Name: "cordon_node", RiskLevel: RiskWrite, NeedApproval: true, Description: "节点调度开关(cordon/uncordon)"},

	// ===== 高危操作（必须管理员审批）=====
	"delete_pod":           {Name: "delete_pod", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 Pod"},
	"delete_deployment":    {Name: "delete_deployment", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 Deployment"},
	"delete_service":       {Name: "delete_service", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 Service"},
	"delete_namespace":     {Name: "delete_namespace", RiskLevel: RiskCritical, NeedApproval: true, Description: "删除 Namespace"},
	"delete_configmap":     {Name: "delete_configmap", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 ConfigMap"},
	"delete_ingress":       {Name: "delete_ingress", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 Ingress"},
	"delete_pvc":           {Name: "delete_pvc", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 PVC"},
	"drain_node":           {Name: "drain_node", RiskLevel: RiskCritical, NeedApproval: true, Description: "排空节点(drain)"},
	"delete_pipeline":      {Name: "delete_pipeline", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除流水线"},
	"delete_statefulset":   {Name: "delete_statefulset", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 StatefulSet"},
	"delete_daemonset":     {Name: "delete_daemonset", RiskLevel: RiskDanger, NeedApproval: true, Description: "删除 DaemonSet"},
}

// GetToolMeta 获取工具元数据
func GetToolMeta(name string) (ToolMeta, bool) {
	meta, ok := toolRegistry[name]
	return meta, ok
}

// IsHighRisk 判断工具是否为高危操作
func IsHighRisk(name string) bool {
	meta, ok := toolRegistry[name]
	if !ok {
		return true // 未注册的工具默认高危
	}
	return meta.NeedApproval
}

// jsonSchema 辅助构建 JSON Schema
func jsonSchema(props map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": props,
		"required":   required,
	}
}

func strProp(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": desc}
}

func intProp(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": desc}
}

// BuildAllTools 构建所有工具定义（传给 OpenAI Function Calling）
func BuildAllTools() []openai.ToolDef {
	return []openai.ToolDef{
		// ==================== 查询类 ====================
		{
			Name: "list_pods", Description: "查询指定命名空间下的 Pod 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"),
				"name":      strProp("Pod名称关键字(可选，模糊搜索)"),
				"cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "get_pod_detail", Description: "获取指定 Pod 的详细信息",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Pod名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "list_deployments", Description: "查询指定命名空间下的 Deployment 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("名称关键字(可选)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "get_deployment_detail", Description: "获取指定 Deployment 的详细信息，包含副本数、镜像、状态等",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "list_services", Description: "查询指定命名空间下的 Service 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("名称关键字(可选)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "get_service_detail", Description: "获取指定 Service 的详细信息",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Service名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "list_namespaces", Description: "查询集群中所有命名空间列表",
			Parameters: jsonSchema(map[string]interface{}{
				"cluster_id": intProp("集群ID"), "name": strProp("名称关键字(可选)"),
			}, []string{"cluster_id"}),
		},
		{
			Name: "list_nodes", Description: "查询集群中所有节点列表",
			Parameters: jsonSchema(map[string]interface{}{
				"cluster_id": intProp("集群ID"), "name": strProp("节点名称关键字(可选)"),
			}, []string{"cluster_id"}),
		},
		{
			Name: "get_node_detail", Description: "获取指定节点的详细信息(CPU/内存/状态/标签/污点)",
			Parameters: jsonSchema(map[string]interface{}{
				"cluster_id": intProp("集群ID"), "name": strProp("节点名称"),
			}, []string{"cluster_id", "name"}),
		},
		{
			Name: "list_ingresses", Description: "查询指定命名空间下的 Ingress 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("名称关键字(可选)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "list_configmaps", Description: "查询指定命名空间下的 ConfigMap 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("名称关键字(可选)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "list_pvcs", Description: "查询指定命名空间下的 PVC 列表",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("名称关键字(可选)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "cluster_id"}),
		},
		{
			Name: "list_clusters", Description: "查询平台管理的所有K8s集群列表",
			Parameters: jsonSchema(map[string]interface{}{}, []string{}),
		},
		{
			Name: "list_pipelines", Description: "查询 CI/CD 流水线列表",
			Parameters: jsonSchema(map[string]interface{}{
				"keyword": strProp("搜索关键字(可选)"), "status": strProp("状态过滤(可选)"),
			}, []string{}),
		},
		{
			Name: "get_pipeline_detail", Description: "获取指定 CI/CD 流水线的详情",
			Parameters: jsonSchema(map[string]interface{}{
				"id": intProp("流水线ID"),
			}, []string{"id"}),
		},
		{
			Name: "get_events", Description: "查询K8s事件列表(可用于排查问题)",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "cluster_id": intProp("集群ID"),
				"kind": strProp("资源类型过滤(可选，如Pod/Deployment)"), "name": strProp("资源名称过滤(可选)"),
			}, []string{"namespace", "cluster_id"}),
		},

		// ==================== 写操作（需确认/审批）====================
		{
			Name: "scale_deployment", Description: "调整 Deployment 的副本数量(扩缩容)",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"),
				"replicas": intProp("目标副本数"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "replicas", "cluster_id"}),
		},
		{
			Name: "restart_deployment", Description: "重启 Deployment(触发滚动更新)",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "rollback_deployment", Description: "将 Deployment 回滚到指定版本",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"),
				"replica_set": strProp("目标 ReplicaSet 名称(可选，不填则回滚到上一版本)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "update_deployment_image", Description: "更新 Deployment 的容器镜像",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"),
				"container": strProp("容器名称"), "image": strProp("新镜像地址(如nginx:1.25)"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "container", "image", "cluster_id"}),
		},
		{
			Name: "cordon_node", Description: "标记节点为不可调度(cordon)或可调度(uncordon)",
			Parameters: jsonSchema(map[string]interface{}{
				"node_name": strProp("节点名称"), "unschedulable": map[string]interface{}{"type": "boolean", "description": "true=不可调度(cordon), false=可调度(uncordon)"},
				"cluster_id": intProp("集群ID"),
			}, []string{"node_name", "unschedulable", "cluster_id"}),
		},
		{
			Name: "trigger_pipeline", Description: "触发 CI/CD 流水线构建",
			Parameters: jsonSchema(map[string]interface{}{
				"id": intProp("流水线ID"),
			}, []string{"id"}),
		},

		// ==================== 高危删除操作 ====================
		{
			Name: "delete_pod", Description: "删除指定的 Pod",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Pod名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "delete_deployment", Description: "删除指定的 Deployment 及其所有 Pod",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Deployment名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "delete_service", Description: "删除指定的 Service",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Service名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "delete_namespace", Description: "删除整个命名空间(极高危！会删除命名空间下所有资源)",
			Parameters: jsonSchema(map[string]interface{}{
				"name": strProp("命名空间名称"), "cluster_id": intProp("集群ID"),
			}, []string{"name", "cluster_id"}),
		},
		{
			Name: "drain_node", Description: "排空节点(驱逐所有Pod并标记不可调度，极高危操作)",
			Parameters: jsonSchema(map[string]interface{}{
				"node_name": strProp("节点名称"), "cluster_id": intProp("集群ID"),
			}, []string{"node_name", "cluster_id"}),
		},
		{
			Name: "delete_configmap", Description: "删除指定的 ConfigMap",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("ConfigMap名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "delete_ingress", Description: "删除指定的 Ingress",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("Ingress名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
		{
			Name: "delete_pvc", Description: "删除指定的 PVC(可能导致数据丢失)",
			Parameters: jsonSchema(map[string]interface{}{
				"namespace": strProp("命名空间"), "name": strProp("PVC名称"), "cluster_id": intProp("集群ID"),
			}, []string{"namespace", "name", "cluster_id"}),
		},
	}
}
