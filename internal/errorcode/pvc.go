package errorcode

// ===== PVC 相关（208xxx）=====
var (
	ErrorPVCNotFound   *Error
	ErrorPVCCreateFail *Error
	ErrorPVCDeleteFail *Error
	ErrorPVCUpdateFail *Error
	ErrorPVCQueryFail  *Error // 列表 / 单查失败
)

func registerPVC() {
	ErrorPVCNotFound = NewError(208001, "PVC 不存在")
	ErrorPVCCreateFail = NewError(208002, "创建 PVC 失败")
	ErrorPVCDeleteFail = NewError(208003, "删除 PVC 失败")
	ErrorPVCUpdateFail = NewError(208004, "更新 PVC 失败")
	ErrorPVCQueryFail = NewError(208005, "查询 PVC 失败")
}
