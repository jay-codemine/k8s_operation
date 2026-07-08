package requests

import "k8soperation/internal/app/models"

// ResourceValidateRequest 资源校验请求
type ResourceValidateRequest struct {
	Env         string                `json:"env" binding:"required"`          // dev/test/staging/prod
	ServiceType string                `json:"service_type" binding:"required"` // java/go/node/python
	Config      models.ResourceConfig `json:"config" binding:"required"`       // 资源配置
}

// ResourceApprovalCreateRequest 资源审批创建请求
type ResourceApprovalCreateRequest struct {
	PipelineID      uint64                `json:"pipeline_id" binding:"required"`
	ReleaseID       uint64                `json:"release_id"`
	Env             string                `json:"env" binding:"required"`
	RequestedConfig models.ResourceConfig `json:"requested_config" binding:"required"`
}
