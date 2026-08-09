package db

import "gorm.io/gorm"

// ScopeNotDeleted 软删除过滤
func ScopeNotDeleted() func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("is_del = 0")
	}
}

// ScopeLikeName 模糊搜索
func ScopeLikeName(field, value string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if value == "" {
			return tx
		}
		return tx.Where(field+" LIKE ?", "%"+value+"%")
	}
}

// ScopeOrderBy 排序
func ScopeOrderBy(field string, desc bool) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if field == "" {
			return tx
		}
		order := field
		if desc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		return tx.Order(order)
	}
}

// PageResult 分页结果
type PageResult[T any] struct {
	List  []T
	Total int64
}

// Paginate 分页
func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 1000 {
			limit = 1000
		}
		return tx.Offset((page - 1) * limit).Limit(limit)
	}
}
