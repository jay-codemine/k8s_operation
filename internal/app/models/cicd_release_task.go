package models

const (
	CicdTaskStatusPending   = "Pending"
	CicdTaskStatusRunning   = "Running"
	CicdTaskStatusSucceeded = "Succeeded"
	CicdTaskStatusFailed    = "Failed"
	CicdTaskStatusCanceled  = "Canceled"
)

// CicdReleaseTask 对应表：cicd_release_task
type CicdReleaseTask struct {
	ID        int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReleaseID int64 `gorm:"column:release_id" json:"release_id"`
	ClusterID int64 `gorm:"column:cluster_id" json:"cluster_id"`

	Status  string `gorm:"column:status" json:"status"`
	Message string `gorm:"column:message" json:"message"`

	PrevImage   string `gorm:"column:prev_image" json:"prev_image"`
	TargetImage string `gorm:"column:target_image" json:"target_image"`

	StartedAt  uint64 `gorm:"column:started_at" json:"started_at"`
	FinishedAt uint64 `gorm:"column:finished_at" json:"finished_at"`

	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt  uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel      uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdReleaseTask) TableName() string { return "cicd_release_task" }
