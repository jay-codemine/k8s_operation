package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	dmai "k8soperation/internal/domain/ai"
	"k8soperation/pkg/openai"
)

// aiLog 获取 AI 专属日志器（兑底到系统日志）
func aiLog() *zap.Logger {
	if global.AILogger != nil {
		return global.AILogger.Logger()
	}
	if global.Logger != nil {
		return global.Logger.Logger()
	}
	return zap.NewNop()
}

// =========================================================================
// AI 助手 Service —— Function Calling 对话 + 自动执行 + 高危审批
// =========================================================================

func (s *Services) aiSvc() *dmai.AIService {
	return dmai.NewAIService(s.db)
}

// AI 客户端单例（懒加载）—— 仅作为无指定 provider/model 时的默认回退
var aiClient *openai.Client

// getAIClientWithModel 通过 Registry 获取指定提供商+模型的客户端
func getAIClientWithModel(providerID, modelID string) (*openai.Client, error) {
	if global.AISetting == nil || !global.AISetting.Enabled {
		return nil, fmt.Errorf("AI 助手未启用")
	}

	// 优先使用 Registry
	if global.AIRegistry != nil {
		// 构建完整 System Prompt：用户自定义描述 + 工具使用指令
		systemPrompt := global.AISetting.SystemPrompt
		if systemPrompt == "" {
			systemPrompt = defaultSystemPrompt
		} else {
			// 用户设置了自定义 prompt，追加工具使用指令确保模型知道如何用工具
			systemPrompt = systemPrompt + toolUsageInstruction
		}
		return global.AIRegistry.GetClient(providerID, modelID, systemPrompt)
	}

	// 回退到旧版单客户端
	return getAIClient()
}

// getAIClient 获取默认 AI 客户端（兼容旧版）
func getAIClient() (*openai.Client, error) {
	if global.AISetting == nil || !global.AISetting.Enabled {
		return nil, fmt.Errorf("AI 助手未启用")
	}
	if aiClient != nil {
		return aiClient, nil
	}

	systemPrompt := global.AISetting.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	aiClient = openai.NewClient(openai.Config{
		APIKey:       global.AISetting.APIKey,
		BaseURL:      global.AISetting.BaseURL,
		Model:        global.AISetting.Model,
		MaxTokens:    global.AISetting.MaxTokens,
		Temperature:  global.AISetting.Temperature,
		SystemPrompt: systemPrompt,
	})
	return aiClient, nil
}

const defaultSystemPrompt = `你是 K8s 管理平台的智能助手，同时也是一个知识渊博的通用 AI 助手。

你的核心能力包括：
【平台操作能力】
1. 查询和管理 Kubernetes 集群资源（Pod、Deployment、Service、Node、Namespace 等）
2. 执行资源操作（创建、删除、扩缩容、重启、回滚、镜像更新）
3. 管理 CI/CD 流水线（查询、触发构建）
4. 节点管理（查询、cordon/uncordon、drain）
5. 集群级资源管理（ConfigMap、Secret、Ingress、PVC 等）

【通用问答能力】
- 你可以回答任何领域的问题，包括但不限于：技术对比、编程知识、操作系统、网络、数据库、DevOps、日常生活等
- 当用户提问非 K8s 平台操作的问题时，直接用你的知识回答，不需要调用工具

重要规则：
- 用简洁专业的中文回答
- 当用户查询平台资源时，优先调用工具函数获取真实数据，而不是编造
- ❗ 用户没有指定 cluster_id 时，必须先调用 list_clusters 获取集群列表，然后告诉用户有哪些集群，让用户确认要操作哪个集群，或者如果只有一个集群则默认使用该集群
- 查询类操作直接执行，写操作和删除操作会需要人工审批
- 将工具返回的数据整理成对用户友好的格式进行回答
- 对于通用知识问题，友好、详细地回答，不要拒绝非 K8s 相关的问题`

// toolUsageInstruction 工具使用指令（始终追加到 System Prompt 后面）
const toolUsageInstruction = `

【工具使用规则 - 必须遵守】
你拥有平台工具函数，当用户请求涉及以下操作时，你必须调用对应的工具函数：
- "查看集群" / "集群列表" → 调用 list_clusters
- "查看Pod" / "Pod列表" / "查看某个namespace的Pod" → 调用 list_pods
- "查看Deployment" / "部署列表" → 调用 list_deployments
- "查看节点" / "Node列表" → 调用 list_nodes
- "查看命名空间" / "Namespace列表" → 调用 list_namespaces
- "查看Service" / "服务列表" → 调用 list_services
- "查看事件" / "Event列表" → 调用 list_events
- 扩缩容、重启、删除等操作 → 调用对应的写操作工具

❗ 集群确认规则（最重要）：
- 几乎所有工具都需要 cluster_id 参数
- 如果用户没有明确指定集群，你必须先调用 list_clusters 获取集群列表
- 如果只有 1 个集群，默认使用该集群的 ID
- 如果有多个集群，请列出集群列表让用户选择
- 绝对不要编造或猜测 cluster_id

禁止行为：
- 当用户要求查看/操作平台资源时，禁止只回复文字说明而不调用工具
- 禁止编造数据，必须通过工具获取真实数据后再回答
- 禁止在不确定 cluster_id 时直接执行操作，必须先查询集群

立即行动：收到用户查询指令后，直接调用工具函数，不要先回复"我来帮你查"之类的话。`

