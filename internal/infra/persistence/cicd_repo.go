package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/cicd"
)

type cicdRepo struct {
	db *gorm.DB
}

func NewCicdRepository(db *gorm.DB) cicd.CicdRepository {
	return &cicdRepo{db: db}
}

// ——— Pipeline ———

func (r *cicdRepo) PipelineSave(ctx context.Context, p *cicd.CicdPipeline) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *cicdRepo) PipelineFindByID(ctx context.Context, id int64) (*cicd.CicdPipeline, error) {
	var p cicd.CicdPipeline
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *cicdRepo) PipelineFindByName(ctx context.Context, name string) (*cicd.CicdPipeline, error) {
	var p cicd.CicdPipeline
	if err := r.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *cicdRepo) PipelineQuery(ctx context.Context, f cicd.PipelineListFilter) ([]*cicd.CicdPipeline, int64, error) {
	db := r.db.WithContext(ctx).Model(&cicd.CicdPipeline{}).Where("is_del = 0")
	if f.Keyword != "" {
		k := "%" + f.Keyword + "%"
		db = db.Where("(name LIKE ? OR description LIKE ?)", k, k)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.Language != "" {
		db = db.Where("language_type = ?", f.Language)
	}
	if f.DeployEnv != "" {
		db = db.Where("deploy_env = ?", f.DeployEnv)
	}
	if f.EnvironmentID > 0 {
		db = db.Where("environment_id = ?", f.EnvironmentID)
	}
	if f.CreatorID > 0 {
		db = db.Where("created_by = ?", f.CreatorID)
	}
	if f.StartTime > 0 {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime > 0 {
		db = db.Where("created_at <= ?", f.EndTime)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []*cicd.CicdPipeline
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *cicdRepo) PipelineUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["modified_at"] = time.Now().Unix()
	return r.db.WithContext(ctx).Model(&cicd.CicdPipeline{}).Where("id = ?", id).Updates(updates).Error
}

func (r *cicdRepo) PipelineDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Model(&cicd.CicdPipeline{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_del": 1, "deleted_at": now}).Error
}

func (r *cicdRepo) PipelineUpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&cicd.CicdPipeline{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *cicdRepo) PipelineUpdateRunComplete(ctx context.Context, id int64, runStatus string) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{"status": runStatus, "last_run_at": now}
	if runStatus == cicd.PipelineRunStatusSuccess {
		updates["last_success_at"] = now
	}
	return r.db.WithContext(ctx).Model(&cicd.CicdPipeline{}).Where("id = ?", id).Updates(updates).Error
}

// ——— PipelineRun ———

func (r *cicdRepo) PipelineRunSave(ctx context.Context, run *cicd.CicdPipelineRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *cicdRepo) PipelineRunFindByID(ctx context.Context, id int64) (*cicd.CicdPipelineRun, error) {
	var run cicd.CicdPipelineRun
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *cicdRepo) PipelineRunFindLatest(ctx context.Context, pipelineID int64) (*cicd.CicdPipelineRun, error) {
	var run cicd.CicdPipelineRun
	if err := r.db.WithContext(ctx).Where("pipeline_id = ?", pipelineID).Order("id DESC").First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *cicdRepo) PipelineRunQuery(ctx context.Context, pipelineID int64, page, pageSize int) ([]*cicd.CicdPipelineRun, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := r.db.WithContext(ctx).Model(&cicd.CicdPipelineRun{}).Where("pipeline_id = ?", pipelineID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*cicd.CicdPipelineRun
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *cicdRepo) PipelineRunUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&cicd.CicdPipelineRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *cicdRepo) PipelineRunUpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{"status": status}
	now := time.Now().Unix()
	if status == cicd.PipelineRunStatusRunning {
		updates["started_at"] = now
	} else if status == cicd.PipelineRunStatusSuccess || status == cicd.PipelineRunStatusFailed {
		updates["finished_at"] = now
	}
	return r.db.WithContext(ctx).Model(&cicd.CicdPipelineRun{}).Where("id = ?", id).Updates(updates).Error
}

// ——— Stage ———

func (r *cicdRepo) StageFindByRunIDAndType(ctx context.Context, runID int64, stageType string) (*cicd.CicdPipelineStage, error) {
	var stage cicd.CicdPipelineStage
	if err := r.db.WithContext(ctx).Where("run_id = ? AND stage_type = ?", runID, stageType).First(&stage).Error; err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *cicdRepo) StageListApprovalAll(ctx context.Context) ([]*cicd.CicdPipelineStage, error) {
	var stages []*cicd.CicdPipelineStage
	err := r.db.WithContext(ctx).Where("stage_type = ?", "approval").Find(&stages).Error
	return stages, err
}

// ——— Approval ———

func (r *cicdRepo) ApprovalSave(ctx context.Context, a *cicd.CicdApproval) (int64, error) {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

func (r *cicdRepo) ApprovalExistsByStageID(ctx context.Context, stageID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&cicd.CicdApproval{}).Where("stage_id = ?", stageID).Count(&count).Error
	return count > 0, err
}

// ——— Environment ———

func (r *cicdRepo) EnvironmentSave(ctx context.Context, e *cicd.CicdEnvironment) (int64, error) {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return 0, err
	}
	return e.ID, nil
}

func (r *cicdRepo) EnvironmentFindByID(ctx context.Context, id int64) (*cicd.CicdEnvironment, error) {
	var env cicd.CicdEnvironment
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&env).Error; err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *cicdRepo) EnvironmentFindByName(ctx context.Context, name string) (*cicd.CicdEnvironment, error) {
	var env cicd.CicdEnvironment
	if err := r.db.WithContext(ctx).Where("name = ? AND is_del = 0", name).First(&env).Error; err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *cicdRepo) EnvironmentUpdate(ctx context.Context, e *cicd.CicdEnvironment) error {
	return r.db.WithContext(ctx).Model(&cicd.CicdEnvironment{}).
		Where("id = ? AND is_del = 0", e.ID).Save(e).Error
}

func (r *cicdRepo) EnvironmentDelete(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Model(&cicd.CicdEnvironment{}).
		Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{"is_del": 1, "deleted_at": now, "modified_at": now}).Error
}
