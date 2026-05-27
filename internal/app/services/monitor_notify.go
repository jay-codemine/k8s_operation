package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ============================================================
// 通知渠道 CRUD
// ============================================================

// NotifyChannelListReq 通知渠道列表请求
type NotifyChannelListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Type    string `form:"type" json:"type"`
	Keyword string `form:"keyword" json:"keyword"`
}

// NotifyChannelListResp 通知渠道列表响应
type NotifyChannelListResp struct {
	Total int64                         `json:"total"`
	Items []models.MonitorNotifyChannel `json:"items"`
}

// ListNotifyChannels 列表
func (s *MonitorCRUDService) ListNotifyChannels(ctx context.Context, req NotifyChannelListReq) (*NotifyChannelListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Model(&models.MonitorNotifyChannel{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorNotifyChannel
	err := db.Order("id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &NotifyChannelListResp{Total: total, Items: items}, nil
}

// GetNotifyChannel 详情
func (s *MonitorCRUDService) GetNotifyChannel(ctx context.Context, id int64) (*models.MonitorNotifyChannel, error) {
	var ch models.MonitorNotifyChannel
	err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// CreateNotifyChannel 创建
func (s *MonitorCRUDService) CreateNotifyChannel(ctx context.Context, ch *models.MonitorNotifyChannel) error {
	return global.DB.WithContext(ctx).Create(ch).Error
}

// UpdateNotifyChannel 更新
func (s *MonitorCRUDService) UpdateNotifyChannel(ctx context.Context, ch *models.MonitorNotifyChannel) error {
	updates := map[string]interface{}{
		"name":             ch.Name,
		"type":             ch.Type,
		"description":      ch.Description,
		"webhook_url":      ch.WebhookURL,
		"security_keyword": ch.SecurityKeyword,
		"at_mobiles":       ch.AtMobiles,
		"at_all":           ch.AtAll,
		"smtp_host":        ch.SMTPHost,
		"smtp_port":        ch.SMTPPort,
		"smtp_user":        ch.SMTPUser,
		"smtp_to":          ch.SMTPTo,
		"msg_template":     ch.MsgTemplate,
		"enabled":          ch.Enabled,
		"send_resolved":    ch.SendResolved,
		"rate_limit":       ch.RateLimit,
	}
	if ch.Secret != "" {
		updates["secret"] = ch.Secret
	}
	if ch.SMTPPass != "" {
		updates["smtp_pass"] = ch.SMTPPass
	}

	return global.DB.WithContext(ctx).Model(ch).
		Where("id = ? AND is_del = 0", ch.ID).
		Updates(updates).Error
}

// DeleteNotifyChannel 删除（软删除）
func (s *MonitorCRUDService) DeleteNotifyChannel(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorNotifyChannel{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// TestNotifyChannel 测试通知渠道
func (s *MonitorCRUDService) TestNotifyChannel(ctx context.Context, ch *models.MonitorNotifyChannel) (bool, string) {
	// 构建测试告警；Summary 中主动带上第一个关键字，
	// 确保测试消息必然通过钉钉安全关键字校验
	summary := "这是一条测试通知，验证告警渠道配置是否正常"
	kwList := splitAndTrim(ch.SecurityKeyword)
	if len(kwList) > 0 {
		summary = fmt.Sprintf("[%s] %s", kwList[0], summary)
	}

	testAlert := &AlertNotification{
		RuleName:    "测试告警 - K8s Operation",
		Severity:    "warning",
		Status:      "firing",
		Summary:     summary,
		Description: "如果您看到此消息，说明告警渠道配置成功。请检查消息格式是否符合预期。",
		Value:       "85.6%",
		FiredAt:     time.Now().Unix(),
		Labels:      map[string]string{"env": "production", "instance": "node-01:9100", "cluster": "k8s-prod"},
	}

	err := SendNotification(ch, testAlert)
	if err != nil {
		return false, fmt.Sprintf("发送失败: %v", err)
	}

	switch len(kwList) {
	case 0:
		return true, "发送成功"
	case 1:
		return true, fmt.Sprintf("发送成功（安全关键字「%s」已注入消息）", kwList[0])
	default:
		return true, fmt.Sprintf("发送成功（配置了 %d 个安全关键字: %s，已注入第一个）", len(kwList), ch.SecurityKeyword)
	}
}

// ============================================================
// 通知发送器
// ============================================================

// AlertNotification 告警通知数据
type AlertNotification struct {
	RuleName    string            `json:"rule_name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"` // firing/resolved
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Value       string            `json:"value"`
	FiredAt     int64             `json:"fired_at"`
	ResolvedAt  int64             `json:"resolved_at"`
	Labels      map[string]string `json:"labels"`
}

// SendNotification 根据渠道类型发送通知
func SendNotification(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	switch ch.Type {
	case "dingtalk":
		return sendDingTalk(ch, alert)
	case "feishu":
		return sendFeishu(ch, alert)
	case "webhook":
		return sendWebhook(ch, alert)
	case "wechat":
		return sendWechat(ch, alert)
	case "email":
		return sendEmail(ch, alert)
	default:
		return fmt.Errorf("不支持的通知类型: %s", ch.Type)
	}
}

// sendDingTalk 发送钉钉通知（企业级告警卡片 — 全中文大厂风格）
func sendDingTalk(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	// 自定义模板优先
	if ch.MsgTemplate != "" {
		rendered, err := renderNotifyTemplate(ch.MsgTemplate, alert)
		if err != nil {
			return err
		}
		title := "🚨 [告警触发] " + alert.RuleName
		if alert.Status == "resolved" {
			title = "✅ [已恢复] " + alert.RuleName
		}
		text := ensureDingTalkKeyword(rendered, ch.SecurityKeyword)
		body := map[string]interface{}{
			"msgtype":  "markdown",
			"markdown": map[string]string{"title": title, "text": text},
			"at":       buildAtConfig(ch),
		}
		return postJSON(ch.WebhookURL, ch.Secret, body)
	}

	isFiring := alert.Status != "resolved"
	firedTime := formatUnixTime(alert.FiredAt)

	var md strings.Builder

	if isFiring {
		// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		// 🔥 告警触发通知（红色主题 · 全中文）
		// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		severityIcon := map[string]string{"critical": "🔴", "warning": "🟡", "info": "🔵"}
		sIcon := severityIcon[alert.Severity]
		if sIcon == "" {
			sIcon = "⚪"
		}
		severityTag := map[string]string{"critical": "P0 · 紧急", "warning": "P1 · 警告", "info": "P2 · 提示"}
		sTag := severityTag[alert.Severity]
		if sTag == "" {
			sTag = alert.Severity
		}

		md.WriteString("## 🚨 告警触发\n\n")
		md.WriteString(fmt.Sprintf("### %s\n\n", alert.RuleName))
		md.WriteString(fmt.Sprintf("> <font color=#FF4D4F>**状态**</font> 🔴 触发中　　<font color=#FF4D4F>**级别**</font> %s %s\n\n", sIcon, sTag))

		md.WriteString("---\n\n")

		// 告警摘要区
		if alert.Summary != "" {
			md.WriteString(fmt.Sprintf("**📋 告警摘要**\n\n%s\n\n", alert.Summary))
		}
		if alert.Description != "" {
			md.WriteString(fmt.Sprintf("**📝 详细描述**\n\n> %s\n\n", alert.Description))
		}

		// 指标数据区
		md.WriteString("---\n\n")
		md.WriteString("**📊 监控数据**\n\n")
		if alert.Value != "" {
			md.WriteString(fmt.Sprintf("- **触发值**　`%s`\n", alert.Value))
		}
		md.WriteString(fmt.Sprintf("- **触发时间**　%s\n", firedTime))

		// 环境标签区
		if len(alert.Labels) > 0 {
			md.WriteString("\n---\n\n")
			md.WriteString("**🏷️ 环境信息**\n\n")
			keyOrder := []string{"cluster", "namespace", "env", "environment", "node", "instance", "pod", "container", "service", "daemonset", "deployment", "statefulset", "job", "alertname", "prometheus", "endpoint"}
			shown := map[string]bool{}
			for _, k := range keyOrder {
				if v, exist := alert.Labels[k]; exist {
					displayVal := v
					if k == "alertname" {
						displayVal = alertNameAlias(v)
					}
					md.WriteString(fmt.Sprintf("- **%s**　`%s`\n", labelAlias(k), displayVal))
					shown[k] = true
				}
			}
			for k, v := range alert.Labels {
				if !shown[k] {
					md.WriteString(fmt.Sprintf("- **%s**　`%s`\n", labelAlias(k), v))
				}
			}
		}

		// 底部
		md.WriteString("\n---\n\n")
		md.WriteString("<font color=#8C8C8C>🛡️ K8s Operation · 智能监控平台</font>\n")

	} else {
		// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		// ✅ 告警恢复通知（绿色主题 · 全中文）
		// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
		md.WriteString("## ✅ 告警恢复\n\n")
		md.WriteString(fmt.Sprintf("### %s\n\n", alert.RuleName))
		md.WriteString("> <font color=#52C41A>**状态**</font> 🟢 已恢复　　告警已自动恢复正常\n\n")

		md.WriteString("---\n\n")

		if alert.Summary != "" {
			md.WriteString(fmt.Sprintf("**📋 原始摘要**\n\n%s\n\n", alert.Summary))
		}

		// 恢复时间线
		md.WriteString("---\n\n")
		md.WriteString("**⏱️ 恢复时间线**\n\n")
		md.WriteString(fmt.Sprintf("- **触发时间**　%s\n", firedTime))
		if alert.ResolvedAt > 0 {
			resolvedTime := formatUnixTime(alert.ResolvedAt)
			dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
			md.WriteString(fmt.Sprintf("- **恢复时间**　%s\n", resolvedTime))
			md.WriteString(fmt.Sprintf("- **持续时长**　%s\n", dur.Round(time.Second)))
		}

		// 环境标签
		if len(alert.Labels) > 0 {
			md.WriteString("\n---\n\n")
			md.WriteString("**🏷️ 环境信息**\n\n")
			keyOrder := []string{"cluster", "namespace", "env", "environment", "node", "instance", "pod", "container", "service", "alertname", "prometheus"}
			shown := map[string]bool{}
			for _, k := range keyOrder {
				if v, exist := alert.Labels[k]; exist {
					displayVal := v
					if k == "alertname" {
						displayVal = alertNameAlias(v)
					}
					md.WriteString(fmt.Sprintf("- **%s**　`%s`\n", labelAlias(k), displayVal))
					shown[k] = true
				}
			}
			for k, v := range alert.Labels {
				if !shown[k] {
					md.WriteString(fmt.Sprintf("- **%s**　`%s`\n", labelAlias(k), v))
				}
			}
		}

		// 底部
		md.WriteString("\n---\n\n")
		md.WriteString("<font color=#8C8C8C>🛡️ K8s Operation · 智能监控平台</font>\n")
	}

	// 构建最终 title（会话列表预览 · 全中文）
	title := "🚨 [告警触发] " + alert.RuleName
	if !isFiring {
		title = "✅ [已恢复] " + alert.RuleName
	}

	text := ensureDingTalkKeyword(md.String(), ch.SecurityKeyword)
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
		"at": buildAtConfig(ch),
	}
	return postJSON(ch.WebhookURL, ch.Secret, body)
}

// sendFeishu 发送飞书通知（富文本 post 卡片）
func sendFeishu(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	if ch.MsgTemplate != "" {
		content, err := renderNotifyTemplate(ch.MsgTemplate, alert)
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"msg_type": "text",
			"content":  map[string]string{"text": content},
		}
		return postJSON(ch.WebhookURL, ch.Secret, body)
	}

	isFiring := alert.Status != "resolved"
	statusIcon := "🚨"
	statusText := "告警触发"
	if !isFiring {
		statusIcon = "✅"
		statusText = "告警恢复"
	}

	severityBadgeMap := map[string]string{
		"critical": "P0-紧急",
		"warning":  "P1-警告",
		"info":     "P2-提示",
	}
	badge, ok := severityBadgeMap[alert.Severity]
	if !ok {
		badge = alert.Severity
	}
	firedTime := formatUnixTime(alert.FiredAt)

	// 构建飞书 post 富文本（每个 []interface{} 为一行内联元素列表）
	newLine := func(text string) []interface{} {
		return []interface{}{map[string]interface{}{"tag": "text", "text": text}}
	}
	sep := newLine("──────────────────────")

	var lines []interface{}
	lines = append(lines, newLine(fmt.Sprintf("%s %s  ·  %s  ·  %s", statusIcon, statusText, badge, firedTime)))
	lines = append(lines, sep)
	if alert.Summary != "" {
		lines = append(lines, newLine(fmt.Sprintf("📋 摘要：%s", alert.Summary)))
	}
	if alert.Description != "" {
		lines = append(lines, newLine(fmt.Sprintf("📝 详情：%s", alert.Description)))
	}
	if alert.Value != "" {
		lines = append(lines, newLine(fmt.Sprintf("📊 当前值：%s", alert.Value)))
	}
	if len(alert.Labels) > 0 {
		lines = append(lines, sep)
		keyOrder := []string{"cluster", "namespace", "env", "environment", "node", "instance", "pod", "container", "service", "daemonset", "deployment", "job", "alertname", "prometheus"}
		shown := map[string]bool{}
		for _, k := range keyOrder {
			if v, exist := alert.Labels[k]; exist {
				displayVal := v
				if k == "alertname" {
					displayVal = alertNameAlias(v)
				}
				lines = append(lines, newLine(fmt.Sprintf("  %s：%s", labelAlias(k), displayVal)))
				shown[k] = true
			}
		}
		for k, v := range alert.Labels {
			if !shown[k] {
				lines = append(lines, newLine(fmt.Sprintf("  %s：%s", labelAlias(k), v)))
			}
		}
	}
	if !isFiring && alert.ResolvedAt > 0 {
		resolvedTime := formatUnixTime(alert.ResolvedAt)
		dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
		lines = append(lines, sep)
		lines = append(lines, newLine(fmt.Sprintf("⏱ 持续时长：%s  ✅ 恢复时间：%s", dur.Round(time.Second), resolvedTime)))
	}
	lines = append(lines, sep)
	lines = append(lines, newLine(fmt.Sprintf("🛡 K8s Operation 智能监控平台  ·  ⏰ %s", firedTime)))

	title := fmt.Sprintf("%s [%s] %s", statusIcon, statusText, alert.RuleName)
	body := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": lines,
				},
			},
		},
	}
	return postJSON(ch.WebhookURL, ch.Secret, body)
}