// needToolCalling 本地意图预判：通过关键词快速判断用户消息是否需要调用平台工具
// 简单问候/闲聊/知识问答 → false（跳过工具，直接 Chat）
// 涉及平台操作的关键词 → true（带工具 ChatWithTools）
func needToolCalling(message string) bool {
	msg := strings.ToLower(message)

	// 平台操作关键词（命中任一则需要工具）
	platformKeywords := []string{
		// 资源类型
		"pod", "deployment", "service", "node", "namespace",
		"ingress", "configmap", "secret", "pvc", "pv",
		"statefulset", "daemonset", "job", "cronjob",
		"replicaset",
		// 中文资源
		"集群", "节点", "命名空间", "容器", "副本", "镜像",
		"工作负载", "存储卷", "配置项", "密钥",
		// 操作动词
		"查看", "查询", "列出", "获取", "搜索",
		"创建", "删除", "更新", "修改", "编辑",
		"扩容", "缩容", "扩缩容", "重启", "回滚",
		"部署", "发布", "上线", "下线",
		"排空", "drain", "cordon", "uncordon",
		"scale", "restart", "rollback", "delete",
		// 流水线
		"流水线", "pipeline", "构建", "cicd", "ci/cd",
		// 事件/日志
		"事件", "event", "日志", "log",
		// 指标
		"cpu", "内存", "memory", "资源占用", "监控",
	}

	for _, kw := range platformKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

// toolCategory 工具类别
type toolCategory int

const (
	toolCatQuery  toolCategory = iota // 查询类
	toolCatWrite                      // 写操作
	toolCatDelete                     // 删除操作
)

// selectRelevantTools 根据用户消息智能筛选相关工具子集，减少发送给 AI 的 token 数量
func selectRelevantTools(message string) []openai.ToolDef {
	msg := strings.ToLower(message)
	allTools := BuildAllTools()

	// 检测意图类别
	isDelete := containsAny(msg, []string{"delete", "删除", "移除", "remove", "drain", "排空"})
	isWrite := containsAny(msg, []string{"scale", "restart", "rollback", "create", "update", "trigger",
		"扩容", "缩容", "扩缩容", "重启", "回滚", "创建", "更新", "触发",
		"发布", "部署", "上线", "下线", "cordon", "uncordon"})

	// 删除操作：只发删除工具 + list_clusters（用于获取 cluster_id）
	if isDelete && !isWrite {
		return filterTools(allTools, func(name string) bool {
			return strings.HasPrefix(name, "delete_") || name == "drain_node" ||
				name == "list_clusters"
		})
	}

	// 写操作：发写工具 + list_clusters
	if isWrite {
		return filterTools(allTools, func(name string) bool {
			meta, ok := GetToolMeta(name)
			if !ok {
				return false
			}
			return meta.NeedApproval || name == "list_clusters"
		})
	}

	// 查询操作：只发查询工具
	return filterTools(allTools, func(name string) bool {
		meta, ok := GetToolMeta(name)
		if !ok {
			return false
		}
		return !meta.NeedApproval // 只返回只读工具
	})
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func filterTools(tools []openai.ToolDef, predicate func(string) bool) []openai.ToolDef {
	var result []openai.ToolDef
	for _, t := range tools {
		if predicate(t.Name) {
			result = append(result, t)
		}
	}
	return result
}

// =========================================================================
// 会话管理
// =========================================================================

// AIConversationCreate 创建新会话
func (s *Services) AIConversationCreate(ctx context.Context, userID uint32, title string) (*models.AIConversation, error) {
	if title == "" {
		title = "新对话"
	}
	conv := &models.AIConversation{
		UserID: userID,
		Title:  title,
		Status: 1,
	}
	if err := s.aiSvc().ConversationCreate(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// AIConversationList 获取用户会话列表
func (s *Services) AIConversationList(ctx context.Context, userID uint32, page, pageSize int) ([]*models.AIConversation, int64, error) {
	return s.aiSvc().ConversationList(ctx, userID, page, pageSize)
}

// AIConversationDelete 归档（软删除）会话
func (s *Services) AIConversationDelete(ctx context.Context, convID, userID uint32) error {
	return s.aiSvc().ConversationDelete(ctx, convID, userID)
}

// AIMessageHistory 获取会话历史消息
func (s *Services) AIMessageHistory(ctx context.Context, convID, userID uint32) ([]*models.AIMessage, error) {
	// 先验证会话归属
	if _, err := s.aiSvc().ConversationGetByID(ctx, convID, userID); err != nil {
		return nil, fmt.Errorf("会话不存在或无权访问")
	}
	return s.aiSvc().MessageListByConversation(ctx, convID)
}

// =========================================================================
// AI 对话（核心逻辑）
// =========================================================================

// AIChatRequest 聊天请求
type AIChatRequest struct {
	ConversationID uint32 `json:"conversation_id"` // 可选，0 = 新建会话
	Message        string `json:"message"`
	ProviderID     string `json:"provider_id,omitempty"` // 可选，指定 AI 提供商
	ModelID        string `json:"model_id,omitempty"`    // 可选，指定模型
	UserID         uint32 `json:"-"`                     // 从 JWT 中获取
}

// AIChatResponse 聊天响应
type AIChatResponse struct {
	ConversationID uint32           `json:"conversation_id"`
	Reply          string           `json:"reply"`
	NeedApproval   bool             `json:"need_approval"`
	ApprovalID     uint32           `json:"approval_id,omitempty"`
	ToolsCalled    []string         `json:"tools_called,omitempty"`  // 本次对话调用了哪些工具
	PendingTools   []PendingToolInfo `json:"pending_tools,omitempty"` // 等待审批的工具
	ContextRound   int              `json:"context_round"`           // 当前对话轮次（本次对话是第几轮）
	HistoryCount   int              `json:"history_count"`           // 本次请求携带的历史消息数（不含本次）
}

// PendingToolInfo 等待审批的工具信息
type PendingToolInfo struct {
	ToolName     string            `json:"tool_name"`
	ApprovalID   uint32            `json:"approval_id"`
	RiskLevel    string            `json:"risk_level"`
	Summary      string            `json:"summary"`
	Namespace    string            `json:"namespace,omitempty"`
	ResourceName string            `json:"resource_name,omitempty"`
	ClusterID    uint32            `json:"cluster_id,omitempty"`
	ToolArgs     map[string]interface{} `json:"tool_args,omitempty"` // 完整参数
}

// AIChat 智能对话（Function Calling + 自动执行 + 高危审批）
func (s *Services) AIChat(ctx context.Context, req *AIChatRequest, factory *ClusterClientFactory) (*AIChatResponse, error) {
	start := time.Now()
	l := aiLog().With(
		zap.Uint32("user_id", req.UserID),
		zap.String("provider", req.ProviderID),
		zap.String("model", req.ModelID),
	)
	l.Info("[AI-Chat] 对话开始",
		zap.Uint32("conv_id", req.ConversationID),
		zap.String("message", truncateMsg(req.Message, 200)),
	)

	client, err := getAIClientWithModel(req.ProviderID, req.ModelID)
	if err != nil {
		l.Error("[AI-Chat] 获取客户端失败", zap.Error(err))
		return nil, err
	}

	// 1. 获取或创建会话
	convID := req.ConversationID
	if convID == 0 {
		conv, err := s.AIConversationCreate(ctx, req.UserID, "")
		if err != nil {
			return nil, fmt.Errorf("创建会话失败: %w", err)
		}
		convID = conv.ID
	}

	// 2. 保存用户消息
	userMsg := &models.AIMessage{
		ConversationID: convID,
		Role:           "user",
		Content:        req.Message,
	}
	if err := s.aiSvc().MessageCreate(ctx, userMsg); err != nil {
		global.Logger.Error("保存用户消息失败", zap.Error(err))
	}

	// 3. 构建历史上下文（含本次 user 消息，因为已先 save 再读 DB）
	messages, err := s.buildHistoryMessages(ctx, convID)
	if err != nil {
		global.Logger.Warn("获取历史消息失败", zap.Error(err))
		messages = []openai.Message{{Role: "user", Content: req.Message}}
	}
	// 统计上下文轮次（user/assistant 成对算一轮，包含本次）
	userMsgCount := 0
	for _, m := range messages {
		if m.Role == "user" {
			userMsgCount++
		}
	}
	resp_historyCount := len(messages) - 1 // 不含本次 user 消息
	if resp_historyCount < 0 {
		resp_historyCount = 0
	}
	l.Info("[AI-Chat] 多轮上下文已构建",
		zap.Int("total_messages", len(messages)),
		zap.Int("history_count", resp_historyCount),
		zap.Int("current_round", userMsgCount),
	)

	// 4. 智能路由：判断是否需要工具调用
	useTools := needToolCalling(req.Message)
	resp := &AIChatResponse{
		ConversationID: convID,
		ContextRound:   userMsgCount,
		HistoryCount:   resp_historyCount,
	}

	if !useTools {
		// 简单对话（问候/闲聊/知识问答）→ 直接 Chat，不发送工具定义，大幅提速
		l.Info("[AI-Chat] 智能路由: 普通对话（跳过工具）")
		reply, chatErr := client.Chat(ctx, messages)
		if chatErr != nil {
			l.Error("[AI-Chat] 普通对话失败",
				zap.Duration("total_latency", time.Since(start)),
				zap.Error(chatErr),
			)
			return nil, fmt.Errorf("AI 请求失败: %w", chatErr)
		}
		resp.Reply = reply
	} else {
	// 平台操作 → 带工具的 Function Calling（智能筛选相关工具子集）
	l.Info("[AI-Chat] 智能路由: 平台操作（启用工具）")
	tools := selectRelevantTools(req.Message)
	l.Info("[AI-Chat] 工具筛选完成",
		zap.Int("selected_tools", len(tools)),
	)

	// Function Calling 循环（最多 5 轮，防止死循环）
	const maxRounds = 5
	for round := 0; round < maxRounds; round++ {
		l.Info("[AI-Chat] Function Calling 轮次",
			zap.Int("round", round+1),
			zap.Int("context_msgs", len(messages)),
		)

		result, err := client.ChatWithTools(ctx, messages, tools)
		if err != nil {
			l.Error("[AI-Chat] AI 请求失败",
				zap.Int("round", round+1),
				zap.Duration("total_latency", time.Since(start)),
				zap.Error(err),
			)
			return nil, fmt.Errorf("AI 请求失败: %w", err)
		}

		// 如果 GPT 没有调用工具，直接返回文本回复
		if len(result.ToolCalls) == 0 {
			resp.Reply = result.Content
			break
		}

		// 将 assistant 的 tool_calls 消息加入上下文（必须携带 ToolCalls，否则 API 报错）
		messages = append(messages, openai.Message{
			Role:      "assistant",
			Content:   result.Content,
			ToolCalls: result.ToolCalls,
		})

		// 逐个处理工具调用
		hasApproval := false
		for _, tc := range result.ToolCalls {
			toolName := tc.Function
			meta, _ := GetToolMeta(toolName)

			// 检查是否需要审批
			if meta.NeedApproval {
				l.Warn("[AI-Chat] 工具需要审批",
					zap.String("tool", toolName),
					zap.String("risk_level", meta.RiskLevel),
					zap.String("args", truncateMsg(tc.Args, 500)),
				)
				// 高危/写操作 → 创建审批请求，不直接执行
				approval, aErr := s.createToolApproval(ctx, convID, req.UserID, tc, meta)
				if aErr != nil {
					global.Logger.Error("创建审批请求失败", zap.Error(aErr))
					// 审批创建失败时，告知 GPT
					messages = append(messages, openai.Message{
						Role:       "tool",
						Content:    fmt.Sprintf(`{"error": "创建审批失败: %s"}`, aErr.Error()),
						ToolCallID: tc.ID,
					})
				} else {
					hasApproval = true
					resp.NeedApproval = true
					resp.ApprovalID = approval.ID
					// 解析工具参数用于前端展示
					var toolArgs map[string]interface{}
					_ = json.Unmarshal([]byte(tc.Args), &toolArgs)
					resp.PendingTools = append(resp.PendingTools, PendingToolInfo{
						ToolName:     toolName,
						ApprovalID:   approval.ID,
						RiskLevel:    meta.RiskLevel,
						Summary:      meta.Description,
						Namespace:    approval.Namespace,
						ResourceName: approval.ResourceName,
						ClusterID:    approval.ClusterID,
						ToolArgs:     toolArgs,
					})
					// 告知 GPT 该操作需要审批
					messages = append(messages, openai.Message{
						Role:       "tool",
						Content:    fmt.Sprintf(`{"status": "pending_approval", "approval_id": %d, "message": "该操作属于%s级别，已提交审批请求(ID:%d)，等待管理员确认后执行"}`, approval.ID, meta.RiskLevel, approval.ID),
						ToolCallID: tc.ID,
					})
				}
			} else {
				// 只读操作 → 直接执行
				toolStart := time.Now()
				toolResult, execErr := s.ExecuteToolCall(ctx, factory, toolName, tc.Args)
				toolLatency := time.Since(toolStart)

				if execErr != nil {
					l.Error("[AI-Chat] 工具执行失败",
						zap.String("tool", toolName),
						zap.Duration("tool_latency", toolLatency),
						zap.String("args", truncateMsg(tc.Args, 500)),
						zap.Error(execErr),
					)
					toolResult = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
				} else {
					l.Info("[AI-Chat] 工具执行成功",
						zap.String("tool", toolName),
						zap.Duration("tool_latency", toolLatency),
						zap.Int("result_length", len(toolResult)),
					)
				}
				resp.ToolsCalled = append(resp.ToolsCalled, toolName)

				// 将执行结果反馈给 GPT
				messages = append(messages, openai.Message{
					Role:       "tool",
					Content:    toolResult,
					ToolCallID: tc.ID,
				})
			}
		}

		// 如果有审批操作，让 GPT 生成最终回复后结束循环
		if hasApproval {
			finalResult, fErr := client.ContinueWithToolResults(ctx, messages, tools)
			if fErr == nil && finalResult.Content != "" {
				resp.Reply = finalResult.Content
			} else {
				// 回退方案
				resp.Reply = s.buildApprovalReplyText(resp.PendingTools)
			}
			break
		}
		// 否则继续循环，让 GPT 根据工具结果继续对话
	}
	} // end if useTools

	// 5. 保存 AI 回复消息
	toolsJSON, _ := json.Marshal(resp.ToolsCalled)
	assistantMsg := &models.AIMessage{
		ConversationID: convID,
		Role:           "assistant",
		Content:        resp.Reply,
		IntentJSON:     string(toolsJSON),
	}
	if err := s.aiSvc().MessageCreate(ctx, assistantMsg); err != nil {
		global.Logger.Error("保存AI回复消息失败", zap.Error(err))
	}

	// 6. 自动更新会话标题（首条消息时）
	if req.ConversationID == 0 {
		title := req.Message
		if len([]rune(title)) > 30 {
			title = string([]rune(title)[:30]) + "..."
		}
		_ = s.aiSvc().ConversationUpdateTitle(ctx, convID, title)
	}

	totalLatency := time.Since(start)
	l.Info("[AI-Chat] 对话完成",
		zap.Uint32("conv_id", convID),
		zap.Duration("total_latency", totalLatency),
		zap.Int("tools_called", len(resp.ToolsCalled)),
		zap.Strings("tools", resp.ToolsCalled),
		zap.Bool("need_approval", resp.NeedApproval),
		zap.Int("reply_length", len(resp.Reply)),
	)

	return resp, nil
}

// buildApprovalReplyText 构建审批提示文本（GPT 生成失败时的回退方案）
func (s *Services) buildApprovalReplyText(pendingTools []PendingToolInfo) string {
	text := "⚠️ **检测到需要审批的操作**\n\n"
	for _, pt := range pendingTools {
		text += fmt.Sprintf("- **%s**（风险等级: %s，审批ID: %d）\n", pt.Summary, pt.RiskLevel, pt.ApprovalID)
	}
	text += "\n请等待管理员审批后自动执行，或在「审批管理」页面查看进度。"
	return text
}

// AIChatStream 流式对话（SSE）—— 流式不支持 Function Calling，回退为普通流式对话
func (s *Services) AIChatStream(ctx context.Context, req *AIChatRequest, callback openai.StreamCallback) (*AIChatResponse, error) {
	start := time.Now()
	l := aiLog().With(
		zap.Uint32("user_id", req.UserID),
		zap.String("provider", req.ProviderID),
		zap.String("model", req.ModelID),
	)
	l.Info("[AI-Stream] 流式对话开始",
		zap.Uint32("conv_id", req.ConversationID),
		zap.String("message", truncateMsg(req.Message, 200)),
	)

	client, err := getAIClientWithModel(req.ProviderID, req.ModelID)
	if err != nil {
		l.Error("[AI-Stream] 获取客户端失败", zap.Error(err))
		return nil, err
	}

	// 获取或创建会话
	convID := req.ConversationID
	if convID == 0 {
		conv, err := s.AIConversationCreate(ctx, req.UserID, "")
		if err != nil {
			return nil, fmt.Errorf("创建会话失败: %w", err)
		}
		convID = conv.ID
	}

	// 保存用户消息
	userMsg := &models.AIMessage{
		ConversationID: convID,
		Role:           "user",
		Content:        req.Message,
	}
	_ = s.aiSvc().MessageCreate(ctx, userMsg)

	// 构建历史上下文
	messages, err := s.buildHistoryMessages(ctx, convID)
	if err != nil {
		messages = []openai.Message{{Role: "user", Content: req.Message}}
	}

	// 流式调用
	fullReply, err := client.ChatStream(ctx, messages, callback)
	if err != nil {
		l.Error("[AI-Stream] 流式调用失败",
			zap.Duration("latency", time.Since(start)),
			zap.Error(err),
		)
		return nil, err
	}

	l.Info("[AI-Stream] 流式对话完成",
		zap.Duration("latency", time.Since(start)),
		zap.Int("reply_length", len(fullReply)),
	)

	// 保存 AI 回复
	assistantMsg := &models.AIMessage{
		ConversationID: convID,
		Role:           "assistant",
		Content:        fullReply,
	}
	_ = s.aiSvc().MessageCreate(ctx, assistantMsg)

	// 自动更新标题
	if req.ConversationID == 0 {
		title := req.Message
		if len([]rune(title)) > 30 {
			title = string([]rune(title)[:30]) + "..."
		}
		_ = s.aiSvc().ConversationUpdateTitle(ctx, convID, title)
	}

	return &AIChatResponse{
		ConversationID: convID,
		Reply:          fullReply,
	}, nil
}

// =========================================================================
// 审批管理
// =========================================================================

// AIApprovalList 获取审批列表（管理员用）
func (s *Services) AIApprovalList(ctx context.Context, status uint8, page, pageSize int) ([]*models.AIApprovalRequest, int64, error) {
	return s.aiSvc().ApprovalListAll(ctx, status, page, pageSize)
}

// AIApprovalMyList 获取我的审批申请
func (s *Services) AIApprovalMyList(ctx context.Context, userID uint32, page, pageSize int) ([]*models.AIApprovalRequest, int64, error) {
	return s.aiSvc().ApprovalListByUser(ctx, userID, page, pageSize)
}

// AIApprovalDetail 获取审批详情
func (s *Services) AIApprovalDetail(ctx context.Context, approvalID uint32) (*models.AIApprovalRequest, []*models.AIApprovalLog, error) {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return nil, nil, err
	}
	logs, err := s.aiSvc().ApprovalLogListByApproval(ctx, approvalID)
	if err != nil {
		return req, nil, err
	}
	return req, logs, nil
}

// AIApprovalApprove 通过审批（通过后自动执行工具调用）
func (s *Services) AIApprovalApprove(ctx context.Context, approvalID, approverID uint32, comment string, factory *ClusterClientFactory, adminOverride bool) error {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("审批请求不存在")
	}
	// 管理页面允许从"已过期"状态强制审批
	if adminOverride {
		if req.Status != models.AIApprovalPending && req.Status != models.AIApprovalExpired {
			return fmt.Errorf("审批已处理（状态: %d），无法重复操作", req.Status)
		}
	} else {
		if req.Status != models.AIApprovalPending {
			return fmt.Errorf("审批已处理")
		}
	}
	// 禁止自审：审批人不能是申请人（管理页面 adminOverride 可跳过）
	if req.RequestUserID == approverID && !adminOverride {
		return fmt.Errorf("不能审批自己提交的操作请求")
	}
	// 检查是否过期（管理页面 adminOverride 可跳过过期限制）
	if !adminOverride && req.ExpireAt > 0 && uint32(time.Now().Unix()) > req.ExpireAt {
		_ = s.aiSvc().ApprovalUpdateStatus(ctx, approvalID, models.AIApprovalExpired, 0, "已过期")
		return fmt.Errorf("审批已过期")
	}

	if err := s.aiSvc().ApprovalUpdateStatus(ctx, approvalID, models.AIApprovalApproved, approverID, comment); err != nil {
		return err
	}

	// 记录审批日志
	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approvalID,
		UserID:     approverID,
		Action:     "approve",
		Comment:    comment,
	})

	// 审批通过后自动执行工具调用
	if req.ToolName != "" && !req.Executed && factory != nil {
		go s.executeApprovedTool(context.Background(), req, factory)
	}

	return nil
}

