// Package tenant 提供多租户 GORM 查询隔离
package tenant

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewScopedDB 创建租户隔离的 DB 实例
// 使用 Query 回调注入表名限定的 WHERE，避免 JOIN 时列名歧义
func NewScopedDB(db *gorm.DB, tenantID uint32) *gorm.DB {
	newDB := db.Session(&gorm.Session{NewDB: true})
	_ = newDB.Callback().Query().Before("gorm:query").Register("tenant:scope", func(d *gorm.DB) {
		tbl := d.Statement.Table
		if tbl == "" && d.Statement.Schema != nil {
			tbl = d.Statement.Schema.Table
		}
		if tbl != "" {
			d.Clauses(clause.Where{
				Exprs: []clause.Expression{
					clause.Eq{
						Column: clause.Column{Table: tbl, Name: "tenant_id"},
						Value:  tenantID,
					},
				},
			})
		}
	})
	return newDB
}
