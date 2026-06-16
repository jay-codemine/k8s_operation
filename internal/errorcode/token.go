package errorcode

var (
	ErrTokenExpired           *Error
	ErrTokenExpiredMaxRefresh *Error
	ErrTokenMalformed         *Error
	ErrTokenInvalid           *Error
	ErrHeaderEmpty            *Error
	ErrHeaderMalformed        *Error
)

func registerToken() {
	ErrTokenExpired = NewError(300001, "令牌已过期")
	ErrTokenExpiredMaxRefresh = NewError(300002, "令牌已过最大刷新时间")
	ErrTokenMalformed = NewError(300003, "请求令牌格式有误")
	ErrTokenInvalid = NewError(300004, "请求令牌无效")
	ErrHeaderEmpty = NewError(300005, "需要认证才能访问！")
	ErrHeaderMalformed = NewError(300006, "请求头中 Authorization 格式有误")
}