// executeApprovedTool 异步执行审批通过的工具调用
func (s *Services) executeApprovedTool(ctx context.Context, approval *models.AIApprovalRequest, factory *ClusterClientFactory) {
	global.Logger.Info("开始执行审批通过的操作",
		zap.Uint32("approval_id", approval.ID),
		zap.String("tool", approval.ToolName))

	result, err := s.ExecuteToolCall(ctx, factory, approval.ToolName, approval.ToolArgsJSON)
	if err != nil {
		result = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		global.Logger.Error("审批操作执行失败",
			zap.Uint32("approval_id", approval.ID), zap.Error(err))
	}

	// 更新执行结果
	_ = s.aiSvc().ApprovalUpdateExecuteResult(ctx, approval.ID, result)

	// 记录执行日志
	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approval.ID,
		UserID:     0, // 系统自动执行
		Action:     "execute",
		Comment:    fmt.Sprintf("审批通过后自动执行: %s", approval.ToolName),
	})

	global.Logger.Info("审批操作执行完成",
		zap.Uint32("approval_id", approval.ID),
		zap.String("tool", approval.ToolName))
}

// AIApprovalReject 拒绝审批
func (s *Services) AIApprovalReject(ctx context.Context, approvalID, approverID uint32, comment string, adminOverride bool) error {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("审批请求不存在")
	}
	// 管理页面允许从"已过期"状态强制拒绝
	if adminOverride {
		if req.Status != models.AIApprovalPending && req.Status != models.AIApprovalExpired {
			return fmt.Errorf("审批已处理（状态: %d），无法重复操作", req.Status)
		}
	} else {
		if req.Status != models.AIApprovalPending {
			return fmt.Errorf("审批已处理")
		}
	}
	// 禁止自审：审批人不能是申请人（管理页面 adminOverride 可跳过）
	if req.RequestUserID == approverID && !adminOverride {
		return fmt.Errorf("不能拒绝自己提交的操作请求，请使用取消功能")
	}

	if err := s.aiSvc().ApprovalUpdateStatus(ctx, approvalID, models.AIApprovalRejected, approverID, comment); err != nil {
		return err
	}

	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approvalID,
		UserID:     approverID,
		Action:     "reject",
		Comment:    comment,
	})

	return nil
}

