package errorcode

// ===== CronJob 相关（205xxx）=====
var (
	ErrorCronJobNotFound   *Error
	ErrorCronJobCreateFail *Error
	ErrorCronJobDeleteFail *Error
	ErrorCronJobUpdateFail *Error
	ErrorCronJobQueryFail  *Error // 列表 / 单查失败
)

func registerCronJob() {
	ErrorCronJobNotFound = NewError(205001, "CronJob 不存在")
	ErrorCronJobCreateFail = NewError(205002, "创建 CronJob 失败")
	ErrorCronJobDeleteFail = NewError(205003, "删除 CronJob 失败")
	ErrorCronJobUpdateFail = NewError(205004, "更新 CronJob 失败")
	ErrorCronJobQueryFail = NewError(205005, "查询 CronJob 失败")
}
