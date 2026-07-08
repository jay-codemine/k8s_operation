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
	ErrTokenExpired = NewError(310001, "令牌已过期")
	ErrTokenExpiredMaxRefresh = NewError(310002, "令牌已过最大刷新时间")
	ErrTokenMalformed = NewError(310003, "请求令牌格式有误")
	ErrTokenInvalid = NewError(310004, "请求令牌无效")
	ErrHeaderEmpty = NewError(310005, "需要认证才能访问！")
	ErrHeaderMalformed = NewError(310006, "请求头中 Authorization 格式有误")
}
