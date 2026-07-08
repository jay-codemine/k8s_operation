package models

import "gorm.io/gorm"

// 软删除
func ScopeNotDeleted() func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("is_del = 0")
	}
}

// 模糊搜索（通用 name）
func ScopeLikeName(field, value string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if value == "" {
			return tx
		}
		return tx.Where(field+" LIKE ?", "%"+value+"%")
	}
}

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

type PageResult[T any] struct {
	List  []T
	Total int64
}

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
		} // 保护
		return tx.Offset((page - 1) * limit).Limit(limit)
	}
}
