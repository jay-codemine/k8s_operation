package appstore

import "k8soperation/pkg/db"

// ============================================================
// 应用商城 - 模型定义
// ============================================================

// AppStoreApp 应用商城应用表
type AppStoreApp struct {
	db.Base
	Name        string `gorm:"column:name;size:128;not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"column:display_name;size:256" json:"display_name"`
	Category    string `gorm:"column:category;size:64;not null;index" json:"category"`
	Version     string `gorm:"column:version;size:64;not null" json:"version"`
	Icon        string `gorm:"column:icon;size:512" json:"icon"`
	Description string `gorm:"column:description;size:1024" json:"description"`
	Provider    string `gorm:"column:provider;size:128" json:"provider"`
	ChartURL    string `gorm:"column:chart_url;size:512" json:"chart_url"`
	DocURL      string `gorm:"column:doc_url;size:512" json:"doc_url"`
	Status      uint8  `gorm:"column:status;default:1" json:"status"`
	Featured    uint8  `gorm:"column:featured;default:0" json:"featured"`
	SortOrder   int    `gorm:"column:sort_order;default:0" json:"sort_order"`
	Tags        string `gorm:"column:tags;size:512" json:"tags"`
	MinK8s      string `gorm:"column:min_k8s;size:32" json:"min_k8s"`
	Namespace   string `gorm:"column:namespace;size:128" json:"namespace"`
	ValuesYAML  string `gorm:"column:values_yaml;type:text" json:"values_yaml"`
}

func (AppStoreApp) TableName() string { return "app_store_apps" }

// AppStoreComponent 应用组件表
type AppStoreComponent struct {
	db.Base
	AppID     uint32 `gorm:"column:app_id;not null;index" json:"app_id"`
	Name      string `gorm:"column:name;size:128;not null" json:"name"`
	Image     string `gorm:"column:image;size:512;not null" json:"image"`
	Replicas  int32  `gorm:"column:replicas;default:1" json:"replicas"`
	Ports     string `gorm:"column:ports;size:512" json:"ports"`
	Args      string `gorm:"column:args;size:1024" json:"args"`
	CPUReq    string `gorm:"column:cpu_req;size:32;default:'50m'" json:"cpu_req"`
	CPULim    string `gorm:"column:cpu_lim;size:32;default:'200m'" json:"cpu_lim"`
	MemReq    string `gorm:"column:mem_req;size:32;default:'64Mi'" json:"mem_req"`
	MemLim    string `gorm:"column:mem_lim;size:32;default:'256Mi'" json:"mem_lim"`
	SortOrder int    `gorm:"column:sort_order;default:0" json:"sort_order"`
}

func (AppStoreComponent) TableName() string { return "app_store_components" }

