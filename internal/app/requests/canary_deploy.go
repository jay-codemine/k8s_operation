package requests

// CanaryPromoteRequest 金丝雀晋升请求
type CanaryPromoteRequest struct {
	PipelineID    int64  `json:"pipeline_id" valid:"required~pipeline_id 不能为空"`
	RunID         int64  `json:"run_id"`
	Namespace     string `json:"namespace" valid:"required~namespace 不能为空"`
	AppName       string `json:"app_name" valid:"required~app_name 不能为空"`
	ContainerName string `json:"container_name"`
}

// CanaryRollbackRequest 金丝雀回滚请求
type CanaryRollbackRequest struct {
	PipelineID int64  `json:"pipeline_id" valid:"required~pipeline_id 不能为空"`
	RunID      int64  `json:"run_id"`
	Namespace  string `json:"namespace" valid:"required~namespace 不能为空"`
	AppName    string `json:"app_name" valid:"required~app_name 不能为空"`
}

// CanaryStatusRequest 金丝雀状态查询
type CanaryStatusRequest struct {
	Namespace string `form:"namespace" valid:"required~namespace 不能为空"`
	AppName   string `form:"app_name" valid:"required~app_name 不能为空"`
}

// CanaryTrafficSplitRequest 调整金丝雀流量比例
type CanaryTrafficSplitRequest struct {
	Namespace    string `json:"namespace" valid:"required~namespace 不能为空"`
	AppName      string `json:"app_name" valid:"required~app_name 不能为空"`
	CanaryRatio  int32  `json:"canary_ratio" valid:"required~canary_ratio 不能为空,range(1|50)~canary_ratio 范围 1-50%"`
	CanaryReplicas int32 `json:"canary_replicas"` // 可选：直接指定副本数
}

// CanaryDeployConfig 金丝雀部署配置（嵌入流水线创建/更新请求）
type CanaryDeployConfig struct {
	EnableCanary       bool   `json:"enable_canary"`
	CanaryReplicas     int32  `json:"canary_replicas"`
	CanaryTrafficRatio int32  `json:"canary_traffic_ratio"`
	CanaryDurationSec  int32  `json:"canary_duration_sec"`
	CanaryAutoPromote  bool   `json:"canary_auto_promote"`
	CanaryAnalysisRules string `json:"canary_analysis_rules"`
}
