package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ==================== 通知服务 ====================

// NotifyConfig 通知配置
type NotifyConfig struct {
	DingTalkWebhook string // 钉钉机器人 Webhook URL
	Enabled         bool   // 是否启用通知
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType  string            `json:"msgtype"`
	Markdown DingTalkMarkdown  `json:"markdown,omitempty"`
	At       *DingTalkAt       `json:"at,omitempty"`
}

type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIds []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// ==================== 部署通知 ====================

// NotifyDeployResult 发送部署结果通知
func (s *Services) NotifyDeployResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, errMsg string) {
	// 检查是否配置了钉钉 Webhook
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	// 构建通知内容
	msg := s.buildDeployNotifyMessage(pipeline, stage, success, errMsg)
	
	// 发送钉钉通知
	go s.sendDingTalkNotify(webhook, msg)
}

// NotifyBuildResult 发送构建结果通知
func (s *Services) NotifyBuildResult(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildBuildNotifyMessage(pipeline, run, success)
	go s.sendDingTalkNotify(webhook, msg)
}

// NotifyApprovalRequired 发送审批提醒通知
func (s *Services) NotifyApprovalRequired(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildApprovalNotifyMessage(pipeline, run)
	go s.sendDingTalkNotify(webhook, msg)
}

// NotifyRollbackResult 发送回滚结果通知
func (s *Services) NotifyRollbackResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, targetRS string, oldImage string, newImage string, userID int64, errMsg string) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildRollbackNotifyMessage(pipeline, stage, success, targetRS, oldImage, newImage, userID, errMsg)
	go s.sendDingTalkNotify(webhook, msg)
}

// NotifyCancelDeployResult 发送取消部署结果通知
func (s *Services) NotifyCancelDeployResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, action string, targetRS string, userID int64) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildCancelDeployNotifyMessage(pipeline, stage, action, targetRS, userID)
	go s.sendDingTalkNotify(webhook, msg)
}

// ==================== 消息构建 ====================

func (s *Services) buildDeployNotifyMessage(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, errMsg string) *DingTalkMessage {
	statusIcon := "✅"
	statusText := "部署成功"
	if !success {
		statusIcon = "❌"
		statusText = "部署失败"
	}

	envText := s.getEnvDisplayName(pipeline.DeployEnv)
	platformURL := s.getPlatformURL()
	
	text := fmt.Sprintf(`### %s %s

**流水线**: %s

**环境**: %s

**命名空间**: %s

**工作负载**: %s/%s

**镜像**: %s

**时间**: %s`,
		statusIcon,
		statusText,
		pipeline.Name,
		envText,
		stage.DeployNamespace,
		stage.DeployWorkloadKind,
		stage.DeployWorkloadName,
		stage.DeployImage,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if !success && errMsg != "" {
		text += fmt.Sprintf("\n\n**错误**: %s", errMsg)
	}

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)", 
			pipeline.JenkinsURL, pipeline.JenkinsJob, pipeline.LastBuildNumber)
	}

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[%s] %s", statusText, pipeline.Name),
			Text:  text,
		},
	}
}

func (s *Services) buildBuildNotifyMessage(pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) *DingTalkMessage {
	statusIcon := "✅"
	statusText := "构建成功"
	if !success {
		statusIcon = "❌"
		statusText = "构建失败"
	}

	platformURL := s.getPlatformURL()

	text := fmt.Sprintf(`### %s %s

**流水线**: %s

**分支**: %s

**构建号**: #%d

**镜像**: %s

**耗时**: %ds

**时间**: %s`,
		statusIcon,
		statusText,
		pipeline.Name,
		run.GitBranch,
		run.BuildNumber,
		run.ImageURL,
		run.DurationSec,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if !success && run.ErrorMessage != "" {
		text += fmt.Sprintf("\n\n**错误**: %s", run.ErrorMessage)
	}

	// 如果需要审批，添加提醒
	if success && pipeline.RequireApproval {
		text += "\n\n⏳ **等待审批**: 请前往平台进行人工审批"
	}

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)", 
			pipeline.JenkinsURL, pipeline.JenkinsJob, run.BuildNumber)
	}

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[%s] %s", statusText, pipeline.Name),
			Text:  text,
		},
	}
}

