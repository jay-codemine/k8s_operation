package database

import (
	"database/sql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect 建立数据库连接
// 参数：
//	dbConfig gorm.Dialector         - GORM 的数据库驱动配置（包含连接方式、DSN 等）
//	_logger gormlogger.Interface    - GORM 的日志实现（控制 SQL 日志输出）

// 返回值：
//
//	*gorm.DB  - GORM 的 ORM 操作对象
//	*sql.DB   - Go 标准库的底层数据库连接对象（可做连接池等配置）
//	error     - 连接或初始化过程中出现的错误
func Connect(dbConfig gorm.Dialector, _logger gormlogger.Interface) (*gorm.DB, *sql.DB, error) {
	// 1. 使用 gorm.Open 建立数据库连接
	// - dbConfig 用来指定驱动和 DSN（如 mysql.Open(...)）
	// - Logger 设置 SQL 日志实现
	db, err := gorm.Open(dbConfig, &gorm.Config{
		Logger: _logger,
	})
	if err != nil {
		// 如果连接失败，返回 nil（*gorm.DB）、nil（*sql.DB）和错误
		return nil, nil, err
	}

	// 2. 从 *gorm.DB 获取底层 *sql.DB
	//    - *sql.DB 可用于连接池配置、执行原生 SQL、Ping 检查等
	sqldb, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 3. 成功返回两个数据库对象和 nil 错误
	return db, sqldb, nil
}