// sendWechat 发送企业微信通知（大厂风格 Markdown 卡片）
func sendWechat(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	if ch.MsgTemplate != "" {
		content, err := renderNotifyTemplate(ch.MsgTemplate, alert)
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"msgtype":  "markdown",
			"markdown": map[string]string{"content": content},
		}
		return postJSON(ch.WebhookURL, "", body)
	}

	isFiring := alert.Status != "resolved"
	firedTime := formatUnixTime(alert.FiredAt)

	var md strings.Builder

	if isFiring {
		// 告警触发 · 全中文
		severityMap := map[string]string{
			"critical": "<font color=\"warning\">🔴 P0 · 紧急</font>",
			"warning":  "<font color=\"warning\">🟡 P1 · 警告</font>",
			"info":     "<font color=\"comment\">🔵 P2 · 提示</font>",
		}
		badge, ok := severityMap[alert.Severity]
		if !ok {
			badge = alert.Severity
		}

		md.WriteString(fmt.Sprintf("## 🚨 告警触发\n"))
		md.WriteString(fmt.Sprintf("### %s\n", alert.RuleName))
		md.WriteString(fmt.Sprintf("> <font color=\"warning\">**触发中**</font>　%s　%s\n\n", badge, firedTime))
		md.WriteString("---\n")
		if alert.Summary != "" {
			md.WriteString(fmt.Sprintf("> **📋 摘要**: %s\n", alert.Summary))
		}
		if alert.Description != "" {
			md.WriteString(fmt.Sprintf("> **📝 详情**: %s\n", alert.Description))
		}
		if alert.Value != "" {
			md.WriteString(fmt.Sprintf("> **📊 触发值**: `%s`\n", alert.Value))
		}
		md.WriteString("\n")

		if len(alert.Labels) > 0 {
			md.WriteString("---\n")
			md.WriteString("**🏷️ 环境信息**\n")
			keyOrder := []string{"cluster", "namespace", "env", "environment", "node", "instance", "pod", "container", "service", "daemonset", "deployment", "statefulset", "job", "alertname", "prometheus", "endpoint"}
			shown := map[string]bool{}
			for _, k := range keyOrder {
				if v, exist := alert.Labels[k]; exist {
					displayVal := v
					if k == "alertname" {
						displayVal = alertNameAlias(v)
					}
					md.WriteString(fmt.Sprintf("> **%s**: `%s`\n", labelAlias(k), displayVal))
					shown[k] = true
				}
			}
			for k, v := range alert.Labels {
				if !shown[k] {
					md.WriteString(fmt.Sprintf("> **%s**: `%s`\n", labelAlias(k), v))
				}
			}
			md.WriteString("\n")
		}
	} else {
		// 告警恢复 · 全中文
		md.WriteString(fmt.Sprintf("## ✅ 告警恢复\n"))
		md.WriteString(fmt.Sprintf("### %s\n", alert.RuleName))
		md.WriteString(fmt.Sprintf("> <font color=\"info\">**已恢复**</font>　告警已自动恢复正常\n\n"))
		md.WriteString("---\n")
		if alert.Summary != "" {
			md.WriteString(fmt.Sprintf("> **📋 原始摘要**: %s\n", alert.Summary))
		}
		md.WriteString("\n---\n")
		md.WriteString("**⏱️ 恢复时间线**\n")
		md.WriteString(fmt.Sprintf("> **触发时间**: %s\n", firedTime))
		if alert.ResolvedAt > 0 {
			resolvedTime := formatUnixTime(alert.ResolvedAt)
			dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
			md.WriteString(fmt.Sprintf("> **恢复时间**: %s\n", resolvedTime))
			md.WriteString(fmt.Sprintf("> **持续时长**: %s\n", dur.Round(time.Second)))
		}
		md.WriteString("\n")
	}

	// 底部签名
	md.WriteString("---\n")
	md.WriteString("🛡️ K8s Operation · 智能监控平台")

	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": md.String(),
		},
	}
	return postJSON(ch.WebhookURL, "", body)
}

