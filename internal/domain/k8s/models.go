package k8s

import "time"

// =========================================================================
// K8s 域数据模型 — 纯数据结构，不依赖 DB/HTTP
// =========================================================================

// ClusterStatus 集群状态
type ClusterStatus uint8

const (
	ClusterStatusOK      uint8 = 0
	ClusterStatusBad     uint8 = 1
	ClusterStatusPending uint8 = 2
)

// Cluster 集群
type Cluster struct {
	ID             uint32 `gorm:"primaryKey;column:id" json:"id"`
	ClusterName    string `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterVersion string `gorm:"column:cluster_version" json:"cluster_version"`
	KubeConfig     string `gorm:"column:kube_config" json:"-"`
	Status         uint8  `gorm:"column:status" json:"status"`
	CreatedAt      uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt     uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt      uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	LastCheckAt    uint64 `gorm:"column:last_check_at" json:"last_check_at"`
	LastError      string `gorm:"column:last_error" json:"last_error"`
	IsDel          uint8  `gorm:"column:is_del" json:"-"`
}

// TableName 返回表名
func (Cluster) TableName() string { return "kube_cluster" }

// AggregateID 实现 domain.AggregateRoot 接口
func (c Cluster) AggregateID() int64 { return int64(c.ID) }

// EventItem K8s 事件
type EventItem struct {
	Namespace       string    `json:"namespace"`
	Kind            string    `json:"kind"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Reason          string    `json:"reason"`
	Message         string    `json:"message"`
	Count           int32     `json:"count"`
	EventTime       time.Time `json:"event_time"`
	SourceComponent string    `json:"source_component,omitempty"`
	SourceInstance  string    `json:"source_instance,omitempty"`
}

// NewEventItem 创建默认事件
func NewEventItem() *EventItem {
	return &EventItem{Type: "Normal"}
}

// NodeMetricItem 节点指标
type NodeMetricItem struct {
	Name            string    `json:"name"`
	Timestamp       time.Time `json:"timestamp"`
	WindowSeconds   int64     `json:"window_seconds"`
	CPUUsageMilli   int64     `json:"cpu_usage_milli"`
	MemUsageBytes   int64     `json:"mem_usage_bytes"`
	CPUAllocMilli   int64     `json:"cpu_alloc_milli,omitempty"`
	MemAllocBytes   int64     `json:"mem_alloc_bytes,omitempty"`
	CPUCapMilli     int64     `json:"cpu_capacity_milli,omitempty"`
	MemCapBytes     int64     `json:"mem_capacity_bytes,omitempty"`
	CPUUsagePercent float64   `json:"cpu_usage_percent,omitempty"`
	MemUsagePercent float64   `json:"mem_usage_percent,omitempty"`
}
