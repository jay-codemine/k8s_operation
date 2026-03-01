package errorcode

var (
	// ========== CICD Build（Jenkins 构建）==========
	ErrorCicdBuildCreateFail   *Error // 创建构建记录失败
	ErrorCicdBuildTriggerFail  *Error // 触发 Jenkins 构建失败
	ErrorCicdBuildQueryFail    *Error // 查询构建记录失败
	ErrorCicdBuildCallbackFail *Error // Jenkins 回调处理失败

	// ========== CICD Release（发布单/任务）==========
	ErrorCicdReleaseCreateFail   *Error // 创建发布单失败
	ErrorCicdReleaseQueryFail    *Error // 查询发布单失败
	ErrorCicdReleaseCancelFail   *Error // 取消发布单失败
	ErrorCicdReleaseRollbackFail *Error // 回滚发布单失败
	ErrorCicdReleaseRetryFail    *Error // 重试发布失败

	// ========== CICD Task（按集群任务执行）==========
	ErrorCicdTaskEnqueueFail *Error // 入队失败（Redis Stream）
	ErrorCicdTaskConsumeFail *Error // 消费队列失败（Worker）
	ErrorCicdTaskExecuteFail *Error // 执行发布任务失败（Patch/Watch）
	ErrorCicdTaskRolloutFail *Error // Rollout 等待失败/超时
	ErrorCicdTaskUpdateFail  *Error // 更新任务状态失败（DB）

	// ========== CICD Security（可选：签名/SBOM）==========
	ErrorCicdSbomGenerateFail *Error // 生成/上传 SBOM 失败
	ErrorCicdSignFail         *Error // 镜像签名失败

	// ========== CICD Git（仓库操作）==========
	ErrorGitBranchesFail *Error // 获取Git分支失败
	ErrorGitValidateFail *Error // 验证Git仓库失败
)

func register_cicd() {
	// 50011x：Build
	ErrorCicdBuildCreateFail = NewError(500110, "创建CICD构建记录失败")
	ErrorCicdBuildTriggerFail = NewError(500111, "触发Jenkins构建失败")
	ErrorCicdBuildQueryFail = NewError(500112, "查询CICD构建记录失败")
	ErrorCicdBuildCallbackFail = NewError(500113, "处理Jenkins回调失败")

	// 50012x：Release
	ErrorCicdReleaseCreateFail = NewError(500120, "创建CICD发布单失败")
	ErrorCicdReleaseQueryFail = NewError(500121, "查询CICD发布单失败")
	ErrorCicdReleaseCancelFail = NewError(500122, "取消CICD发布失败")
	ErrorCicdReleaseRollbackFail = NewError(500123, "回滚CICD发布失败")
	ErrorCicdReleaseRetryFail = NewError(500124, "重试CICD发布失败")

	// 50013x：Task
	ErrorCicdTaskEnqueueFail = NewError(500130, "CICD任务入队失败")
	ErrorCicdTaskConsumeFail = NewError(500131, "CICD任务消费失败")
	ErrorCicdTaskExecuteFail = NewError(500132, "执行CICD发布任务失败")
	ErrorCicdTaskRolloutFail = NewError(500133, "等待CICD发布Rollout失败")
	ErrorCicdTaskUpdateFail = NewError(500134, "更新CICD任务状态失败")

	// 50014x：Security（可选）
	ErrorCicdSbomGenerateFail = NewError(500140, "生成/上传SBOM失败")
	ErrorCicdSignFail = NewError(500141, "镜像签名失败")

	// 50015x：Git仓库操作
	ErrorGitBranchesFail = NewError(500150, "获取Git分支列表失败")
	ErrorGitValidateFail = NewError(500151, "验证Git仓库连接失败")
}
