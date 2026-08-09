package cicd

import "context"

// CicdRepository CICD 核心仓储接口
type CicdRepository interface {
	// Pipeline
	PipelineSave(ctx context.Context, p *CicdPipeline) error
	PipelineFindByID(ctx context.Context, id int64) (*CicdPipeline, error)
	PipelineFindByName(ctx context.Context, name string) (*CicdPipeline, error)
	PipelineQuery(ctx context.Context, f PipelineListFilter) ([]*CicdPipeline, int64, error)
	PipelineUpdate(ctx context.Context, id int64, updates map[string]interface{}) error
	PipelineDelete(ctx context.Context, id int64) error
	PipelineUpdateStatus(ctx context.Context, id int64, status string) error
	PipelineUpdateRunComplete(ctx context.Context, id int64, runStatus string) error

	// PipelineRun
	PipelineRunSave(ctx context.Context, run *CicdPipelineRun) error
	PipelineRunFindByID(ctx context.Context, id int64) (*CicdPipelineRun, error)
	PipelineRunFindLatest(ctx context.Context, pipelineID int64) (*CicdPipelineRun, error)
	PipelineRunQuery(ctx context.Context, pipelineID int64, page, pageSize int) ([]*CicdPipelineRun, int64, error)
	PipelineRunUpdate(ctx context.Context, id int64, updates map[string]interface{}) error
	PipelineRunUpdateStatus(ctx context.Context, id int64, status string) error

	// Stage
	StageFindByRunIDAndType(ctx context.Context, runID int64, stageType string) (*CicdPipelineStage, error)
	StageListApprovalAll(ctx context.Context) ([]*CicdPipelineStage, error)

	// Approval
	ApprovalSave(ctx context.Context, a *CicdApproval) (int64, error)
	ApprovalExistsByStageID(ctx context.Context, stageID int64) (bool, error)

	// Environment
	EnvironmentSave(ctx context.Context, e *CicdEnvironment) (int64, error)
	EnvironmentFindByID(ctx context.Context, id int64) (*CicdEnvironment, error)
	EnvironmentFindByName(ctx context.Context, name string) (*CicdEnvironment, error)
	EnvironmentUpdate(ctx context.Context, e *CicdEnvironment) error
	EnvironmentDelete(ctx context.Context, id int64) error
}
