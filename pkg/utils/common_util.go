package utils

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func NormalizeNamespace(ns string) string {
	// 检查输入的命名空间字符串是否为空
	if ns == "" {
		// 如果为空，则返回所有命名空间的字符串
		return metav1.NamespaceAll
	}
	// 如果不为空，则直接返回输入的命名空间字符串
	return ns
}

func ClampLimit(n int64) int64 {
	// 检查输入值n的范围，并返回相应的结果
	if n <= 0 {
		// 如果n小于等于0，返回固定值50
		return 50
	}
	if n > 500 {
		// 如果n大于500，返回固定值500
		return 500
	}
	// 如果n在1到500之间，直接返回n的值
	return n
}

func ValueOrZero(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
