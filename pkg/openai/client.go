// Package openai 封装 OpenAI/ChatGPT 调用能力，支持 Function Calling
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"

	goopenai "github.com/sashabaranov/go-openai"
)

// Config OpenAI 客户端配置
type Config struct {
	APIKey       string // OpenAI API Key
	BaseURL      string // 自定义 API 地址（兼容国内代理/Azure）
	Model        string // 模型名称, 默认 gpt-4o
	MaxTokens    int    // 最大 Token 数
	Temperature  float32
	SystemPrompt string // 全局 System Prompt
}

// AILog AI 专属日志器接口（由 global.AILogger 注入）
var AILog aiLogger

type aiLogger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

// noopLogger 空日志器（未注入时兜底）
type noopLogger struct{}

func (noopLogger) Debug(string, ...zap.Field) {}
func (noopLogger) Info(string, ...zap.Field)  {}
func (noopLogger) Warn(string, ...zap.Field)  {}
func (noopLogger) Error(string, ...zap.Field) {}

func log() aiLogger {
	if AILog != nil {
		return AILog
	}
	return noopLogger{}
}

// Client OpenAI 客户端
type Client struct {
	cli    *goopenai.Client
	config Config
}

// Message 对话消息
type Message struct {
	Role       string     `json:"role"`                  // system / user / assistant / tool
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // tool 角色时需要
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // assistant 角色时携带的工具调用
}

// ToolCall 工具调用信息
type ToolCall struct {
	ID       string `json:"id"`
	Function string `json:"function"`
	Args     string `json:"args"` // JSON 字符串
}

// ToolDef 工具定义（对应 OpenAI Function Calling）
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON Schema
}

// ChatResult 带 Function Calling 的对话结果
type ChatResult struct {
	Content   string     `json:"content"`              // 文本回复
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // 需要调用的工具
}

// NewClient 创建 OpenAI 客户端
func NewClient(cfg Config) *Client {
	if cfg.Model == "" {
		cfg.Model = goopenai.GPT4o
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}

	var cli *goopenai.Client
	if cfg.BaseURL != "" {
		clientCfg := goopenai.DefaultConfig(cfg.APIKey)
		clientCfg.BaseURL = cfg.BaseURL
		cli = goopenai.NewClientWithConfig(clientCfg)
	} else {
		cli = goopenai.NewClient(cfg.APIKey)
	}

	return &Client{cli: cli, config: cfg}
}

// Chat 普通对话（非流式）
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	msgs := c.buildMessages(messages)
	start := time.Now()

	log().Info("[AI-API] Chat 请求开始",
		zap.String("model", c.config.Model),
		zap.String("base_url", c.config.BaseURL),
		zap.Int("msg_count", len(msgs)),
		zap.Int("max_tokens", c.config.MaxTokens),
	)

	resp, err := c.cli.CreateChatCompletion(ctx, goopenai.ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    msgs,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	})
	latency := time.Since(start)

	if err != nil {
		log().Error("[AI-API] Chat 请求失败",
			zap.String("model", c.config.Model),
			zap.Duration("latency", latency),
			zap.Error(err),
		)
		return "", fmt.Errorf("openai chat error: %w", err)
	}

	if len(resp.Choices) == 0 {
		log().Warn("[AI-API] Chat 空响应",
			zap.String("model", c.config.Model),
			zap.Duration("latency", latency),
		)
		return "", fmt.Errorf("openai: empty response")
	}

	replyLen := len(resp.Choices[0].Message.Content)
	log().Info("[AI-API] Chat 请求成功",
		zap.String("model", c.config.Model),
		zap.Duration("latency", latency),
		zap.Int("reply_length", replyLen),
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
	)

	return resp.Choices[0].Message.Content, nil
}

// ChatWithTools 带 Function Calling 的对话
func (c *Client) ChatWithTools(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResult, error) {
	msgs := c.buildMessages(messages)
	start := time.Now()

	log().Info("[AI-API] ChatWithTools 请求开始",
		zap.String("model", c.config.Model),
		zap.String("base_url", c.config.BaseURL),
		zap.Int("msg_count", len(msgs)),
		zap.Int("tool_count", len(tools)),
	)

	// 构建 OpenAI Tools
	var oaiTools []goopenai.Tool
	for _, t := range tools {
		paramsBytes, _ := json.Marshal(t.Parameters)
		oaiTools = append(oaiTools, goopenai.Tool{
			Type: goopenai.ToolTypeFunction,
			Function: &goopenai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  json.RawMessage(paramsBytes),
			},
		})
	}

	resp, err := c.cli.CreateChatCompletion(ctx, goopenai.ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    msgs,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Tools:       oaiTools,
	})
	latency := time.Since(start)

	if err != nil {
		log().Error("[AI-API] ChatWithTools 请求失败",
			zap.String("model", c.config.Model),
			zap.Duration("latency", latency),
			zap.Error(err),
		)
		return nil, fmt.Errorf("openai chat with tools error: %w", err)
	}

	if len(resp.Choices) == 0 {
		log().Warn("[AI-API] ChatWithTools 空响应",
			zap.String("model", c.config.Model),
			zap.Duration("latency", latency),
		)
		return nil, fmt.Errorf("openai: empty response")
	}

	choice := resp.Choices[0]
	result := &ChatResult{
		Content: choice.Message.Content,
	}

	// 提取 tool_calls
	var toolNames []string
	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:       tc.ID,
			Function: tc.Function.Name,
			Args:     tc.Function.Arguments,
		})
		toolNames = append(toolNames, tc.Function.Name)
	}

	log().Info("[AI-API] ChatWithTools 请求成功",
		zap.String("model", c.config.Model),
		zap.Duration("latency", latency),
		zap.Int("tool_calls", len(result.ToolCalls)),
		zap.Strings("tools", toolNames),
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
	)

	return result, nil
}

