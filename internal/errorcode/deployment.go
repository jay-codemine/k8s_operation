package errorcode

var (
	ErrorK8sDeploymentCreateFail   *Error
	ErrorK8sDeploymentDeleteFail   *Error
	ErrorK8sDeploymentListFail     *Error
	ErrorK8sDeploymentDetailFail   *Error
	ErrorK8sDeploymentUpdateFail   *Error
	ErrorK8sDeploymentRollbackFail *Error
	ErrorK8sDeploymentScaleFail    *Error
	ErrorK8sDeploymentRestartFail  *Error
	ErrorK8sDeploymentGetPodFail   *Error
)

func register_k8s_Deployment() {
	ErrorK8sDeploymentCreateFail = NewError(500021, "创建K8s Deployment失败")
	ErrorK8sDeploymentDeleteFail = NewError(500022, "删除K8s Deployment失败")
	ErrorK8sDeploymentListFail = NewError(500023, "获取K8s Deployment列表失败")
	ErrorK8sDeploymentDetailFail = NewError(500024, "获取K8s Deployment详情失败")
	ErrorK8sDeploymentUpdateFail = NewError(500025, "更新K8s Deployment失败")
	ErrorK8sDeploymentRollbackFail = NewError(500026, "回滚K8s Deployment失败")
	ErrorK8sDeploymentScaleFail = NewError(500027, "扩缩容K8s Deployment失败")
	ErrorK8sDeploymentRestartFail = NewError(500028, "重启K8s Deployment失败")
	ErrorK8sDeploymentGetPodFail = NewError(500029, "获取K8s Deployment Pod列表失败")
}
