package initialize

// ---------------------------------------------
// 工具函数：白名单校验
// ---------------------------------------------
// isAllowedTable 表名白名单
func isAllowedTable(t string) bool {
	switch t {
	case "user", "projects": // 允许的表
		return true
	default:
		return false
	}
}

// isAllowedColumn 按表名定义允许的字段列表
func isAllowedColumn(table, col string) bool {
	allowed := map[string][]string{
		"user":     {"email", "username", "phone"},
		"projects": {"slug", "name"},
	}
	cols, ok := allowed[table]
	if !ok {
		return false
	}
	for _, c := range cols {
		if c == col {
			return true
		}
	}
	return false
}
