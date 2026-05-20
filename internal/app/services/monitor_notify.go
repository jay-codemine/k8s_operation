package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	return global.DB.WithContext(ctx).Model(ch).
		Where("id = ? AND is_del = 0", ch.ID).
		Updates(map[string]interface{}{
			"name":          ch.Name,
			"type":          ch.Type,
			"description":   ch.Description,
			"webhook_url":   ch.WebhookURL,
			"secret":        ch.Secret,
			"at_mobiles":    ch.AtMobiles,
			"at_all":        ch.AtAll,
			"smtp_host":     ch.SMTPHost,
			"smtp_port":     ch.SMTPPort,
			"smtp_user":     ch.SMTPUser,
			"smtp_pass":     ch.SMTPPass,
			"smtp_to":       ch.SMTPTo,
			"msg_template":  ch.MsgTemplate,
			"enabled":       ch.Enabled,
			"send_resolved": ch.SendResolved,
			"rate_limit":    ch.RateLimit,
		}).Error
}

// DeleteNotifyChannel 删除（软删除）
func (s *MonitorCRUDService) DeleteNotifyChannel(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorNotifyChannel{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// TestNotifyChannel 测试通知渠道
func (s *MonitorCRUDService) TestNotifyChannel(ctx context.Context, ch *models.MonitorNotifyChannel) (bool, string) {
	testAlert := &AlertNotification{
		RuleName:    "测试告警 - K8s Operation",
		Severity:    "warning",
		Status:      "firing",
		Summary:     "这是一条测试通知，验证告警渠道配置是否正常",
		Description: "如果您看到此消息，说明告警渠道配置成功",
		Value:       "85.6%",
		FiredAt:     time.Now().Unix(),
		Labels:      map[string]string{"env": "production", "instance": "node-01:9100", "cluster": "k8s-prod"},
	}

	err := SendNotification(ch, testAlert)
	if err != nil {
		return false, fmt.Sprintf("发送失败: %v", err)
	}
	return true, "发送成功"
}

// ============================================================
// 通知发送器
// ============================================================

// AlertNotification 告警通知数据
type AlertNotification struct {
	RuleName    string            `json:"rule_name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`       // firing/resolved
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
	default:
		return fmt.Errorf("不支持的通知类型: %s", ch.Type)
	}
}

// sendDingTalk 发送钉钉通知（大厂风格 ActionCard 模板）
func sendDingTalk(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	statusEmoji := "🔥"
	statusText := "告警触发"
	statusColor := "#FF4D4F"
	if alert.Status == "resolved" {
		statusEmoji = "✅"
		statusText = "告警恢复"
		statusColor = "#52C41A"
	}

	severityMap := map[string]string{
		"critical": "🔴 P0-Critical",
		"warning":  "🟡 P1-Warning",
		"info":     "🔵 P2-Info",
	}

	title := fmt.Sprintf("%s [%s] %s", statusEmoji, strings.ToUpper(alert.Status), alert.RuleName)
	firedTime := time.Unix(alert.FiredAt, 0).Format("2006-01-02 15:04:05")

	var mdBuilder strings.Builder
	// 头部分割线 + 状态块
	mdBuilder.WriteString(fmt.Sprintf("### %s %s\n\n", statusEmoji, alert.RuleName))
	mdBuilder.WriteString("---\n\n")
	mdBuilder.WriteString(fmt.Sprintf("> **状态**: <font color=%s>%s</font>\n\n", statusColor, statusText))
	mdBuilder.WriteString(fmt.Sprintf("> **级别**: %s\n\n", severityMap[alert.Severity]))

	// 核心信息区
	if alert.Summary != "" {
		mdBuilder.WriteString(fmt.Sprintf("> **摘要**: %s\n\n", alert.Summary))
	}
	if alert.Description != "" {
		mdBuilder.WriteString(fmt.Sprintf("> **详情**: %s\n\n", alert.Description))
	}
	if alert.Value != "" {
		mdBuilder.WriteString(fmt.Sprintf("> **触发值**: `%s`\n\n", alert.Value))
	}

	// 时间信息
	mdBuilder.WriteString("---\n\n")
	mdBuilder.WriteString(fmt.Sprintf("⏰ **触发时间**: %s\n\n", firedTime))
	if alert.Status == "resolved" && alert.ResolvedAt > 0 {
		resolvedTime := time.Unix(alert.ResolvedAt, 0).Format("2006-01-02 15:04:05")
		duration := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
		mdBuilder.WriteString(fmt.Sprintf("✅ **恢复时间**: %s（持续 %s）\n\n", resolvedTime, duration.Round(time.Second)))
	}

	// 标签信息
	if len(alert.Labels) > 0 {
		mdBuilder.WriteString("**🏷️ 标签信息:**\n\n")
		for k, v := range alert.Labels {
			mdBuilder.WriteString(fmt.Sprintf("- `%s` = `%s`\n", k, v))
		}
		mdBuilder.WriteString("\n")
	}

	// 底部签名
	mdBuilder.WriteString("---\n\n")
	mdBuilder.WriteString("🛡️ K8s Operation 监控平台 | 自动告警\n")

	// 构造请求体
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  mdBuilder.String(),
		},
	}

	// @指定人
	at := map[string]interface{}{
		"isAtAll": ch.AtAll,
	}
	if ch.AtMobiles != "" {
		at["atMobiles"] = strings.Split(ch.AtMobiles, ",")
	}
	body["at"] = at

	return postJSON(ch.WebhookURL, ch.Secret, body)
}

