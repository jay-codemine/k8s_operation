package utils

import "strconv"

// StrTo 是一个自定义类型，基于 string 类型
// 它提供了一系列方便的方法来进行字符串到各种数值类型的转换
type StrTo string

// 类型转换相关方法
// 字符串
// String 方法将 StrTo 类型转换为 string 类型
func (s StrTo) String() string {
	return string(s)
}

// Int 方法尝试将 StrTo 转换为 int 类型
// 如果转换失败，返回错误
func (s StrTo) Int() (int, error) {
	i, err := strconv.Atoi(s.String())
	return i, err
}

// MustInt 方法将 StrTo 转换为 int 类型
// 如果转换失败，忽略错误并返回 0
func (s StrTo) MustInt() int {
	i, _ := s.Int()
	return i
}

// UInt32 方法尝试将 StrTo 转换为 uint32 类型
// 如果转换失败，返回错误
func (s StrTo) UInt32() (uint32, error) {
	i, err := strconv.Atoi(s.String())
	return uint32(i), err
}

// MustUint32 方法将 StrTo 转换为 uint32 类型
// 如果转换失败，忽略错误并返回 0
func (s StrTo) MustUint32() uint32 {
	i, _ := s.UInt32()
	return i
}
