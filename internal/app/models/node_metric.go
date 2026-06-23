package models

import "time"

type NodeMetricItem struct {
	// 结构体字段定义，用于节点资源使用情况的数据结构
	Name          string    `json:"name"`            // 节点名称
	Timestamp     time.Time `json:"timestamp"`       // 时间戳，记录数据采集的时间点
	WindowSeconds int64     `json:"window_seconds"`  // 时间窗口大小，单位为秒
	CPUUsageMilli int64     `json:"cpu_usage_milli"` // CPU使用量，单位为毫核(millicores)
	MemUsageBytes int64     `json:"mem_usage_bytes"` // 内存使用量，单位为字节
	// CPUAllocMilli 表示已分配的CPU资源，单位为毫核(millicores)
	CPUAllocMilli int64 `json:"cpu_alloc_milli,omitempty"`
	// MemAllocBytes 表示已分配的内存大小，单位为字节
	MemAllocBytes int64 `json:"mem_alloc_bytes,omitempty"`
	// CPUCapMilli 表示CPU容量，单位为毫核(millicores)
	CPUCapMilli int64 `json:"cpu_capacity_milli,omitempty"`
	// MemCapBytes 表示内存容量，单位为字节
	MemCapBytes int64 `json:"mem_capacity_bytes,omitempty"`
	// CPUUsagePercent 表示CPU使用率，以百分比形式表示
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"`
	// MemUsagePercent 表示内存使用率，以百分比形式表示
	MemUsagePercent float64 `json:"mem_usage_percent,omitempty"`
}
