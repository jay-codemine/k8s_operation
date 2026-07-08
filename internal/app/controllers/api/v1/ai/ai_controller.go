package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/openai"
)

// AIAssistantController AI 助手控制器
type AIAssistantController struct {
	factory *services.ClusterClientFactory
}

func NewAIAssistantController() *AIAssistantController {
	svc := services.NewServices()
	return &AIAssistantController{
		factory: services.NewClusterClientFactory(svc),
	}
}

// getUserID 从上下文获取当前用户ID
func getUserID(ctx *gin.Context) uint32 {
	uid, _ := ctx.Get("user_id")
	switch v := uid.(type) {
	case int64:
		return uint32(v)
	case uint32:
		return v
	case float64:
		return uint32(v)
	default:
		return 0
	}
}

// Chat 普通对话（非流式）
// @Summary AI 对话
// @Description 向 AI 助手发送消息并获取回复，高危操作会自动触发审批流程
// @Tags AI 助手
// @Accept json
// @Produce json
// @Param body body services.AIChatRequest true "聊天请求"
// @Success 200 {object} services.AIChatResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/ai/chat [post]
func (c *AIAssistantController) Chat(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	// 检查 AI 是否可用
	if global.AISetting == nil || !global.AISetting.Enabled {
		resp.ToErrorResponse(errorcode.AIConfigMissing)
		return
	}

	var req services.AIChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	if req.Message == "" {
		resp.ToErrorResponse(errorcode.AIMessageEmpty)
		return
	}

	req.UserID = userID

	// 使用独立 context，防止前端超时导致请求被取消
	// 工具调用场景可能需要多轮 AI 请求，给足超时时间
	aiCtx, aiCancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer aiCancel()

	svc := services.NewServices()
	result, err := svc.AIChat(aiCtx, &req, c.factory)
	if err != nil {
		global.Logger.Error("AI 对话失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AIRequestFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(result)
}

// ChatStream 流式对话（SSE）
// @Summary AI 流式对话
// @Description 向 AI 助手发送消息，通过 SSE 流式接收回复
// @Tags AI 助手
// @Accept json
// @Produce text/event-stream
// @Param body body services.AIChatRequest true "聊天请求"
// @Router /api/v1/ai/chat/stream [post]
func (c *AIAssistantController) ChatStream(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.SSEvent("error", "用户未登录")
		return
	}

	if global.AISetting == nil || !global.AISetting.Enabled {
		ctx.SSEvent("error", "AI 功能未启用")
		return
	}

	var req services.AIChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.SSEvent("error", err.Error())
		return
	}
	if req.Message == "" {
		ctx.SSEvent("error", "消息不能为空")
		return
	}

	req.UserID = userID

	// 设置 SSE 响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")

	// SSE 流式使用 request context（连接断开自动停止）
	svc := services.NewServices()
	result, err := svc.AIChatStream(ctx.Request.Context(), &req, func(chunk string) error {
		ctx.SSEvent("message", chunk)
		ctx.Writer.Flush()
		return nil
	})
	if err != nil {
		ctx.SSEvent("error", err.Error())
		ctx.Writer.Flush()
		return
	}

	// 发送完成事件
	ctx.SSEvent("done", gin.H{
		"conversation_id": result.ConversationID,
		"full_reply":      result.Reply,
	})
	ctx.Writer.Flush()
}

// ConversationList 获取会话列表
// @Summary 获取 AI 会话列表
// @Tags AI 助手
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/conversations [get]
func (c *AIAssistantController) ConversationList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	svc := services.NewServices()
	list, total, err := svc.AIConversationList(ctx.Request.Context(), userID, page, pageSize)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(list, total)
}

// ConversationMessages 获取会话消息历史
// @Summary 获取会话消息列表
// @Tags AI 助手
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/conversations/{id}/messages [get]
func (c *AIAssistantController) ConversationMessages(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	convID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if convID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的会话ID"))
		return
	}

	svc := services.NewServices()
	messages, err := svc.AIMessageHistory(ctx.Request.Context(), uint32(convID), userID)
	if err != nil {
		resp.ToErrorResponse(errorcode.AIConversationNotFound.WithDetails(err.Error()))
		return
	}

	resp.Success(messages)
}

// ConversationDelete 删除（归档）会话
// @Summary 删除 AI 会话
// @Tags AI 助手
// @Param id path int true "会话ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/conversations/{id} [delete]
func (c *AIAssistantController) ConversationDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	convID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if convID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的会话ID"))
		return
	}

	svc := services.NewServices()
	if err := svc.AIConversationDelete(ctx.Request.Context(), uint32(convID), userID); err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "会话已删除"})
}

