package cicd

import (
	"fmt"
	"time"
)

// ========== 聚合根：CicdRelease → CicdReleaseStage → CicdReleaseTask ==========

// AggregateID 实现 domain.AggregateRoot 接口
func (r *CicdRelease) AggregateID() int64 { return r.ID }

// AddStage 在发布下创建阶段（聚合边界：Stage 只能通过 Release 创建）
func (r *CicdRelease) AddStage(name string, order int) (*CicdReleaseStage, error) {
	if r.ID == 0 {
		return nil, fmt.Errorf("发布尚未持久化，无法添加阶段")
	}
	if r.Status == CicdReleaseStatusSucceeded || r.Status == CicdReleaseStatusCanceled {
		return nil, fmt.Errorf("发布已%s，无法添加阶段", r.Status)
	}
	now := uint64(time.Now().Unix())
	return &CicdReleaseStage{
		ReleaseID:  r.ID,
		StageName:  name,
		StageOrder: order,
		Status:     CicdReleaseStageStatusPending,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// Start 开始发布
func (r *CicdRelease) Start() error {
	if r.Status != CicdReleaseStatusPending && r.Status != CicdReleaseStatusAwaitingApproval {
		return fmt.Errorf("只能开始待处理的发布，当前状态: %s", r.Status)
	}
	r.Status = CicdReleaseStatusRunning
	return nil
}

// Succeed 标记发布成功
func (r *CicdRelease) Succeed() error {
	if r.Status != CicdReleaseStatusRunning {
		return fmt.Errorf("只能标记运行中的发布为成功，当前状态: %s", r.Status)
	}
	r.Status = CicdReleaseStatusSucceeded
	return nil
}

// Fail 标记发布失败
func (r *CicdRelease) Fail(reason string) error {
	if r.Status == CicdReleaseStatusSucceeded || r.Status == CicdReleaseStatusCanceled {
		return fmt.Errorf("发布已终结，无法标记失败")
	}
	r.Status = CicdReleaseStatusFailed
	r.Message = reason
	return nil
}

// Cancel 取消发布
func (r *CicdRelease) Cancel(reason string) error {
	if r.Status == CicdReleaseStatusSucceeded || r.Status == CicdReleaseStatusCanceled {
		return fmt.Errorf("发布已终结，无法取消")
	}
	r.Status = CicdReleaseStatusCanceled
	r.Message = reason
	return nil
}

// ========== 聚合子实体：CicdReleaseStage ==========

// Start 开始执行阶段
func (s *CicdReleaseStage) Start() error {
	if s.Status != CicdReleaseStageStatusPending && s.Status != CicdReleaseStageStatusWaiting {
		return fmt.Errorf("阶段状态 %s 不允许开始", s.Status)
	}
	s.Status = CicdReleaseStageStatusRunning
	s.StartTime = uint64(time.Now().Unix())
	return nil
}

// MarkSuccess 标记阶段成功
func (s *CicdReleaseStage) MarkSuccess() error {
	if s.Status != CicdReleaseStageStatusRunning {
		return fmt.Errorf("只能标记运行中的阶段为成功")
	}
	s.Status = CicdReleaseStageStatusSuccess
	s.EndTime = uint64(time.Now().Unix())
	return nil
}

// MarkFailed 标记阶段失败
func (s *CicdReleaseStage) MarkFailed(msg string) error {
	if s.Status != CicdReleaseStageStatusRunning {
		return fmt.Errorf("只能标记运行中的阶段为失败")
	}
	s.Status = CicdReleaseStageStatusFailed
	s.Message = msg
	s.EndTime = uint64(time.Now().Unix())
	return nil
}
