package errorcode

// ===== License 授权（110xxx）=====
var (
	LicenseRequired       *Error // 未激活授权（前端拦截此码跳转激活页）
	LicenseExpired        *Error // 授权已过期
	LicenseInvalid        *Error // License 无效（格式/签名错误）
	LicenseMachineErr     *Error // 机器码不匹配
	LicenseActivateFailed *Error // 激活失败（写盘等内部错误）
)

func registerLicense() {
	LicenseRequired = NewError(110001, "平台未授权，请先激活 License")
	LicenseExpired = NewError(110002, "License 已过期，请联系供应商续期")
	LicenseInvalid = NewError(110003, "License 无效")
	LicenseMachineErr = NewError(110004, "License 与本机机器码不匹配")
	LicenseActivateFailed = NewError(110005, "License 激活失败")
}
