package errorcode

// ===== DaemonSet 相关（202xxx）=====
var (
	ErrorDaemonSetNotFound   *Error
	ErrorDaemonSetCreateFail *Error
	ErrorDaemonSetDeleteFail *Error
	ErrorDaemonSetUpdateFail *Error
	ErrorDaemonSetQueryFail  *Error // 列表 / 单查失败
)

func registerDaemonSet() {
	ErrorDaemonSetNotFound = NewError(202001, "DaemonSet 不存在")
	ErrorDaemonSetCreateFail = NewError(202002, "创建 DaemonSet 失败")
	ErrorDaemonSetDeleteFail = NewError(202003, "删除 DaemonSet 失败")
	ErrorDaemonSetUpdateFail = NewError(202004, "更新 DaemonSet 失败")
	ErrorDaemonSetQueryFail = NewError(202005, "查询 DaemonSet 失败")
}
