package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// EncodeMD5 MD5加密
// EncodeMD5 是一个将输入字符串进行MD5加密的函数
// 参数:
//
//	value - 需要进行MD5加密的字符串
//
// 返回值:
//
//	string - 返回经过MD5加密后的32位十六进制字符串
func EncodeMD5(value string) string {
	// 创建MD5哈希实例
	m := md5.New()
	// 将输入字符串的字节写入哈希实例
	m.Write([]byte(value))
	// 计算MD5哈希值并返回其十六进制字符串表示
	return hex.EncodeToString(m.Sum(nil))
}
