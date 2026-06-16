package errorcode

// 说明：将变量声明为包级全局，实际赋值在 registerPod() 里完成，
// 便于控制 AllowOverride 等策略，也避免 import 时的初始化顺序问题。
var (
	ErrorK8sPodUpdateFail         *Error
	ErrorK8sPodDeleteFail         *Error
	ErrorK8sPodListFail           *Error
	ErrorK8sPodDetailFail         *Error
	ErrorK8sGetContainerName      *Error
	ErrorK8sGetContainerImage     *Error
	ErrorK8sGetInitContainerName  *Error
	ErrorK8sGetInitContainerImage *Error
	ErrorK8sGetContainerLog       *Error
)

// 内部注册函数（由 Register() 调用）
func register_k8s_Pod() {
	// 如果你有“是否允许覆盖”的开关，这里统一由 NewError/内部 register 方法处理
	ErrorK8sPodUpdateFail = NewError(500011, "更新K8s Pod失败")
	ErrorK8sPodDeleteFail = NewError(500012, "删除K8s Pod失败")
	ErrorK8sPodListFail = NewError(500013, "获取K8s Pod列表失败")
	ErrorK8sPodDetailFail = NewError(500014, "获取K8s Pod详情失败")
	ErrorK8sGetContainerName = NewError(500015, "获取K8s Pod容器名失败")
	ErrorK8sGetContainerImage = NewError(500016, "获取K8s Pod容器镜像失败")
	ErrorK8sGetInitContainerName = NewError(500017, "获取K8s Pod Init容器名失败")
	ErrorK8sGetInitContainerImage = NewError(500018, "获取K8s Pod Init容器镜像失败")
	ErrorK8sGetContainerLog = NewError(500019, "获取K8s Pod 容器日志失败")
}
