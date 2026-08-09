package cicd

// CicdPipelineTarget 流水线-环境部署目标映射表
type CicdPipelineTarget struct {
	ID         int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID int64 `gorm:"column:pipeline_id;index:idx_pt_pipeline_env" json:"pipeline_id"`
	Env        string `gorm:"column:env;size:32;index:idx_pt_pipeline_env" json:"env"`

	ClusterID    int64  `gorm:"column:cluster_id" json:"cluster_id"`
	Namespace    string `gorm:"column:namespace;size:100" json:"namespace"`
	WorkloadKind string `gorm:"column:workload_kind;size:32" json:"workload_kind"`
	WorkloadName string `gorm:"column:workload_name;size:200" json:"workload_name"`
	Container    string `gorm:"column:container;size:200" json:"container"`

	AutoDeploy      bool   `gorm:"column:auto_deploy" json:"auto_deploy"`
	RequireApproval bool   `gorm:"column:require_approval" json:"require_approval"`
	PromoteFrom     string `gorm:"column:promote_from;size:32" json:"promote_from"`
	SortOrder       int    `gorm:"column:sort_order" json:"sort_order"`

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdPipelineTarget) TableName() string { return "cicd_pipeline_target" }

// CicdPipelineTargetView 环境目标视图
type CicdPipelineTargetView struct {
	CicdPipelineTarget
	ClusterName string `json:"cluster_name"`
}

// PromotionEnvNode 晋级链上单个环境节点
type PromotionEnvNode struct {
	Env             string `json:"env"`
	ClusterID       int64  `json:"cluster_id"`
	ClusterName     string `json:"cluster_name"`
	Namespace       string `json:"namespace"`
	WorkloadKind    string `json:"workload_kind"`
	WorkloadName    string `json:"workload_name"`
	Container       string `json:"container"`
	AutoDeploy      bool   `json:"auto_deploy"`
	RequireApproval bool   `json:"require_approval"`
	PromoteFrom     string `json:"promote_from"`
	SortOrder       int    `json:"sort_order"`

	CurrentReleaseID     int64  `json:"current_release_id"`
	CurrentImageRepo     string `json:"current_image_repo"`
	CurrentImageTag      string `json:"current_image_tag"`
	CurrentImageDigest   string `json:"current_image_digest"`
	CurrentReleaseStatus string `json:"current_release_status"`
	CurrentDeployTime    uint64 `json:"current_deploy_time"`
	Configured           bool   `json:"configured"`
}
