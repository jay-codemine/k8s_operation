package models

// CicdReleaseStageStatus 阶段状态定义
type CicdReleaseStageStatus string

const (
	CicdReleaseStageStatusPending   CicdReleaseStageStatus = "pending"
	CicdReleaseStageStatusRunning   CicdReleaseStageStatus = "running"
	CicdReleaseStageStatusSuccess   CicdReleaseStageStatus = "success"
	CicdReleaseStageStatusFailed    CicdReleaseStageStatus = "failed"
	CicdReleaseStageStatusSkipped   CicdReleaseStageStatus = "skipped"
	CicdReleaseStageStatusWaiting   CicdReleaseStageStatus = "waiting"
	CicdReleaseStageStatusAborted   CicdReleaseStageStatus = "aborted"
)

// CicdReleaseStage 对应表：cicd_release_stage
type CicdReleaseStage struct {
	ID         int64                  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReleaseID  int64                  `gorm:"column:release_id" json:"release_id"`
	StageName  string                 `gorm:"column:stage_name" json:"stage_name"`
	StageOrder int                    `gorm:"column:stage_order" json:"stage_order"`
	Status     CicdReleaseStageStatus `gorm:"column:status" json:"status"`
	Message    string                 `gorm:"column:message" json:"message"`
	Logs       string                 `gorm:"column:logs;type:longtext" json:"logs"`

	StartTime  uint64 `gorm:"column:start_time" json:"start_time"`
	EndTime    uint64 `gorm:"column:end_time" json:"end_time"`
	Duration   uint64 `gorm:"column:duration" json:"duration"`
	BuildNumber string `gorm:"column:build_number" json:"build_number"`

	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdReleaseStage) TableName() string { return "cicd_release_stage" }