package errorcode

// ===== 资源/对象（300xxx）=====
var (
	ResourceNotFound      *Error
	ResourceConflict      *Error
	ResourceStatusInvalid *Error
)

func registerResource() {
	ResourceNotFound = NewError(300001, "资源不存在")
	ResourceConflict = NewError(300002, "资源冲突（重复/唯一键冲突）")
	ResourceStatusInvalid = NewError(300003, "资源当前状态不允许此操作")
}
