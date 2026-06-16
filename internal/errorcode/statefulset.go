package errorcode

// ===== StatefulSet 相关（203xxx）=====
var (
	ErrorStatefulSetNotFound   *Error
	ErrorStatefulSetCreateFail *Error
	ErrorStatefulSetDeleteFail *Error
	ErrorStatefulSetUpdateFail *Error
	ErrorStatefulSetQueryFail  *Error // 列表 / 单查失败
)

func registerStatefulSet() {
	ErrorStatefulSetNotFound = NewError(203001, "StatefulSet 不存在")
	ErrorStatefulSetCreateFail = NewError(203002, "创建 StatefulSet 失败")
	ErrorStatefulSetDeleteFail = NewError(203003, "删除 StatefulSet 失败")
	ErrorStatefulSetUpdateFail = NewError(203004, "更新 StatefulSet 失败")
	ErrorStatefulSetQueryFail = NewError(203005, "查询 StatefulSet 失败")
}
