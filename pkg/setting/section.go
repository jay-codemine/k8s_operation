package setting

import (
	"time"
)

// ServerSettingS 服务端配置结构体
type ServerSettingS struct {
	RunMode         string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	// ReadinessTimeout 就绪探针 DB Ping 超时(秒)，需小于 K8s readinessProbe 的 timeoutSeconds
	// 留空/<=0 时代码兜底为 2s
	ReadinessTimeout time.Duration
}

// ReadSection 读取配置信息
//
//	 k: 模块名
//	 v: 配置的结构体
//		example:
//			ServerSettingS.ReadSection("Server",&ServerSettingS)
//
// 返回值:
//
//	error: 如果读取过程中发生错误，则返回错误信息
func (s *Setting) ReadSection(k string, v interface{}) error {
	// 使用 vp 的 UnmarshalKey 方法将键 k 对应的值解析到 v 中
	// 环境变量已在 NewSetting 中于 YAML 文本层面展开，无需额外处理
	if err := s.vp.UnmarshalKey(k, v); err != nil {
		return err
	}
	return nil
}

// DatabaseSettingS 定义数据库配置结构体，用于存储数据库连接相关的各项参数
type DatabaseSettingS struct {
	DBType         string        // 数据库类型，如：mysql、postgres等
	Username       string        // 数据库用户名
	Password       string        // 数据库密码
	Host           string        // 数据库主机地址
	Port           string        // 数据库端口号
	DBName         string        // 数据库名称
	Charset        string        // 数据库字符集，如：utf8mb4
	ParseTime      bool          // 是否解析时间，true表示自动解析时间类型
	MaxIdleConns   int           // 数据库连接池最大空闲连接数
	MaxOpenConns   int           // MaxOpenConns 表示数据库连接池中最大打开的连接数
	MaxLifeSeconds int // 数据库连接池中连接的最大生命周期，单位为秒
}

type AppSettingS struct {
	LogLevel               string
	LogType                string
	LogFileName            string
	LogMaxSize             int
	LogMaxBackup           int
	LogMaxAge              int
	LogCompress            bool
	BusinessLogFileName    string // 新增
	MirrorBusinessToSystem bool   // 新增
	JWTExpireTime          int
	JWTSigningKey          string
	JWTMaxRefreshTime      int
	TIMEZONE               string
	AppName                string
	GlobalKubeConfigPath   string
	// 新增
	EnableLogStreaming bool
	LogTailDefault     int64
	LogTailMax         int64
	LogLimitBytes      int64
	DefaultClusterID   uint32
	AutoInitK8s        bool // 开机自启初始化k8s集群
	AllowEmptyStart    bool // 是否允许无集群空启动（生产环境建议 false）
	SkipAutoMigrate    bool // 跳过自动建表（已通过 init.sql 初始化完成时设为 true）
}

type ErrorCodeSettingS struct {
	AllowOverride bool
}

type PodLogSetting struct {
	EnableStreaming bool
	TailDefault     int64
	TailMax         int64
	LimitBytes      int64
	Timestamps      bool
	Previous        bool
}

// CacheSettingS 缓存配置
// CacheSettingS 定义了缓存配置的结构体，包含缓存服务器的各项设置参数
type CacheSettingS struct {
	Type       string   // 缓存类型，如 redis、memcached 等
	Name       string   // 缓存名称
	Address    string   // 单节点地址，格式如 "host:port"
	Addresses  []string // Redis Cluster 节点地址列表，格式如 ["host1:6379", "host2:6379"]
	Username   string   // 缓存服务器用户名（如果需要认证）
	Password   string   // 缓存服务器密码（如果需要认证）
	MaxConnect int      // 最大连接数，控制与缓存服务器的并发连接数量
	Network    string   // 网络类型，如 "tcp"、"tcp4"、"tcp6" 等
	Secret     string   // 加密密钥，用于加密缓存数据
}

type NodeEvictionConfig struct {
	DefaultGraceSeconds   int64 `yaml:"defaultGraceSeconds"`
	MaxGraceSeconds       int64 `yaml:"maxGraceSeconds"`
	DefaultTimeoutSeconds int   `yaml:"defaultTimeoutSeconds"`
	IgnoreDaemonSets      bool  `yaml:"ignoreDaemonSets"`
	DeleteEmptyDir        bool  `yaml:"deleteEmptyDir"`
}

