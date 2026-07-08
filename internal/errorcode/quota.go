package errorcode

// ===== 限流/配额（500xxx）=====
var (
	RateLimited   *Error
	QuotaExceeded *Error
)

func registerQuota() {
	RateLimited = NewError(500001, "已触发限流/频控")
	QuotaExceeded = NewError(500002, "配额已用尽")
}
