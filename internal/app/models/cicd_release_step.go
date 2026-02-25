package models

// CicdReleaseStep 对应表：cicd_release_step
type CicdReleaseStep struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID    int64  `gorm:"column:task_id" json:"task_id"`
	Step      string `gorm:"column:step" json:"step"`       // Patch/Watch/Done
	Status    string `gorm:"column:status" json:"status"`   // Running/Succeeded/Failed
	Message   string `gorm:"column:message" json:"message"` // 详情
	CreatedAt uint64 `gorm:"column:created_at" json:"created_at"`
}

func (CicdReleaseStep) TableName() string { return "cicd_release_step" }
