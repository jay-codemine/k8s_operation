package errorcode

// ===== 请求质量（400xxx）=====
var (
	ValidationFailed     *Error
	UnsupportedMediaType *Error
	PayloadTooLarge      *Error
	UnprocessableEntity  *Error
	BadRequestSyntax     *Error
	ErrorAuthLoginFail   *Error
)

func registerRequest() {
	ValidationFailed = NewError(400001, "参数校验失败")
	UnsupportedMediaType = NewError(400002, "不支持的内容类型")
	PayloadTooLarge = NewError(400003, "请求体过大")
	UnprocessableEntity = NewError(400004, "请求语义错误，无法处理")
	BadRequestSyntax = NewError(400005, "请求格式错误")
	ErrorAuthLoginFail = NewError(400001, "登录失败,用户名或密码错误")
}
