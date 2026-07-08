package errorcode

// ===== Ingress 相关（207xxx）=====
var (
	ErrorIngressNotFound   *Error
	ErrorIngressCreateFail *Error
	ErrorIngressDeleteFail *Error
	ErrorIngressUpdateFail *Error
	ErrorIngressQueryFail  *Error // 列表 / 单查失败
)

func registerIngress() {
	ErrorIngressNotFound = NewError(207001, "Ingress 不存在")
	ErrorIngressCreateFail = NewError(207002, "创建 Ingress 失败")
	ErrorIngressDeleteFail = NewError(207003, "删除 Ingress 失败")
	ErrorIngressUpdateFail = NewError(207004, "更新 Ingress 失败")
	ErrorIngressQueryFail = NewError(207005, "查询 Ingress 失败")
}
