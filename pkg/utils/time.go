package utils

import "time"

// NowUnix 返回当前秒级时间戳
func NowUnix() uint64 {
	return uint64(time.Now().Unix())
}

// NowUnix32 返回当前秒级时间戳(uint32)
func NowUnix32() uint32 {
	return uint32(time.Now().Unix())
}
