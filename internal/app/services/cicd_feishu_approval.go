package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ==================== 飞书审批通知服务 ====================
// 实现开发自助发布 + 飞书审批模式：
// 1. 构建成功后自动发送飞书交互卡片到审批群
// 2. 审批人在飞书中点击「通过/拒绝」按钮
// 3. 飞书通过回调 URL 通知平台执行审批动作
// 4. 平台自动完成审批流程并触发部署
// =========================================================

// FeishuApprovalCallbackRequest 飞书审批回调请求结构
type FeishuApprovalCallbackRequest struct {
	Token  string `json:"token" binding:"required"`  // 审批唯一 Token
	Action string `json:"action" binding:"required"` // approve / reject
	Reason string `json:"reason"`                    // 审批意见（可选）
}

// NotifyFeishuApproval 发送飞书审批交互卡片
// 当构建成功需要审批时调用此方法
func (s *Services) NotifyFeishuApproval(ctx context.Context, pipeline *models.CicdPipeline, run *models.CicdPipelineRun, approval *models.CicdApproval) {
	webhook := s.getFeishuWebhook()
	if webhook == "" {
		global.Logger.Debug("[飞书审批] Webhook 未配置，跳过飞书通知")
		return
	}

	// 生成唯一的审批 Token
	token, err := generateApprovalToken()
	if err != nil {
		global.Logger.Error("[飞书审批] 生成Token失败", zap.Error(err))
		return
	}

	// 保存 Token 到审批记录
	_ = s.cicdSvc().ApprovalUpdate(ctx, approval.ID, map[string]interface{}{
		"feishu_token": token,
	})

	// 构建飞书交互卡片消息
	msg := s.buildFeishuApprovalCard(pipeline, run, approval, token)

	// 异步发送
	go func() {
		if err := s.sendFeishuMessage(webhook, msg); err != nil {
			global.Logger.Error("[飞书审批] 发送通知失败",
				zap.String("pipeline", pipeline.Name),
				zap.Error(err),
			)
		} else {
			global.Logger.Info("[飞书审批] 通知发送成功",
				zap.String("pipeline", pipeline.Name),
				zap.Int64("approval_id", approval.ID),
			)
		}
	}()
}