// sendWebhook 发送通用 Webhook 通知
func sendWebhook(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	if ch.MsgTemplate != "" {
		content, err := renderNotifyTemplate(ch.MsgTemplate, alert)
		if err != nil {
			return err
		}
		return postJSON(ch.WebhookURL, "", map[string]string{"text": content})
	}
	return postJSON(ch.WebhookURL, "", alert)
}

// sendEmail 发送邮件通知
func sendEmail(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	if ch.SMTPHost == "" || ch.SMTPUser == "" || ch.SMTPPass == "" || ch.SMTPTo == "" {
		return fmt.Errorf("邮件渠道 SMTP 配置不完整")
	}

	port := ch.SMTPPort
	if port == 0 {
		port = 465
	}

	recipients := splitAndTrim(ch.SMTPTo)
	if len(recipients) == 0 {
		return fmt.Errorf("邮件收件人为空")
	}

	// 邮件主题 · 全中文
	severityCN := map[string]string{"critical": "紧急", "warning": "警告", "info": "提示"}
	sLabel := severityCN[alert.Severity]
	if sLabel == "" {
		sLabel = alert.Severity
	}
	subject := fmt.Sprintf("[%s] %s", sLabel, alert.RuleName)
	if alert.Status == "resolved" {
		subject = fmt.Sprintf("[已恢复][%s] %s", sLabel, alert.RuleName)
	}

	firedTime := time.Unix(alert.FiredAt, 0).Format("2006-01-02 15:04:05")
	var body strings.Builder
	if ch.MsgTemplate != "" {
		rendered, err := renderNotifyTemplate(ch.MsgTemplate, alert)
		if err != nil {
			return err
		}
		body.WriteString(rendered)
	} else {
		// 全中文邮件正文
		statusCN := "🔴 触发中"
		if alert.Status == "resolved" {
			statusCN = "🟢 已恢复"
		}
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		body.WriteString(fmt.Sprintf("  规则名称: %s\n", alert.RuleName))
		body.WriteString(fmt.Sprintf("  当前状态: %s\n", statusCN))
		body.WriteString(fmt.Sprintf("  告警级别: %s\n", severityText(alert.Severity)))
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		if alert.Summary != "" {
			body.WriteString(fmt.Sprintf("📋 告警摘要: %s\n", alert.Summary))
		}
		if alert.Description != "" {
			body.WriteString(fmt.Sprintf("📝 详细描述: %s\n", alert.Description))
		}
		if alert.Value != "" {
			body.WriteString(fmt.Sprintf("📊 触发值: %s\n", alert.Value))
		}
		body.WriteString(fmt.Sprintf("\n⏰ 触发时间: %s\n", firedTime))
		if alert.Status == "resolved" && alert.ResolvedAt > 0 {
			body.WriteString(fmt.Sprintf("✅ 恢复时间: %s\n", time.Unix(alert.ResolvedAt, 0).Format("2006-01-02 15:04:05")))
			dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
			body.WriteString(fmt.Sprintf("⏱️ 持续时长: %s\n", dur.Round(time.Second)))
		}
		if len(alert.Labels) > 0 {
			body.WriteString("\n环境标签:\n")
			for k, v := range alert.Labels {
				displayVal := v
				if k == "alertname" {
					displayVal = alertNameAlias(v)
				}
				body.WriteString(fmt.Sprintf("- %s = %s\n", labelAlias(k), displayVal))
			}
		}
		body.WriteString("\n--\nK8s Operation 智能监控平台\n")
	}

	addr := fmt.Sprintf("%s:%d", ch.SMTPHost, port)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		ch.SMTPUser,
		strings.Join(recipients, ","),
		subject,
		body.String(),
	))

	auth := smtp.PlainAuth("", ch.SMTPUser, ch.SMTPPass, ch.SMTPHost)
	if port == 465 {
		return sendMailTLS(addr, ch.SMTPHost, auth, ch.SMTPUser, recipients, msg)
	}
	return smtp.SendMail(addr, auth, ch.SMTPUser, recipients, msg)
}

