package models

// 制品类型常量
const (
	ArtifactTypeJar      = "jar"      // Java JAR 包
	ArtifactTypeWar      = "war"      // Java WAR 包
	ArtifactTypeBinary   = "binary"   // Go 二进制
	ArtifactTypeDist     = "dist"     // 前端构建产物（dist.tar.gz）
	ArtifactTypeWheel    = "wheel"    // Python wheel 包
	ArtifactTypeImage    = "image"    // Docker 镜像（记录引用，不存储文件）
	ArtifactTypeArchive  = "archive"  // 通用压缩包
)

// 制品状态常量
const (
	ArtifactStatusUploading = "uploading" // 上传中
	ArtifactStatusReady     = "ready"     // 就绪可用
	ArtifactStatusExpired   = "expired"   // 已过期
	ArtifactStatusDeleted   = "deleted"   // 已删除
)

// CicdArtifact 制品库 - 构建产物管理
// 复合索引设计：
//   idx_list_query  → 加速 ArtifactList 分页查询（is_del + pipeline_id + status + created_at DESC）
//   idx_run_status  → 加速 ArtifactListByRunID + 状态筛选
type CicdArtifact struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID int64  `gorm:"column:pipeline_id;index:idx_list_query" json:"pipeline_id"` // 关联流水线
	RunID      int64  `gorm:"column:run_id;index:idx_run_status" json:"run_id"`            // 关联运行记录
	BuildNumber int   `gorm:"column:build_number" json:"build_number"`                    // Jenkins 构建号

	// 制品基本信息
	Name         string `gorm:"column:name;size:200" json:"name"`                                          // 制品名称（如 order-service-1.0.0.jar）
	ArtifactType string `gorm:"column:artifact_type;size:20;index:idx_list_query" json:"artifact_type"`     // 制品类型
	Version      string `gorm:"column:version;size:100" json:"version"`                                    // 版本号
	LanguageType string `gorm:"column:language_type;size:20;index:idx_list_query" json:"language_type"`     // 语言类型

	// 存储信息
	FilePath   string `gorm:"column:file_path;size:500" json:"file_path"`     // 文件存储路径
	FileSize   int64  `gorm:"column:file_size" json:"file_size"`              // 文件大小（字节）
	Sha256     string `gorm:"column:sha256;size:64" json:"sha256"`            // SHA256 校验和
	StorageType string `gorm:"column:storage_type;size:20;default:'local'" json:"storage_type"` // 存储类型：local/s3/oss

	// Git 信息（构建来源追溯）
	GitRepo    string `gorm:"column:git_repo;size:500" json:"git_repo"`
	GitBranch  string `gorm:"column:git_branch;size:100" json:"git_branch"`
	GitCommit  string `gorm:"column:git_commit;size:40" json:"git_commit"`

	// 镜像信息（如果制品已打包为镜像）
	ImageRepo   string `gorm:"column:image_repo;size:500" json:"image_repo"`
	ImageTag    string `gorm:"column:image_tag;size:200" json:"image_tag"`
	ImageDigest string `gorm:"column:image_digest;size:100" json:"image_digest"`

	// 构建元数据
	BuildDuration int    `gorm:"column:build_duration" json:"build_duration"` // 构建耗时（秒）
	BuildLog      string `gorm:"column:build_log;type:text" json:"build_log"` // 构建摘要日志
	Metadata      JSONMap `gorm:"column:metadata;type:json" json:"metadata"`  // 扩展元数据

	// 状态
	Status        string `gorm:"column:status;size:20;default:'ready';index:idx_list_query;index:idx_run_status" json:"status"`
	DownloadCount int    `gorm:"column:download_count;default:0" json:"download_count"` // 下载次数
	
	// 元数据
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