// AIApprovalCancel 取消审批（申请人自己取消）
func (s *Services) AIApprovalCancel(ctx context.Context, approvalID, userID uint32) error {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("审批请求不存在")
	}
	if req.RequestUserID != userID {
		return fmt.Errorf("只有申请人可以取消")
	}
	if req.Status != models.AIApprovalPending {
		return fmt.Errorf("审批已处理")
	}

	if err := s.aiSvc().ApprovalUpdateStatus(ctx, approvalID, models.AIApprovalCanceled, 0, "申请人取消"); err != nil {
		return err
	}

	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approvalID,
		UserID:     userID,
		Action:     "cancel",
		Comment:    "申请人主动取消",
	})

	return nil
}

// AIApprovalPendingCount 获取待审批数量（全局，管理员用）
func (s *Services) AIApprovalPendingCount(ctx context.Context) (int64, error) {
	_, total, err := s.aiSvc().ApprovalListPending(ctx, 1, 1)
	return total, err
}

// AIApprovalMyPendingCount 获取我的待审批数量（普通用户用）
func (s *Services) AIApprovalMyPendingCount(ctx context.Context, userID uint32) (int64, error) {
	return s.aiSvc().ApprovalMyPendingCount(ctx, userID)
}

// AIApprovalDelete 删除审批记录（管理员专用）
func (s *Services) AIApprovalDelete(ctx context.Context, approvalID uint32, userID uint32, isAdmin bool) error {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("审批请求不存在")
	}
	// 管理员可删除任何记录；普通用户只能删除自己提交且非已通过的
	if !isAdmin {
		if req.RequestUserID != userID {
			return fmt.Errorf("无权删除他人的审批记录")
		}
		if req.Status == models.AIApprovalApproved {
			return fmt.Errorf("已通过的审批不可删除")
		}
	}

	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approvalID,
		UserID:     userID,
		Action:     "delete",
		Comment:    "手动删除审批记录",
	})

	return s.aiSvc().ApprovalDelete(ctx, approvalID)
}

