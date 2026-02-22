package errorcode

var (
	ErrorUserCreateFail       *Error
	ErrorUserDeleteFail       *Error
	ErrorUserUpdateFail       *Error
	ErrorUserListFail         *Error
	ErrorUserNotFound         *Error
	ErrorUserPasswordNotMatch *Error
)

func registerUser() {
	ErrorUserCreateFail = NewError(200001, "创建用户失败")
	ErrorUserDeleteFail = NewError(200002, "删除用户失败,用户不存在")
	ErrorUserUpdateFail = NewError(200003, "更新用户失败,用户不存在")
	ErrorUserListFail = NewError(200004, "列出用户失败")

	ErrorUserNotFound = NewError(200005, "用户名不存在")
	ErrorUserPasswordNotMatch = NewError(200006, "两次密码不一致")
}
