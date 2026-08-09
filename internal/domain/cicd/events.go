package cicd

import "k8soperation/internal/domain/events"

// ——— CICD 领域事件 ———

// ReleaseStarted 发布开始事件
type ReleaseStarted struct {
	events.BaseEvent
	ReleaseID  int64
	PipelineID int64
	Env        string
}

func NewReleaseStarted(releaseID, pipelineID int64, env string) ReleaseStarted {
	return ReleaseStarted{
		BaseEvent:  events.NewBaseEvent("cicd.release.started"),
		ReleaseID:  releaseID,
		PipelineID: pipelineID,
		Env:        env,
	}
}

// ReleaseCompleted 发布完成事件
type ReleaseCompleted struct {
	events.BaseEvent
	ReleaseID  int64
	PipelineID int64
	Status     string
}

func NewReleaseCompleted(releaseID, pipelineID int64, status string) ReleaseCompleted {
	return ReleaseCompleted{
		BaseEvent:  events.NewBaseEvent("cicd.release.completed"),
		ReleaseID:  releaseID,
		PipelineID: pipelineID,
		Status:     status,
	}
}

// PipelineTriggered 流水线触发事件
type PipelineTriggered struct {
	events.BaseEvent
	PipelineID int64
	RunID      int64
	TriggerBy  string
}

func NewPipelineTriggered(pipelineID, runID int64, triggerBy string) PipelineTriggered {
	return PipelineTriggered{
		BaseEvent:  events.NewBaseEvent("cicd.pipeline.triggered"),
		PipelineID: pipelineID,
		RunID:      runID,
		TriggerBy:  triggerBy,
	}
}
