package models

const (
	CicdReleaseStatusPending          = "Pending"
	CicdReleaseStatusAwaitingApproval = "AwaitingApproval" // 等待审批
	CicdReleaseStatusQueued           = "Queued"
	CicdReleaseStatusRunning          = "Running"
	CicdReleaseStatusSucceeded        = "Succeeded"
	CicdReleaseStatusFailed           = "Failed"
	CicdReleaseStatusCanceled         = "Canceled"
	CicdReleaseStatusRollback         = "Rollback" // 已回滚
)

// CicdRelease 对应表：cicd_release
type CicdRelease struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppName   string `gorm:"column:app_name" json:"app_name"`
	Namespace string `gorm:"column:namespace" json:"namespace"`

	WorkloadKind  string `gorm:"column:workload_kind" json:"workload_kind"`
	WorkloadName  string `gorm:"column:workload_name" json:"workload_name"`
	ContainerName string `gorm:"column:container_name" json:"container_name"`

	Strategy    string `gorm:"column:strategy" json:"strategy"`
	TimeoutSec  uint32 `gorm:"column:timeout_sec" json:"timeout_sec"`
	Concurrency uint32 `gorm:"column:concurrency" json:"concurrency"` // 并发数

	Status        string `gorm:"column:status" json:"status"`
	Message       string `gorm:"column:message" json:"message"`
	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	RequestID     string `gorm:"column:request_id" json:"request_id"`

	BuildID     int64   `gorm:"column:build_id" json:"build_id"`
	ImageRepo   string  `gorm:"column:image_repo" json:"image_repo"`
	ImageTag    string  `gorm:"column:image_tag" json:"image_tag"`
	ImageDigest *string `gorm:"column:image_digest" json:"image_digest"` // nullable

	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt  uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel      uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdRelease) TableName() string { return "cicd_release" }

// CicdReleaseWithDeployMode 带部署模式的发布单（前端列表用）
type CicdReleaseWithDeployMode struct {
	CicdRelease
	DeployMode string `json:"deploy_mode"`
}

// ==================== GitOps 发布统计与搜索 ====================

// GitOpsReleaseStats GitOps 发布统计数据
type GitOpsReleaseStats struct {
	Total       int64   `json:"total"`
	Synced      int64   `json:"synced"`
	Failed      int64   `json:"failed"`
	Running     int64   `json:"running"`
	PendingSync int64   `json:"pending_sync"`
	TodayCount  int64   `json:"today_count"`
	AvgSyncSec  float64 `json:"avg_sync_sec"`
	ActiveApps  int64   `json:"active_apps"`
	SuccessRate float64 `json:"success_rate"`
}

// GitOpsReleaseSearchRequest 增强搜索请求
type GitOpsReleaseSearchRequest struct {
	Keyword    string `form:"keyword" json:"keyword"`
	AppName    string `form:"app_name" json:"app_name"`
	Status     string `form:"status" json:"status"`
	SyncStatus string `form:"sync_status" json:"sync_status"`
	Env        string `form:"env" json:"env"`
	DateFrom   string `form:"date_from" json:"date_from"`
	DateTo     string `form:"date_to" json:"date_to"`
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
}

// GitOpsReleaseItem 前端展示用 GitOps 发布项
type GitOpsReleaseItem struct {
	ID           int64  `json:"id"`
	AppName      string `json:"app_name"`
	PipelineID   int64  `json:"pipeline_id"`
	PipelineName string `json:"pipeline_name"`
	Status       string `json:"status"`
	SyncStatus   string `json:"sync_status"`
	SyncRevision string `json:"sync_revision"`
	ArgoApp      string `json:"argo_app"`
	Workflow     string `json:"workflow"`
	ImageRepo    string `json:"image_repo"`
	ImageTag     string `json:"image_tag"`
	Namespace    string `json:"namespace"`
	DeployMode   string `json:"deploy_mode"`
	CreatedAt    uint64 `json:"created_at"`
}