// QuickAsk 快捷问答（不需要创建/关联会话，用于全局提问入口）
// @Summary AI 快捷问答
// @Description 无需会话上下文的快捷提问，适用于全局 AI 入口
// @Tags AI 助手
// @Accept json
// @Produce json
// @Param body body object true "请求体" example({"message":"如何查看Pod日志？"})
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/quick-ask [post]
func (c *AIAssistantController) QuickAsk(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if global.AISetting == nil || !global.AISetting.Enabled {
		resp.ToErrorResponse(errorcode.AIConfigMissing)
		return
	}

	var body struct {
		Message    string `json:"message" binding:"required"`
		ProviderID string `json:"provider_id"`
		ModelID    string `json:"model_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		resp.ToErrorResponse(errorcode.AIMessageEmpty)
		return
	}

	// 通过 Registry 获取客户端
	var client *openai.Client
	var clientErr error
	if global.AIRegistry != nil {
		client, clientErr = global.AIRegistry.GetClient(body.ProviderID, body.ModelID,
			`你是 K8s 管理平台的智能助手，同时也是一个知识渊博的通用 AI 助手。请简洁专业地用中文回答用户的问题，包括但不限于 Kubernetes 运维、技术对比、编程知识、操作系统、日常生活等各类话题。`)
	} else {
		client = openai.NewClient(openai.Config{
			APIKey:      global.AISetting.APIKey,
			BaseURL:     global.AISetting.BaseURL,
			Model:       global.AISetting.Model,
			MaxTokens:   global.AISetting.MaxTokens,
			Temperature: global.AISetting.Temperature,
			SystemPrompt: `你是 K8s 管理平台的智能助手，同时也是一个知识渊博的通用 AI 助手。请简洁专业地用中文回答用户的问题，包括但不限于 Kubernetes 运维、技术对比、编程知识、操作系统、日常生活等各类话题。`,
		})
	}
	if clientErr != nil {
		resp.ToErrorResponse(errorcode.AIRequestFailed.WithDetails(clientErr.Error()))
		return
	}

	// 使用独立 context，防止前端超时导致请求被取消
	aiCtx, aiCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer aiCancel()

	messages := []openai.Message{
		{Role: "user", Content: body.Message},
	}
	reply, callErr := client.Chat(aiCtx, messages)
	if callErr != nil {
		global.Logger.Error("AI 快捷问答失败", zap.Error(callErr))
		resp.ToErrorResponse(errorcode.AIRequestFailed.WithDetails(callErr.Error()))
		return
	}

	resp.Success(gin.H{
		"reply": reply,
	})
}

// Status 获取 AI 助手状态
func (c *AIAssistantController) Status(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	enabled := global.AISetting != nil && global.AISetting.Enabled

	// 获取当前模型信息
	defaultProvider := ""
	defaultModel := ""
	providerCount := 0
	if global.AIRegistry != nil {
		defaultProvider = global.AIRegistry.GetDefaultProviderID()
		defaultModel = global.AIRegistry.GetDefaultModelID()
		providerCount = len(global.AIRegistry.ListProviders())
	}

	// 获取待审批数量
	var pendingCount int64
	if enabled {
		svc := services.NewServices()
		pendingCount, _ = svc.AIApprovalPendingCount(ctx.Request.Context())
	}

	resp.Success(gin.H{
		"enabled":           enabled,
		"default_provider":  defaultProvider,
		"default_model":     defaultModel,
		"provider_count":    providerCount,
		"pending_approvals": pendingCount,
		"features": gin.H{
			"chat":           enabled,
			"stream":         enabled,
			"intent_detect":  enabled,
			"approval":       enabled,
			"quick_ask":      enabled,
			"multi_provider": enabled && providerCount > 1,
		},
	})
}

// Models 获取可用 AI 提供商和模型列表
func (c *AIAssistantController) Models(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if global.AISetting == nil || !global.AISetting.Enabled || global.AIRegistry == nil {
		resp.ToErrorResponse(errorcode.AIConfigMissing)
		return
	}

	providers := global.AIRegistry.ListProviders()
	resp.Success(gin.H{
		"providers":        providers,
		"default_provider": global.AIRegistry.GetDefaultProviderID(),
		"default_model":    global.AIRegistry.GetDefaultModelID(),
	})
}

// IntentAnalyze 单独的意图分析接口
// @Summary AI 意图分析
// @Description 分析用户输入的操作意图，返回风险等级和是否需要审批
// @Tags AI 助手
// @Accept json
// @Produce json
// @Param body body object true "请求体"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/intent [post]
func (c *AIAssistantController) IntentAnalyze(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if global.AISetting == nil || !global.AISetting.Enabled {
		resp.ToErrorResponse(errorcode.AIConfigMissing)
		return
	}

	var body struct {
		Message    string `json:"message" binding:"required"`
		ProviderID string `json:"provider_id"`
		ModelID    string `json:"model_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		resp.ToErrorResponse(errorcode.AIMessageEmpty)
		return
	}

	// 通过 Registry 获取客户端
	var client *openai.Client
	var clientErr error
	if global.AIRegistry != nil {
		client, clientErr = global.AIRegistry.GetClient(body.ProviderID, body.ModelID, "")
	} else {
		client = openai.NewClient(openai.Config{
			APIKey:      global.AISetting.APIKey,
			BaseURL:     global.AISetting.BaseURL,
			Model:       global.AISetting.Model,
			MaxTokens:   1024,
			Temperature: 0.3,
		})
	}
	if clientErr != nil {
		resp.ToErrorResponse(errorcode.AIRequestFailed.WithDetails(clientErr.Error()))
		return
	}

	// 使用独立 context，防止前端超时导致请求被取消
	aiCtx2, aiCancel2 := context.WithTimeout(context.Background(), 120*time.Second)
	defer aiCancel2()

	intentJSON, err := client.AnalyzeIntent(aiCtx2, body.Message)
	if err != nil {
		resp.ToErrorResponse(errorcode.AIIntentParseFailed.WithDetails(err.Error()))
		return
	}

	// 尝试解析为结构化数据
	var intent map[string]interface{}
	if parseErr := parseJSON(intentJSON, &intent); parseErr != nil {
		resp.Success(gin.H{
			"raw":    intentJSON,
			"parsed": false,
		})
		return
	}

	resp.Success(gin.H{
		"intent": intent,
		"parsed": true,
	})
}

func parseJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// Logs 获取 AI 助手最近日志（方便排查问题）
// @Summary AI 日志查询
// @Description 获取 AI 助手的最近日志记录，用于排查大模型问题
// @Tags AI 助手
// @Produce json
// @Param lines query int false "返回行数" default(100)
// @Param level query string false "级别过滤(error/warn/info)" default("")
// @Param keyword query string false "关键字搜索"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/logs [get]
func (c *AIAssistantController) Logs(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	// 参数
	maxLines, _ := strconv.Atoi(ctx.DefaultQuery("lines", "100"))
	if maxLines <= 0 || maxLines > 500 {
		maxLines = 100
	}
	levelFilter := strings.ToLower(ctx.Query("level"))
	keyword := ctx.Query("keyword")

	logFile := "storage/logs/ai.log"
	entries, err := readRecentLogLines(logFile, maxLines, levelFilter, keyword)
	if err != nil {
		global.Logger.Warn("AI 日志读取失败", zap.Error(err))
		resp.Success(gin.H{
			"entries": []string{},
			"total":   0,
			"message": "AI 日志文件不存在或无法读取",
		})
		return
	}

	resp.Success(gin.H{
		"entries":  entries,
		"total":    len(entries),
		"log_file": logFile,
	})
}

// readRecentLogLines 读取日志文件最近 N 行，支持级别/关键字过滤
func readRecentLogLines(filePath string, maxLines int, levelFilter, keyword string) ([]map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 读取所有行
	var allLines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 最大 1MB 一行
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		allLines = append(allLines, line)
	}

	// 从尾部取最近 N 行
	startIdx := 0
	if len(allLines) > maxLines*3 { // 取 3 倍多以便过滤后还能够
		startIdx = len(allLines) - maxLines*3
	}

	var results []map[string]interface{}
	for i := startIdx; i < len(allLines); i++ {
		line := allLines[i]

		// 尝试解析为 JSON
		var entry map[string]interface{}
		if json.Unmarshal([]byte(line), &entry) != nil {
			// 非 JSON 格式，包装为纯文本
			entry = map[string]interface{}{"raw": line}
		}

		// 级别过滤
		if levelFilter != "" {
			if lvl, ok := entry["level"].(string); ok {
				if strings.ToLower(lvl) != levelFilter {
					continue
				}
			}
		}

		// 关键字过滤
		if keyword != "" {
			if !strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
				continue
			}
		}

		results = append(results, entry)
	}

	// 只返回最后 maxLines 条
	if len(results) > maxLines {
		results = results[len(results)-maxLines:]
	}

	return results, nil
}
