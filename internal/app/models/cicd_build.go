package models

// CicdBuild 对应表：cicd_build
type CicdBuild struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppName   string `gorm:"column:app_name" json:"app_name"`
	GitURL    string `gorm:"column:git_url" json:"git_url"`
	GitBranch string `gorm:"column:git_branch" json:"git_branch"`
	GitCommit string `gorm:"column:git_commit" json:"git_commit"`

	JenkinsJob         string `gorm:"column:jenkins_job" json:"jenkins_job"`
	JenkinsQueueID     int64  `gorm:"column:jenkins_queue_id" json:"jenkins_queue_id"`
	JenkinsBuildNumber int    `gorm:"column:jenkins_build_number" json:"jenkins_build_number"`
	JenkinsBuildURL    string `gorm:"column:jenkins_build_url" json:"jenkins_build_url"`

	Status  string `gorm:"column:status" json:"status"`
	Message string `gorm:"column:message" json:"message"`

	ImageRepo   string  `gorm:"column:image_repo" json:"image_repo"`
	ImageTag    string  `gorm:"column:image_tag" json:"image_tag"`
	ImageDigest *string `gorm:"column:image_digest" json:"image_digest"`

	SbomRef string `gorm:"column:sbom_ref" json:"sbom_ref"`
	SignRef string `gorm:"column:sign_ref" json:"sign_ref"`

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdBuild) TableName() string { return "cicd_build" }
