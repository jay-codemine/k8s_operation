package dataselect

import (
	"sort"
	"strings"
	"time"
)

// DataCell 定义接口
// GetCreation 获取创建时间
// GetName 获取对象名称
type DataCell interface {
	GetCreation() time.Time
	GetName() string
}

// DataSelector 用于封装排除、过滤以及分页的数据类型
type DataSelector struct {
	GenericDataList []DataCell
	DataSelect      *DataSelectQuery
}

// DataSelectQuery 定义过滤和分页的结构体
// 过滤使用的 Name
// 分页使用的 Limit 和 Page
// Limit 是单页数据条数
// Page 是第几页
type DataSelectQuery struct {
	Filter   *FilterQuery
	Paginate *PaginateQuery
}

type FilterQuery struct {
	Name string
}

type PaginateQuery struct {
	Limit int
	Page  int
}

// 实现自定义结构的排序
// 需要重写 Len(),Swap(),Less()

// Len 用户获取数据的长度
func (d *DataSelector) Len() int {
	return len(d.GenericDataList)
}

// Swap 用于数据比较大小之后的位置变更
// i,j 数组下标
func (d *DataSelector) Swap(i, j int) {
	d.GenericDataList[i], d.GenericDataList[j] = d.GenericDataList[j], d.GenericDataList[i]
}

// Less 用于比较大小
// i,j 数组下标
func (d *DataSelector) Less(i, j int) bool {
	a := d.GenericDataList[i].GetCreation()
	b := d.GenericDataList[j].GetCreation()
	return b.Before(a)
}

// Sort 用于触发排序
func (d *DataSelector) Sort() *DataSelector {
	sort.Sort(d)
	return d
}

// Filter 用于过滤数据
// 比较数据的 Name 属性，若包含，则返回
func (d *DataSelector) Filter() *DataSelector {
	if d.DataSelect == nil || d.DataSelect.Filter == nil {
		return d
	}

	// 取过滤关键字，去掉首尾空格并转小写
	kw := strings.ToLower(strings.TrimSpace(d.DataSelect.Filter.Name))
	if kw == "" {
		// 关键字为空，直接返回全部数据
		return d
	}

	var filtered []DataCell
	for _, item := range d.GenericDataList {
		// 名称同样转小写，保证大小写不敏感
		name := strings.ToLower(item.GetName())
		// 检查名称中是否包含关键字
		if strings.Contains(name, kw) {
			// 如果包含关键字，则将该项添加到过滤后的列表中
			filtered = append(filtered, item)
		}
	}

	// 将过滤后的列表赋值给通用数据列表属性
	d.GenericDataList = filtered
	// 返回处理后的数据对象
	return d
}

// Paginate 用于数组分页，根据 Limit 和 Page 传参
func (d *DataSelector) Paginate() *DataSelector {
	// ===== 安全校验 =====
	if d.DataSelect == nil || d.DataSelect.Paginate == nil {
		return d
	}

	limit := d.DataSelect.Paginate.Limit
	page := d.DataSelect.Paginate.Page

	if limit <= 0 || page <= 0 {
		return d
	}

	total := len(d.GenericDataList)
	startIndex := limit * (page - 1)

	// 关键修复：起始索引越界，直接返回空列表
	if startIndex >= total {
		d.GenericDataList = []DataCell{}
		return d
	}

	endIndex := startIndex + limit
	if endIndex > total {
		endIndex = total
	}

	d.GenericDataList = d.GenericDataList[startIndex:endIndex]
	return d
}

// TotalCount 返回过滤后的总数（分页前的数量）
func (d *DataSelector) TotalCount() int {
	// 注意：这里直接返回当前列表长度即可，
	// 一般在调用方应该是先 Filter() 再调 TotalCount()
	return len(d.GenericDataList)
}