// AppStoreComponentRequest 组件创建/更新请求
type AppStoreComponentRequest struct {
	ID        uint32 `json:"id"`
	AppID     uint32 `json:"app_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Image     string `json:"image" binding:"required"`
	Replicas  int32  `json:"replicas"`
	Ports     string `json:"ports"`
	Args      string `json:"args"`
	CPUReq    string `json:"cpu_req"`
	CPULim    string `json:"cpu_lim"`
	MemReq    string `json:"mem_req"`
	MemLim    string `json:"mem_lim"`
	SortOrder int    `json:"sort_order"`
}

// AppStoreComponentBatchDeleteRequest 批量删除组件请求
type AppStoreComponentBatchDeleteRequest struct {
	IDs []uint32 `json:"ids" binding:"required"`
}

// AppStoreComponentSortRequest 组件排序请求
type AppStoreComponentSortRequest struct {
	Items []ComponentSortItem `json:"items" binding:"required"`
}

// ComponentSortItem 单个组件排序项
type ComponentSortItem struct {
	ID        uint32 `json:"id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

// ============================================================
// Request / Response DTO
// ============================================================

// AppStoreListRequest 应用列表请求
type AppStoreListRequest struct {
	Category string `form:"category" json:"category"`
	Keyword  string `form:"keyword" json:"keyword"`
	Status   int    `form:"status" json:"status"`
	Featured int    `form:"featured" json:"featured"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

// AppStoreCreateRequest 创建应用请求
type AppStoreCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name"`
	Category    string `json:"category" binding:"required"`
	Version     string `json:"version" binding:"required"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
	ChartURL    string `json:"chart_url"`
	DocURL      string `json:"doc_url"`
	Status      uint8  `json:"status"`
	Featured    uint8  `json:"featured"`
	SortOrder   int    `json:"sort_order"`
	Tags        string `json:"tags"`
	MinK8s      string `json:"min_k8s"`
	Namespace   string `json:"namespace"`
	ValuesYAML  string `json:"values_yaml"`
}

// AppStoreUpdateRequest 更新应用请求
type AppStoreUpdateRequest struct {
	ID          uint32 `json:"id" binding:"required"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Category    string `json:"category"`
	Version     string `json:"version"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
	ChartURL    string `json:"chart_url"`
	DocURL      string `json:"doc_url"`
	Status      uint8  `json:"status"`
	Featured    uint8  `json:"featured"`
	SortOrder   int    `json:"sort_order"`
	Tags        string `json:"tags"`
	MinK8s      string `json:"min_k8s"`
	Namespace   string `json:"namespace"`
	ValuesYAML  string `json:"values_yaml"`
}

// AppStoreInstallRequest 安装应用请求
type AppStoreInstallRequest struct {
	AppID     uint32 `json:"app_id" binding:"required"`
	ClusterID uint32 `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	RelName   string `json:"release_name" binding:"required"`
	Values    string `json:"values"`
}

// AppStoreInstall 安装记录表
type AppStoreInstall struct {
	db.Base
	AppID       uint32 `gorm:"column:app_id;not null;index" json:"app_id"`
	AppName     string `gorm:"column:app_name;size:128;not null" json:"app_name"`
	ClusterID   uint32 `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	ClusterName string `gorm:"column:cluster_name;size:128" json:"cluster_name"`
	Namespace   string `gorm:"column:namespace;size:128;not null" json:"namespace"`
	ReleaseName string `gorm:"column:release_name;size:128;not null" json:"release_name"`
	Version     string `gorm:"column:version;size:64" json:"version"`
	Values      string `gorm:"column:values;type:text" json:"values"`
	Status      uint8  `gorm:"column:status;default:1" json:"status"`
	Message     string `gorm:"column:message;size:1024" json:"message"`
	Operator    string `gorm:"column:operator;size:64" json:"operator"`
}

func (AppStoreInstall) TableName() string { return "app_store_installs" }

// 安装状态常量
const (
	InstallStatusInstalling   uint8 = 1
	InstallStatusInstalled    uint8 = 2
	InstallStatusFailed       uint8 = 3
	InstallStatusUninstalling uint8 = 4
	InstallStatusUninstalled  uint8 = 5
	InstallStatusPartialReady uint8 = 6
)

// AppStoreInstallListRequest 安装记录列表请求
type AppStoreInstallListRequest struct {
	AppID     uint32 `form:"app_id" json:"app_id"`
	ClusterID uint32 `form:"cluster_id" json:"cluster_id"`
	Status    int    `form:"status" json:"status"`
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
}

// AppStoreCategoryCount 分类统计
type AppStoreCategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ============================================================
// 安装状态实时查询 Response
// ============================================================

// AppInstallStatusResponse 安装状态实时查询响应
type AppInstallStatusResponse struct {
	InstallID   uint32 `json:"install_id"`
	AppName     string `json:"app_name"`
	ClusterName string `json:"cluster_name"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"release_name"`
	Version     string `json:"version"`

	DbStatus  uint8  `json:"db_status"`
	DbMessage string `json:"db_message"`

	ClusterReachable bool   `json:"cluster_reachable"`
	ClusterError     string `json:"cluster_error,omitempty"`

	DeploymentStatus  string `json:"deployment_status"`
	DeploymentMessage string `json:"deployment_message,omitempty"`
	DesiredReplicas   int    `json:"desired_replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	UpdatedReplicas   int    `json:"updated_replicas"`
	AvailableReplicas int    `json:"available_replicas"`

	ServiceStatus string   `json:"service_status"`
	ServiceType   string   `json:"service_type,omitempty"`
	ServicePorts  []string `json:"service_ports,omitempty"`
	ClusterIP     string   `json:"cluster_ip,omitempty"`

	Pods       []PodStatusInfo       `json:"pods,omitempty"`
	ConfigMaps []ConfigMapStatusInfo `json:"configmaps,omitempty"`

	NamespaceOverview *NamespaceOverview       `json:"namespace_overview,omitempty"`
	AllDeployments    []DeploymentStatusInfo   `json:"all_deployments,omitempty"`
	AllServices       []ServiceStatusInfo      `json:"all_services,omitempty"`
	Events            []K8sEventInfo           `json:"events,omitempty"`
}

// NamespaceOverview 命名空间资源概览
type NamespaceOverview struct {
	TotalDeployments int `json:"total_deployments"`
	TotalServices    int `json:"total_services"`
	TotalPods        int `json:"total_pods"`
	TotalConfigMaps  int `json:"total_configmaps"`
	RunningPods      int `json:"running_pods"`
	PendingPods      int `json:"pending_pods"`
	FailedPods       int `json:"failed_pods"`
}

// DeploymentStatusInfo 命名空间内 Deployment 概要
type DeploymentStatusInfo struct {
	Name              string `json:"name"`
	Replicas          int    `json:"replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	UpdatedReplicas   int    `json:"updated_replicas"`
	AvailableReplicas int    `json:"available_replicas"`
	Status            string `json:"status"`
	Image             string `json:"image,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
}

// ServiceStatusInfo 命名空间内 Service 概要
type ServiceStatusInfo struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	ClusterIP string   `json:"cluster_ip"`
	Ports     []string `json:"ports,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
}

// K8sEventInfo K8s 事件信息
type K8sEventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Object    string `json:"object"`
	Count     int32  `json:"count"`
	FirstTime string `json:"first_time,omitempty"`
	LastTime  string `json:"last_time,omitempty"`
}

// PodStatusInfo Pod 状态详情
type PodStatusInfo struct {
	Name       string                `json:"name"`
	Phase      string                `json:"phase"`
	NodeName   string                `json:"node_name"`
	PodIP      string                `json:"pod_ip"`
	StartTime  string                `json:"start_time,omitempty"`
	Restarts   int                   `json:"restarts"`
	Containers []ContainerStatusInfo `json:"containers,omitempty"`
}

// ContainerStatusInfo 容器状态详情
type ContainerStatusInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	State        string `json:"state"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	RestartCount int    `json:"restart_count"`
}

// ConfigMapStatusInfo ConfigMap 状态信息
type ConfigMapStatusInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
}

// AppStoreInstallUpdateRequest 编辑安装
type AppStoreInstallUpdateRequest struct {
	Replicas *int32 `json:"replicas"`
	Image    string `json:"image"`
	CPUReq   string `json:"cpu_request"`
	CPULim   string `json:"cpu_limit"`
	MemReq   string `json:"memory_request"`
	MemLim   string `json:"memory_limit"`
}
