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
	registerCluster()  // k8s_error code
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
	// 后续可以继续扩展
}
