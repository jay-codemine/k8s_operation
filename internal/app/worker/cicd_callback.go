package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
)

// CicdCallback 回调通知服务
type CicdCallback struct {
	client     *http.Client
	webhookURL string // 可配置的 webhook 地址
}

// CallbackPayload 回调请求体
type CallbackPayload struct {
	TaskID    int64  `json:"task_id"`
	ReleaseID int64  `json:"release_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// ReleaseCallbackPayload 发布单完成回调
type ReleaseCallbackPayload struct {
	ReleaseID int64  `json:"release_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// NewCicdCallback 创建回调服务
func NewCicdCallback() *CicdCallback {
	return &CicdCallback{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		webhookURL: "", // 从配置中读取，这里先留空
	}
}

// SetWebhookURL 设置 Webhook URL
func (c *CicdCallback) SetWebhookURL(url string) {
	c.webhookURL = url
}

// NotifyTaskComplete 通知任务完成
func (c *CicdCallback) NotifyTaskComplete(ctx context.Context, taskID, releaseID int64, success bool, message string) {
	if c.webhookURL == "" {
		// 没有配置 webhook，跳过通知
		return
	}

	payload := &CallbackPayload{
		TaskID:    taskID,
		ReleaseID: releaseID,
		Success:   success,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	go c.sendCallback(ctx, payload)
}

// NotifyReleaseComplete 通知发布单完成
func (c *CicdCallback) NotifyReleaseComplete(ctx context.Context, releaseID int64, status, message string) {
	if c.webhookURL == "" {
		return
	}

	payload := &ReleaseCallbackPayload{
		ReleaseID: releaseID,
		Status:    status,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	go c.sendReleaseCallback(ctx, payload)
}

// sendCallback 发送任务回调
func (c *CicdCallback) sendCallback(ctx context.Context, payload *CallbackPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		global.Logger.Error("marshal callback payload failed", zap.Error(err))
		return
	}

	c.doSendWithRetry(ctx, body, 3)
}

// sendReleaseCallback 发送发布单回调
func (c *CicdCallback) sendReleaseCallback(ctx context.Context, payload *ReleaseCallbackPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		global.Logger.Error("marshal release callback payload failed", zap.Error(err))
		return
	}

	c.doSendWithRetry(ctx, body, 3)
}

// doSendWithRetry 带重试的发送
func (c *CicdCallback) doSendWithRetry(ctx context.Context, body []byte, maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		err := c.doSend(ctx, body)
		if err == nil {
			return
		}

		global.Logger.Warn("send callback failed, will retry",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", maxRetries),
			zap.Error(err))

		// 指数退避
		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	global.Logger.Error("send callback failed after retries",
		zap.Int("max_retries", maxRetries),
		zap.String("url", c.webhookURL))
}

// doSend 发送 HTTP 请求
func (c *CicdCallback) doSend(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		global.Logger.Info("callback sent successfully",
			zap.String("url", c.webhookURL),
			zap.Int("status", resp.StatusCode))
		return nil
	}

	global.Logger.Warn("callback response error",
		zap.String("url", c.webhookURL),
		zap.Int("status", resp.StatusCode))

	return nil // 非 2xx 也不算失败，只是记录日志
}