func (s *Services) buildApprovalNotifyMessage(pipeline *models.CicdPipeline, run *models.CicdPipelineRun) *DingTalkMessage {
	envText := s.getEnvDisplayName(pipeline.DeployEnv)
	platformURL := s.getPlatformURL()
	
	text := fmt.Sprintf(`### ⏳ 待审批

**流水线**: %s

**环境**: %s

**分支**: %s

**构建号**: #%d

**镜像**: %s

**时间**: %s`,
		pipeline.Name,
		envText,
		run.GitBranch,
		run.BuildNumber,
		run.ImageURL,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("✅ [点击进行审批](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建日志](%s/job/%s/%d/console)", 
			pipeline.JenkinsURL, pipeline.JenkinsJob, run.BuildNumber)
	}

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[待审批] %s", pipeline.Name),
			Text:  text,
		},
		At: &DingTalkAt{
			IsAtAll: false, // 可以配置 @ 指定人员
		},
	}
}

// buildRollbackNotifyMessage 构建回滚通知消息
func (s *Services) buildRollbackNotifyMessage(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, targetRS string, oldImage string, newImage string, userID int64, errMsg string) *DingTalkMessage {
	statusIcon := "↩️"
	statusText := "回滚成功"
	if !success {
		statusIcon = "❌"
		statusText = "回滚失败"
	}

	envText := s.getEnvDisplayName(pipeline.DeployEnv)
	platformURL := s.getPlatformURL()

	text := fmt.Sprintf(`### %s %s

**流水线**: %s

**环境**: %s

**命名空间**: %s

**工作负载**: %s/%s

**目标版本**: %s

**回滚前镜像**: %s

**回滚后镜像**: %s

**操作人 ID**: %d

**时间**: %s`,
		statusIcon,
		statusText,
		pipeline.Name,
		envText,
		stage.DeployNamespace,
		stage.DeployWorkloadKind,
		stage.DeployWorkloadName,
		targetRS,
		oldImage,
		newImage,
		userID,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if !success && errMsg != "" {
		text += fmt.Sprintf("\n\n**错误**: %s", errMsg)
	}

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[通知] %s - %s", statusText, pipeline.Name),
			Text:  text,
		},
	}
}

// buildCancelDeployNotifyMessage 构建取消部署通知消息
func (s *Services) buildCancelDeployNotifyMessage(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, action string, targetRS string, userID int64) *DingTalkMessage {
	var statusIcon, statusText, actionDesc string
	if action == "cancelled" {
		statusIcon = "⏹️"
		statusText = "部署已取消"
		actionDesc = "取消操作（未执行）"
	} else {
		statusIcon = "↩️"
		statusText = "部署已回滚"
		actionDesc = fmt.Sprintf("取消并回滚到 %s", targetRS)
	}

	envText := s.getEnvDisplayName(pipeline.DeployEnv)
	platformURL := s.getPlatformURL()

	text := fmt.Sprintf(`### %s %s

**流水线**: %s

**环境**: %s

**命名空间**: %s

**工作负载**: %s/%s

**操作**: %s

**操作人 ID**: %d

**时间**: %s`,
		statusIcon,
		statusText,
		pipeline.Name,
		envText,
		stage.DeployNamespace,
		stage.DeployWorkloadKind,
		stage.DeployWorkloadName,
		actionDesc,
		userID,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[通知] %s - %s", statusText, pipeline.Name),
			Text:  text,
		},
	}
}

// ==================== 辅助函数 ====================

func (s *Services) getDingTalkWebhook(pipeline *models.CicdPipeline) string {
	// 优先从流水线配置获取（未来可扩展字段）
	// 否则从全局配置获取
	if global.JenkinsSetting != nil && global.JenkinsSetting.DingTalkWebhook != "" {
		return global.JenkinsSetting.DingTalkWebhook
	}
	return ""
}

func (s *Services) getPlatformURL() string {
	// 优先使用配置的前端页面地址
	if global.JenkinsSetting != nil && global.JenkinsSetting.PlatformURL != "" {
		return global.JenkinsSetting.PlatformURL
	}
	// 回退到回调地址（后端 API）
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		return global.JenkinsSetting.CallbackURL
	}
	return ""
}

func (s *Services) getEnvDisplayName(env string) string {
	switch env {
	case models.DeployEnvDev:
		return "🔧 开发环境"
	case models.DeployEnvStaging:
		return "🧪 测试环境"
	case models.DeployEnvProd:
		return "🚀 生产环境"
	default:
		return env
	}
}

func (s *Services) sendDingTalkNotify(webhook string, msg *DingTalkMessage) {
	if webhook == "" || msg == nil {
		return
	}

	body, err := json.Marshal(msg)
	if err != nil {
		global.Logger.Error("[通知] 序列化钉钉消息失败", zap.Error(err))
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		global.Logger.Error("[通知] 发送钉钉消息失败", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		global.Logger.Warn("[通知] 钉钉返回非200状态码",
			zap.Int("status_code", resp.StatusCode),
		)
		return
	}

	global.Logger.Info("[通知] 钉钉消息发送成功",
		zap.String("title", msg.Markdown.Title),
	)
}
