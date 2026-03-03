package models

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"  // 待审批
	ApprovalStatusApproved = "approved" // 已通过
	ApprovalStatusRejected = "rejected" // 已拒绝
	ApprovalStatusExpired  = "expired"  // 已过期
)

// CicdEnvironment 对应表：cicd_environment
type CicdEnvironment struct {
	ID              int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string  `gorm:"column:name" json:"name"`                       // 环境名称(dev/staging/prod)
	DisplayName     string  `gorm:"column:display_name" json:"display_name"`       // 显示名称
	Description     string  `gorm:"column:description" json:"description"`         // 描述
	ClusterID       int64   `gorm:"column:cluster_id" json:"cluster_id"`           // 关联集群ID
	Namespace       string  `gorm:"column:namespace" json:"namespace"`             // 默认命名空间
	Color           string  `gorm:"column:color" json:"color"`                     // 环境颜色标识
	SortOrder       int     `gorm:"column:sort_order" json:"sort_order"`           // 排序
	RequireApproval bool    `gorm:"column:require_approval" json:"require_approval"` // 是否需要审批
	ApprovalUsers   JSONMap `gorm:"column:approval_users;type:json" json:"approval_users"` // 审批人员列表
	CreatedUserID   int64   `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt       uint64  `gorm:"column:created_at" json:"created_at"`
	ModifiedAt      uint64  `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt       uint64  `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel           uint8   `gorm:"column:is_del" json:"is_del"`
}

func (CicdEnvironment) TableName() string { return "cicd_environment" }

// CicdApproval 对应表：cicd_approval
type CicdApproval struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID    int64  `gorm:"column:pipeline_id" json:"pipeline_id"`       // 流水线ID
	PipelineRunID int64  `gorm:"column:pipeline_run_id" json:"pipeline_run_id"` // 运行记录ID
	ReleaseID     int64  `gorm:"column:release_id" json:"release_id"`         // 发布单ID
	EnvName       string `gorm:"column:env_name" json:"env_name"`             // 目标环境
	Status        string `gorm:"column:status" json:"status"`                 // 状态
	Image         string `gorm:"column:image" json:"image"`                   // 待部署镜像
	ImageDigest   string `gorm:"column:image_digest" json:"image_digest"`     // 镜像摘要
	RequestUserID int64  `gorm:"column:request_user_id" json:"request_user_id"` // 申请人
	RequestReason string `gorm:"column:request_reason" json:"request_reason"` // 申请原因
	ApproveUserID int64  `gorm:"column:approve_user_id" json:"approve_user_id"` // 审批人
	ApproveReason string `gorm:"column:approve_reason" json:"approve_reason"` // 审批意见
	ApproveTime   uint64 `gorm:"column:approve_time" json:"approve_time"`     // 审批时间
	ExpireTime    uint64 `gorm:"column:expire_time" json:"expire_time"`       // 过期时间
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdApproval) TableName() string { return "cicd_approval" }

// EnvironmentListItem 环境列表项（包含集群名称）
type EnvironmentListItem struct {
	CicdEnvironment
	ClusterName string `json:"cluster_name"` // 集群名称
}
