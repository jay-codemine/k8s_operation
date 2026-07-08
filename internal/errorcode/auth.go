package errorcode

// ===== 鉴权/账号（200xxx）=====
var (
	Unauthorized         *Error
	PermissionDenied     *Error
	UserFrozenOrDisabled *Error
)

func registerAuth() {
	Unauthorized = NewError(200001, "未认证或认证失效")
	PermissionDenied = NewError(200002, "无权限执行该操作")
	UserFrozenOrDisabled = NewError(200003, "用户状态异常")
}
