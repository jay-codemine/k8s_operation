package models

import (
	"database/sql/driver"
	"encoding/json"
)

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"  // 待审批
	ApprovalStatusApproved = "approved" // 已通过
	ApprovalStatusRejected = "rejected" // 已拒绝
	ApprovalStatusExpired  = "expired"  // 已过期
)

// ApprovalLevel 审批级别配置（JSON 数组中的一项）
type ApprovalLevel struct {
	Level   int     `json:"level"`    // 审批级别（1=一级审批，2=二级审批...）
	Role    string  `json:"role"`     // 角色标识（如 test_lead / dev_lead / ops_lead / sre）
	Label   string  `json:"label"`    // 审批级别显示名称（如"测试负责人"）
	UserIDs []int64 `json:"user_ids"` // 该级别审批人 ID 列表
}

// ApprovalLevels 审批级别配置列表（GORM JSON 序列化）
type ApprovalLevels []ApprovalLevel

// Scan implements sql.Scanner
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

// Value implements driver.Valuer
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
	ID              int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string         `gorm:"column:name;size:50;index" json:"name"`                       // 环境名称(dev/staging/prod)
	DisplayName     string         `gorm:"column:display_name;size:100" json:"display_name"`       // 显示名称
	Description     string         `gorm:"column:description;size:500" json:"description"`         // 描述
	ClusterID       int64          `gorm:"column:cluster_id;index" json:"cluster_id"`           // 关联集群ID
	Namespace       string         `gorm:"column:namespace;size:100" json:"namespace"`             // 默认命名空间
	Color           string         `gorm:"column:color;size:20" json:"color"`                     // 环境颜色标识
	SortOrder       int            `gorm:"column:sort_order" json:"sort_order"`           // 排序
	RequireApproval bool           `gorm:"column:require_approval" json:"require_approval"` // 是否需要审批
	AutoRollbackOnFail bool        `gorm:"column:auto_rollback_on_fail" json:"auto_rollback_on_fail"` // 部署失败时自动回滚到上一版本（建议生产环境开启）
	ApprovalUsers   JSONMap        `gorm:"column:approval_users;type:json" json:"approval_users"` // 审批人员列表（兼容旧逻辑）
	ApprovalLevels  ApprovalLevels `gorm:"column:approval_levels;type:json" json:"approval_levels"` // 多级审批配置（新）
	CreatedUserID   int64          `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt       uint64         `gorm:"column:created_at" json:"created_at"`
	ModifiedAt      uint64         `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt       uint64         `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel           uint8          `gorm:"column:is_del" json:"is_del"`
}

func (CicdEnvironment) TableName() string { return "cicd_environment" }

// CicdApproval 对应表：cicd_approval
type CicdApproval struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID    int64  `gorm:"column:pipeline_id" json:"pipeline_id"`       // 流水线ID
	PipelineRunID int64  `gorm:"column:pipeline_run_id" json:"pipeline_run_id"` // 运行记录ID
	StageID       int64  `gorm:"column:stage_id" json:"stage_id"`             // 关联流水线阶段ID
	ReleaseID     int64  `gorm:"column:release_id" json:"release_id"`         // 发布单ID
	EnvName       string `gorm:"column:env_name;size:50" json:"env_name"`             // 目标环境
	Status        string `gorm:"column:status;size:50" json:"status"`                 // 状态
	Image         string `gorm:"column:image;size:500" json:"image"`                   // 待部署镜像
	ImageDigest   string `gorm:"column:image_digest;size:100" json:"image_digest"`     // 镜像摘要
	RequestUserID int64  `gorm:"column:request_user_id" json:"request_user_id"` // 申请人
	RequestReason string `gorm:"column:request_reason;size:500" json:"request_reason"` // 申请原因
	ApproveUserID int64  `gorm:"column:approve_user_id" json:"approve_user_id"` // 审批人
	ApproveReason string `gorm:"column:approve_reason;size:500" json:"approve_reason"` // 审批意见
	ApproveTime   uint64 `gorm:"column:approve_time" json:"approve_time"`     // 审批时间
	ExpireTime    uint64 `gorm:"column:expire_time" json:"expire_time"`       // 过期时间
	FeishuToken   string `gorm:"column:feishu_token;size:64" json:"feishu_token"` // 飞书审批回调Token（唯一标识，用于回调验证）
	ApprovalLevel int    `gorm:"column:approval_level;default:1" json:"approval_level"` // 当前审批级别（1=一级，2=二级...）
	TotalLevels   int    `gorm:"column:total_levels;default:1" json:"total_levels"`     // 总审批级数
	LevelLabel    string `gorm:"column:level_label;size:64" json:"level_label"`         // 当前级别显示名称
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdApproval) TableName() string { return "cicd_approval" }

// ApprovalListItem 审批列表项（包含流水线名称和申请人名称）
type ApprovalListItem struct {
	CicdApproval
	PipelineName    string `json:"pipeline_name"`     // 流水线名称
	RequestUsername string `json:"request_username"`  // 申请人用户名
	ApproveUsername string `json:"approve_username"`  // 审批人用户名
}

// EnvironmentListItem 环境列表项（包含集群名称）
type EnvironmentListItem struct {
	CicdEnvironment
	ClusterName string `json:"cluster_name"` // 集群名称
}