type notifyTemplateData struct {
	RuleName     string
	Severity     string
	SeverityText string
	Status       string
	StatusText   string
	Summary      string
	Description  string
	Value        string
	FiredAt      string
	ResolvedAt   string
	Labels       map[string]string
	LabelsText   string
	Platform     string
}

func renderNotifyTemplate(tpl string, alert *AlertNotification) (string, error) {
	t, err := template.New("notify").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("消息模板解析失败: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, buildNotifyTemplateData(alert)); err != nil {
		return "", fmt.Errorf("消息模板渲染失败: %w", err)
	}
	return buf.String(), nil
}

func buildNotifyTemplateData(alert *AlertNotification) notifyTemplateData {
	resolvedAt := ""
	if alert.ResolvedAt > 0 {
		resolvedAt = formatUnixTime(alert.ResolvedAt)
	}
	return notifyTemplateData{
		RuleName:     alert.RuleName,
		Severity:     alert.Severity,
		SeverityText: severityText(alert.Severity),
		Status:       alert.Status,
		StatusText:   statusText(alert.Status),
		Summary:      alert.Summary,
		Description:  alert.Description,
		Value:        alert.Value,
		FiredAt:      formatUnixTime(alert.FiredAt),
		ResolvedAt:   resolvedAt,
		Labels:       alert.Labels,
		LabelsText:   formatLabels(alert.Labels),
		Platform:     "K8s Operation",
	}
}

