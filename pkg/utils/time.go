package utils

import "time"

// NowUnix 返回当前秒级时间戳
func NowUnix() uint32 {
	return uint32(time.Now().Unix())
}
