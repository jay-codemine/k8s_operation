// Package tenant 提供多租户 GORM 查询隔离
package tenant

import (
	"gorm.io/gorm"
)

// NewScopedDB 创建租户隔离的 DB 实例
// 后续所有查询自动带 WHERE tenant_id = tid
// 使用 Session 确保条件在 Model()/Find() 等链式调用中持久化
func NewScopedDB(db *gorm.DB, tenantID uint32) *gorm.DB {
	return db.Session(&gorm.Session{}).Where("tenant_id = ?", tenantID)
}