// ensureDingTalkKeyword 确保消息内容包含至少一个安全关键字。
// keywords 支持逗号分隔的多个关键字（对应钉钉「自定义关键词」安全设置，最多10个）。
// 逻辑：若消息已包含任意一个关键字则直接返回；否则将第一个关键字注入到消息头部。
func ensureDingTalkKeyword(content, keywords string) string {
	keywords = strings.TrimSpace(keywords)
	if keywords == "" {
		return content
	}

	kwList := splitAndTrim(keywords)
	// 检查是否已包含任意一个关键字（大小写敏感，与钉钉行为一致）
	for _, kw := range kwList {
		if strings.Contains(content, kw) {
			return content
		}
	}

	// 无匹配，将第一个关键字注入到消息头部
	if len(kwList) > 0 {
		return kwList[0] + "\n\n" + content
	}
	return content
}

func severityText(severity string) string {
	switch severity {
	case "critical":
		return "P0-紧急"
	case "warning":
		return "P1-警告"
	case "info":
		return "P2-提示"
	default:
		return severity
	}
}

func statusText(status string) string {
	if status == "resolved" {
		return "告警恢复"
	}
	return "告警触发"
}

func formatUnixTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	var builder strings.Builder
	for k, v := range labels {
		builder.WriteString(fmt.Sprintf("- %s = %s\n", k, v))
	}
	return strings.TrimSpace(builder.String())
}

func sendMailTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		writer.Close()
		return err
	}
	return writer.Close()
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// labelAlias 将常见 Prometheus 标签 key 映射为中文别名
func labelAlias(k string) string {
	aliases := map[string]string{
		// 基础环境
		"cluster":     "集群",
		"namespace":   "命名空间",
		"env":         "环境",
		"environment": "环境",
		"region":      "地域",
		"zone":        "可用区",
		"dc":          "数据中心",
		// 采集相关
		"job":      "采集任务",
		"instance": "实例",
		"endpoint": "端点",
		// 节点
		"node":     "节点",
		"nodename": "节点名",
		// 工作负载
		"pod":                       "Pod",
		"container":                 "容器",
		"service":                   "服务",
		"deployment":                "部署",
		"daemonset":                 "守护进程集",
		"statefulset":               "有状态副本集",
		"replicaset":                "副本集",
		"cronjob":                   "定时任务",
		"job_name":                  "任务名",
		"horizontalpodautoscaler":   "弹性伸缩",
		"persistentvolumeclaim":     "持久卷声明",
		"persistentvolume":          "持久卷",
		"ingress":                   "入口网关",
		"configmap":                 "配置字典",
		"secret":                    "密钥对象",
		// 告警元数据
		"alertname":  "告警规则",
		"severity":   "级别",
		"prometheus": "数据源",
		"replica":    "副本",
		// 应用层
		"handler":    "接口路径",
		"method":     "请求方法",
		"code":       "状态码",
		"status":     "状态",
		"path":       "路径",
		"host":       "主机",
		"url":        "地址",
		"protocol":   "协议",
		"port":       "端口",
		"device":     "磁盘设备",
		"mountpoint": "挂载点",
		"fstype":     "文件系统",
		"interface":  "网卡",
		"mode":       "模式",
		"type":       "类型",
		"reason":     "原因",
		"phase":      "阶段",
		"condition":  "条件",
		"resource":   "资源",
		"unit":       "单元",
		"le":         "分位桶",
		"quantile":   "分位数",
	}
	if alias, ok := aliases[k]; ok {
		return alias
	}
	return k
}

