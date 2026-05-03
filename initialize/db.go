package initialize

import (
	"context"
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/database"
	"time"
)

// SetupDB 初始化DB
/*
SetupDB 函数用于初始化和配置数据库连接
根据全局配置中的数据库类型设置相应的数据库连接
目前支持 MySQL 数据库
返回值: error - 如果连接或配置过程中出现错误则返回错误信息
*/
func SetupDB() error {
	// 拼接 DSN，加上超时参数（防止连不通时卡很久）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=Local&timeout=1s&readTimeout=2s&writeTimeout=2s",
		global.DatabaseSetting.Username,
		global.DatabaseSetting.Password,
		global.DatabaseSetting.Host,
		global.DatabaseSetting.Port,
		global.DatabaseSetting.DBName,
		global.DatabaseSetting.Charset,
		global.DatabaseSetting.ParseTime,
	)

	// 创建 gorm dialector
	dbConfig := mysql.New(mysql.Config{DSN: dsn})

	// 连接数据库
	var err error
	global.DB, global.SQLDB, err = database.Connect(dbConfig, logger.Default.LogMode(logger.Info))
	if err != nil {
		return fmt.Errorf("connect db failed: %w", err)
	}

	// 连接池设置
	global.SQLDB.SetMaxOpenConns(global.DatabaseSetting.MaxOpenConns)
	global.SQLDB.SetMaxIdleConns(global.DatabaseSetting.MaxIdleConns)
	global.SQLDB.SetConnMaxLifetime(time.Duration(global.DatabaseSetting.MaxLifeSeconds) * time.Second)

	// 快速 Ping 测试连接，最多等 1 秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := global.SQLDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	// 自动迁移表结构
	if err := autoMigrateTables(); err != nil {
		return fmt.Errorf("auto migrate tables failed: %w", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("init default data failed: %w", err)
	}

	return nil
}

// autoMigrateTables 自动迁移表结构
func autoMigrateTables() error {
	// 平台基础表
	if err := global.DB.AutoMigrate(
		&models.PlatformSettings{},
		&models.AppStoreApp{},
		&models.AppStoreInstall{},
		&models.AppStoreComponent{},
		&models.CicdApproval{},
		&models.CicdBuildAgent{},
	); err != nil {
		return fmt.Errorf("migrate base tables: %w", err)
	}

	// cicd_pipeline 表字段补全（不用 AutoMigrate 避免 GORM 与已有 UNIQUE KEY 冲突）
	if err := ensurePipelineColumns(); err != nil {
		log.Printf("[AutoMigrate] cicd_pipeline 字段补全失败: %v", err)
	}

	// AI 助手模块（逐表迁移，确保每张表都成功）
	aiModels := []struct {
		name  string
		model interface{}
	}{
		{"ai_conversations", &models.AIConversation{}},
		{"ai_messages", &models.AIMessage{}},
		{"ai_approval_requests", &models.AIApprovalRequest{}},
		{"ai_approval_logs", &models.AIApprovalLog{}},
	}
	for _, m := range aiModels {
		if err := global.DB.AutoMigrate(m.model); err != nil {
			log.Printf("[AutoMigrate] 创建表 %s 失败: %v", m.name, err)
			return fmt.Errorf("migrate %s: %w", m.name, err)
		}
		log.Printf("[AutoMigrate] 表 %s 迁移成功", m.name)
	}

	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	ctx := context.Background()
	d := dao.NewDao(global.DB)

	// 初始化平台设置默认值
	if err := d.PlatformSettingsInitDefaults(ctx); err != nil {
		return fmt.Errorf("init platform settings failed: %w", err)
	}

	// 初始化应用商城种子数据
	svc := services.NewServices()
	if err := svc.AppStoreSeed(ctx); err != nil {
		return fmt.Errorf("init appstore seed data failed: %w", err)
	}

	// 初始化构建探针种子数据（OTEL Java Agent）
	if err := svc.BuildAgentSeedOTEL(ctx); err != nil {
		log.Printf("[InitData] 构建探针种子数据初始化失败: %v", err)
	}

	return nil
}

// ensurePipelineColumns 检查并补全 cicd_pipeline 表缺失的列
// 不使用 AutoMigrate 是因为 GORM 会尝试将 varchar 改为 longtext，与 UNIQUE KEY 冲突
func ensurePipelineColumns() error {
	// 检查 cicd_pipeline 表是否存在
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline'").Scan(&count)
	if count == 0 {
		return nil // 表不存在，跳过（全新安装由 SQL 初始化脚本负责）
	}

	// 所有可能缺失的列（模型新增但旧表可能没有的）
	type colDef struct {
		name string
		sql  string
	}
	columns := []colDef{
		{"language_type", "ALTER TABLE `cicd_pipeline` ADD COLUMN `language_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '语言类型' AFTER `jenkins_credential_id`"},
		{"enable_sonar", "ALTER TABLE `cicd_pipeline` ADD COLUMN `enable_sonar` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用SonarQube代码扫描' AFTER `require_approval`"},
		{"last_deploy_image", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像' AFTER `enable_sonar`"},
		{"last_deploy_digest", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要' AFTER `last_deploy_image`"},
		{"last_deploy_time", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间' AFTER `last_deploy_digest`"},
		{"last_deploy_status", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态' AFTER `last_deploy_time`"},
		{"last_deploy_version", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本' AFTER `last_deploy_status`"},
	}

	for _, col := range columns {
		var exists int64
		global.DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline' AND column_name = ?", col.name).Scan(&exists)
		if exists == 0 {
			if err := global.DB.Exec(col.sql).Error; err != nil {
				return fmt.Errorf("add column %s: %w", col.name, err)
			}
			log.Printf("[AutoMigrate] cicd_pipeline 补全列: %s", col.name)
		}
	}
	return nil
}
