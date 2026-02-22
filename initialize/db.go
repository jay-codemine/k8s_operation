package initialize

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"k8soperation/global"
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

	return nil
}