type PodConfig struct {
	Eviction PodEvictionConfig `yaml:"eviction"`
}
type PodEvictionConfig struct {
	DefaultGraceSeconds int64 `yaml:"default_grace_seconds"` // -1 表示用 Pod 自己的 terminationGracePeriodSeconds
}

type NodeConfig struct {
	Eviction NodeEvictionConfig `yaml:"eviction"`
}

type ClusterClientConfig struct {
	TTL       time.Duration `mapstructure:"TTL" yaml:"TTL"`
	TTLJitter time.Duration `mapstructure:"TTLJitter" yaml:"TTLJitter"`
}

// SecuritySettingS 安全配置结构体
type SecuritySettingS struct {
	// KubeConfig 加密密钥，用于加密存储在数据库中的 kubeconfig
	KubeConfigEncryptKey string `mapstructure:"KubeConfigEncryptKey"`
	// 密码加密强度（bcrypt cost），范围 4-31，默认 10
	PasswordBcryptCost int `mapstructure:"PasswordBcryptCost"`
	// 是否自动加密旧数据（启动时检查并加密未加密的数据）
	AutoEncryptLegacyData bool `mapstructure:"AutoEncryptLegacyData"`
}

// JenkinsSettingS Jenkins 配置结构体
// 用于平台驱动 CI/CD 流水线
type JenkinsSettingS struct {
	URL            string `mapstructure:"URL"`            // Jenkins 服务器地址
	Username       string `mapstructure:"Username"`       // Jenkins 用户名
	APIToken       string `mapstructure:"APIToken"`       // Jenkins API Token
	TriggerTimeout int    `mapstructure:"TriggerTimeout"` // 触发构建等待超时(秒)
	CallbackURL    string `mapstructure:"CallbackURL"`    // 平台回调地址（后端 API）
	PlatformURL    string `mapstructure:"PlatformURL"`    // 前端页面地址（用于通知链接）
	DefaultBranch  string `mapstructure:"DefaultBranch"`  // 默认 Git 分支（未指定时使用，默认 master）
	// 镜像仓库配置
	DefaultImageRegistry string `mapstructure:"DefaultImageRegistry"` // 默认镜像仓库前缀，如 registry.cn-hangzhou.aliyuncs.com/k8s-gos
	// 凭证 ID 配置（传给 Jenkins 模板，可自定义覆盖默认值）
	GitCredentialID      string `mapstructure:"GitCredentialID"`      // Git 凭证 ID（默认 gitee-id）
	RegistryCredentialID string `mapstructure:"RegistryCredentialID"` // 镜像仓库凭证 ID（默认 harbor-registry）
	HMACCredentialID     string `mapstructure:"HMACCredentialID"`     // HMAC 签名凭证 ID（默认 hmac-secret）
	// 回调机制配置
	HMACSecret   string `mapstructure:"HMACSecret"`   // HMAC 签名密钥（用于验证回调请求）
	PollInterval int    `mapstructure:"PollInterval"` // 轮询间隔(秒)，默认15
	MaxBuildTime int    `mapstructure:"MaxBuildTime"` // 最大构建时间(分钟)，超时判定失败，默认30
	MaxConcurrentBuilds int `mapstructure:"MaxConcurrentBuilds"` // 每个 Job 最大并发构建数，默认10
	// 通知配置
	EnableDingTalk  bool   `mapstructure:"EnableDingTalk"`  // 是否启用钉钉通知
	DingTalkWebhook string `mapstructure:"DingTalkWebhook"` // 钉钉机器人 Webhook URL
	EnableFeishu    bool   `mapstructure:"EnableFeishu"`    // 是否启用飞书通知
	FeishuWebhook   string `mapstructure:"FeishuWebhook"`   // 飞书机器人 Webhook URL
	FeishuSecret    string `mapstructure:"FeishuSecret"`    // 飞书机器人签名密钥
}

