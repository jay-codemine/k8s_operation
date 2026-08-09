package cicd

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// CicdResourceTemplate 资源档位模板
type CicdResourceTemplate struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(64);not null" json:"name"`
	ServiceType string `gorm:"type:varchar(32);not null" json:"service_type"`
	Env         string `gorm:"type:varchar(16);not null" json:"env"`

	ReplicasDefault int    `gorm:"default:1" json:"replicas_default"`
	ReplicasMin     int    `gorm:"default:1" json:"replicas_min"`
	ReplicasMax     int    `gorm:"default:10" json:"replicas_max"`
	CPURequest      string `gorm:"type:varchar(16);default:'200m'" json:"cpu_request"`
	CPULimit        string `gorm:"type:varchar(16);default:'500m'" json:"cpu_limit"`
	MemoryRequest   string `gorm:"type:varchar(16);default:'256Mi'" json:"memory_request"`
	MemoryLimit     string `gorm:"type:varchar(16);default:'512Mi'" json:"memory_limit"`

	HPAEnabled     bool `gorm:"default:false" json:"hpa_enabled"`
	HPAMinReplicas int  `gorm:"default:2" json:"hpa_min_replicas"`
	HPAMaxReplicas int  `gorm:"default:10" json:"hpa_max_replicas"`
	HPACPUTarget   int  `gorm:"default:70" json:"hpa_cpu_target"`

	Description string `gorm:"type:varchar(255)" json:"description"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	CreatedAt   uint64 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt  uint64 `gorm:"autoUpdateTime" json:"modified_at"`
	DeletedAt   uint64 `gorm:"default:0" json:"-"`
}

func (CicdResourceTemplate) TableName() string { return "cicd_resource_template" }

// CicdEnvResourceRule 环境资源规则
type CicdEnvResourceRule struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Env         string `gorm:"type:varchar(16);not null" json:"env"`
	ServiceType string `gorm:"type:varchar(32);default:''" json:"service_type"`

	CPULimitMax     string `gorm:"type:varchar(16);default:'4'" json:"cpu_limit_max"`
	MemoryLimitMax  string `gorm:"type:varchar(16);default:'8Gi'" json:"memory_limit_max"`
	ReplicasMax     int    `gorm:"default:10" json:"replicas_max"`
	CPURequestMin   string `gorm:"type:varchar(16);default:''" json:"cpu_request_min"`
	MemoryRequestMin string `gorm:"type:varchar(16);default:''" json:"memory_request_min"`
	ReplicasMin     int    `gorm:"default:1" json:"replicas_min"`

	RequireApproval bool   `gorm:"default:false" json:"require_approval"`
	ApprovalRole    string `gorm:"type:varchar(64);default:''" json:"approval_role"`

	Description string `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   uint64 `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt  uint64 `gorm:"autoUpdateTime" json:"modified_at"`
}

func (CicdEnvResourceRule) TableName() string { return "cicd_env_resource_rule" }

// ApprovalStatusCancelled 审批已取消状态
const ApprovalStatusCancelled = "cancelled"

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// CicdDeployApproval 发布审批记录
type CicdDeployApproval struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PipelineID uint64 `gorm:"not null;index" json:"pipeline_id"`
	ReleaseID  uint64 `gorm:"default:0" json:"release_id"`
	Env        string `gorm:"type:varchar(16);not null" json:"env"`

	RequestedConfig ResourceConfig `gorm:"type:text;serializer:json" json:"requested_config"`
	CurrentConfig   ResourceConfig `gorm:"type:text;serializer:json" json:"current_config"`

	RiskLevel    RiskLevel `gorm:"type:varchar(16);default:'low'" json:"risk_level"`
	RiskWarnings []string  `gorm:"type:text;serializer:json" json:"risk_warnings"`

	Status         string `gorm:"type:varchar(16);default:'pending';index" json:"status"`
	ApplicantID    uint64 `gorm:"not null;index" json:"applicant_id"`
	ApplicantName  string `gorm:"type:varchar(64);not null" json:"applicant_name"`
	ApproverID     uint64 `gorm:"default:0" json:"approver_id"`
	ApproverName   string `gorm:"type:varchar(64);default:''" json:"approver_name"`
	ApproveComment string `gorm:"type:varchar(500);default:''" json:"approve_comment"`

	AppliedAt  uint64 `gorm:"default:0" json:"applied_at"`
	ApprovedAt uint64 `gorm:"default:0" json:"approved_at"`
	ExpiredAt  uint64 `gorm:"default:0" json:"expired_at"`
}

func (CicdDeployApproval) TableName() string { return "cicd_deploy_approval" }

// ChangeType 变更类型
type ChangeType string

const (
	ChangeTypeCreate   ChangeType = "create"
	ChangeTypeUpdate   ChangeType = "update"
	ChangeTypeScale    ChangeType = "scale"
	ChangeTypeRollback ChangeType = "rollback"
)

// CicdResourceChangeLog 资源配置变更日志
type CicdResourceChangeLog struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PipelineID uint64 `gorm:"not null;index" json:"pipeline_id"`
	Env        string `gorm:"type:varchar(16);not null" json:"env"`

	ChangeType   ChangeType     `gorm:"type:varchar(16);not null" json:"change_type"`
	BeforeConfig ResourceConfig `gorm:"type:text;serializer:json" json:"before_config"`
	AfterConfig  ResourceConfig `gorm:"type:text;serializer:json" json:"after_config"`

	OperatorID   uint64 `gorm:"not null" json:"operator_id"`
	OperatorName string `gorm:"type:varchar(64);not null" json:"operator_name"`
	Reason       string `gorm:"type:varchar(500);default:''" json:"reason"`

	CreatedAt uint64 `gorm:"autoCreateTime;index" json:"created_at"`
}

func (CicdResourceChangeLog) TableName() string { return "cicd_resource_change_log" }

// ResourceConfig 资源配置
type ResourceConfig struct {
	Replicas  int              `json:"replicas"`
	Strategy  string           `json:"strategy"`
	Resources ResourceRequests `json:"resources"`
	HPA       *HPAConfig       `json:"hpa,omitempty"`
}

// ResourceRequests 资源请求和限制
type ResourceRequests struct {
	Requests ResourceValues `json:"requests"`
	Limits   ResourceValues `json:"limits"`
}

// ResourceValues CPU和内存值
type ResourceValues struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// HPAConfig HPA配置
type HPAConfig struct {
	Enabled     bool `json:"enabled"`
	MinReplicas int  `json:"min_replicas"`
	MaxReplicas int  `json:"max_replicas"`
	CPUTarget   int  `json:"cpu_target"`
}

func (r ResourceConfig) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ResourceConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

// ResourceValidationResult 资源校验结果
type ResourceValidationResult struct {
	Valid        bool      `json:"valid"`
	Warnings     []string  `json:"warnings"`
	Errors       []string  `json:"errors"`
	NeedApproval bool      `json:"need_approval"`
	ApprovalRole string    `json:"approval_role"`
	RiskLevel    RiskLevel `json:"risk_level"`
	Suggestion   string    `json:"suggestion"`
}