// alertNameAlias 将常见 Prometheus/kube-prometheus-stack 告警名翻译为中文
// 用于通知展示时让标签更可读
func alertNameAlias(name string) string {
	aliases := map[string]string{
		"KubeDaemonSetRolloutStuck":        "DaemonSet滚动更新卡住",
		"KubeDeploymentReplicasMismatch":   "Deployment副本数不匹配",
		"KubeStatefulSetReplicasMismatch":  "StatefulSet副本数不匹配",
		"KubePodNotReady":                  "Pod未就绪",
		"KubePodCrashLooping":              "Pod频繁重启",
		"KubeJobFailed":                    "Job执行失败",
		"KubeJobNotCompleted":              "Job未完成",
		"KubeHpaReplicasMismatch":          "HPA副本数不匹配",
		"KubeHpaMaxedOut":                  "HPA已达最大副本数",
		"KubeCPUOvercommit":                "CPU资源超分配",
		"KubeMemoryOvercommit":             "内存资源超分配",
		"KubeQuotaAlmostFull":              "资源配额即将用尽",
		"KubeQuotaFullyUsed":               "资源配额已用尽",
		"KubeQuotaExceeded":                "资源配额超限",
		"KubePersistentVolumeFillingUp":    "持久卷即将满载",
		"KubePersistentVolumeInodesFillingUp": "持久卷Inode即将用尽",
		"KubeContainerWaiting":             "容器等待中",
		"KubeNodeNotReady":                 "节点未就绪",
		"KubeNodeUnreachable":              "节点不可达",
		"KubeNodeReadinessFlapping":        "节点就绪状态抖动",
		"KubeClientErrors":                 "K8s客户端错误",
		"KubeAPILatencyHigh":               "K8s API延迟过高",
		"KubeletTooManyPods":               "Kubelet Pod数过多",
		"NodeFilesystemAlmostOutOfSpace":   "节点磁盘空间不足",
		"NodeFilesystemSpaceFillingUp":     "节点磁盘空间即将用尽",
		"NodeRAIDDegraded":                 "节点RAID降级",
		"NodeClockSkewDetected":            "节点时钟偏移",
		"NodeNetworkInterfaceFlapping":     "网卡状态抖动",
		"NodeHighNumberConntrackEntriesUsed": "连接跟踪表即将满",
		"TargetDown":                       "监控目标离线",
		"Watchdog":                         "监控看门狗",
		"PrometheusJobMissing":             "Prometheus采集任务丢失",
		"AlertmanagerFailedReload":         "Alertmanager重载失败",
		"AlertmanagerMembersInconsistent":  "Alertmanager集群不一致",
		"PrometheusTargetLimitHit":         "Prometheus目标数达上限",
		"PrometheusTSDBCompactionsFailing": "Prometheus压缩失败",
		"PrometheusRuleFailures":           "Prometheus规则执行失败",
	}
	if alias, ok := aliases[name]; ok {
		return alias
	}
	return name
}

// buildAtConfig 构建钉钉 @人员配置
func buildAtConfig(ch *models.MonitorNotifyChannel) map[string]interface{} {
	at := map[string]interface{}{"isAtAll": ch.AtAll}
	if ch.AtMobiles != "" {
		at["atMobiles"] = strings.Split(ch.AtMobiles, ",")
	}
	return at
}

// postJSON 发送 JSON POST 请求，并解析钉钉/飞书/企业微信的业务错误码
func postJSON(url, secret string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 如果有 secret，计算钉钉签名并追加到 URL
	if secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		stringToSign := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(stringToSign))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		url = fmt.Sprintf("%s%stimestamp=%s&sign=%s", url, sep, timestamp, sign)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体（上限 2KB，避免超大响应）
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 钉钉/飞书/企业微信即使 HTTP 200 也会在 body 里返回业务错误码
	// 钉钉: {"errcode":0,"errmsg":"ok"}
	// 飞书: {"code":0,"msg":"success"}
	// 企业微信: {"errcode":0,"errmsg":"ok"}
	var apiResp struct {
		ErrCode int    `json:"errcode"` // 钉钉 / 企业微信
		ErrMsg  string `json:"errmsg"`
		Code    int    `json:"code"` // 飞书
		Msg     string `json:"msg"`
	}
	if jsonErr := json.Unmarshal(respBody, &apiResp); jsonErr == nil {
		// 钉钉 / 企业微信
		if apiResp.ErrCode != 0 {
			hint := dingTalkErrHint(apiResp.ErrCode)
			if hint != "" {
				return fmt.Errorf("API错误 errcode=%d (%s): %s", apiResp.ErrCode, hint, apiResp.ErrMsg)
			}
			return fmt.Errorf("API错误 errcode=%d: %s", apiResp.ErrCode, apiResp.ErrMsg)
		}
		// 飞书
		if apiResp.Code != 0 {
			return fmt.Errorf("API错误 code=%d: %s", apiResp.Code, apiResp.Msg)
		}
	}
	return nil
}

