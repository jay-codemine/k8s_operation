package utils

import (
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"sort"
	"time"
)

func TryEventsV1First() bool {
	// 返回全局变量 global.SupportsEventsV1 的值
	// 这个值用于判断系统是否支持 Events V1 版本
	return global.SupportsEventsV1
}

func ApplySinceAndSort(list []models.EventItem, sinceSeconds int64) []models.EventItem {
	// 先做时间过滤（复用底层数组，减少分配）
	// 如果sinceSeconds大于0，则执行过滤操作
	if sinceSeconds > 0 {
		// 计算截止时间，当前时间减去sinceSeconds秒
		cutoff := time.Now().Add(-time.Duration(sinceSeconds) * time.Second)
		n := 0 // 用于记录符合条件的元素数量
		// 遍历列表，过滤掉EventTime为零或早于截止时间的元素
		for _, it := range list {
			if it.EventTime.IsZero() || it.EventTime.After(cutoff) {
				list[n] = it // 将符合条件的元素移到列表前面
				n++          // 增加符合条件的元素计数
			}
		}
		// 通过切片操作截断列表，只保留符合条件的元素
		list = list[:n]
	}

	// 再做排序：按时间降序；零时间排在最后
	sort.SliceStable(list, func(i, j int) bool {
		ti, tj := list[i].EventTime, list[j].EventTime
		if ti.IsZero() && tj.IsZero() {
			return false
		} // 两个都零：保持原序
		if ti.IsZero() {
			return false
		} // i 为零：放后面
		if tj.IsZero() {
			return true
		} // j 为零：i 在前
		return ti.After(tj) // 都非零：时间新在前
	})

	return list
}
