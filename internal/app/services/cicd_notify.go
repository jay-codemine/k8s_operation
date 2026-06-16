package services

import (
	"bytes"
	"context"
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

// NotifyBuildStarted 发送构建开始通知
func (s *Services) NotifyBuildStarted(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, buildNumber int) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildBuildStartedMessage(pipeline, run, buildNumber)
	go s.sendDingTalkNotify(webhook, msg)
}

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

// NotifyBuildResult 发送构建结果通知（异步）
func (s *Services) NotifyBuildResult(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildBuildNotifyMessage(pipeline, run, success)
	go s.sendDingTalkNotify(webhook, msg)
}

// NotifyBuildResultSync 发送构建结果通知（同步，用于后台 Worker 中确保通知可追踪）
func (s *Services) NotifyBuildResultSync(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, success bool) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		global.Logger.Warn("[通知] 钉钉 Webhook 未配置，跳过通知",
			zap.String("pipeline_name", pipeline.Name),
		)
		return
	}

	msg := s.buildBuildNotifyMessage(pipeline, run, success)
	s.sendDingTalkNotify(webhook, msg)
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

// notifyAutoDeployResult 发送自动部署结果通知（内部使用）
// 用于 Jenkins 回调后的自动部署场景
func (s *Services) notifyAutoDeployResult(ctx context.Context, pipeline *models.CicdPipeline, image string, success bool, errMsg string) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildAutoDeployNotifyMessage(pipeline, image, success, errMsg)
	s.sendDingTalkNotify(webhook, msg)
}

// buildAutoDeployNotifyMessage 构建自动部署通知消息
func (s *Services) buildAutoDeployNotifyMessage(pipeline *models.CicdPipeline, image string, success bool, errMsg string) *DingTalkMessage {
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

// notifyLegacyDeployResult 发送旧版配置自动部署结果通知
func (s *Services) notifyLegacyDeployResult(ctx context.Context, pipeline *models.CicdPipeline, namespace, deploymentName, image string, success bool, errMsg string) {
	webhook := s.getDingTalkWebhook(pipeline)
	if webhook == "" {
		return
	}

	msg := s.buildLegacyDeployNotifyMessage(pipeline, namespace, deploymentName, image, success, errMsg)
	s.sendDingTalkNotify(webhook, msg)
}

// buildLegacyDeployNotifyMessage 构建旧版配置部署通知消息
func (s *Services) buildLegacyDeployNotifyMessage(pipeline *models.CicdPipeline, namespace, deploymentName, image string, success bool, errMsg string) *DingTalkMessage {
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

// ==================== 消息构建 ====================

// buildBuildStartedMessage 构建构建开始通知消息
func (s *Services) buildBuildStartedMessage(pipeline *models.CicdPipeline, run *models.CicdPipelineRun, buildNumber int) *DingTalkMessage {
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

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("[构建开始] %s #%d", pipeline.Name, buildNumber),
			Text:  text,
		},
	}
}

func (s *Services) buildDeployNotifyMessage(pipeline *models.CicdPipeline, stage *models.CicdPipelineStage, success bool, errMsg string) *DingTalkMessage {
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

	// 添加快捷链接
	text += "\n\n---\n"
	if platformURL != "" {
		// 跳转到流水线详情页的执行阶段，并自动选中审批阶段
		text += fmt.Sprintf("✅ [点击进行审批](%s/cicd/pipelines/%d?tab=stages&auto_select=approval)\n\n", platformURL, pipeline.ID)
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
	// 如果已有环境配置，直接使用
	if env != "" {
		return s.getEnvDisplayName(env)
	}
	
	// 尝试从集群名称中解析环境
	if clusterID > 0 {
		var cluster models.K8sCluster
		if err := global.DB.Where("id = ?", clusterID).First(&cluster).Error; err == nil {
			clusterName := cluster.ClusterName
			// 从集群名称中识别环境关键词
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

	// 读取响应体，记录钉钉返回的完整信息
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
