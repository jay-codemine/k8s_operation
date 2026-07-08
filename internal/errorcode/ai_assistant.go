package errorcode

// ===== AI 助手 (800xxx) =====
var (
	AIServiceUnavailable *Error // AI 服务不可用
	AIRequestFailed      *Error // AI 请求失败
	AIConversationNotFound *Error // 会话不存在
	AIMessageEmpty       *Error // 消息不能为空
	AIStreamError        *Error // 流式响应错误
	AIIntentParseFailed  *Error // 意图解析失败
	AIApprovalNotFound   *Error // 审批请求不存在
	AIApprovalProcessed  *Error // 审批已处理
	AIApprovalExpired    *Error // 审批已过期
	AIApprovalForbidden  *Error // 无权审批
	AIApprovalRequired   *Error // 该操作需要审批
	AIConfigMissing      *Error // AI 配置缺失
)

func registerAIAssistant() {
	AIServiceUnavailable   = NewError(800001, "AI 服务暂不可用，请检查配置")
	AIRequestFailed        = NewError(800002, "AI 请求失败")
	AIConversationNotFound = NewError(800003, "对话会话不存在")
	AIMessageEmpty         = NewError(800004, "消息内容不能为空")
	AIStreamError          = NewError(800005, "AI 流式响应异常")
	AIIntentParseFailed    = NewError(800006, "AI 意图解析失败")
	AIApprovalNotFound     = NewError(800007, "审批请求不存在")
	AIApprovalProcessed    = NewError(800008, "该审批已处理，无法重复操作")
	AIApprovalExpired      = NewError(800009, "审批请求已过期")
	AIApprovalForbidden    = NewError(800010, "无权进行审批操作")
	AIApprovalRequired     = NewError(800011, "该操作为高危操作，需要管理员审批")
	AIConfigMissing        = NewError(800012, "AI 功能未配置，请在系统设置中配置 API Key")
}
