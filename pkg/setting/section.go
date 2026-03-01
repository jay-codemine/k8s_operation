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
	// 如果解析过程中发生错误，则直接返回该错误
	if err := s.vp.UnmarshalKey(k, v); err != nil {
		return err
	}
	// 如果解析成功，返回 nil 表示无错误
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
	MaxLifeSeconds time.Duration // 数据库连接池中连接的最大生命周期，单位为秒
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
	Type       string // 缓存类型，如 redis、memcached 等
	Name       string // 缓存名称
	Address    string // 缓存服务器地址，格式如 "host:port"
	Username   string // 缓存服务器用户名（如果需要认证）
	Password   string // 缓存服务器密码（如果需要认证）
	MaxConnect int    // 最大连接数，控制与缓存服务器的并发连接数量
	Network    string // 网络类型，如 "tcp"、"tcp4"、"tcp6" 等
	Secret     string // 加密密钥，用于加密缓存数据
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

// JenkinsSettingS Jenkins 配置结构体
// 用于平台驱动 CI/CD 流水线
type JenkinsSettingS struct {
	URL            string `mapstructure:"URL"`            // Jenkins 服务器地址
	Username       string `mapstructure:"Username"`       // Jenkins 用户名
	APIToken       string `mapstructure:"APIToken"`       // Jenkins API Token
	TriggerTimeout int    `mapstructure:"TriggerTimeout"` // 触发构建等待超时(秒)
	CallbackURL    string `mapstructure:"CallbackURL"`    // 平台回调地址（后端 API）
	PlatformURL    string `mapstructure:"PlatformURL"`    // 前端页面地址（用于通知链接）
	// 回调机制配置
	HMACSecret   string `mapstructure:"HMACSecret"`   // HMAC 签名密钥（用于验证回调请求）
	PollInterval int    `mapstructure:"PollInterval"` // 轮询间隔(秒)，默认15
	MaxBuildTime int    `mapstructure:"MaxBuildTime"` // 最大构建时间(分钟)，超时判定失败，默认30
	// 通知配置
	DingTalkWebhook string `mapstructure:"DingTalkWebhook"` // 钉钉机器人 Webhook URL
}
