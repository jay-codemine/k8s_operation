package cicd

import (
	"database/sql/driver"
	"encoding/json"
)

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
	ApprovalStatusExpired  = "expired"
)

// ApprovalLevel 审批级别配置
type ApprovalLevel struct {
	Level   int     `json:"level"`
	Role    string  `json:"role"`
	Label   string  `json:"label"`
	UserIDs []int64 `json:"user_ids"`
}

// ApprovalLevels 审批级别配置列表
type ApprovalLevels []ApprovalLevel

func (a *ApprovalLevels) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		*a = nil
		return nil
	}
	if len(bytes) == 0 || string(bytes) == "null" {
		*a = nil
		return nil
	}
	return json.Unmarshal(bytes, a)
}

func (a ApprovalLevels) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// CicdEnvironment 对应表：cicd_environment
type CicdEnvironment struct {
	ID                 int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name               string         `gorm:"column:name;size:50;index" json:"name"`
	DisplayName        string         `gorm:"column:display_name;size:100" json:"display_name"`
	Description        string         `gorm:"column:description;size:500" json:"description"`
	ClusterID          int64          `gorm:"column:cluster_id;index" json:"cluster_id"`
	Namespace          string         `gorm:"column:namespace;size:100" json:"namespace"`
	Color              string         `gorm:"column:color;size:20" json:"color"`
	SortOrder          int            `gorm:"column:sort_order" json:"sort_order"`
	RequireApproval    bool           `gorm:"column:require_approval" json:"require_approval"`
	AutoRollbackOnFail bool           `gorm:"column:auto_rollback_on_fail" json:"auto_rollback_on_fail"`
	ApprovalUsers      JSONMap        `gorm:"column:approval_users;type:json" json:"approval_users"`
	ApprovalLevels     ApprovalLevels `gorm:"column:approval_levels;type:json" json:"approval_levels"`
	CreatedUserID      int64          `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt          uint64         `gorm:"column:created_at" json:"created_at"`
	ModifiedAt         uint64         `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt          uint64         `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel              uint8          `gorm:"column:is_del" json:"is_del"`
}

func (CicdEnvironment) TableName() string { return "cicd_environment" }

// CicdApproval 对应表：cicd_approval
type CicdApproval struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID    int64  `gorm:"column:pipeline_id" json:"pipeline_id"`
	PipelineRunID int64  `gorm:"column:pipeline_run_id" json:"pipeline_run_id"`
	StageID       int64  `gorm:"column:stage_id" json:"stage_id"`
	ReleaseID     int64  `gorm:"column:release_id" json:"release_id"`
	EnvName       string `gorm:"column:env_name;size:50" json:"env_name"`
	Status        string `gorm:"column:status;size:50" json:"status"`
	Image         string `gorm:"column:image;size:500" json:"image"`
	ImageDigest   string `gorm:"column:image_digest;size:100" json:"image_digest"`
	RequestUserID int64  `gorm:"column:request_user_id" json:"request_user_id"`
	RequestReason string `gorm:"column:request_reason;size:500" json:"request_reason"`
	ApproveUserID int64  `gorm:"column:approve_user_id" json:"approve_user_id"`
	ApproveReason string `gorm:"column:approve_reason;size:500" json:"approve_reason"`
	ApproveTime   uint64 `gorm:"column:approve_time" json:"approve_time"`
	ExpireTime    uint64 `gorm:"column:expire_time" json:"expire_time"`
	FeishuToken   string `gorm:"column:feishu_token;size:64" json:"feishu_token"`
	ApprovalLevel int    `gorm:"column:approval_level;default:1" json:"approval_level"`
	TotalLevels   int    `gorm:"column:total_levels;default:1" json:"total_levels"`
	LevelLabel    string `gorm:"column:level_label;size:64" json:"level_label"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdApproval) TableName() string { return "cicd_approval" }

// ApprovalListItem 审批列表项
type ApprovalListItem struct {
	CicdApproval
	PipelineName    string `json:"pipeline_name"`
	RequestUsername string `json:"request_username"`
	ApproveUsername string `json:"approve_username"`
}

// EnvironmentListItem 环境列表项
type EnvironmentListItem struct {
	CicdEnvironment
	ClusterName string `json:"cluster_name"`
}
