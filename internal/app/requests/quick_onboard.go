package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ========== 快速接入应用 ==========

func NewQuickOnboardRequest() *QuickOnboardRequest {
	return &QuickOnboardRequest{
		Namespace:    "default",
		WorkloadKind: "Deployment",
		Replicas:     1,
		ServiceType:  "ClusterIP",
	}
}

// QuickOnboardRequest 一键接入：创建 K8s 资源 + 接入流水线 + 可选首次部署
type QuickOnboardRequest struct {
	// --- 基础信息 ---
	AppName       string `json:"app_name" valid:"app_name"`
	Namespace     string `json:"namespace" valid:"namespace"`
	WorkloadKind  string `json:"workload_kind" valid:"workload_kind"` // Deployment|StatefulSet|DaemonSet|CronJob|Job
	WorkloadName  string `json:"workload_name" valid:"workload_name"` // 可选，默认=app_name
	ContainerName string `json:"container_name" valid:"container_name"`

	// --- 镜像 ---
	Image string `json:"image" valid:"image"`

	// --- 副本数（Deployment/StatefulSet 有效）---
	Replicas int32 `json:"replicas"`

	// --- 端口 ---
	Ports []PortDef `json:"ports"` // [{name:"http",port:8080}]

	// --- 资源配置 ---
	CPUReq string `json:"cpu_req"` // 如 "100m"
	CPULim string `json:"cpu_lim"` // 如 "500m"
	MemReq string `json:"mem_req"` // 如 "128Mi"
	MemLim string `json:"mem_lim"` // 如 "256Mi"

	// --- 环境变量 ---
	EnvVars []EnvVarDef `json:"env_vars"` // [{name:"ENV",value:"test"}]

	// --- CronJob 专属 ---
	CronSchedule          string `json:"cron_schedule"`            // 如 "*/5 * * * *"
	CronConcurrencyPolicy string `json:"cron_concurrency_policy"`  // Allow/Forbid/Replace

	// --- Job 专属 ---
	JobCompletions           *int32 `json:"job_completions"`            // 完成次数
	JobParallelism           *int32 `json:"job_parallelism"`            // 并行度
	JobBackoffLimit          *int32 `json:"job_backoff_limit"`          // 重试次数
	JobTTLSecondsAfterFinished *int32 `json:"job_ttl_seconds_after_finished"` // 完成后保留时间

	// --- Service 配置 ---
	ServiceType  string `json:"service_type"`  // ClusterIP|NodePort|LoadBalancer，空=不创建
	ServicePorts []PortDef `json:"service_ports"` // 可选，默认用 ports

	// --- 集群 ---
	ClusterID uint32 `json:"cluster_id" valid:"cluster_id"`

	// --- 可选：是否立即部署 ---
	AutoDeploy bool `json:"auto_deploy"`

	// --- 可选：Git 仓库（用于创建流水线）---
	GitRepo   string `json:"git_repo"`
	GitBranch string `json:"git_branch"`

	// --- ConfigMap ---
	ConfigMapData      []KVDef `json:"configmap_data"`
	ConfigMapMountPath string  `json:"configmap_mount_path"` // 如 /etc/config

	// --- Secret ---
	SecretData      []KVDef `json:"secret_data"`
	SecretMountPath string  `json:"secret_mount_path"` // 如 /etc/secrets

	// --- PVC 存储 ---
	PVCName        string `json:"pvc_name"`
	PVCSize        string `json:"pvc_size"`         // 如 "10Gi"
	PVCStorageClass string `json:"pvc_storage_class"` // storage class 名称
	PVCAccessMode  string `json:"pvc_access_mode"`   // ReadWriteOnce/ReadWriteMany/ReadOnlyMany
	PVCMountPath   string `json:"pvc_mount_path"`    // 如 /data
}

// KVDef 键值对（用于 ConfigMap / Secret / Env）
type KVDef struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PortDef 端口定义
type PortDef struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"` // TCP/UDP，默认 TCP
}

// EnvVarDef 环境变量定义
type EnvVarDef struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ValidQuickOnboardRequest(data interface{}, ctx *gin.Context) map[string][]string {
	req := data.(*QuickOnboardRequest)

	rules := govalidator.MapData{
		"app_name":      []string{"required"},
		"namespace":     []string{"required"},
		"workload_kind": []string{"required", "in:Deployment,StatefulSet,DaemonSet,CronJob,Job"},
		"image":         []string{"required"},
		"cluster_id":    []string{"required"},
	}

	messages := govalidator.MapData{
		"app_name":      []string{"required:应用名称不能为空"},
		"namespace":     []string{"required:命名空间不能为空"},
		"workload_kind": []string{"required:工作负载类型不能为空", "in:工作负载类型必须是 Deployment/StatefulSet/DaemonSet/CronJob/Job"},
		"image":         []string{"required:镜像地址不能为空"},
		"cluster_id":    []string{"required:目标集群不能为空"},
	}

	msgs := valid.ValidateOptions(data, rules, messages)

	// CronJob 额外校验
	if req.WorkloadKind == "CronJob" && req.CronSchedule == "" {
		if msgs == nil {
			msgs = make(map[string][]string)
		}
		msgs["cron_schedule"] = []string{"CronJob 必须设置定时调度表达式"}
	}

	return msgs
}
