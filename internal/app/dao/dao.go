package dao

import (
	"gorm.io/gorm"
)

type Dao struct {
	db *gorm.DB
}

func NewDao(db *gorm.DB) *Dao {
	return &Dao{db: db}
}

// DB 返回底层 *gorm.DB，供 Service 层执行复杂查询（领域化过渡期使用）
func (d *Dao) DB() *gorm.DB {
	return d.db
}
