package errorcode

// ===== 集群相关（200xxx）=====
var (
	ErrorClusterNotFound   *Error
	ErrorClusterUnhealthy  *Error
	ErrorClusterForbidden  *Error
	ErrorClusterInitFailed *Error
	ErrorClusterDeleteFail *Error // 删除失败
	ErrorClusterUpdateFail *Error // 更新失败
	ErrorClusterQueryFail  *Error // 查询失败（列表/单查）
	ErrorClusterCreateFail *Error // 创建失败
)

func registerCluster() {
	ErrorClusterNotFound = NewError(200001, "集群名字不存在")
	ErrorClusterUnhealthy = NewError(200002, "集群不可用")
	ErrorClusterForbidden = NewError(200003, "没有访问该集群的权限")
	ErrorClusterInitFailed = NewError(200004, "K8s 集群初始化失败")

	ErrorClusterDeleteFail = NewError(200005, "删除集群失败")
	ErrorClusterUpdateFail = NewError(200006, "更新集群失败")
	ErrorClusterQueryFail = NewError(200007, "查询集群失败")
	ErrorClusterCreateFail = NewError(200008, "创建集群失败")
}
