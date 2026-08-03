// Package tenant 提供多租户 GORM 查询隔离
package tenant

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

// NewScopedDB 创建租户隔离的 DB 实例
// 通过 Session（无 NewDB）+ Where 持久化 tenant_id 过滤条件
// 注意：Session 使用 clone=2，确保 WHERE 在链式调用中通过 Statement clone 传播
// 同时将 tenantID 存入 Statement.Context，供 JOIN 查询函数通过 GetTenantID 获取
func NewScopedDB(db *gorm.DB, tenantID uint32) *gorm.DB {
	ctx := context.WithValue(db.Statement.Context, ctxKey{}, tenantID)
	return db.Session(&gorm.Session{
		Context: ctx,
	}).Where("tenant_id = ?", tenantID)
}

// GetTenantID 从 DB Statement 的 Context 中提取租户 ID
// JOIN 查询使用 global.DB 显式过滤时可调用此函数获取 tenant_id 值
func GetTenantID(db *gorm.DB) (uint32, bool) {
	if db.Statement == nil || db.Statement.Context == nil {
		return 0, false
	}
	tid, ok := db.Statement.Context.Value(ctxKey{}).(uint32)
	return tid, ok
}
