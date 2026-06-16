package errorcode

// ===== CRD/CR 动态资源管理错误码（207xxx）=====
var (
	ErrorCRDListFail          *Error
	ErrorCRDGetFail           *Error
	ErrorCRDDeleteFail        *Error
	ErrorCRDDeleteProtected   *Error
	ErrorCRListFail           *Error
	ErrorCRGetFail            *Error
	ErrorCRCreateFail         *Error
	ErrorCRUpdateFail         *Error
	ErrorCRDeleteFail         *Error
	ErrorCRDeleteProtected    *Error
	ErrorCRDryRunFail         *Error
	ErrorCRYamlParseFail      *Error
	ErrorCRGVRInvalid         *Error
)

func registerCRD() {
	ErrorCRDListFail = NewError(215001, "获取 CRD 列表失败")
	ErrorCRDGetFail = NewError(215002, "获取 CRD 详情失败")
	ErrorCRDDeleteFail = NewError(215003, "删除 CRD 失败")
	ErrorCRDDeleteProtected = NewError(215004, "CRD 受删除保护，无法删除")
	ErrorCRListFail = NewError(215101, "获取 CR 实例列表失败")
	ErrorCRGetFail = NewError(215102, "获取 CR 实例详情失败")
	ErrorCRCreateFail = NewError(215103, "创建 CR 实例失败")
	ErrorCRUpdateFail = NewError(215104, "更新 CR 实例失败")
	ErrorCRDeleteFail = NewError(215105, "删除 CR 实例失败")
	ErrorCRDeleteProtected = NewError(215106, "CR 实例受删除保护，无法删除")
	ErrorCRDryRunFail = NewError(215201, "DryRun 校验失败")
	ErrorCRYamlParseFail = NewError(215202, "YAML 解析失败")
	ErrorCRGVRInvalid = NewError(215203, "资源类型参数无效(group/version/resource)")
}
