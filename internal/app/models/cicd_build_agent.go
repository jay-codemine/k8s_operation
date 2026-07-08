package models

// ==================== 构建探针（Build Agent）类型常量 ====================
// 用于 CI/CD 构建过程中注入到 Docker 镜像的探针/代理
// 如：OpenTelemetry Java Agent、SkyWalking Agent、Arthas 等
const (
	AgentCategoryObservability = "observability" // 可观测性（OTEL、SkyWalking、Pinpoint）
	AgentCategoryDiagnostics   = "diagnostics"   // 诊断工具（Arthas、JProfiler Agent）
	AgentCategorySecurity      = "security"       // 安全扫描（RASP、Falco Probe）
	AgentCategoryCustom        = "custom"         // 自定义探针
)

// 探针适用语言
const (
	AgentScopeJava   = "java"
	AgentScopeGo     = "go"
	AgentScopePython = "python"
	AgentScopeAll    = "all" // 通用，不限语言
)

// 探针状态
const (
	AgentStatusActive   = "active"   // 启用中
	AgentStatusInactive = "inactive" // 已停用
)

// CicdBuildAgent 构建探针表 - 管理 CI/CD 构建中注入的 Agent JAR / Binary
type CicdBuildAgent struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:100;uniqueIndex" json:"name"`                  // 探针名称（如 opentelemetry-javaagent）
	DisplayName string `gorm:"column:display_name;size:200" json:"display_name"`              // 显示名称（如 OpenTelemetry Java Agent）
	Description string `gorm:"column:description;size:1000" json:"description"`               // 描述说明
	Category    string `gorm:"column:category;size:30;index" json:"category"`                 // 分类：observability/diagnostics/security/custom
	Scope       string `gorm:"column:scope;size:20;index" json:"scope"`                       // 适用语言：java/go/python/all
	Version     string `gorm:"column:version;size:50" json:"version"`                         // 当前版本（如 1.33.0）
	FileName    string `gorm:"column:file_name;size:300" json:"file_name"`                    // 文件名（如 opentelemetry-javaagent.jar）
	FilePath    string `gorm:"column:file_path;size:500" json:"file_path"`                    // 存储路径
	FileSize    int64  `gorm:"column:file_size" json:"file_size"`                             // 文件大小（字节）
	Sha256      string `gorm:"column:sha256;size:64" json:"sha256"`                           // SHA256 校验和
	DownloadURL string `gorm:"column:download_url;size:500" json:"download_url"`              // 官方下载地址（用于版本更新参考）
	DocURL      string `gorm:"column:doc_url;size:500" json:"doc_url"`                        // 文档地址
	Icon        string `gorm:"column:icon;size:50" json:"icon"`                               // 图标 emoji 或 URL

	// Dockerfile 注入配置
	DockerCopyDest string `gorm:"column:docker_copy_dest;size:200" json:"docker_copy_dest"` // 镜像内目标路径（如 /app/opentelemetry-javaagent.jar）
	EnvKey         string `gorm:"column:env_key;size:100" json:"env_key"`                   // 注入的环境变量名（如 OTEL_OPTS）
	EnvValue       string `gorm:"column:env_value;size:2000" json:"env_value"`              // 环境变量默认值模板

	// 状态
	Status        string `gorm:"column:status;size:20;default:'active'" json:"status"` // active/inactive
	DownloadCount int    `gorm:"column:download_count;default:0" json:"download_count"`
	UsedCount     int    `gorm:"column:used_count;default:0" json:"used_count"` // 被流水线引用次数

	// 元数据
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
