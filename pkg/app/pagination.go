package app

import (
	"github.com/gin-gonic/gin" // 引入Gin框架，用于处理HTTP请求
	"k8soperation/global"      // 引入全局变量配置
	"k8soperation/pkg/utils"
)

// GetPage 从请求上下文中获取页码参数
// 如果页码无效(<=0)，则返回默认值1
func GetPage(ctx *gin.Context) int {
	// 从查询参数中获取page，并将其转换为整数
	page := utils.StrTo(ctx.Query("page")).MustInt()
	// 检查页码是否有效，无效则返回1
	if page <= 0 {
		return 1
	}

	return page
}

// GetPageSize 从请求上下文中获取每页大小参数
// 如果页大小无效(<=0)，则返回默认值
// 如果页大小超过最大限制，则返回最大限制值
func GetPageSize(ctx *gin.Context) int {
	// 从查询参数中获取 page_size，并将其转换为整数
	page_size := utils.StrTo(ctx.Query("page_size")).MustInt()
	// 检查页大小是否有效，无效则返回默认值
	if page_size <= 0 {
		return global.DefaultPageSize
	}
	// 检查是否超过最大页大小限制
	// 检查输入的页面大小是否超过全局设置的最大页面大小
	// 如果超过最大值，则返回最大页面大小；否则返回原始输入的页面大小
	if page_size > global.MaxPageSize {
		return global.MaxPageSize
	}

	// 返回页面大小
	return page_size
}

// GetPageOffSet 根据页码和页大小计算偏移量
// 偏移量用于数据库查询时的分页计算
func GetPageOffSet(page, pageSize int) int {
	result := 0
	// 计算偏移量：(当前页码-1) * 每页大小
	if page > 0 {
		// 计算分页偏移量，公式为：(当前页码 - 1) * 每页大小
		result = (page - 1) * pageSize

		// 返回计算得到的偏移量
		return result
	}
	// 返回默认的result值（如果前面的条件不满足）
	return result
}
