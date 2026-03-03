package initialize

import (
	"errors"
	"fmt"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/global"
	"strconv"
	"strings"
)

func SetupValidator() error {
	// 注册一个名为 "not_exists" 的自定义规则
	// 用途：验证请求数据在数据库中必须不存在，常用于唯一性校验
	// 使用方式：
	//   not_exists:user,email           → 校验 user 表中 email 字段不能重复
	//   not_exists:user,email,32        → 校验 user 表中 email 字段不能重复，但排除 id=32 的记录（用于更新时自己不和自己冲突）
	govalidator.AddCustomRule("not_exists", func(field, rule, message string, value interface{}) error {
		// field   → 字段名，例如 "username"
		// rule    → 完整规则字符串，例如 "not_exists:user,username,except_id=1"
		// message → 自定义错误提示，如果 messages 配置了则传入
		// value   → 字段值，例如用户传的 "joker"

		// 0) 校验规则前缀，保证规则合法
		if !strings.HasPrefix(rule, "not_exists:") {
			return errors.New("not_exists 规则格式错误")
		}

		// 1) 解析参数：表名、字段名、可选 except_id
		raw := strings.TrimSpace(strings.TrimPrefix(rule, "not_exists:"))
		parts := strings.Split(raw, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i]) // 去掉每个参数首尾空格
		}
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("not_exists 规则参数错误，期望 not_exists:表名,字段名[,except_id=ID]")
		}
		tableName := parts[0] // 表名
		dbField := parts[1]   // 字段名

		// 2) 白名单校验（防止 SQL 注入，只允许预定义的表和字段）
		if !isAllowedTable(tableName) || !isAllowedColumn(tableName, dbField) {
			return errors.New("非法的表名或字段名")
		}

		// 3) 获取字段值，并处理“空值短路”逻辑
		val := strings.TrimSpace(fmt.Sprint(value))
		if val == "" {
			// 如果值为空，不在这里报错，交给 required 规则去处理
			return nil
		}

		// 4) except_id 参数解析，支持 "123" 或 "except_id=123" 两种形式
		var exceptIDStr string
		if len(parts) >= 3 && parts[2] != "" {
			if strings.Contains(parts[2], "=") {
				kv := strings.SplitN(parts[2], "=", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "except_id" {
					exceptIDStr = strings.TrimSpace(kv[1]) // 取出等号右边的值
				}
			} else {
				exceptIDStr = parts[2] // 简写模式，直接就是 ID
			}
		}

		// 尝试把 except_id 转成 int64，避免脏值
		var exceptID any
		if exceptIDStr != "" {
			if id64, err := strconv.ParseInt(exceptIDStr, 10, 64); err == nil {
				exceptID = id64
			} else {
				return errors.New("except_id 必须为数字")
			}
		}

		// 5) 构造查询语句
		q := global.DB.Table(tableName).Where(fmt.Sprintf("%s = ?", dbField), val)
		// 如果有逻辑删除，可以加上过滤条件：
		// q = q.Where("is_del = 0")

		// 如果设置了 except_id，排除掉这条记录（用于更新时忽略自己）
		if exceptID != nil {
			q = q.Where("id <> ?", exceptID)
		}

		// 6) 执行查询并判断结果
		var count int64
		if err := q.Count(&count).Error; err != nil {
			// 如果数据库查询失败，返回通用错误提示
			return errors.New("系统繁忙，请稍后再试")
		}
		if count > 0 {
			// 如果已存在记录，返回错误
			if message != "" {
				// 使用用户自定义的错误提示
				return errors.New(message)
			}
			// 默认提示：字段已存在
			return fmt.Errorf("用户 %s 已注册", val)
		}
		// 校验通过
		return nil
	})

	return nil
}
