package errorcode

// Register 统一注册所有错误码
func Register() {
	registerCommon()
	registerAuth()
	registerResource()
	registerRequest()
	registerQuota()
	registerDependency()
	registerBiz()
	registerToken()
	registerUser()
	registerCluster()   // k8s_error code
	registerK8sResource() // K8s 资源通用错误码
	registerPod()      // kube_pod error code
	register_k8s_Pod() // k8s_pod error code
	register_k8s_Deployment()
	registerService()
	registerDaemonSet()
	registerStatefulSet()
	registerJob()
	registerCronJob()
	registerIngress()
	registerPVC()
	// CICD
	register_cicd()
	register_pipeline()
	// RBAC
	registerRBAC()
	// 应用商城
	registerAppStore()
	// AI 助手
	registerAIAssistant()
	// CRD/CR 动态资源
	registerCRD()
	// HPA/VPA 弹性扩缩容
	register_k8s_Autoscaler()
	// License 授权
	registerLicense()
	// 后续可以继续扩展
}
