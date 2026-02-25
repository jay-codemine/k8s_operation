package errorcode

import (
	"fmt"
	"net/http"
)

// Error 自定义错误结构体，用于封装错误信息
type Error struct {
	code    int      // 错误码，用于标识错误类型
	msg     string   // 错误信息，用于描述错误内容
	details []string // 错误详情，用于存储更详细的错误信息列表
}

// 错误码注册表
var codes = map[int]string{}

// 由外部设置的“是否允许覆盖”开关（避免包初始化时依赖 global）
var allowOverride bool

// 在 main/init 流程里由外部调用，设置开关
func SetAllowOverride(b bool) { allowOverride = b }

// NewError 注册并返回错误对象
func NewError(code int, msg string) *Error {
	if !allowOverride {
		if _, ok := codes[code]; ok {
			panic(fmt.Sprintf("错误码%d已经存在,请更换一个", code))
		}
	}
	codes[code] = msg
	return &Error{code: code, msg: msg}
}

// Error 实现了 Go 的 error 接口
// 调用 fmt.Println(err) 时会触发这个方法
// 返回格式: "错误码: xxx，错误信息: xxx"
func (e *Error) Error() string {
	if len(e.details) > 0 {
		return fmt.Sprintf("错误码: %d，错误信息: %s，详情: %v", e.Code(), e.Msg(), e.details)
	}
	return fmt.Sprintf("错误码: %d，错误信息: %s", e.Code(), e.Msg())
}

// Code 返回业务错误码
func (e *Error) Code() int {
	return e.code
}

// Msg 返回错误信息（模板字符串或固定说明）
// 一般用于固定提示，比如 "入参错误"
func (e *Error) Msg() string {
	return e.msg
}

// Msgf 返回格式化后的错误信息
func (e *Error) Msgf(args ...interface{}) string {
	if len(args) == 0 {
		return e.msg
	}
	return fmt.Sprintf(e.msg, args...)
}

// Details 返回错误详情列表
// - 详情可以包含更具体的上下文，比如具体哪个字段缺失
func (e *Error) Details() []string {
	return e.details
}

// WithDetails 生成一个带有额外详情的新错误对象
//   - 不会修改原始错误（避免污染全局错误常量）
//   - 每次返回一个新的 Error 拷贝，保证线程安全
//   - 示例:
//     err := InvalidParams.withDetails("字段 username 缺失", "请求ID=12345")
//     err.Details() => ["字段 username 缺失", "请求ID=12345"]
func (e *Error) WithDetails(details ...string) *Error {
	newError := *e                                          // 拷贝原对象，避免修改原始 Error
	newError.details = []string{}                           // 初始化详情字段
	newError.details = append(newError.details, details...) // 追加新详情
	return &newError
}

// StatusCode 将业务错误码映射到 HTTP 状态码
// - 用于在 API 返回时，自动转换为合适的 HTTP 状态码
// - 例如: InvalidParams => 400, NotFound => 404, Success => 200
// StatusCode 方法返回与错误代码对应的 HTTP 状态码
func (e *Error) StatusCode() int {
	switch e.Code() {
	case Success.Code():
		return http.StatusOK // 200 成功
	case ServerError.Code():
		return http.StatusInternalServerError // 500 内部错误
	case InvalidParams.Code():
		return http.StatusBadRequest // 400 参数错误
	case NotFound.Code():
		return http.StatusNotFound // 404 资源不存在
	case UnauthorizedAuthNotExist.Code(),
		UnauthorizedTokenError.Code(),
		UnauthorizedTokenGenerate.Code(),
		UnauthorizedTokenTimeout.Code(),
		UserNotLogin.Code():
		return http.StatusUnauthorized // 401 鉴权失败/未认证/未登录
	case TooManyRequests.Code():
		return http.StatusTooManyRequests // 429 请求过多
	default:
		// 2xxxxx 系列为业务错误，返回 400
		if e.Code() >= 200000 && e.Code() < 300000 {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError // 默认返回 500
	}
}

//var httpMap = map[int]int{
//	Success.code:         http.StatusOK,
//	ServerError.code:     http.StatusInternalServerError,
//	InvalidParams.code:   http.StatusBadRequest,
//	NotFound.code:        http.StatusNotFound,
//	UserNotLogin.code:    http.StatusUnauthorized,
//	TooManyRequests.code: http.StatusTooManyRequests,
//	// ……按需补充
//}
//
//func (e *Error) StatusCode() int {
//	if sc, ok := httpMap[e.code]; ok {
//		return sc
//	}
//	return http.StatusInternalServerError
//}
