package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ==================== 通知服务 ====================

// NotifyConfig 通知配置
type NotifyConfig struct {
	DingTalkWebhook string // 钉钉机器人 Webhook URL
	FeishuWebhook   string // 飞书机器人 Webhook URL
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

// FeishuMessage 飞书消息结构（富文本 post 卡片）
type FeishuMessage struct {
	MsgType string             `json:"msg_type"`
	Content FeishuMessageContent `json:"content"`
}

type FeishuMessageContent struct {
	Post FeishuPost `json:"post"`
}

type FeishuPost struct {
	ZhCn FeishuPostContent `json:"zh_cn"`
}

type FeishuPostContent struct {
	Title   string          `json:"title"`
	Content [][]FeishuPostTag `json:"content"`
}

type FeishuPostTag struct {
	Tag    string `json:"tag"`
	Text   string `json:"text,omitempty"`
	Href   string `json:"href,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// ==================== 部署通知 ====================

// NotifyBuildStarted 发送构建开始通知
func (s *Services) NotifyBuildStarted(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, buildNumber int) {
	title, text := s.buildBuildStartedText(pipeline, run, buildNumber)

	// 钉钉通知
	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	// 飞书通知
	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// NotifyDeployResult 发送部署结果通知
func (s *Services) NotifyDeployResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, errMsg string) {
	title, text := s.buildDeployResultText(pipeline, stage, success, errMsg)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// NotifyBuildResult 发送构建结果通知（异步）
func (s *Services) NotifyBuildResult(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) {
	title, text := s.buildBuildResultText(pipeline, run, success)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// NotifyBuildResultSync 发送构建结果通知（同步，用于后台 Worker 中确保通知可追踪）
func (s *Services) NotifyBuildResultSync(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) {
	title, text := s.buildBuildResultText(pipeline, run, success)

	// 同步发送钉钉通知
	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		s.sendDingTalkNotify(webhook, msg)
	} else {
		global.Logger.Warn("[通知] 钉钉 Webhook 未配置，跳过钉钉通知",
			zap.String("pipeline_name", pipeline.Name),
		)
	}

	// 同步发送飞书通知
	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		s.sendFeishuNotify(webhook, msg)
	} else {
		global.Logger.Debug("[通知] 飞书 Webhook 未配置，跳过飞书通知",
			zap.String("pipeline_name", pipeline.Name),
		)
	}
}

// NotifyApprovalRequired 发送审批提醒通知
func (s *Services) NotifyApprovalRequired(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun) {
	title, text := s.buildApprovalText(pipeline, run)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// NotifyRollbackResult 发送回滚结果通知
func (s *Services) NotifyRollbackResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, targetRS string, oldImage string, newImage string, userID int64, errMsg string) {
	title, text := s.buildRollbackText(pipeline, stage, success, targetRS, oldImage, newImage, userID, errMsg)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// NotifyCancelDeployResult 发送取消部署结果通知
func (s *Services) NotifyCancelDeployResult(ctx context.Context, pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, action string, targetRS string, userID int64) {
	title, text := s.buildCancelDeployText(pipeline, stage, action, targetRS, userID)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}

// notifyAutoDeployResult 发送自动部署结果通知（内部使用）
// 用于 Jenkins 回调后的自动部署场景
func (s *Services) notifyAutoDeployResult(ctx context.Context, pipeline *models.CicdPipeline, image string, success bool, errMsg string) {
	title, text := s.buildAutoDeployText(pipeline, image, success, errMsg)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		s.sendFeishuNotify(webhook, msg)
	}
}

// buildAutoDeployText 自动部署通知文本
func (s *Services) buildAutoDeployText(pipeline *models.CicdPipeline, image string, success bool, errMsg string) (string, string) {
	statusIcon := "✅"
	statusText := "自动部署成功"
	if !success {
		statusIcon = "❌"
		statusText = "自动部署失败"
	}

	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
	platformURL := s.getPlatformURL()

	workloadKind := pipeline.TargetWorkloadKind
	if workloadKind == "" {
		workloadKind = "Deployment"
	}

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
		pipeline.TargetNamespace,
		workloadKind,
		pipeline.TargetWorkloadName,
		image,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if !success && errMsg != "" {
		text += fmt.Sprintf("\n\n**错误**: %s", errMsg)
	}

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, pipeline.LastBuildNumber)
	}

	return fmt.Sprintf("[%s] %s", statusText, pipeline.Name), text
}

// notifyLegacyDeployResult 发送旧版配置自动部署结果通知
func (s *Services) notifyLegacyDeployResult(ctx context.Context, pipeline *models.CicdPipeline, namespace, deploymentName, image string, success bool, errMsg string) {
	title, text := s.buildLegacyDeployText(pipeline, namespace, deploymentName, image, success, errMsg)

	if webhook := s.getDingTalkWebhook(pipeline); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		s.sendDingTalkNotify(webhook, msg)
	}

	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		s.sendFeishuNotify(webhook, msg)
	}
}

// buildLegacyDeployText 旧版配置部署通知文本
func (s *Services) buildLegacyDeployText(pipeline *models.CicdPipeline, namespace, deploymentName, image string, success bool, errMsg string) (string, string) {
	statusIcon := "✅"
	statusText := "自动部署成功"
	if !success {
		statusIcon = "❌"
		statusText = "自动部署失败"
	}

	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
	platformURL := s.getPlatformURL()

	text := fmt.Sprintf(`### %s %s

**流水线**: %s

**环境**: %s

**命名空间**: %s

**工作负载**: Deployment/%s

**镜像**: %s

**时间**: %s`,
		statusIcon,
		statusText,
		pipeline.Name,
		envText,
		namespace,
		deploymentName,
		image,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if !success && errMsg != "" {
		text += fmt.Sprintf("\n\n**错误**: %s", errMsg)
	}

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, pipeline.LastBuildNumber)
	}

	return fmt.Sprintf("[%s] %s", statusText, pipeline.Name), text
}

// ==================== 消息构建（统一返回 title, text） ====================

// buildBuildStartedText 构建开始通知文本
func (s *Services) buildBuildStartedText(pipeline *models.CicdPipeline, run *models.CicdPipelineRun, buildNumber int) (string, string) {
	platformURL := s.getPlatformURL()
	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)

	text := fmt.Sprintf(`### 🚀 构建已触发

**流水线**: %s

**环境**: %s

**分支**: %s

**构建号**: #%d

**触发时间**: %s`,
		pipeline.Name,
		envText,
		run.GitBranch,
		buildNumber,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建日志](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, buildNumber)
	}

	return fmt.Sprintf("[构建开始] %s #%d", pipeline.Name, buildNumber), text
}

// buildDeployResultText 部署结果通知文本
func (s *Services) buildDeployResultText(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, errMsg string) (string, string) {
	statusIcon := "✅"
	statusText := "部署成功"
	if !success {
		statusIcon = "❌"
		statusText = "部署失败"
	}

	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
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

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, pipeline.LastBuildNumber)
	}

	return fmt.Sprintf("[%s] %s", statusText, pipeline.Name), text
}

// buildBuildResultText 构建结果通知文本
func (s *Services) buildBuildResultText(pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) (string, string) {
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

	if success && pipeline.RequireApproval {
		text += "\n\n⏳ **等待审批**: 请前往平台进行人工审批"
	}

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, run.BuildNumber)
	}

	return fmt.Sprintf("[%s] %s", statusText, pipeline.Name), text
}

// buildApprovalText 审批提醒通知文本
func (s *Services) buildApprovalText(pipeline *models.CicdPipeline, run *models.CicdPipelineRun) (string, string) {
	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
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

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("✅ [点击进行审批](%s/cicd/pipelines/%d?tab=stages&auto_select=approval)\n\n", platformURL, pipeline.ID)
	}
	if pipeline.JenkinsURL != "" && pipeline.JenkinsJob != "" {
		text += fmt.Sprintf("🛠 [查看 Jenkins 构建日志](%s/job/%s/%d/console)",
			pipeline.JenkinsURL, pipeline.JenkinsJob, run.BuildNumber)
	}

	return fmt.Sprintf("[待审批] %s", pipeline.Name), text
}

// buildRollbackText 回滚通知文本
func (s *Services) buildRollbackText(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, targetRS string, oldImage string, newImage string, userID int64, errMsg string) (string, string) {
	statusIcon := "↩️"
	statusText := "回滚成功"
	if !success {
		statusIcon = "❌"
		statusText = "回滚失败"
	}

	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
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

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}

	return fmt.Sprintf("[通知] %s - %s", statusText, pipeline.Name), text
}

// buildCancelDeployText 取消部署通知文本
func (s *Services) buildCancelDeployText(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, action string, targetRS string, userID int64) (string, string) {
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

	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
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

	text += "\n\n---\n"
	if platformURL != "" {
		text += fmt.Sprintf("🔗 [查看流水线详情](%s/cicd/pipelines/%d?tab=stages)\n\n", platformURL, pipeline.ID)
	}

	return fmt.Sprintf("[通知] %s - %s", statusText, pipeline.Name), text
}

// ==================== 辅助函数 ====================

// getDingTalkWebhook 获取钉钉 Webhook（需开关 EnableDingTalk=true 且 Webhook 非空）
func (s *Services) getDingTalkWebhook(pipeline *models.CicdPipeline) string {
	if global.JenkinsSetting != nil && global.JenkinsSetting.EnableDingTalk && global.JenkinsSetting.DingTalkWebhook != "" {
		return global.JenkinsSetting.DingTalkWebhook
	}
	return ""
}

// getFeishuWebhook 获取飞书 Webhook（CICD 通知用）
// 定义在 cicd_feishu_approval.go 中，这里避免重复定义

// getFeishuSecret 获取飞书签名密钥
// 定义在 cicd_feishu_approval.go 中，这里避免重复定义

func (s *Services) getPlatformURL() string {
	if global.JenkinsSetting != nil && global.JenkinsSetting.PlatformURL != "" {
		return global.JenkinsSetting.PlatformURL
	}
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		return global.JenkinsSetting.CallbackURL
	}
	return ""
}

func (s *Services) getEnvDisplayName(env string) string {
	switch env {
	case models.DeployEnvDev:
		return "🔧 开发环境"
	case models.DeployEnvTest:
		return "🧪 测试环境"
	case models.DeployEnvStaging:
		return "📦 预发环境"
	case models.DeployEnvProd, "production":
		return "🚀 生产环境"
	case "":
		return "未设置"
	default:
		return env
	}
}

// getEnvDisplayNameWithCluster 获取环境显示名称（支持从集群名称解析）
func (s *Services) getEnvDisplayNameWithCluster(env string, clusterID int64) string {
	if env != "" {
		return s.getEnvDisplayName(env)
	}

	if clusterID > 0 {
		var cluster models.K8sCluster
		if err := global.DB.Where("id = ?", clusterID).First(&cluster).Error; err == nil {
			clusterName := cluster.ClusterName
			if strings.Contains(clusterName, "生产") || strings.Contains(clusterName, "prod") {
				return "🚀 生产环境"
			}
			if strings.Contains(clusterName, "预发") || strings.Contains(clusterName, "staging") {
				return "📦 预发环境"
			}
			if strings.Contains(clusterName, "测试") || strings.Contains(clusterName, "test") {
				return "🧪 测试环境"
			}
			if strings.Contains(clusterName, "开发") || strings.Contains(clusterName, "dev") {
				return "🔧 开发环境"
			}
		}
	}

	return "未设置"
}

// ==================== 消息转换 ====================

// textToDingTalkMessage 将通用 title/text 转为钉钉消息
func (s *Services) textToDingTalkMessage(title, text string) *DingTalkMessage {
	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: title,
			Text:  text,
		},
	}
}

// textToFeishuMessage 将通用 title/text 转为飞书富文本消息
// 飞书不支持 Markdown，使用 post 富文本卡片
func (s *Services) textToFeishuMessage(title, text string) *FeishuMessage {
	// 将 Markdown 格式的文本简单转换为飞书富文本格式
	// 飞书 post 格式: [[{tag,text}, {tag,text}, ...], [{tag,text}, ...]]
	var content [][]FeishuPostTag

	// 按行解析，将每行转为对应的飞书标签
	lines := strings.Split(text, "\n")
	var currentLine []FeishuPostTag

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(currentLine) > 0 {
				content = append(content, currentLine)
				currentLine = nil
			}
			continue
		}

		// 处理 ### 标题
		if strings.HasPrefix(trimmed, "### ") {
			if len(currentLine) > 0 {
				content = append(content, currentLine)
				currentLine = nil
			}
			headerText := strings.TrimPrefix(trimmed, "### ")
			content = append(content, []FeishuPostTag{{
				Tag:  "text",
				Text: headerText + "\n",
			}})
			continue
		}

		// 处理 --- 分割线
		if trimmed == "---" {
			if len(currentLine) > 0 {
				content = append(content, currentLine)
				currentLine = nil
			}
			content = append(content, []FeishuPostTag{{
				Tag:  "text",
				Text: "————————————————\n",
			}})
			continue
		}

		// 处理链接 [text](url)
		linkRegex := strings.NewReplacer()
		_ = linkRegex // placeholder

		// 处理 **bold** 粗体
		result := trimmed
		result = strings.ReplaceAll(result, "**", "")

		currentLine = append(currentLine, FeishuPostTag{
			Tag:  "text",
			Text: result + "\n",
		})
	}

	if len(currentLine) > 0 {
		content = append(content, currentLine)
	}

	// 确保至少有一行内容
	if len(content) == 0 {
		content = [][]FeishuPostTag{{{Tag: "text", Text: title}}}
	}

	return &FeishuMessage{
		MsgType: "post",
		Content: FeishuMessageContent{
			Post: FeishuPost{
				ZhCn: FeishuPostContent{
					Title:   title,
					Content: content,
				},
			},
		},
	}
}

// ==================== 发送方法 ====================

func (s *Services) sendDingTalkNotify(webhook string, msg *DingTalkMessage) {
	if webhook == "" || msg == nil {
		global.Logger.Warn("[通知] sendDingTalkNotify 跳过: webhook 或 msg 为空")
		return
	}

	body, err := json.Marshal(msg)
	if err != nil {
		global.Logger.Error("[通知] 序列化钉钉消息失败", zap.Error(err))
		return
	}

	global.Logger.Info("[通知] 准备发送钉钉消息",
		zap.String("title", msg.Markdown.Title),
		zap.Int("body_len", len(body)),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		global.Logger.Error("[通知] 发送钉钉消息失败",
			zap.String("title", msg.Markdown.Title),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

	if resp.StatusCode != http.StatusOK {
		global.Logger.Error("[通知] 钉钉返回非200状态码",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
			zap.String("title", msg.Markdown.Title),
		)
		return
	}

	global.Logger.Info("[通知] 钉钉消息发送成功",
		zap.String("title", msg.Markdown.Title),
		zap.String("response", string(respBody)),
	)
}

// sendFeishuNotify 发送飞书通知
func (s *Services) sendFeishuNotify(webhook string, msg *FeishuMessage) {
	if webhook == "" || msg == nil {
		global.Logger.Warn("[通知] sendFeishuNotify 跳过: webhook 或 msg 为空")
		return
	}

	body, err := json.Marshal(msg)
	if err != nil {
		global.Logger.Error("[通知] 序列化飞书消息失败", zap.Error(err))
		return
	}

	// 如果配置了签名密钥，生成签名
	secret := s.getFeishuSecret()
	finalURL := webhook
	if secret != "" {
		timestamp := time.Now().Unix()
		stringToSign := fmt.Sprintf("%d", timestamp)
		mac := hmac.New(sha256.New, []byte(stringToSign))
		mac.Write([]byte(secret))
		signData := mac.Sum(nil)
		signature := base64.StdEncoding.EncodeToString(signData)
		finalURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhook, timestamp, signature)
	}

	global.Logger.Info("[通知] 准备发送飞书消息",
		zap.String("title", msg.Content.Post.ZhCn.Title),
		zap.Int("body_len", len(body)),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(finalURL, "application/json", bytes.NewReader(body))
	if err != nil {
		global.Logger.Error("[通知] 发送飞书消息失败",
			zap.String("title", msg.Content.Post.ZhCn.Title),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

	if resp.StatusCode != http.StatusOK {
		global.Logger.Error("[通知] 飞书返回非200状态码",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return
	}

	global.Logger.Info("[通知] 飞书消息发送成功",
		zap.String("response", string(respBody)),
	)
}
