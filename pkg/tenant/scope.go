// Package tenant 提供多租户 GORM 查询隔离
package tenant

import (
	"context"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ctxKey struct{}

const (
	// tenantColumn 租户列名，所有多租户表统一使用
	tenantColumn = "tenant_id"
	// fillCallbackName INSERT 时自动填充 tenant_id 的 GORM 回调名
	fillCallbackName = "tenant:fill_tenant_id"
)

// NewScopedDB 创建租户隔离的 DB 实例
// 通过 Session（无 NewDB）+ Where 持久化 tenant_id 过滤条件
// 注意：Session 使用 clone=2，确保 WHERE 在链式调用中通过 Statement clone 传播
// 同时将 tenantID 存入 Statement.Context，供 JOIN 查询函数通过 GetTenantID 获取
//
// 过滤条件用 clause.CurrentTable 限定表名而非裸列名：JOIN 查询里两张表都有
// tenant_id 时，裸列名会让 MySQL 报 Error 1052 ambiguous（整个接口 500）。
// CurrentTable 在构建 SQL 时渲染为 Statement.Table，Table("xxx AS e") 这类
// 带别名的查询会被 GORM 解析成别名 e，因此渲染结果是 e.tenant_id。
func NewScopedDB(db *gorm.DB, tenantID uint32) *gorm.DB {
	ctx := context.WithValue(db.Statement.Context, ctxKey{}, tenantID)
	return db.Session(&gorm.Session{
		Context: ctx,
	}).Where(clause.Eq{
		Column: clause.Column{Table: clause.CurrentTable, Name: tenantColumn},
		Value:  tenantID,
	})
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

// RegisterCallbacks 注册 INSERT 时自动填充 tenant_id 的 GORM 回调，全局只需在 DB 初始化后调用一次。
// NewScopedDB 的 WHERE 只约束查询条件，对 INSERT 的列值无效；若不填充，
// models.Base.TenantID 的零值会被 GORM 省略、由 MySQL 的 DEFAULT 1 兜住，
// 结果是非默认租户创建的数据全部写进 1 号租户，该租户自己再也查不到。
func RegisterCallbacks(db *gorm.DB) error {
	return db.Callback().Create().Before("gorm:create").Register(fillCallbackName, fillTenantID)
}

// fillTenantID 把当前租户 ID 写入待插入记录的 tenant_id 字段，仅在该字段为零值时生效
// （调用方已显式指定租户时不覆盖，例如超级管理员跨租户写入）
func fillTenantID(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Schema == nil {
		return
	}
	// 无租户上下文（后台任务、初始化流程直接用 global.DB）时不介入
	tid, ok := GetTenantID(db)
	if !ok || tid == 0 {
		return
	}
	field := db.Statement.Schema.LookUpField(tenantColumn)
	if field == nil {
		return
	}

	ctx := db.Statement.Context
	setIfZero := func(rv reflect.Value) {
		if _, isZero := field.ValueOf(ctx, rv); isZero {
			_ = field.Set(ctx, rv, tid)
		}
	}

	switch db.Statement.ReflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			setIfZero(reflect.Indirect(db.Statement.ReflectValue.Index(i)))
		}
	case reflect.Struct:
		setIfZero(db.Statement.ReflectValue)
	}
}
