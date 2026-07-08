package global

import (
	"database/sql"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	SQLDB *sql.DB
)
