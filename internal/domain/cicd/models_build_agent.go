package cicd

// 构建探针类别常量
const (
	AgentCategoryObservability = "observability"
	AgentCategoryDiagnostics   = "diagnostics"
	AgentCategorySecurity      = "security"
	AgentCategoryCustom        = "custom"
)

// 探针适用语言
const (
	AgentScopeJava   = "java"
	AgentScopeGo     = "go"
	AgentScopePython = "python"
	AgentScopeAll    = "all"
)

// 探针状态
const (
	AgentStatusActive   = "active"
	AgentStatusInactive = "inactive"
)

// CicdBuildAgent 构建探针表
type CicdBuildAgent struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:100;uniqueIndex" json:"name"`
	DisplayName string `gorm:"column:display_name;size:200" json:"display_name"`
	Description string `gorm:"column:description;size:1000" json:"description"`
	Category    string `gorm:"column:category;size:30;index" json:"category"`
	Scope       string `gorm:"column:scope;size:20;index" json:"scope"`
	Version     string `gorm:"column:version;size:50" json:"version"`
	FileName    string `gorm:"column:file_name;size:300" json:"file_name"`
	FilePath    string `gorm:"column:file_path;size:500" json:"file_path"`
	FileSize    int64  `gorm:"column:file_size" json:"file_size"`
	Sha256      string `gorm:"column:sha256;size:64" json:"sha256"`
	DownloadURL string `gorm:"column:download_url;size:500" json:"download_url"`
	DocURL      string `gorm:"column:doc_url;size:500" json:"doc_url"`
	Icon        string `gorm:"column:icon;size:50" json:"icon"`

	DockerCopyDest string `gorm:"column:docker_copy_dest;size:200" json:"docker_copy_dest"`
	EnvKey         string `gorm:"column:env_key;size:100" json:"env_key"`
	EnvValue       string `gorm:"column:env_value;size:2000" json:"env_value"`

	Status        string `gorm:"column:status;size:20;default:'active'" json:"status"`
	DownloadCount int    `gorm:"column:download_count;default:0" json:"download_count"`
	UsedCount     int    `gorm:"column:used_count;default:0" json:"used_count"`

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del;default:0" json:"is_del"`
}

func (CicdBuildAgent) TableName() string { return "cicd_build_agent" }

// ValidAgentCategories 合法的探针分类
var ValidAgentCategories = []string{
	AgentCategoryObservability, AgentCategoryDiagnostics, AgentCategorySecurity, AgentCategoryCustom,
}

// ValidAgentScopes 合法的适用语言
var ValidAgentScopes = []string{
	AgentScopeJava, AgentScopeGo, AgentScopePython, AgentScopeAll,
}
