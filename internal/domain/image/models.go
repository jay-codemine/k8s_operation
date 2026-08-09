package image

// ImageCleanupPolicy 镜像清理策略
type ImageCleanupPolicy struct {
	ID                int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	RegistryID        int64  `gorm:"type:bigint;not null;index:idx_registry_id" json:"registry_id"`
	Name              string `gorm:"type:varchar(100);not null" json:"name"`
	Enabled           bool   `gorm:"type:tinyint(1);default:1" json:"enabled"`
	RepositoryPattern string `gorm:"type:varchar(200);default:'*'" json:"repository_pattern"`
	TagPattern        string `gorm:"type:varchar(200);default:'*'" json:"tag_pattern"`
	KeepLastCount     int    `gorm:"type:int;default:5" json:"keep_last_count"`
	KeepDays          int    `gorm:"type:int;default:30" json:"keep_days"`
	CronExpression    string `gorm:"type:varchar(50);default:'0 2 * * *'" json:"cron_expression"`
	LastRunAt         int64  `gorm:"type:bigint" json:"last_run_at"`
	LastRunResult     string `gorm:"type:varchar(500)" json:"last_run_result"`
	DeletedCount      int64  `gorm:"type:bigint;default:0" json:"deleted_count"`
	Description       string `gorm:"type:varchar(500)" json:"description"`
	CreatedBy         int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt         int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt        int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel             int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (ImageCleanupPolicy) TableName() string { return "image_cleanup_policy" }

// ImageCleanupLog 清理任务日志
type ImageCleanupLog struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PolicyID     int64  `gorm:"type:bigint;not null;index:idx_policy_id" json:"policy_id"`
	RegistryID   int64  `gorm:"type:bigint;not null" json:"registry_id"`
	StartTime    int64  `gorm:"type:bigint;not null;index:idx_start_time" json:"start_time"`
	EndTime      int64  `gorm:"type:bigint" json:"end_time"`
	Status       string `gorm:"type:varchar(20);default:'running'" json:"status"`
	ScannedCount int    `gorm:"type:int;default:0" json:"scanned_count"`
	DeletedCount int    `gorm:"type:int;default:0" json:"deleted_count"`
	FreedSize    int64  `gorm:"type:bigint;default:0" json:"freed_size"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`
	Details      string `gorm:"type:json" json:"details"`
}

func (ImageCleanupLog) TableName() string { return "image_cleanup_log" }

// ImageRegistry 镜像仓库配置
type ImageRegistry struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"type:varchar(100);not null;uniqueIndex:idx_registry_name" json:"name"`
	Type            string `gorm:"type:varchar(50);not null;default:'docker'" json:"type"`
	URL             string `gorm:"type:varchar(500);not null" json:"url"`
	Username        string `gorm:"type:varchar(100)" json:"username"`
	Password        string `gorm:"type:varchar(500)" json:"-"`
	AccessKeyID     string `gorm:"type:varchar(100)" json:"access_key_id"`
	AccessKeySecret string `gorm:"type:varchar(200)" json:"-"`
	Region          string `gorm:"type:varchar(50)" json:"region"`
	Insecure        bool   `gorm:"type:tinyint(1);default:0" json:"insecure"`
	Description     string `gorm:"type:varchar(500)" json:"description"`
	IsDefault       bool   `gorm:"type:tinyint(1);default:0" json:"is_default"`
	Status          string `gorm:"type:varchar(50);default:'unknown'" json:"status"`
	LastCheckAt     int64  `gorm:"type:bigint" json:"last_check_at"`
	LastError       string `gorm:"type:varchar(500)" json:"last_error"`
	CreatedBy       int64  `gorm:"type:bigint" json:"created_by"`
	CreatedAt       int64  `gorm:"type:bigint;autoCreateTime" json:"created_at"`
	ModifiedAt      int64  `gorm:"type:bigint;autoUpdateTime" json:"modified_at"`
	IsDel           int    `gorm:"type:tinyint(1);default:0" json:"-"`
}

func (ImageRegistry) TableName() string { return "image_registry" }
