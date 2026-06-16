package models

// ============================================================
// 应用商城 - Model 定义
// 参考 Rancher Apps / Kuboard 应用商店 设计
// ============================================================

// AppStoreApp 应用商城应用表（对应 app_store_apps）
type AppStoreApp struct {
	Base
	Name        string `gorm:"column:name;size:128;not null;uniqueIndex" json:"name" description:"应用名称"`
	DisplayName string `gorm:"column:display_name;size:256" json:"display_name" description:"显示名称"`
	Category    string `gorm:"column:category;size:64;not null;index" json:"category" description:"分类: 网络/监控/日志/存储/安全/GitOps/数据库/消息队列"`
	Version     string `gorm:"column:version;size:64;not null" json:"version" description:"版本号"`
	Icon        string `gorm:"column:icon;size:512" json:"icon" description:"图标URL或内置icon名"`
	Description string `gorm:"column:description;size:1024" json:"description" description:"应用描述"`
	Provider    string `gorm:"column:provider;size:128" json:"provider" description:"提供方: official/community/third-party"`
	ChartURL    string `gorm:"column:chart_url;size:512" json:"chart_url" description:"Helm Chart地址"`
	DocURL      string `gorm:"column:doc_url;size:512" json:"doc_url" description:"文档地址"`
	Status      uint8  `gorm:"column:status;default:1" json:"status" description:"状态: 1-可用 2-维护中 3-已下架"`
	Featured    uint8  `gorm:"column:featured;default:0" json:"featured" description:"是否推荐: 0-否 1-是"`
	SortOrder   int    `gorm:"column:sort_order;default:0" json:"sort_order" description:"排序权重"`
	Tags        string `gorm:"column:tags;size:512" json:"tags" description:"标签,逗号分隔"`
	MinK8s      string `gorm:"column:min_k8s;size:32" json:"min_k8s" description:"最低K8s版本要求"`
	Namespace   string `gorm:"column:namespace;size:128" json:"namespace" description:"默认安装命名空间"`
	ValuesYAML  string `gorm:"column:values_yaml;type:text" json:"values_yaml" description:"默认values.yaml"`
}

func (AppStoreApp) TableName() string {
	return "app_store_apps"
}

// AppStoreComponent 应用组件表（对应 app_store_components）
// 每个应用可包含多个组件，安装时按组件列表创建 Deployment + Service
// 参考 Rancher App Catalog / Helm Chart 的子组件拆分方式
type AppStoreComponent struct {
	Base
	AppID     uint32 `gorm:"column:app_id;not null;index" json:"app_id" description:"关联应用ID"`
	Name      string `gorm:"column:name;size:128;not null" json:"name" description:"组件名称(如 prometheus/grafana)"`
	Image     string `gorm:"column:image;size:512;not null" json:"image" description:"容器镜像"`
	Replicas  int32  `gorm:"column:replicas;default:1" json:"replicas" description:"副本数"`
	Ports     string `gorm:"column:ports;size:512" json:"ports" description:"端口定义JSON,如: [{\"name\":\"http\",\"port\":9090}]"`
	Args      string `gorm:"column:args;size:1024" json:"args" description:"容器启动参数JSON,如: [\"--config.file=...\"]"`
	CPUReq    string `gorm:"column:cpu_req;size:32;default:'50m'" json:"cpu_req" description:"CPU Request"`
	CPULim    string `gorm:"column:cpu_lim;size:32;default:'200m'" json:"cpu_lim" description:"CPU Limit"`
	MemReq    string `gorm:"column:mem_req;size:32;default:'64Mi'" json:"mem_req" description:"Memory Request"`
	MemLim    string `gorm:"column:mem_lim;size:32;default:'256Mi'" json:"mem_lim" description:"Memory Limit"`
	SortOrder int    `gorm:"column:sort_order;default:0" json:"sort_order" description:"排序(越大越靠前,主组件设最大)"`
}

func (AppStoreComponent) TableName() string {
	return "app_store_components"
}

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
	Category string `form:"category" json:"category"`   // 分类筛选
	Keyword  string `form:"keyword" json:"keyword"`     // 搜索关键词
	Status   int    `form:"status" json:"status"`       // 状态筛选
	Featured int    `form:"featured" json:"featured"`   // 推荐筛选: -1不限/0否/1是
	Page     int    `form:"page" json:"page"`           // 页码
	PageSize int    `form:"page_size" json:"page_size"` // 每页数量
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
	RelName   string `json:"release_name" binding:"required"` // Helm release name
	Values    string `json:"values"`                          // 自定义values.yaml
}

