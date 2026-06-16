package utils

import (
	"k8soperation/global"
	"time"
)

// TimenowInTimezone 获取当前时间，支持时区
// TimenowInTimezone 返回指定时区的当前时间
// 函数会从全局配置中获取时区设置，并返回该时区的当前时间
// 返回值: time.Time - 指定时区的当前时间
func TimenowInTimezone() time.Time {
	// 从全局配置中加载时区设置
	// global.AppSetting.TIMEZONE 应该包含时区名称，如 "Asia/Shanghai"
	chinaTimezone, _ := time.LoadLocation(global.AppSetting.TIMEZONE)
	// 获取当前时间并转换为指定时区的时间
	return time.Now().In(chinaTimezone)
}
