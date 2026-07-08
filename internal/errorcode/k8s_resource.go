package errorcode

// ===== K8s 资源通用操作错误码（201xxx）=====
var (
	ErrorK8sResourceListFail   *Error
	ErrorK8sResourceGetFail    *Error
	ErrorK8sResourceCreateFail *Error
	ErrorK8sResourceUpdateFail *Error
	ErrorK8sResourceDeleteFail *Error
)

func registerK8sResource() {
	ErrorK8sResourceListFail = NewError(201001, "获取资源列表失败")
	ErrorK8sResourceGetFail = NewError(201002, "获取资源详情失败")
	ErrorK8sResourceCreateFail = NewError(201003, "创建资源失败")
	ErrorK8sResourceUpdateFail = NewError(201004, "更新资源失败")
	ErrorK8sResourceDeleteFail = NewError(201005, "删除资源失败")
}