// AIApprovalUpdate 更新审批备注（仅待审批状态可编辑）
func (s *Services) AIApprovalUpdate(ctx context.Context, approvalID uint32, userID uint32, comment string, isAdmin bool) error {
	req, err := s.aiSvc().ApprovalGetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("审批请求不存在")
	}
	if !isAdmin && req.RequestUserID != userID {
		return fmt.Errorf("无权编辑他人的审批记录")
	}
	if req.Status != models.AIApprovalPending {
		return fmt.Errorf("只有待审批状态的记录可编辑")
	}

	updates := map[string]interface{}{
		"approve_comment": comment,
	}
	if err := s.aiSvc().ApprovalUpdate(ctx, approvalID, updates); err != nil {
		return err
	}

	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: approvalID,
		UserID:     userID,
		Action:     "update",
		Comment:    fmt.Sprintf("更新备注: %s", comment),
	})
	return nil
}

// AIApprovalStats 获取审批统计数据
func (s *Services) AIApprovalStats(ctx context.Context) (map[string]int64, error) {
	return s.aiSvc().ApprovalStats(ctx)
}

// IsApprovalAdmin 检查用户是否有审批权限
// 只有 super_admin / platform_admin / cluster_admin 角色才能审批
func (s *Services) IsApprovalAdmin(userID int64) bool {
	rbacSvc := s.rbacSvc()
	if rbacSvc.IsSuperAdmin(userID) {
		return true
	}
	roles, err := rbacSvc.UserRoleList(userID)
	if err != nil {
		return false
	}
	for _, role := range roles {
		switch role.RoleType {
		case models.RoleTypeSuperAdmin, "platform_admin", models.RoleTypeClusterAdmin:
			return true
		}
	}
	return false
}