// GitOpsSettingS GitOps 配置结构体（ArgoCD + Argo Workflows）
// 可选配置 — 未配置时 GitOps 模式不可用，Jenkins 模式不受影响
type GitOpsSettingS struct {
	// ArgoCD 配置
	ArgoCDURL       string `mapstructure:"ArgoCDURL"`       // ArgoCD API Server 地址
	ArgoCDAuthToken string `mapstructure:"ArgoCDAuthToken"` // ArgoCD Auth Token (优先级高于 Username/Password)
	// Argo Workflows 配置
	ArgoWorkflowsURL   string `mapstructure:"ArgoWorkflowsURL"`   // Argo Workflows Server 地址
	ArgoWorkflowsToken string `mapstructure:"ArgoWorkflowsToken"` // Argo Workflows Auth Token
	// 回调配置
	CallbackURL  string `mapstructure:"CallbackURL"`  // 平台回调地址（Argo Workflows → 平台）
	HMACSecret   string `mapstructure:"HMACSecret"`   // HMAC 签名密钥（用于验证 webhook 回调）
	// Git Manifest 仓库配置
	GitManifestRepo string `mapstructure:"GitManifestRepo"` // 默认 Git Manifest 仓库
	GitManifestPath string `mapstructure:"GitManifestPath"` // 默认 manifests 路径
	GitCredentialID string `mapstructure:"GitCredentialID"` // Git 凭证 ID
	// 通知配置
	EnableDingTalk  bool   `mapstructure:"EnableDingTalk"`  // 是否启用钉钉通知
	DingTalkWebhook string `mapstructure:"DingTalkWebhook"` // 钉钉机器人 Webhook URL
	EnableFeishu    bool   `mapstructure:"EnableFeishu"`    // 是否启用飞书通知
	FeishuWebhook   string `mapstructure:"FeishuWebhook"`   // 飞书机器人 Webhook URL
	FeishuSecret    string `mapstructure:"FeishuSecret"`    // 飞书机器人签名密钥
	// 同步轮询配置
	SyncPollInterval int `mapstructure:"SyncPollInterval"` // ArgoCD 同步状态轮询间隔(秒), 默认 10
	SyncMaxWaitTime  int `mapstructure:"SyncMaxWaitTime"`  // 最大等待同步完成时间(秒), 默认 600
}

// =============================================================================
// 平台系统设置配置结构体（混合模式）
// =============================================================================

// PlatformSettingsS 平台系统设置主结构
type PlatformSettingsS struct {
	Basic        PlatformBasicSettings        `mapstructure:"Basic"`
	Security     PlatformSecuritySettings     `mapstructure:"Security"`
	Alert        PlatformAlertSettings        `mapstructure:"Alert"`
	Notification PlatformNotificationSettings `mapstructure:"Notification"`
	About        PlatformAboutSettings        `mapstructure:"About"`
}

// PlatformBasicSettings 基础设置
type PlatformBasicSettings struct {
	DefaultPage    string `mapstructure:"DefaultPage"`    // 默认进入页
	DefaultCluster string `mapstructure:"DefaultCluster"` // 默认集群
	Language       string `mapstructure:"Language"`       // 界面语言
	Timezone       string `mapstructure:"Timezone"`       // 时区设置
}

// PlatformSecuritySettings 安全设置
type PlatformSecuritySettings struct {
	SessionTimeout int    `mapstructure:"SessionTimeout"` // 会话超时(分钟)
	Enable2FA      bool   `mapstructure:"Enable2FA"`      // 双因素认证
	PasswordPolicy string `mapstructure:"PasswordPolicy"` // 密码策略
	AuditRetention int    `mapstructure:"AuditRetention"` // 审计日志保留天数
}

// PlatformAlertSettings 告警设置
type PlatformAlertSettings struct {
	CPUThreshold  int `mapstructure:"CPUThreshold"`  // CPU 阈值
	MemThreshold  int `mapstructure:"MemThreshold"`  // 内存阈值
	DiskThreshold int `mapstructure:"DiskThreshold"` // 磁盘阈值
	AlertSilence  int `mapstructure:"AlertSilence"`  // 告警静默期(分钟)
}

