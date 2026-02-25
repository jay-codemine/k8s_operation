package errorcode

// ===== 通用（100xxx）=====
var (
	Success       *Error
	ServerError   *Error
	InvalidParams *Error
	NotFound      *Error

	UnauthorizedAuthNotExist  *Error
	UnauthorizedTokenError    *Error
	UnauthorizedTokenTimeout  *Error
	UnauthorizedTokenGenerate *Error
	TooManyRequests           *Error
	UserNotLogin              *Error
)

func registerCommon() {
	Success = NewError(0, "成功")
	ServerError = NewError(100001, "服务内部错误")
	InvalidParams = NewError(100002, "输入参数有错误")
	NotFound = NewError(100003, "找不到")

	UnauthorizedAuthNotExist = NewError(100004, "鉴权失败，找不到对应的 AppKey/AppSecret")
	UnauthorizedTokenError = NewError(100005, "鉴权失败，Token 错误")
	UnauthorizedTokenTimeout = NewError(100006, "鉴权失败，Token 超时")
	UnauthorizedTokenGenerate = NewError(100007, "鉴权失败，Token 生成失败")
	TooManyRequests = NewError(100008, "请求过多")
	UserNotLogin = NewError(100009, "用户未登录")
}
