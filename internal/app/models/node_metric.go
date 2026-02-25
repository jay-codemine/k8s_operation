package models

import "time"

type NodeMetricItem struct {
	// 结构体字段定义，用于资源使用情况的数据结构
	Name          string    `json:"name"`            // 资源名称或标识符
	Timestamp     time.Time `json:"timestamp"`       // 时间戳，记录数据采集的时间点
	WindowSeconds int64     `json:"window_seconds"`  // 时间窗口大小，单位为秒
	CPUUsageMilli int64     `json:"cpu_usage_milli"` // CPU使用时间，单位为毫秒
	MemUsageBytes int64     `json:"mem_usage_bytes"` // 内存使用量，单位为字节
	// CPUAllocMilli 表示已分配的CPU时间，单位为毫秒
	// json:"cpu_alloc_milli,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	CPUAllocMilli int64 `json:"cpu_alloc_milli,omitempty"`
	// MemAllocBytes 表示已分配的内存大小，单位为字节
	// json:"mem_alloc_bytes,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	MemAllocBytes int64 `json:"mem_alloc_bytes,omitempty"`
	// CPUCapMilli 表示CPU容量，单位为毫秒
	// json:"cpu_capacity_milli,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	CPUCapMilli int64 `json:"cpu_capacity_milli,omitempty"`
	// MemCapBytes 表示内存容量，单位为字节
	// json:"mem_capacity_bytes,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	MemCapBytes int64 `json:"mem_capacity_bytes,omitempty"`
	// CPUUsagePercent 表示CPU使用率，以百分比形式表示
	// json:"cpu_usage_percent,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"`
	// MemUsagePercent 表示内存使用率，以百分比形式表示
	// json:"mem_usage_percent,omitempty" 表示在JSON序列化时使用该字段名，omitempty表示当字段值为空时省略该字段
	MemUsagePercent float64 `json:"mem_usage_percent,omitempty"`
}