// PlatformNotificationSettings 通知设置
type PlatformNotificationSettings struct {
	EnableEmail    bool                 `mapstructure:"EnableEmail"`
	SMTP           SMTPSettings         `mapstructure:"SMTP"`
	EnableDingTalk bool                 `mapstructure:"EnableDingTalk"`
	DingTalk       DingTalkSettings     `mapstructure:"DingTalk"`
	EnableWebhook  bool                 `mapstructure:"EnableWebhook"`
	Webhook        WebhookSettings      `mapstructure:"Webhook"`
}

// SMTPSettings SMTP 配置
type SMTPSettings struct {
	Server   string `mapstructure:"Server"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"` // 敏感，不存数据库
	From     string `mapstructure:"From"`
	TLS      bool   `mapstructure:"TLS"`
}

// DingTalkSettings 钉钉配置
type DingTalkSettings struct {
	Webhook string `mapstructure:"Webhook"` // 敏感，不存数据库
	Secret  string `mapstructure:"Secret"`  // 加签密钥
}

// WebhookSettings 自定义 Webhook 配置
type WebhookSettings struct {
	URL     string            `mapstructure:"URL"`     // 敏感，不存数据库
	Headers map[string]string `mapstructure:"Headers"` // 自定义请求头
}

// PlatformAboutSettings 关于平台（只读）
type PlatformAboutSettings struct {
	Version    string `mapstructure:"Version"`
	BuildDate  string `mapstructure:"BuildDate"`
	GoVersion  string `mapstructure:"GoVersion"`
	VueVersion string `mapstructure:"VueVersion"`
	DBType     string `mapstructure:"DBType"`
	K8sSupport string `mapstructure:"K8sSupport"`
}

// MonitoringSettingS 监控配置结构体
type MonitoringSettingS struct {
	Enabled       bool   `mapstructure:"Enabled"`       // 是否启用监控
	PrometheusURL string `mapstructure:"PrometheusURL"` // Prometheus 地址
	QueryTimeout  int    `mapstructure:"QueryTimeout"`  // 查询超时（秒）
}

// ==================== LDAP 配置 ====================

// LDAPSettingS LDAP 认证配置
type LDAPSettingS struct {
	Enabled       bool   `mapstructure:"Enabled"`       // 是否启用 LDAP 认证
	Host          string `mapstructure:"Host"`          // LDAP 服务器地址
	Port          int    `mapstructure:"Port"`          // LDAP 端口（389/636）
	UseTLS        bool   `mapstructure:"UseTLS"`        // 是否启用 TLS（636端口用 true）
	SkipVerify    bool   `mapstructure:"SkipVerify"`    // 跳过 TLS 证书验证
	BindDN        string `mapstructure:"BindDN"`        // 管理员 Bind DN
	BindPassword  string `mapstructure:"BindPassword"`  // 管理员密码
	BaseDN        string `mapstructure:"BaseDN"`        // 用户搜索基础 DN
	UserFilter    string `mapstructure:"UserFilter"`    // 用户搜索过滤器，如 (uid=%s) 或 (sAMAccountName=%s)
	GroupBaseDN   string `mapstructure:"GroupBaseDN"`   // 组搜索基础 DN
	GroupFilter   string `mapstructure:"GroupFilter"`   // 组搜索过滤器，如 (memberUid=%s) 或 (member=%s)
	GroupAttr     string `mapstructure:"GroupAttr"`     // 组名属性（默认 cn）
	// 属性映射
	AttrUsername    string `mapstructure:"AttrUsername"`    // 用户名属性（默认 uid / sAMAccountName）
	AttrEmail       string `mapstructure:"AttrEmail"`       // 邮箱属性（默认 mail）
	AttrPhone       string `mapstructure:"AttrPhone"`       // 手机属性（默认 mobile）
	AttrDisplayName string `mapstructure:"AttrDisplayName"` // 显示名属性（默认 cn / displayName）
	// 同步策略
	SyncOnLogin   bool `mapstructure:"SyncOnLogin"`   // 登录时同步角色（推荐 true）
	AutoCreate    bool `mapstructure:"AutoCreate"`    // 首次登录自动创建用户
	LocalFallback bool `mapstructure:"LocalFallback"` // LDAP 不可用时是否回退到本地认证
	// 组-角色映射
	GroupRoleMapping []LDAPGroupRoleMapping `mapstructure:"GroupRoleMapping"`
}