// sendFeishu 发送飞书通知（富文本卡片）
func sendFeishu(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	statusEmoji := "🔥"
	statusText := "告警触发"
	if alert.Status == "resolved" {
		statusEmoji = "✅"
		statusText = "告警恢复"
	}

	severityMap := map[string]string{
		"critical": "🔴 P0-Critical",
		"warning":  "🟡 P1-Warning",
		"info":     "🔵 P2-Info",
	}
	firedTime := time.Unix(alert.FiredAt, 0).Format("2006-01-02 15:04:05")

	// 构建飞书富文本内容
	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s【%s】%s\n", statusEmoji, statusText, alert.RuleName))
	content.WriteString(fmt.Sprintf("\n━━━━━━━━━━━━━━━━\n"))
	content.WriteString(fmt.Sprintf("\n📌 级别: %s\n", severityMap[alert.Severity]))
	if alert.Summary != "" {
		content.WriteString(fmt.Sprintf("📝 摘要: %s\n", alert.Summary))
	}
	if alert.Value != "" {
		content.WriteString(fmt.Sprintf("📊 触发值: %s\n", alert.Value))
	}
	content.WriteString(fmt.Sprintf("⏰ 时间: %s\n", firedTime))

	if alert.Status == "resolved" && alert.ResolvedAt > 0 {
		resolvedTime := time.Unix(alert.ResolvedAt, 0).Format("2006-01-02 15:04:05")
		dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
		content.WriteString(fmt.Sprintf("✅ 恢复: %s（持续 %s）\n", resolvedTime, dur.Round(time.Second)))
	}

	if len(alert.Labels) > 0 {
		content.WriteString("\n🏷️ 标签:\n")
		for k, v := range alert.Labels {
			content.WriteString(fmt.Sprintf("  • %s = %s\n", k, v))
		}
	}
	content.WriteString("\n━━━━━━━━━━━━━━━━\n")
	content.WriteString("🛡️ K8s Operation 监控平台")

	body := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content.String(),
		},
	}

	return postJSON(ch.WebhookURL, ch.Secret, body)
}

// sendWechat 发送企业微信通知（Markdown 卡片）
func sendWechat(ch *models.MonitorNotifyChannel, alert *AlertNotification) error {
	statusEmoji := "🔥"
	statusText := "告警触发"
	statusColor := "warning"
	if alert.Status == "resolved" {
		statusEmoji = "✅"
		statusText = "告警恢复"
		statusColor = "info"
	}

	severityMap := map[string]string{
		"critical": "<font color=\"warning\">P0-Critical</font>",
		"warning":  "<font color=\"warning\">P1-Warning</font>",
		"info":     "<font color=\"comment\">P2-Info</font>",
	}
	firedTime := time.Unix(alert.FiredAt, 0).Format("2006-01-02 15:04:05")

	var md strings.Builder
	md.WriteString(fmt.Sprintf("## %s %s\n", statusEmoji, alert.RuleName))
	md.WriteString(fmt.Sprintf("> <font color=\"%s\">%s</font>\n\n", statusColor, statusText))
	md.WriteString(fmt.Sprintf("> **级别**: %s\n", severityMap[alert.Severity]))
	if alert.Summary != "" {
		md.WriteString(fmt.Sprintf("> **摘要**: %s\n", alert.Summary))
	}
	if alert.Value != "" {
		md.WriteString(fmt.Sprintf("> **触发值**: `%s`\n", alert.Value))
	}
	md.WriteString(fmt.Sprintf("> **时间**: %s\n\n", firedTime))

	if alert.Status == "resolved" && alert.ResolvedAt > 0 {
		resolvedTime := time.Unix(alert.ResolvedAt, 0).Format("2006-01-02 15:04:05")
		dur := time.Unix(alert.ResolvedAt, 0).Sub(time.Unix(alert.FiredAt, 0))
		md.WriteString(fmt.Sprintf("> **恢复时间**: %s（持续 %s）\n\n", resolvedTime, dur.Round(time.Second)))
	}

	if len(alert.Labels) > 0 {
		md.WriteString("**标签**:\n")
		for k, v := range alert.Labels {
			md.WriteString(fmt.Sprintf("> `%s`=`%s`\n", k, v))
		}
	}
	md.WriteString("\n---\n⚡ K8s Operation 监控平台 | 自动告警")

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
	return postJSON(ch.WebhookURL, "", alert)
}

// postJSON 发送 JSON POST 请求
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("返回状态码: %d", resp.StatusCode)
	}
	return nil
}
