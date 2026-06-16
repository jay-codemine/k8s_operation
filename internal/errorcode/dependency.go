package errorcode

// ===== 依赖/三方（600xxx）=====
var (
	DBError             *Error
	CacheError          *Error
	UpstreamTimeout     *Error
	UpstreamBadResponse *Error
	RPCInvokeError      *Error
)

func registerDependency() {
	DBError = NewError(600001, "数据库操作失败")
	CacheError = NewError(600002, "缓存操作失败")
	UpstreamTimeout = NewError(600003, "上游接口超时")
	UpstreamBadResponse = NewError(600004, "上游响应异常")
	RPCInvokeError = NewError(600005, "RPC 调用失败")
}
