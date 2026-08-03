package ai_assistant

import (
	"github.com/gin-gonic/gin"

	ai "k8soperation/internal/app/controllers/api/v1/ai"
	"k8soperation/internal/app/services"
)

type AIAssistantRouter struct {
	factory *services.ClusterClientFactory
}

// NewAIAssistantRouterWithFactory 注入启动期创建的共享集群客户端工厂。
// AI 对话的工具调用要连目标集群，必须复用共享工厂的客户端缓存与租户校验，
// 不能让 controller 自建。
func NewAIAssistantRouterWithFactory(factory *services.ClusterClientFactory) *AIAssistantRouter {
	return &AIAssistantRouter{factory: factory}
}

// Inject 注入 AI 助手 & 审批管理 & AIOps 路由
// 路由前缀: /api/v1/ai/...
func (r *AIAssistantRouter) Inject(router *gin.RouterGroup) {
	chatCtrl := ai.NewAIAssistantController(r.factory)
	approvalCtrl := ai.NewAIApprovalController(r.factory)
	aiopsCtrl := ai.NewAIOpsController()

	g := router.Group("/ai")
	{
		// ===== AI 助手状态 =====
		g.GET("/status", chatCtrl.Status)
		g.GET("/models", chatCtrl.Models) // 获取可用提供商+模型列表
		g.GET("/logs", chatCtrl.Logs)     // AI 日志查询（排查问题用）

		// ===== AI 对话 =====
		g.POST("/chat", chatCtrl.Chat)                // 普通对话
		g.POST("/chat/stream", chatCtrl.ChatStream)    // 流式对话 (SSE)
		g.POST("/quick-ask", chatCtrl.QuickAsk)        // 快捷问答
		g.POST("/intent", chatCtrl.IntentAnalyze)      // 意图分析

		// ===== 会话管理 =====
		g.GET("/conversations", chatCtrl.ConversationList)              // 会话列表
		g.GET("/conversations/:id/messages", chatCtrl.ConversationMessages) // 消息历史
		g.DELETE("/conversations/:id", chatCtrl.ConversationDelete)     // 删除会话

		// ===== 审批管理 =====
		g.GET("/approvals", approvalCtrl.List)                         // 审批列表（管理员）
		g.GET("/approvals/mine", approvalCtrl.MyList)                  // 我的审批申请
		g.GET("/approvals/pending-count", approvalCtrl.PendingCount)   // 待审批数量
		g.GET("/approvals/stats", approvalCtrl.Stats)                  // 审批统计数据
		g.GET("/approvals/:id", approvalCtrl.Detail)                   // 审批详情
		g.POST("/approvals/:id/approve", approvalCtrl.Approve)        // 通过审批
		g.POST("/approvals/:id/reject", approvalCtrl.Reject)          // 拒绝审批
		g.POST("/approvals/:id/cancel", approvalCtrl.Cancel)          // 取消审批
		g.PUT("/approvals/:id", approvalCtrl.Update)                   // 更新审批备注
		g.DELETE("/approvals/:id", approvalCtrl.Delete)                // 删除审批记录

		// ===== AIOps 智能运维 =====
		ops := g.Group("/ops")
		{
			ops.GET("/dashboard", aiopsCtrl.GetDashboard)                     // AIOps 仪表盘
			ops.GET("/records", aiopsCtrl.GetAnalysisRecords)                 // 分析记录列表
			ops.GET("/channels", aiopsCtrl.GetNotifyChannels)                 // 可用通知渠道
			ops.POST("/alert/analyze", aiopsCtrl.AnalyzeAlert)               // AI 告警分析
			ops.POST("/log/diagnose", aiopsCtrl.DiagnoseLogs)                // AI 日志诊断
			ops.POST("/inspection/run", aiopsCtrl.RunInspection)             // 手动触发巡检
			ops.GET("/inspection/list", aiopsCtrl.GetInspectionReports)      // 巡检报告列表
			ops.GET("/inspection/:id", aiopsCtrl.GetInspectionDetail)        // 巡检报告详情
			ops.GET("/inspection/:id/export", aiopsCtrl.ExportReport)        // 导出巡检报告
			ops.POST("/inspection/:id/notify", aiopsCtrl.NotifyReport)       // 发送巡检报告通知
		}
	}
}
