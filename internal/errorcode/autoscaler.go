package errorcode

var (
	// HPA
	ErrorK8sHPAListFail           *Error
	ErrorK8sHPADetailFail         *Error
	ErrorK8sHPACreateFail         *Error
	ErrorK8sHPAUpdateFail         *Error
	ErrorK8sHPADeleteFail         *Error
	ErrorK8sHPAScaleFail          *Error
	ErrorK8sHPACreateFromYamlFail *Error

	// VPA
	ErrorK8sVPANotInstalled       *Error
	ErrorK8sVPAListFail           *Error
	ErrorK8sVPADetailFail         *Error
	ErrorK8sVPACreateFail         *Error
	ErrorK8sVPAUpdateFail         *Error
	ErrorK8sVPADeleteFail         *Error
	ErrorK8sVPACreateFromYamlFail *Error
)

func register_k8s_Autoscaler() {
	// HPA: 5005xx
	ErrorK8sHPAListFail = NewError(500551, "获取K8s HPA列表失败")
	ErrorK8sHPADetailFail = NewError(500552, "获取K8s HPA详情失败")
	ErrorK8sHPACreateFail = NewError(500553, "创建K8s HPA失败")
	ErrorK8sHPAUpdateFail = NewError(500554, "更新K8s HPA失败")
	ErrorK8sHPADeleteFail = NewError(500555, "删除K8s HPA失败")
	ErrorK8sHPAScaleFail = NewError(500556, "修改K8s HPA副本数失败")
	ErrorK8sHPACreateFromYamlFail = NewError(500557, "通过YAML创建K8s HPA失败")

	// VPA: 5006xx
	ErrorK8sVPANotInstalled = NewError(500561, "当前集群未安装 VPA Operator，请先部署 vertical-pod-autoscaler")
	ErrorK8sVPAListFail = NewError(500562, "获取K8s VPA列表失败")
	ErrorK8sVPADetailFail = NewError(500563, "获取K8s VPA详情失败")
	ErrorK8sVPACreateFail = NewError(500564, "创建K8s VPA失败")
	ErrorK8sVPAUpdateFail = NewError(500565, "更新K8s VPA失败")
	ErrorK8sVPADeleteFail = NewError(500566, "删除K8s VPA失败")
	ErrorK8sVPACreateFromYamlFail = NewError(500567, "通过YAML创建K8s VPA失败")
}
