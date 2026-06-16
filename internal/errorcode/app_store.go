package errorcode

// ===== 应用商城（109xxx）=====
var (
	AppStoreListFailed   *Error
	AppStoreNotFound     *Error
	AppStoreCreateFailed *Error
	AppStoreUpdateFailed *Error
	AppStoreDeleteFailed *Error
	AppStoreNameExists   *Error

	// 安装相关
	AppStoreInstallFailed    *Error
	AppStoreAlreadyInstalled *Error
	AppStoreUninstallFailed  *Error
	AppStoreInstallNotFound  *Error
	AppStoreClusterNotFound  *Error
)

func registerAppStore() {
	AppStoreListFailed = NewError(109001, "获取应用列表失败")
	AppStoreNotFound = NewError(109002, "应用不存在")
	AppStoreCreateFailed = NewError(109003, "创建应用失败")
	AppStoreUpdateFailed = NewError(109004, "更新应用失败")
	AppStoreDeleteFailed = NewError(109005, "删除应用失败")
	AppStoreNameExists = NewError(109006, "应用名称已存在")

	AppStoreInstallFailed = NewError(109010, "安装应用失败")
	AppStoreAlreadyInstalled = NewError(109011, "应用已在该集群命名空间安装")
	AppStoreUninstallFailed = NewError(109012, "卸载应用失败")
	AppStoreInstallNotFound = NewError(109013, "安装记录不存在")
	AppStoreClusterNotFound = NewError(109014, "目标集群不存在或不可用")
}
