package utils

import (
	"k8soperation/global"
	"time"
)

// TimenowInTimezone 返回指定时区的当前时间
// 函数会从全局配置中获取时区设置，并返回该时区的当前时间
// 如果配置为空或加载失败，默认使用 Asia/Shanghai (UTC+8)
// 返回值: time.Time - 指定时区的当前时间
func TimenowInTimezone() time.Time {
	tz := ""
	if global.AppSetting != nil {
		tz = global.AppSetting.TIMEZONE
	}
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		// 兜底：硬编码 UTC+8，避免 panic
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Now().In(loc)
}
