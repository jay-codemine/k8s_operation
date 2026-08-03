package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"

	"k8soperation/pkg/valid"
)

// 注意：`valid` tag 的值是 Rules map 的键（valid.ValidateOptions 里 TagIdentifier: "valid"），
// 不是规则本身。之前这几个结构体写成 valid:"required~xxx 不能为空"，键永远匹配不上 Rules，
// 校验等于没跑 —— 空 namespace/app_name 会一路传到 client-go，
// 换来一个 "resource name may not be empty" 的 500。

// CanaryPromoteRequest 金丝雀晋升请求
type CanaryPromoteRequest struct {
	PipelineID    int64  `json:"pipeline_id" valid:"pipeline_id"`
	RunID         int64  `json:"run_id"`
	Namespace     string `json:"namespace" valid:"namespace"`
	AppName       string `json:"app_name" valid:"app_name"`
	ContainerName string `json:"container_name"`
}

func ValidCanaryPromoteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"pipeline_id": []string{"required", "numeric", "min:1"},
		"namespace":   []string{"required"},
		"app_name":    []string{"required"},
	}
	messages := govalidator.MapData{
		"pipeline_id": []string{"required:流水线ID不能为空", "min:流水线ID必须大于0"},
		"namespace":   []string{"required:命名空间不能为空"},
		"app_name":    []string{"required:应用名称不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// CanaryRollbackRequest 金丝雀回滚请求
type CanaryRollbackRequest struct {
	PipelineID int64  `json:"pipeline_id" valid:"pipeline_id"`
	RunID      int64  `json:"run_id"`
	Namespace  string `json:"namespace" valid:"namespace"`
	AppName    string `json:"app_name" valid:"app_name"`
}

func ValidCanaryRollbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"pipeline_id": []string{"required", "numeric", "min:1"},
		"namespace":   []string{"required"},
		"app_name":    []string{"required"},
	}
	messages := govalidator.MapData{
		"pipeline_id": []string{"required:流水线ID不能为空", "min:流水线ID必须大于0"},
		"namespace":   []string{"required:命名空间不能为空"},
		"app_name":    []string{"required:应用名称不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// CanaryStatusRequest 金丝雀状态查询
type CanaryStatusRequest struct {
	Namespace string `form:"namespace" valid:"namespace"`
	AppName   string `form:"app_name" valid:"app_name"`
}

func ValidCanaryStatusRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"app_name":  []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required:命名空间不能为空"},
		"app_name":  []string{"required:应用名称不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// CanaryTrafficSplitRequest 调整金丝雀流量比例
type CanaryTrafficSplitRequest struct {
	Namespace      string `json:"namespace" valid:"namespace"`
	AppName        string `json:"app_name" valid:"app_name"`
	CanaryRatio    int32  `json:"canary_ratio" valid:"canary_ratio"`
	CanaryReplicas int32  `json:"canary_replicas"` // 可选：直接指定副本数
}

func ValidCanaryTrafficSplitRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":    []string{"required"},
		"app_name":     []string{"required"},
		"canary_ratio": []string{"required", "numeric", "min:1", "max:50"},
	}
	messages := govalidator.MapData{
		"namespace":    []string{"required:命名空间不能为空"},
		"app_name":     []string{"required:应用名称不能为空"},
		"canary_ratio": []string{"required:流量比例不能为空", "min:流量比例范围 1-50%", "max:流量比例范围 1-50%"},
	}
	return valid.ValidateOptions(data, rules, messages)
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
