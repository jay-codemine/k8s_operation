package models

import "k8soperation/pkg/db"

// 重新导出共享 DB 工具函数

var (
	ScopeNotDeleted = db.ScopeNotDeleted
	ScopeLikeName   = db.ScopeLikeName
	ScopeOrderBy    = db.ScopeOrderBy
	Paginate        = db.Paginate
)

type PageResult[T any] = db.PageResult[T]
