package errorcode

// ===== Service 相关（206xxx）=====
var (
	ErrorServiceNotFound   *Error
	ErrorServiceCreateFail *Error
	ErrorServiceDeleteFail *Error
	ErrorServiceUpdateFail *Error
	ErrorServiceQueryFail  *Error // 列表 / 单查失败
)

func registerService() {
	ErrorServiceNotFound = NewError(206001, "Service 不存在")
	ErrorServiceCreateFail = NewError(206002, "创建 Service 失败")
	ErrorServiceDeleteFail = NewError(206003, "删除 Service 失败")
	ErrorServiceUpdateFail = NewError(206004, "更新 Service 失败")
	ErrorServiceQueryFail = NewError(206005, "查询 Service 失败")
}
