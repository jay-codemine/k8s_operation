package utils

import "github.com/gin-gonic/gin"

// 保证返回 [] 而不是 null
func EnsureArray(v interface{}) interface{} {
	if v == nil {
		return []interface{}{}
	}
	return v
}

// 保证返回 {} 而不是 null
func EnsureObject(v interface{}) interface{} {
	if v == nil {
		return gin.H{}
	}
	return v
}
