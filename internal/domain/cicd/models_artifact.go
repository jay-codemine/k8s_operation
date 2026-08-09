package cicd

// 制品类型常量
const (
	ArtifactTypeJar     = "jar"
	ArtifactTypeWar     = "war"
	ArtifactTypeBinary  = "binary"
	ArtifactTypeDist    = "dist"
	ArtifactTypeWheel   = "wheel"
	ArtifactTypeImage   = "image"
	ArtifactTypeArchive = "archive"
)

// 制品状态常量
const (
	ArtifactStatusUploading = "uploading"
	ArtifactStatusReady     = "ready"
	ArtifactStatusExpired   = "expired"
	ArtifactStatusDeleted   = "deleted"
)

// CicdArtifact 制品库 - 构建产物管理
type CicdArtifact struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID  int64  `gorm:"column:pipeline_id;index:idx_list_query" json:"pipeline_id"`
	RunID       int64  `gorm:"column:run_id;index:idx_run_status" json:"run_id"`
	BuildNumber int    `gorm:"column:build_number" json:"build_number"`

	Name         string `gorm:"column:name;size:200" json:"name"`
	ArtifactType string `gorm:"column:artifact_type;size:20;index:idx_list_query" json:"artifact_type"`
	Version      string `gorm:"column:version;size:100" json:"version"`
	LanguageType string `gorm:"column:language_type;size:20;index:idx_list_query" json:"language_type"`

	FilePath    string `gorm:"column:file_path;size:500" json:"file_path"`
	FileSize    int64  `gorm:"column:file_size" json:"file_size"`
	Sha256      string `gorm:"column:sha256;size:64" json:"sha256"`
	StorageType string `gorm:"column:storage_type;size:20;default:'local'" json:"storage_type"`

	GitRepo   string `gorm:"column:git_repo;size:500" json:"git_repo"`
	GitBranch string `gorm:"column:git_branch;size:100" json:"git_branch"`
	GitCommit string `gorm:"column:git_commit;size:40" json:"git_commit"`

	ImageRepo   string `gorm:"column:image_repo;size:500" json:"image_repo"`
	ImageTag    string `gorm:"column:image_tag;size:200" json:"image_tag"`
	ImageDigest string `gorm:"column:image_digest;size:100" json:"image_digest"`

	BuildDuration int    `gorm:"column:build_duration" json:"build_duration"`
	BuildLog      string `gorm:"column:build_log;type:text" json:"build_log"`
	Metadata      JSONMap `gorm:"column:metadata;type:json" json:"metadata"`

	Status        string `gorm:"column:status;size:20;default:'ready';index:idx_list_query;index:idx_run_status" json:"status"`
	DownloadCount int    `gorm:"column:download_count;default:0" json:"download_count"`

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at;index:idx_list_query" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del;index:idx_list_query" json:"is_del"`
}

func (CicdArtifact) TableName() string { return "cicd_artifact" }

// ArtifactTypeByLanguage 根据语言类型推导默认制品类型
var ArtifactTypeByLanguage = map[string]string{
	LanguageTypeJava:     ArtifactTypeJar,
	LanguageTypeGo:       ArtifactTypeBinary,
	LanguageTypeFrontend: ArtifactTypeDist,
	LanguageTypePython:   ArtifactTypeWheel,
}