// AppStoreInstall 安装记录表（对应 app_store_installs）
type AppStoreInstall struct {
	Base
	AppID       uint32 `gorm:"column:app_id;not null;index" json:"app_id" description:"应用ID"`
	AppName     string `gorm:"column:app_name;size:128;not null" json:"app_name" description:"应用名称(冗余)"`
	ClusterID   uint32 `gorm:"column:cluster_id;not null;index" json:"cluster_id" description:"集群ID"`
	ClusterName string `gorm:"column:cluster_name;size:128" json:"cluster_name" description:"集群名称(冗余)"`
	Namespace   string `gorm:"column:namespace;size:128;not null" json:"namespace" description:"安装命名空间"`
	ReleaseName string `gorm:"column:release_name;size:128;not null" json:"release_name" description:"Release名称"`
	Version     string `gorm:"column:version;size:64" json:"version" description:"安装版本"`
	Values      string `gorm:"column:values;type:text" json:"values" description:"自定义values"`
	Status      uint8  `gorm:"column:status;default:1" json:"status" description:"状态: 1-安装中 2-已安装 3-安装失败 4-卸载中 5-已卸载"`
	Message     string `gorm:"column:message;size:1024" json:"message" description:"状态消息"`
	Operator    string `gorm:"column:operator;size:64" json:"operator" description:"操作人"`
}

func (AppStoreInstall) TableName() string {
	return "app_store_installs"
}

// 安装状态常量
const (
	InstallStatusInstalling    uint8 = 1 // 安装中
	InstallStatusInstalled     uint8 = 2 // 已安装（所有组件 Ready）
	InstallStatusFailed        uint8 = 3 // 安装失败
	InstallStatusUninstalling  uint8 = 4 // 卸载中
	InstallStatusUninstalled   uint8 = 5 // 已卸载
	InstallStatusPartialReady  uint8 = 6 // 部分就绪（部分组件未 Ready，降级运行）
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

// AppInstallStatusResponse 安装状态实时查询响应（包含 K8s 集群真实状态）
type AppInstallStatusResponse struct {
	InstallID   uint32 `json:"install_id"`
	AppName     string `json:"app_name"`
	ClusterName string `json:"cluster_name"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"release_name"`
	Version     string `json:"version"`

	// 数据库状态
	DbStatus  uint8  `json:"db_status"`
	DbMessage string `json:"db_message"`

	// 集群连接
	ClusterReachable bool   `json:"cluster_reachable"`
	ClusterError     string `json:"cluster_error,omitempty"`

	// Deployment 状态（主 release）
	DeploymentStatus  string `json:"deployment_status"` // Running/Pending/Failed/NotFound/Error
	DeploymentMessage string `json:"deployment_message,omitempty"`
	DesiredReplicas   int    `json:"desired_replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	UpdatedReplicas   int    `json:"updated_replicas"`
	AvailableReplicas int    `json:"available_replicas"`

	// Service 状态（主 release）
	ServiceStatus string   `json:"service_status"` // Active/NotFound
	ServiceType   string   `json:"service_type,omitempty"`
	ServicePorts  []string `json:"service_ports,omitempty"`
	ClusterIP     string   `json:"cluster_ip,omitempty"`

	// Pod 详情
	Pods []PodStatusInfo `json:"pods,omitempty"`

	// ConfigMap 列表
	ConfigMaps []ConfigMapStatusInfo `json:"configmaps,omitempty"`

	// 命名空间级别资源概览
	NamespaceOverview *NamespaceOverview `json:"namespace_overview,omitempty"`

	// 命名空间内所有 Deployment
	AllDeployments []DeploymentStatusInfo `json:"all_deployments,omitempty"`

	// 命名空间内所有 Service
	AllServices []ServiceStatusInfo `json:"all_services,omitempty"`

	// 最近 Events
	Events []K8sEventInfo `json:"events,omitempty"`
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
	Status            string `json:"status"` // Running/Pending/Failed
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
	Type      string `json:"type"`      // Normal/Warning
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Object    string `json:"object"`    // 涉及的对象 kind/name
	Count     int32  `json:"count"`
	FirstTime string `json:"first_time,omitempty"`
	LastTime  string `json:"last_time,omitempty"`
}

// PodStatusInfo Pod 状态详情
type PodStatusInfo struct {
	Name       string                `json:"name"`
	Phase      string                `json:"phase"` // Running/Pending/Succeeded/Failed/Unknown
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
	State        string `json:"state"` // Running/Waiting/Terminated
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

// AppStoreInstallUpdateRequest 编辑安装（更新 Deployment 参数）
type AppStoreInstallUpdateRequest struct {
	Replicas *int32 `json:"replicas"`          // 副本数
	Image    string `json:"image"`             // 镜像地址
	CPUReq   string `json:"cpu_request"`       // CPU Request
	CPULim   string `json:"cpu_limit"`         // CPU Limit
	MemReq   string `json:"memory_request"`    // Memory Request
	MemLim   string `json:"memory_limit"`      // Memory Limit
}
