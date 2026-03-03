package errorcode

var (
	// Pipeline 相关错误码 (5002xx)
	ErrorPipelineCreateFail   *Error // 创建流水线失败
	ErrorPipelineQueryFail    *Error // 查询流水线失败
	ErrorPipelineUpdateFail   *Error // 更新流水线失败
	ErrorPipelineDeleteFail   *Error // 删除流水线失败
	ErrorPipelineRunFail      *Error // 运行流水线失败
	ErrorPipelineStopFail     *Error // 停止流水线失败
	ErrorPipelineLogsFail     *Error // 获取流水线日志失败
	ErrorPipelineCallbackFail *Error // 处理回调失败
	ErrorPipelineNotFound     *Error // 流水线不存在
	ErrorPipelineRunning      *Error // 流水线正在运行
	ErrorPipelineDisabled     *Error // 流水线已禁用
)

func register_pipeline() {
	// 5002xx：Pipeline
	ErrorPipelineCreateFail = NewError(500200, "创建流水线失败")
	ErrorPipelineQueryFail = NewError(500201, "查询流水线失败")
	ErrorPipelineUpdateFail = NewError(500202, "更新流水线失败")
	ErrorPipelineDeleteFail = NewError(500203, "删除流水线失败")
	ErrorPipelineRunFail = NewError(500204, "运行流水线失败")
	ErrorPipelineStopFail = NewError(500205, "停止流水线失败")
	ErrorPipelineLogsFail = NewError(500206, "获取流水线日志失败")
	ErrorPipelineCallbackFail = NewError(500207, "处理回调失败")
	ErrorPipelineNotFound = NewError(500208, "流水线不存在")
	ErrorPipelineRunning = NewError(500209, "流水线正在运行中")
	ErrorPipelineDisabled = NewError(500210, "流水线已禁用")
}
