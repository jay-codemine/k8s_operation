package models

// CicdPipelineTarget 对应表：cicd_pipeline_target
//
// 一条流水线（cicd_pipeline）与「多个部署环境」的映射表，是「一次构建、跨环境晋级」
// (build once, promote everywhere) 的关键数据结构。
//
// 背景：cicd_pipeline 自身只带单个 Target*（cluster/namespace/workload/container/env），
// 这导致每个环境都要重新触发流水线从 0 编译。引入本表后，一条流水线可预先绑定
// dev/test/staging/prod 等多个环境各自的部署目标；晋级时直接复用已构建的不可变镜像
// (repo@sha256:digest) 部署到目标环境，无需重新构建。
type CicdPipelineTarget struct {
	ID         int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID int64 `gorm:"column:pipeline_id;index:idx_pt_pipeline_env" json:"pipeline_id"` // 关联流水线ID

	// Env 环境标识(dev/test/staging/prod)，与 pipeline_id 组成唯一约束
	Env string `gorm:"column:env;size:32;index:idx_pt_pipeline_env" json:"env"`

	// 部署目标（该环境下这条流水线部署到哪里）
	ClusterID    int64  `gorm:"column:cluster_id" json:"cluster_id"`                          // 目标集群ID
	Namespace    string `gorm:"column:namespace;size:100" json:"namespace"`                  // 目标命名空间
	WorkloadKind string `gorm:"column:workload_kind;size:32" json:"workload_kind"`           // 工作负载类型(Deployment/StatefulSet/DaemonSet)
	WorkloadName string `gorm:"column:workload_name;size:200" json:"workload_name"`          // 工作负载名称
	Container    string `gorm:"column:container;size:200" json:"container"`                  // 容器名称

	// 晋级/审批策略
	AutoDeploy      bool   `gorm:"column:auto_deploy" json:"auto_deploy"`           // CI 构建成功后是否自动部署到本环境(通常仅 dev)
	RequireApproval bool   `gorm:"column:require_approval" json:"require_approval"` // 晋级到本环境是否需要审批(通常 staging/prod)
	PromoteFrom     string `gorm:"column:promote_from;size:32" json:"promote_from"` // 本环境镜像的上游来源环境(如 prod 来自 staging)，用于晋级链约束与可视化
	SortOrder       int    `gorm:"column:sort_order" json:"sort_order"`             // 晋级顺序/展示顺序(dev=1,test=2,staging=3,prod=4)

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdPipelineTarget) TableName() string { return "cicd_pipeline_target" }

// CicdPipelineTargetView 环境目标视图（携带集群名称，供前端展示）
type CicdPipelineTargetView struct {
	CicdPipelineTarget
	ClusterName string `json:"cluster_name"` // 集群名称
}

// PromotionEnvNode 晋级链上单个环境节点（目标配置 + 当前部署的镜像/发布单）
type PromotionEnvNode struct {
	Env             string `json:"env"`               // 环境标识
	ClusterID       int64  `json:"cluster_id"`        // 目标集群ID
	ClusterName     string `json:"cluster_name"`      // 集群名称
	Namespace       string `json:"namespace"`         // 命名空间
	WorkloadKind    string `json:"workload_kind"`     // 工作负载类型
	WorkloadName    string `json:"workload_name"`     // 工作负载名称
	Container       string `json:"container"`         // 容器名称
	AutoDeploy      bool   `json:"auto_deploy"`       // 是否自动部署
	RequireApproval bool   `json:"require_approval"`  // 是否需要审批
	PromoteFrom     string `json:"promote_from"`      // 上游来源环境
	SortOrder       int    `json:"sort_order"`        // 顺序

	// 当前环境最新一次发布信息（用于晋级链可视化）
	CurrentReleaseID     int64  `json:"current_release_id"`     // 最新发布单ID
	CurrentImageRepo     string `json:"current_image_repo"`     // 当前镜像仓库
	CurrentImageTag      string `json:"current_image_tag"`      // 当前镜像tag
	CurrentImageDigest   string `json:"current_image_digest"`   // 当前镜像digest
	CurrentReleaseStatus string `json:"current_release_status"` // 最新发布单状态
	CurrentDeployTime    uint64 `json:"current_deploy_time"`    // 最新部署时间
	Configured           bool   `json:"configured"`             // 该环境是否已配置部署目标
}
