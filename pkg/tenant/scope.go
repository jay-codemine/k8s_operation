// Package tenant 提供多租户 GORM 查询 Scope
package tenant

import (
	"gorm.io/gorm"
)

// ScopeKey context key for tenant-scoped DB
type contextKey string

const DBKey contextKey = "tenant_db"

// Scope 返回一个 GORM Scope，自动为所有查询添加 WHERE tenant_id = ?
// 用法：db.Scopes(tenant.Scope(tid)).Find(&users)
func Scope(tenantID uint32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// NewScopedDB 创建一个带租户过滤的 DB 实例
// 后续所有查询自动带 WHERE tenant_id = tid
func NewScopedDB(db *gorm.DB, tenantID uint32) *gorm.DB {
	return db.Scopes(Scope(tenantID))
}
