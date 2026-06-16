package errorcode

// ===== 业务域（700xxx+）=====
var (
	BizRuleViolated *Error
)

func registerBiz() {
	BizRuleViolated = NewError(700001, "触发业务规则校验")
}