// dingTalkErrHint 返回钉钉常见错误码的中文说明
func dingTalkErrHint(code int) string {
	switch code {
	case 310000:
		return "安全验证失败：消息内容不含安全关键字，请检查「安全关键字」配置"
	case 310001:
		return "消息被频率限制拦截，稍后重试"
	case 300001:
		return "Webhook access_token 无效或已过期"
	case 130101:
		return "消息体格式错误"
	default:
		return ""
	}
}

// ============================================================
// 通知路由策略 CRUD
// 大厂级设计：多策略优先级路由，自动将告警分发到不同群
// ============================================================

// RoutePolicyListReq 路由策略列表请求
type RoutePolicyListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Keyword string `form:"keyword" json:"keyword"`
}

// RoutePolicyListResp 路由策略列表响应
type RoutePolicyListResp struct {
	Total int64                              `json:"total"`
	Items []models.MonitorNotifyRoutePolicy `json:"items"`
}

// ListRoutePolicies 列表查询
func (s *MonitorCRUDService) ListRoutePolicies(ctx context.Context, req RoutePolicyListReq) (*RoutePolicyListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Model(&models.MonitorNotifyRoutePolicy{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorNotifyRoutePolicy
	err := db.Order("priority ASC, id ASC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &RoutePolicyListResp{Total: total, Items: items}, nil
}

// GetRoutePolicy 详情
func (s *MonitorCRUDService) GetRoutePolicy(ctx context.Context, id int64) (*models.MonitorNotifyRoutePolicy, error) {
	var policy models.MonitorNotifyRoutePolicy
	err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// CreateRoutePolicy 创建路由策略
func (s *MonitorCRUDService) CreateRoutePolicy(ctx context.Context, policy *models.MonitorNotifyRoutePolicy) error {
	// 如果设为默认策略，取消其他默认
	if policy.IsDefault {
		global.DB.WithContext(ctx).Model(&models.MonitorNotifyRoutePolicy{}).
			Where("is_default = 1 AND is_del = 0").
			Update("is_default", false)
	}
	return global.DB.WithContext(ctx).Create(policy).Error
}

// UpdateRoutePolicy 更新路由策略
func (s *MonitorCRUDService) UpdateRoutePolicy(ctx context.Context, policy *models.MonitorNotifyRoutePolicy) error {
	// 如果设为默认策略，取消其他默认
	if policy.IsDefault {
		global.DB.WithContext(ctx).Model(&models.MonitorNotifyRoutePolicy{}).
			Where("is_default = 1 AND is_del = 0 AND id != ?", policy.ID).
			Update("is_default", false)
	}
	return global.DB.WithContext(ctx).Model(policy).
		Where("id = ? AND is_del = 0", policy.ID).
		Updates(map[string]interface{}{
			"name":        policy.Name,
			"description": policy.Description,
			"priority":    policy.Priority,
			"channel_ids": policy.ChannelIDs,
			"match_mode":  policy.MatchMode,
			"severities":  policy.Severities,
			"groups":      policy.Groups,
			"label_match": policy.LabelMatch,
			"is_default":  policy.IsDefault,
			"enabled":     policy.Enabled,
		}).Error
}

// DeleteRoutePolicy 删除路由策略
func (s *MonitorCRUDService) DeleteRoutePolicy(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorNotifyRoutePolicy{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// ResolveRoutePolicyChannels 根据路由策略解析告警规则应发送到的渠道
// 用于 Worker 评估时：当 rule.NotifyChannels 为空时，通过路由策略自动匹配
func (s *MonitorCRUDService) ResolveRoutePolicyChannels(ctx context.Context, rule *models.MonitorAlertRule) string {
	var policies []models.MonitorNotifyRoutePolicy
	global.DB.WithContext(ctx).
		Where("enabled = 1 AND is_del = 0").
		Order("priority ASC, id ASC").
		Find(&policies)

	if len(policies) == 0 {
		return ""
	}

	// 解析规则标签
	labels := make(map[string]string)
	if rule.Labels != "" {
		_ = json.Unmarshal([]byte(rule.Labels), &labels)
	}

	return s.matchRoutePolicy(policies, rule.Severity, rule.Group, labels)
}