// HandleFeishuApprovalCallback 处理飞书审批回调
func (s *Services) HandleFeishuApprovalCallback(ctx context.Context, req *FeishuApprovalCallbackRequest) error {
	if req.Token == "" {
		return errors.New("审批Token不能为空")
	}
	if req.Action != "approve" && req.Action != "reject" {
		return errors.New("无效的审批操作，仅支持 approve/reject")
	}

	// 根据 Token 查找审批记录
	approval, err := s.cicdSvc().ApprovalGetByFeishuToken(ctx, req.Token)
	if err != nil {
		return errors.New("审批记录不存在或Token无效")
	}

	// 检查审批状态
	if approval.Status != models.ApprovalStatusPending {
		return fmt.Errorf("该审批已处理（当前状态: %s），无法重复操作", approval.Status)
	}

	// 检查是否过期
	if approval.ExpireTime > 0 && uint64(time.Now().Unix()) > approval.ExpireTime {
		_ = s.cicdSvc().ApprovalUpdateStatus(ctx, approval.ID, models.ApprovalStatusExpired, 0, "")
		return errors.New("该审批申请已过期")
	}

	// 执行审批操作
	var status string
	if req.Action == "approve" {
		status = models.ApprovalStatusApproved
	} else {
		status = models.ApprovalStatusRejected
	}

	// 使用系统用户ID（飞书回调无登录态，用 0 表示飞书审批）
	var approveUserID int64 = 0

	if err := s.cicdSvc().ApprovalUpdateStatus(ctx, approval.ID, status, approveUserID, req.Reason); err != nil {
		return err
	}

	global.Logger.Info("[飞书审批] 回调处理完成",
		zap.Int64("approval_id", approval.ID),
		zap.String("action", req.Action),
		zap.String("reason", req.Reason),
	)

	// 同步更新关联的流水线审批阶段
	if approval.StageID > 0 {
		if status == models.ApprovalStatusApproved {
			// ====== 多级审批级联逻辑 ======
			if approval.ApprovalLevel < approval.TotalLevels {
				// 还有下一级，创建下一级审批
				nextID, nextErr := s.CreateNextLevelApproval(ctx, approval)
				if nextErr == nil && nextID > 0 {
					nextApproval, _ := s.cicdSvc().ApprovalGetByID(ctx, nextID)
					if nextApproval != nil {
						pipeline, _ := s.cicdSvc().PipelineGetByID(ctx, nextApproval.PipelineID)
						run, _ := s.cicdSvc().PipelineRunGetByID(ctx, nextApproval.PipelineRunID)
						if pipeline != nil && run != nil {
							s.NotifyFeishuApproval(ctx, pipeline, run, nextApproval)
						}
					}
				}
				// 发送当前级通过结果通知
				go s.notifyFeishuApprovalResult(ctx, approval, req.Action, req.Reason)
				return nil // 不触发部署，等待下一级
			}

			// 最后一级通过：更新阶段审批状态，启动部署
			_ = s.cicdSvc().StageUpdateApproval(ctx, approval.StageID, approveUserID, "approved", req.Reason)

			// 启动部署阶段
			stage, stageErr := s.cicdSvc().StageGetByID(ctx, approval.StageID)
			if stageErr == nil && stage != nil {
				deployStage, dErr := s.cicdSvc().StageGetByRunIDAndType(ctx, stage.RunID, models.StageTypeDeploy)
				if dErr == nil && deployStage != nil {
					run, _ := s.cicdSvc().PipelineRunGetByID(ctx, stage.RunID)
					if run != nil && run.ImageURL != "" {
						_ = s.cicdSvc().StageUpdate(ctx, deployStage.ID, map[string]interface{}{
							"status":       models.StageStatusPending,
							"deploy_image": run.ImageURL,
						})
						global.Logger.Info("[飞书审批] 全部审批通过，部署阶段已就绪",
							zap.Int64("deploy_stage_id", deployStage.ID),
							zap.String("image", run.ImageURL),
						)
					}
				}
			}
		} else {
			// 拒绝：标记后续阶段为跳过
			_ = s.cicdSvc().StageUpdateApproval(ctx, approval.StageID, approveUserID, "rejected", req.Reason)

			stage, stageErr := s.cicdSvc().StageGetByID(ctx, approval.StageID)
			if stageErr == nil && stage != nil {
				stages, _ := s.cicdSvc().StageListByRunID(ctx, stage.RunID)
				for _, stg := range stages {
					if stg.StageOrder > stage.StageOrder && stg.Status == models.StageStatusPending {
						_ = s.cicdSvc().StageUpdateStatus(ctx, stg.ID, models.StageStatusSkipped)
					}
				}
				// 更新流水线运行状态为失败
				_ = s.cicdSvc().PipelineRunUpdateStatus(ctx, stage.RunID, models.PipelineRunStatusFailed)
				run, _ := s.cicdSvc().PipelineRunGetByID(ctx, stage.RunID)
				if run != nil {
					_ = s.cicdSvc().PipelineUpdateRunComplete(ctx, run.PipelineID, models.PipelineRunStatusFailed)
				}
			}
		}
	}

	// 发送审批结果通知到飞书群
	go s.notifyFeishuApprovalResult(ctx, approval, req.Action, req.Reason)

	return nil
}

// ==================== 飞书消息构建 ====================

// buildFeishuApprovalCard 构建飞书交互卡片（用于审批）
func (s *Services) buildFeishuApprovalCard(pipeline *models.CicdPipeline, run *models.CicdPipelineRun, approval *models.CicdApproval, token string) map[string]interface{} {
	envText := s.getEnvDisplayNameWithCluster(pipeline.DeployEnv, pipeline.TargetClusterID)
	platformURL := s.getPlatformURL()
	callbackURL := s.getCallbackURL()

	// 构建审批回调 URL
	approveURL := fmt.Sprintf("%s/api/v1/k8s/cicd/approval/feishu-callback?token=%s&action=approve", callbackURL, token)
	rejectURL := fmt.Sprintf("%s/api/v1/k8s/cicd/approval/feishu-callback?token=%s&action=reject", callbackURL, token)

	// 飞书交互卡片 (Interactive Card)
	// 构建审批级别信息
	levelInfo := ""
	if approval.TotalLevels > 1 {
		levelInfo = fmt.Sprintf(" (%d/%d: %s)", approval.ApprovalLevel, approval.TotalLevels, approval.LevelLabel)
	}

	card := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
			},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("🔔 发布审批%s - %s", levelInfo, pipeline.Name),
				},
				"template": "orange",
			},
			"elements": []interface{}{
				// 信息区域
				map[string]interface{}{
					"tag": "div",
					"fields": []interface{}{
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**流水线**\n%s", pipeline.Name),
							},
						},
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**环境**\n%s", envText),
							},
						},
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**分支**\n%s", run.GitBranch),
							},
						},
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**构建号**\n#%d", run.BuildNumber),
							},
						},
					},
				},
				// 镜像信息
				map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**镜像**: `%s`", run.ImageURL),
					},
				},
				// 部署目标
				map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**部署目标**: %s/%s → %s/%s", pipeline.TargetNamespace, pipeline.TargetWorkloadName, pipeline.TargetWorkloadKind, pipeline.TargetContainer),
					},
				},
				// 分割线
				map[string]interface{}{
					"tag": "hr",
				},
				// 操作按钮区域
				map[string]interface{}{
					"tag": "action",
					"actions": []interface{}{
						// 通过按钮
						map[string]interface{}{
							"tag": "button",
							"text": map[string]interface{}{
								"tag":     "plain_text",
								"content": "✅ 通过",
							},
							"type": "primary",
							"url":  approveURL,
						},
						// 拒绝按钮
						map[string]interface{}{
							"tag": "button",
							"text": map[string]interface{}{
								"tag":     "plain_text",
								"content": "❌ 拒绝",
							},
							"type": "danger",
							"url":  rejectURL,
						},
						// 查看详情
						map[string]interface{}{
							"tag": "button",
							"text": map[string]interface{}{
								"tag":     "plain_text",
								"content": "🔗 查看详情",
							},
							"url": fmt.Sprintf("%s/cicd/pipelines/%d?tab=stages", platformURL, pipeline.ID),
						},
					},
				},
				// 提示信息
				map[string]interface{}{
					"tag": "note",
					"elements": []interface{}{
						map[string]interface{}{
							"tag":     "plain_text",
							"content": fmt.Sprintf("⏳ 审批将在 7 天后过期 | 申请时间: %s", time.Now().Format("2006-01-02 15:04:05")),
						},
					},
				},
			},
		},
	}

	return card
}

