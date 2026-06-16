package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ==================== 资源模板 DAO ====================

// ResourceTemplateList 获取资源模板列表
func (d *Dao) ResourceTemplateList(ctx context.Context, env, serviceType string) ([]models.CicdResourceTemplate, error) {
	var list []models.CicdResourceTemplate
	db := global.DB.WithContext(ctx).Where("deleted_at = 0")
	
	if env != "" {
		db = db.Where("env = ?", env)
	}
	if serviceType != "" {
		db = db.Where("service_type = ?", serviceType)
	}
	
	err := db.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// ResourceTemplateGetByID 根据ID获取模板
func (d *Dao) ResourceTemplateGetByID(ctx context.Context, id uint64) (*models.CicdResourceTemplate, error) {
	var tpl models.CicdResourceTemplate
	err := global.DB.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// ResourceTemplateGetDefault 获取默认模板
func (d *Dao) ResourceTemplateGetDefault(ctx context.Context, env, serviceType string) (*models.CicdResourceTemplate, error) {
	var tpl models.CicdResourceTemplate
	err := global.DB.WithContext(ctx).
		Where("env = ? AND service_type = ? AND is_default = 1 AND deleted_at = 0", env, serviceType).
		First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// ResourceTemplateCreate 创建模板
func (d *Dao) ResourceTemplateCreate(ctx context.Context, tpl *models.CicdResourceTemplate) error {
	tpl.CreatedAt = uint64(time.Now().Unix())
	tpl.ModifiedAt = uint64(time.Now().Unix())
	return global.DB.WithContext(ctx).Create(tpl).Error
}

// ResourceTemplateUpdate 更新模板
func (d *Dao) ResourceTemplateUpdate(ctx context.Context, tpl *models.CicdResourceTemplate) error {
	tpl.ModifiedAt = uint64(time.Now().Unix())
	return global.DB.WithContext(ctx).Save(tpl).Error
}

// ResourceTemplateDelete 软删除模板
func (d *Dao) ResourceTemplateDelete(ctx context.Context, id uint64) error {
	return global.DB.WithContext(ctx).Model(&models.CicdResourceTemplate{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now().Unix()).Error
}

// ==================== 环境资源规则 DAO ====================

// EnvResourceRuleList 获取环境规则列表
func (d *Dao) EnvResourceRuleList(ctx context.Context, env string) ([]models.CicdEnvResourceRule, error) {
	var list []models.CicdEnvResourceRule
	db := global.DB.WithContext(ctx)
	
	if env != "" {
		db = db.Where("env = ?", env)
	}
	
	err := db.Order("env ASC, service_type ASC").Find(&list).Error
	return list, err
}

// EnvResourceRuleGet 获取环境规则（优先匹配服务类型，其次通用）
func (d *Dao) EnvResourceRuleGet(ctx context.Context, env, serviceType string) (*models.CicdEnvResourceRule, error) {
	var rule models.CicdEnvResourceRule
	
	// 优先匹配特定服务类型的规则
	err := global.DB.WithContext(ctx).
		Where("env = ? AND service_type = ?", env, serviceType).
		First(&rule).Error
	if err == nil {
		return &rule, nil
	}
	
	// 回退到通用规则
	if err == gorm.ErrRecordNotFound {
		err = global.DB.WithContext(ctx).
			Where("env = ? AND service_type = ''", env).
			First(&rule).Error
	}
	
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// EnvResourceRuleGetByID 根据ID获取规则
func (d *Dao) EnvResourceRuleGetByID(ctx context.Context, id uint64) (*models.CicdEnvResourceRule, error) {
	var rule models.CicdEnvResourceRule
	err := global.DB.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// EnvResourceRuleUpdate 更新规则
func (d *Dao) EnvResourceRuleUpdate(ctx context.Context, rule *models.CicdEnvResourceRule) error {
	rule.ModifiedAt = uint64(time.Now().Unix())
	return global.DB.WithContext(ctx).Save(rule).Error
}

// ==================== 发布审批 DAO ====================

// DeployApprovalCreate 创建审批记录
func (d *Dao) DeployApprovalCreate(ctx context.Context, approval *models.CicdDeployApproval) error {
	approval.AppliedAt = uint64(time.Now().Unix())
	approval.ExpiredAt = uint64(time.Now().Add(72 * time.Hour).Unix()) // 72小时过期
	return global.DB.WithContext(ctx).Create(approval).Error
}

// DeployApprovalGetByID 根据ID获取审批记录
func (d *Dao) DeployApprovalGetByID(ctx context.Context, id uint64) (*models.CicdDeployApproval, error) {
	var approval models.CicdDeployApproval
	err := global.DB.WithContext(ctx).Where("id = ?", id).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// DeployApprovalList 获取审批列表
func (d *Dao) DeployApprovalList(ctx context.Context, status string, page, pageSize int) ([]models.CicdDeployApproval, int64, error) {
	var list []models.CicdDeployApproval
	var total int64
	
	db := global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	
	db.Count(&total)
	
	offset := (page - 1) * pageSize
	err := db.Order("applied_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DeployApprovalListByApplicant 获取申请人的审批列表
func (d *Dao) DeployApprovalListByApplicant(ctx context.Context, applicantID uint64, page, pageSize int) ([]models.CicdDeployApproval, int64, error) {
	var list []models.CicdDeployApproval
	var total int64
	
	db := global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{}).Where("applicant_id = ?", applicantID)
	db.Count(&total)
	
	offset := (page - 1) * pageSize
	err := db.Order("applied_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DeployApprovalApprove 通过审批
func (d *Dao) DeployApprovalApprove(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{}).
		Where("id = ? AND status = ?", id, models.ApprovalStatusPending).
		Updates(map[string]interface{}{
			"status":          models.ApprovalStatusApproved,
			"approver_id":     approverID,
			"approver_name":   approverName,
			"approve_comment": comment,
			"approved_at":     time.Now().Unix(),
		}).Error
}

// DeployApprovalReject 拒绝审批
func (d *Dao) DeployApprovalReject(ctx context.Context, id uint64, approverID uint64, approverName, comment string) error {
	return global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{}).
		Where("id = ? AND status = ?", id, models.ApprovalStatusPending).
		Updates(map[string]interface{}{
			"status":          models.ApprovalStatusRejected,
			"approver_id":     approverID,
			"approver_name":   approverName,
			"approve_comment": comment,
			"approved_at":     time.Now().Unix(),
		}).Error
}

// DeployApprovalCancel 取消审批
func (d *Dao) DeployApprovalCancel(ctx context.Context, id uint64, applicantID uint64) error {
	return global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{}).
		Where("id = ? AND applicant_id = ? AND status = ?", id, applicantID, models.ApprovalStatusPending).
		Update("status", models.ApprovalStatusCancelled).Error
}

// DeployApprovalExpireOld 过期旧的审批记录
func (d *Dao) DeployApprovalExpireOld(ctx context.Context) (int64, error) {
	result := global.DB.WithContext(ctx).Model(&models.CicdDeployApproval{}).
		Where("status = ? AND expired_at < ?", models.ApprovalStatusPending, time.Now().Unix()).
		Update("status", models.ApprovalStatusExpired)
	return result.RowsAffected, result.Error
}

// ==================== 变更日志 DAO ====================

// ResourceChangeLogCreate 创建变更日志
func (d *Dao) ResourceChangeLogCreate(ctx context.Context, log *models.CicdResourceChangeLog) error {
	log.CreatedAt = uint64(time.Now().Unix())
	return global.DB.WithContext(ctx).Create(log).Error
}

// ResourceChangeLogList 获取变更日志列表
func (d *Dao) ResourceChangeLogList(ctx context.Context, pipelineID uint64, env string, page, pageSize int) ([]models.CicdResourceChangeLog, int64, error) {
	var list []models.CicdResourceChangeLog
	var total int64
	
	db := global.DB.WithContext(ctx).Model(&models.CicdResourceChangeLog{}).Where("pipeline_id = ?", pipelineID)
	if env != "" {
		db = db.Where("env = ?", env)
	}
	db.Count(&total)
	
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
