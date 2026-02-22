package models

const (
	CicdReleaseStatusPending   = "Pending"
	CicdReleaseStatusQueued    = "Queued"
	CicdReleaseStatusRunning   = "Running"
	CicdReleaseStatusSucceeded = "Succeeded"
	CicdReleaseStatusFailed    = "Failed"
	CicdReleaseStatusCanceled  = "Canceled"
	CicdReleaseStatusRollback  = "Rollback" // 已回滚
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
