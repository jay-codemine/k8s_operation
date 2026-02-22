package errorcode

// ===== Job 相关（204xxx）=====
var (
	ErrorJobNotFound   *Error
	ErrorJobCreateFail *Error
	ErrorJobDeleteFail *Error
	ErrorJobUpdateFail *Error
	ErrorJobQueryFail  *Error // 列表 / 单查失败
)

func registerJob() {
	ErrorJobNotFound = NewError(204001, "Job 不存在")
	ErrorJobCreateFail = NewError(204002, "创建 Job 失败")
	ErrorJobDeleteFail = NewError(204003, "删除 Job 失败")
	ErrorJobUpdateFail = NewError(204004, "更新 Job 失败")
	ErrorJobQueryFail = NewError(204005, "查询 Job 失败")
}
