package services

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"k8soperation/global"
)

func (s *Services) tryFinalizeRelease(ctx context.Context, releaseID int64) {
	st, err := s.dao.CicdTaskStatsByRelease(ctx, releaseID)
	if err != nil {
		return
	}

	// 1) 任意失败 => release Failed（CAS 防并发覆盖）
	if st.Failed > 0 {
		// 聚合失败任务的具体错误信息，返回给前端方便用户排查
		failMsg := s.aggregateFailedTaskMessages(ctx, releaseID)
		_, _ = s.dao.CicdReleaseUpdateStatusCAS(ctx, releaseID,
			[]string{"Pending", "Queued", "Running"},
			"Failed",
			failMsg,
		)
		// 发送部署失败通知
		s.sendReleaseFinalizeNotification(ctx, releaseID, false, failMsg)
		// 若目标环境开启「失败自动回滚」，紧急恢复至部署前版本（best-effort）
		s.maybeAutoRollbackOnFail(ctx, releaseID)
		return
	}

	// 2) 还有没结束的 => 不动
	if st.Pending > 0 || st.Queued > 0 || st.Running > 0 {
		return
	}

	// 3) 全部结束且无失败 => Succeeded
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{"Pending", "Queued", "Running"},
		"Succeeded",
		"all tasks succeeded",
	)
	// 发送部署成功通知
	s.sendReleaseFinalizeNotification(ctx, releaseID, true, "部署成功")
}

// aggregateFailedTaskMessages 聚合失败任务的错误信息，生成人类可读的失败原因
func (s *Services) aggregateFailedTaskMessages(ctx context.Context, releaseID int64) string {
	tasks, err := s.dao.CicdTasksByReleaseID(ctx, releaseID)
	if err != nil {
		return "部分任务执行失败"
	}

	var failedMsgs []string
	for _, t := range tasks {
		if t.Status == "Failed" && t.Message != "" {
			failedMsgs = append(failedMsgs, t.Message)
		}
	}

	if len(failedMsgs) == 0 {
		return "部分任务执行失败"
	}

	// 如果只有一个失败任务，直接返回其消息
	if len(failedMsgs) == 1 {
		return failedMsgs[0]
	}

	// 多个失败任务，拼接并截断避免过长
	result := strings.Join(failedMsgs, "; ")
	if len(result) > 500 {
		result = result[:500] + "..."
	}
	return fmt.Sprintf("%d个任务失败: %s", len(failedMsgs), result)
}

// sendReleaseFinalizeNotification 发送发布单完结通知（飞书/钉钉）
func (s *Services) sendReleaseFinalizeNotification(ctx context.Context, releaseID int64, success bool, message string) {
	// 获取发布单信息
	rel, err := s.dao.CicdReleaseGetByID(ctx, releaseID)
	if err != nil {
		global.Logger.Warn("[通知] 获取发布单失败，跳过通知",
			zap.Int64("release_id", releaseID),
			zap.Error(err),
		)
		return
	}

	// 构建通知文本
	statusEmoji := "✅"
	statusText := "部署成功"
	if !success {
		statusEmoji = "❌"
		statusText = "部署失败"
	}

	// 获取流水线信息（如果有关联）
	pipelineName := "-"
	buildURL := ""
	platformURL := s.getPlatformURL()
	if rel.BuildID > 0 {
		run, runErr := s.dao.PipelineRunGetByID(ctx, rel.BuildID)
		if runErr == nil && run != nil {
			pipeline, pErr := s.dao.PipelineGetByID(ctx, run.PipelineID)
			if pErr == nil && pipeline != nil {
				pipelineName = pipeline.Name
				if platformURL != "" {
					buildURL = platformURL + "/cicd/pipelines"
				}
			}
		}
	}

	targetImage := rel.ImageRepo
	if rel.ImageTag != "" {
		if targetImage != "" {
			targetImage += ":" + rel.ImageTag
		} else {
			targetImage = rel.ImageTag
		}
	}
	if rel.ImageDigest != nil && *rel.ImageDigest != "" {
		if idx := strings.LastIndex(targetImage, ":"); idx > 0 {
			targetImage = targetImage[:idx] + "@" + *rel.ImageDigest
		} else {
			targetImage = targetImage + "@" + *rel.ImageDigest
		}
	}

	title := statusEmoji + " " + statusText + " - " + rel.AppName
	text := "### " + title + "\n\n"
	text += "**发布单**: #" + fmt.Sprintf("%d", releaseID) + "\n"
	text += "**应用名称**: " + rel.AppName + "\n"
	text += "**流水线**: " + pipelineName + "\n"
	text += "**命名空间**: " + rel.Namespace + "\n"
	text += "**目标镜像**: " + targetImage + "\n"
	text += "**部署策略**: " + rel.Strategy + "\n"
	text += "**状态**: " + statusText + "\n"
	text += "**消息**: " + message + "\n"

	if buildURL != "" {
		text += "\n[查看详情](" + buildURL + ")\n"
	}

	// 发送钉钉通知
	if webhook := s.getDingTalkWebhook(nil); webhook != "" {
		msg := s.textToDingTalkMessage(title, text)
		go s.sendDingTalkNotify(webhook, msg)
	}

	// 发送飞书通知
	if webhook := s.getFeishuWebhook(); webhook != "" {
		msg := s.textToFeishuMessage(title, text)
		go s.sendFeishuNotify(webhook, msg)
	}
}