// =========================================================================
// 内部辅助方法
// =========================================================================

// buildHistoryMessages 构建历史消息上下文（多轮对话核心）
// 规则：
//  1. 跳过 tool 角色消息（中间态 Function Calling 结果，对应 assistant.tool_calls 已丢失，API 会 400）
//  2. 跳过 assistant 空 content 消息（仅含 tool_calls 的中间态，直接发送会污染上下文）
//  3. 滑动窗口截取最近 N 轮（默认 20 轮，可通过 AISetting.MaxHistoryRound 配置）
func (s *Services) buildHistoryMessages(ctx context.Context, convID uint32) ([]openai.Message, error) {
	history, err := s.aiSvc().MessageListByConversation(ctx, convID)
	if err != nil {
		return nil, err
	}

	maxRounds := 20
	if global.AISetting != nil && global.AISetting.MaxHistoryRound > 0 {
		maxRounds = global.AISetting.MaxHistoryRound
	}

	// 先过滤无效消息（tool / 空 assistant），再做窗口截取
	var filtered []*models.AIMessage
	for _, msg := range history {
		if msg.Role == "tool" {
			continue
		}
		if msg.Role == "assistant" && strings.TrimSpace(msg.Content) == "" {
			// 仅含 tool_calls 的中间态 assistant，content 为空，跳过
			continue
		}
		filtered = append(filtered, msg)
	}

	// 只取最近 N 轮（一轮 = user + assistant）
	start := 0
	if len(filtered) > maxRounds*2 {
		start = len(filtered) - maxRounds*2
	}

	var messages []openai.Message
	for _, msg := range filtered[start:] {
		messages = append(messages, openai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages, nil
}

// createToolApproval 为 Function Calling 工具调用创建审批请求
func (s *Services) createToolApproval(ctx context.Context, convID, userID uint32, tc openai.ToolCall, meta ToolMeta) (*models.AIApprovalRequest, error) {
	expireMinutes := 1440 // 默认 24 小时
	if global.AISetting != nil && global.AISetting.ApprovalExpire > 0 {
		expireMinutes = global.AISetting.ApprovalExpire
	}

	// 从参数中提取关键信息
	var args map[string]interface{}
	_ = json.Unmarshal([]byte(tc.Args), &args)

	namespace := ""
	if ns, ok := args["namespace"]; ok {
		namespace = fmt.Sprintf("%v", ns)
	}
	resourceName := ""
	if rn, ok := args["name"]; ok {
		resourceName = fmt.Sprintf("%v", rn)
	}
	var clusterID uint32
	if cid, ok := args["cluster_id"]; ok {
		switch v := cid.(type) {
		case float64:
			clusterID = uint32(v)
		}
	}

	req := &models.AIApprovalRequest{
		ConversationID: convID,
		RequestUserID:  userID,
		Intent:         tc.Function,
		Resource:       meta.Description,
		ResourceName:   resourceName,
		Namespace:      namespace,
		ClusterID:      clusterID,
		RiskLevel:      meta.RiskLevel,
		OperationJSON:  tc.Args,
		ToolName:       tc.Function,
		ToolArgsJSON:   tc.Args,
		ToolCallID:     tc.ID,
		Summary:        fmt.Sprintf("%s: %s/%s", meta.Description, namespace, resourceName),
		Status:         models.AIApprovalPending,
		ExpireAt:       uint32(time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix()),
	}
	if err := s.aiSvc().ApprovalCreate(ctx, req); err != nil {
		return nil, err
	}

	// 记录审批创建日志
	_ = s.aiSvc().ApprovalLogCreate(ctx, &models.AIApprovalLog{
		ApprovalID: req.ID,
		UserID:     userID,
		Action:     "create",
		Comment:    fmt.Sprintf("AI Function Calling 检测到%s级操作，自动创建审批: %s", meta.RiskLevel, meta.Description),
	})

	return req, nil
}

// truncateMsg 截断消息用于日志（避免过长日志）
func truncateMsg(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