// CanarySettingS 金丝雀部署配置
type CanarySettingS struct {
	Enabled             bool  `mapstructure:"Enabled"`
	DefaultReplicas     int32 `mapstructure:"DefaultReplicas"`
	DefaultTrafficRatio int32 `mapstructure:"DefaultTrafficRatio"`
	DefaultDurationSec  int32 `mapstructure:"DefaultDurationSec"`
	AutoPromote         bool  `mapstructure:"AutoPromote"`
	HealthCheckInterval int32 `mapstructure:"HealthCheckInterval"`
	MaxTrafficRatio     int32 `mapstructure:"MaxTrafficRatio"`
}

// LogControlS 日志控制配置
type LogControlS struct {
	SuppressPrometheusQueryWarn bool `mapstructure:"SuppressPrometheusQueryWarn" yaml:"SuppressPrometheusQueryWarn"`
	SuppressK8sClusterInitWarn  bool `mapstructure:"SuppressK8sClusterInitWarn" yaml:"SuppressK8sClusterInitWarn"`
	SuppressK8sClusterInitError bool `mapstructure:"SuppressK8sClusterInitError" yaml:"SuppressK8sClusterInitError"`
}

// LDAPGroupRoleMapping LDAP 组到平台角色的映射
type LDAPGroupRoleMapping struct {
	LDAPGroup     string `mapstructure:"LDAPGroup"`     // LDAP 组名（cn 值）
	PlatformRole  string `mapstructure:"PlatformRole"`  // 平台角色标识
	ClusterAccess string `mapstructure:"ClusterAccess"` // 集群域权限级别: none/read/write/admin
	CICDAccess    string `mapstructure:"CICDAccess"`    // 发布域权限级别: none/read/write/admin
}

// AIAssistantSettingS AI 助手配置（支持多模型提供商）
type AIAssistantSettingS struct {
	Enabled         bool              `mapstructure:"Enabled"`         // 是否启用 AI 助手
	DefaultProvider string            `mapstructure:"DefaultProvider"` // 默认提供商 ID
	SystemPrompt    string            `mapstructure:"SystemPrompt"`    // 全局 System Prompt
	ApprovalExpire  int               `mapstructure:"ApprovalExpire"`  // 审批过期时间(分钟)，默认 30
	MaxHistoryRound int               `mapstructure:"MaxHistoryRound"` // 会话最大历史轮数，默认 20
	Providers       []AIProviderConfig `mapstructure:"Providers"`       // 多提供商配置

	// === 兼容旧版单提供商配置（如果 Providers 为空则回退） ===
	APIKey      string  `mapstructure:"APIKey"`
	BaseURL     string  `mapstructure:"BaseURL"`
	Model       string  `mapstructure:"Model"`
	MaxTokens   int     `mapstructure:"MaxTokens"`
	Temperature float32 `mapstructure:"Temperature"`
}

// AIProviderConfig AI 提供商配置
type AIProviderConfig struct {
	ID          string          `mapstructure:"ID"`          // 提供商唯一标识: openai / deepseek / zhipu / qwen / moonshot
	Name        string          `mapstructure:"Name"`        // 显示名称
	Icon        string          `mapstructure:"Icon"`        // 图标标识
	APIKey      string          `mapstructure:"APIKey"`      // API Key
	BaseURL     string          `mapstructure:"BaseURL"`     // API 地址
	MaxTokens   int             `mapstructure:"MaxTokens"`   // 默认最大 Token
	Temperature float32         `mapstructure:"Temperature"` // 默认温度
	Models      []AIModelConfig `mapstructure:"Models"`      // 可用模型列表
}

// AIModelConfig 模型配置
type AIModelConfig struct {
	ID          string `mapstructure:"ID"`          // 模型 ID（传给 API 的）
	Name        string `mapstructure:"Name"`        // 显示名称
	MaxTokens   int    `mapstructure:"MaxTokens"`   // 最大 Token（可选，覆盖 Provider 默认值）
	Description string `mapstructure:"Description"` // 模型描述
	Capability  string `mapstructure:"Capability"`  // 能力标签: chat / reasoning / code / vision
}
