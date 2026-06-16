package errorcode

// ===== Pod 相关（201xxx）=====
var (
	ErrorPodNotFound              *Error
	ErrorPodCreateFail            *Error
	ErrorPodDeleteFail            *Error
	ErrorPodUpdateFail            *Error
	ErrorPodQueryFail             *Error // 列表 / 单查失败
	ErrorPodLogFail               *Error // 获取日志失败
	ErrorPodContainerNotReady     *Error // 容器还未就绪
	ErrorK8sPodPatchFail          *Error // 更新镜像失败
	ErrorK8sGetPodMetrics         *Error // 获取 Pod metrics 失败
	ErrorMetricsServerUnavailable *Error // metrics-server 不可用
)

func registerPod() {
	ErrorPodNotFound = NewError(211001, "Pod 不存在")
	ErrorPodCreateFail = NewError(211002, "创建 Pod 失败")
	ErrorPodDeleteFail = NewError(211003, "删除 Pod 失败")
	ErrorPodUpdateFail = NewError(211004, "更新 Pod 失败")
	ErrorPodQueryFail = NewError(211005, "查询 Pod 失败")
	ErrorPodLogFail = NewError(211006, "获取 Pod 日志失败")
	ErrorPodContainerNotReady = NewError(211009, "容器还未就绪，请稍后再试")
	ErrorK8sPodPatchFail = NewError(211010, "更新 Pod 镜像失败")
	ErrorK8sGetPodMetrics = NewError(211007, "获取 Pod 资源使用情况失败")
	ErrorMetricsServerUnavailable = NewError(211008, "metrics-server 未安装或不可用")
}