// ContinueWithToolResults 将工具执行结果反馈给 GPT 继续对话
func (c *Client) ContinueWithToolResults(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResult, error) {
	return c.ChatWithTools(ctx, messages, tools)
}

// StreamCallback SSE 流式回调函数
type StreamCallback func(chunk string) error

// ChatStream 流式对话（SSE）
func (c *Client) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) (string, error) {
	msgs := c.buildMessages(messages)
	start := time.Now()

	log().Info("[AI-API] ChatStream 请求开始",
		zap.String("model", c.config.Model),
		zap.String("base_url", c.config.BaseURL),
		zap.Int("msg_count", len(msgs)),
	)

	stream, err := c.cli.CreateChatCompletionStream(ctx, goopenai.ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    msgs,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
	})
	if err != nil {
		log().Error("[AI-API] ChatStream 连接失败",
			zap.String("model", c.config.Model),
			zap.Duration("latency", time.Since(start)),
			zap.Error(err),
		)
		return "", fmt.Errorf("openai stream error: %w", err)
	}
	defer stream.Close()

	var sb strings.Builder
	chunkCount := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log().Error("[AI-API] ChatStream 接收中断",
				zap.String("model", c.config.Model),
				zap.Duration("latency", time.Since(start)),
				zap.Int("chunks_received", chunkCount),
				zap.Error(err),
			)
			return sb.String(), fmt.Errorf("openai stream recv error: %w", err)
		}
		if len(resp.Choices) > 0 {
			chunk := resp.Choices[0].Delta.Content
			sb.WriteString(chunk)
			chunkCount++
			if callback != nil {
				if err := callback(chunk); err != nil {
					return sb.String(), err
				}
			}
		}
	}

	log().Info("[AI-API] ChatStream 完成",
		zap.String("model", c.config.Model),
		zap.Duration("latency", time.Since(start)),
		zap.Int("chunks", chunkCount),
		zap.Int("reply_length", sb.Len()),
	)

	return sb.String(), nil
}

// AnalyzeIntent 意图识别（让 GPT 分析用户意图并返回 JSON）
func (c *Client) AnalyzeIntent(ctx context.Context, userMessage string) (string, error) {
	intentPrompt := `你是一个 K8s 管理平台的 AI 助手，请分析用户输入的意图，返回严格的 JSON 格式:
{
  "intent": "操作类型",
  "resource": "资源类型",
  "action": "具体动作",
  "params": {},
  "risk_level": "low/medium/high/critical",
  "need_approval": false,
  "summary": "操作摘要"
}

意图类型说明:
- query: 查询类（查看 Pod、Deployment 等状态）
- create: 创建资源
- update: 更新/扩缩容/滚动更新
- delete: 删除资源（高危）
- scale: 扩缩容
- restart: 重启
- drain: 节点排空（高危）
- rollback: 回滚
- cicd: CI/CD 流水线操作
- config: 配置变更
- chat: 普通咨询

高危操作（need_approval=true）:
- 删除 Namespace
- 删除 Deployment/StatefulSet/DaemonSet
- 节点排空/驱逐（drain/cordon）
- 删除 PV/PVC
- 生产环境的任何写操作
- 集群级别配置变更
- RBAC 权限变更`

	messages := []Message{
		{Role: "system", Content: intentPrompt},
		{Role: "user", Content: userMessage},
	}
	return c.Chat(ctx, messages)
}

func (c *Client) buildMessages(messages []Message) []goopenai.ChatCompletionMessage {
	var msgs []goopenai.ChatCompletionMessage

	// 如果配置了全局 System Prompt，添加到消息最前面
	if c.config.SystemPrompt != "" {
		msgs = append(msgs, goopenai.ChatCompletionMessage{
			Role:    goopenai.ChatMessageRoleSystem,
			Content: c.config.SystemPrompt,
		})
	}

	for _, m := range messages {
		msg := goopenai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		// assistant 消息携带 tool_calls（Function Calling 协议要求）
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, goopenai.ToolCall{
					ID:   tc.ID,
					Type: goopenai.ToolTypeFunction,
					Function: goopenai.FunctionCall{
						Name:      tc.Function,
						Arguments: tc.Args,
					},
				})
			}
		}
		msgs = append(msgs, msg)
	}
	return msgs
}