// notifyFeishuApprovalResult 审批结果通知
func (s *Services) notifyFeishuApprovalResult(ctx context.Context, approval *models.CicdApproval, action, reason string) {
	webhook := s.getFeishuWebhook()
	if webhook == "" {
		return
	}

	var icon, statusText string
	if action == "approve" {
		icon = "✅"
		statusText = "已通过"
	} else {
		icon = "❌"
		statusText = "已拒绝"
	}

	// 获取流水线名称
	pipelineName := "未知流水线"
	if approval.PipelineID > 0 {
		pipeline, err := s.cicdSvc().PipelineGetByID(ctx, approval.PipelineID)
		if err == nil && pipeline != nil {
			pipelineName = pipeline.Name
		}
	}

	content := fmt.Sprintf("%s 审批%s\n\n流水线: %s\n镜像: %s",
		icon, statusText, pipelineName, approval.Image)
	if reason != "" {
		content += fmt.Sprintf("\n意见: %s", reason)
	}

	msg := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}

	if err := s.sendFeishuMessage(webhook, msg); err != nil {
		global.Logger.Error("[飞书审批] 结果通知发送失败", zap.Error(err))
	}
}

// ==================== 辅助函数 ====================

// getFeishuWebhook 获取飞书 Webhook 地址（需开关 EnableFeishu=true 且 Webhook 非空）
func (s *Services) getFeishuWebhook() string {
	if global.JenkinsSetting != nil && global.JenkinsSetting.EnableFeishu && global.JenkinsSetting.FeishuWebhook != "" {
		return global.JenkinsSetting.FeishuWebhook
	}
	return ""
}

// getFeishuSecret 获取飞书签名密钥
func (s *Services) getFeishuSecret() string {
	if global.JenkinsSetting != nil {
		return global.JenkinsSetting.FeishuSecret
	}
	return ""
}

// getCallbackURL 获取平台回调地址
func (s *Services) getCallbackURL() string {
	if global.JenkinsSetting != nil && global.JenkinsSetting.CallbackURL != "" {
		return global.JenkinsSetting.CallbackURL
	}
	return ""
}

// sendFeishuMessage 发送飞书消息（支持签名）
func (s *Services) sendFeishuMessage(webhook string, payload map[string]interface{}) error {
	// 如果配置了密钥，计算飞书签名
	if global.JenkinsSetting != nil && global.JenkinsSetting.FeishuSecret != "" {
		secret := global.JenkinsSetting.FeishuSecret
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		stringToSign := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(stringToSign))
		mac.Write([]byte{})
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		payload["timestamp"] = timestamp
		payload["sign"] = sign
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("发送飞书消息失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书HTTP状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 检查飞书业务错误码
	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if jsonErr := json.Unmarshal(respBody, &apiResp); jsonErr == nil && apiResp.Code != 0 {
		return fmt.Errorf("飞书API错误 code=%d: %s", apiResp.Code, apiResp.Msg)
	}

	return nil
}

// generateApprovalToken 生成唯一的审批 Token（32字节随机十六进制）
func generateApprovalToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
