package models

import (
	"gorm.io/gorm"
)

// ========== 审计日志模型 ==========

// AuditLog 平台操作审计日志
type AuditLog struct {
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64  `gorm:"not null;index:idx_audit_user" json:"user_id"`
	Username        string `gorm:"size:191;not null" json:"username"`
	UserIP          string `gorm:"size:50" json:"user_ip"`
	UserAgent       string `gorm:"size:500" json:"user_agent"`
	Action          string `gorm:"size:100;not null;index:idx_audit_action" json:"action"`           // create/update/delete/login/logout/exec/approve
	ActionDisplay   string `gorm:"size:191" json:"action_display"`                                   // 操作显示名称
	Module          string `gorm:"size:100;not null;index:idx_audit_module" json:"module"`           // cluster/workload/cicd/rbac/platform/auth/ai
	TargetType      string `gorm:"size:100;index:idx_audit_target" json:"target_type"`               // deployment/pod/pipeline/user/role...
	TargetID        string `gorm:"size:100;index:idx_audit_target" json:"target_id"`                 // 目标ID
	TargetName      string `gorm:"size:191" json:"target_name"`                                      // 目标名称
	RequestURI      string `gorm:"size:500" json:"request_uri"`
	RequestMethod   string `gorm:"size:10" json:"request_method"`
	RequestBody     string `gorm:"type:text" json:"request_body,omitempty"`
	ResponseCode    int    `gorm:"default:0" json:"response_code"`
	ResponseMessage string `gorm:"size:500" json:"response_message"`
	Detail          string `gorm:"type:json" json:"detail,omitempty"`       // 操作详情JSON
	Extra           string `gorm:"type:json" json:"extra,omitempty"`        // 额外信息JSON
	ClusterID       *int64 `gorm:"index:idx_audit_cluster" json:"cluster_id,omitempty"`
	ClusterName     string `gorm:"size:191" json:"cluster_name,omitempty"`
	Namespace       string `gorm:"size:100" json:"namespace,omitempty"`
	PipelineID      *int64 `gorm:"index:idx_audit_pipeline" json:"pipeline_id,omitempty"`
	PipelineName    string `gorm:"size:191" json:"pipeline_name,omitempty"`
	ProjectID       *int64 `json:"project_id,omitempty"`
	ProjectName     string `gorm:"size:191" json:"project_name,omitempty"`
	Status          string `gorm:"size:50;not null;default:'success';index:idx_audit_status" json:"status"` // success/failed
	ErrorMessage    string `gorm:"size:1000" json:"error_message,omitempty"`
	DurationMs      int    `gorm:"default:0" json:"duration_ms"`
	CreatedAt       int64  `gorm:"not null;index:idx_audit_created" json:"created_at"`
}

func (a *AuditLog) TableName() string {
	return "audit_log"
}

// ========== 审计日志保留策略 ==========

// AuditRetentionPolicy 审计日志保留策略
type AuditRetentionPolicy struct {
	RetentionDays int  `json:"retention_days"` // 保留天数，0 = 永久保留
	IsPermanent   bool `json:"is_permanent"`   // 是否永久保留
}

// ========== 查询请求/响应 ==========

// AuditLogQuery 审计日志查询参数
type AuditLogQuery struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	UserID     int64  `form:"user_id" json:"user_id"`
	Username   string `form:"username" json:"username"`
	Action     string `form:"action" json:"action"`
	Module     string `form:"module" json:"module"`
	TargetType string `form:"target_type" json:"target_type"`
	Status     string `form:"status" json:"status"`
	ClusterID  int64  `form:"cluster_id" json:"cluster_id"`
	Keyword    string `form:"keyword" json:"keyword"`
	StartTime  int64  `form:"start_time" json:"start_time"` // unix timestamp
	EndTime    int64  `form:"end_time" json:"end_time"`     // unix timestamp
	SortField  string `form:"sort_field" json:"sort_field"`
	SortOrder  string `form:"sort_order" json:"sort_order"` // asc/desc
}

// AuditLogListResponse 审计日志列表响应
type AuditLogListResponse struct {
	List       []*AuditLog      `json:"list"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Statistics *AuditStatistics `json:"statistics,omitempty"`
}

// AuditStatistics 审计统计信息
type AuditStatistics struct {
	TotalToday    int64            `json:"total_today"`
	TotalWeek     int64            `json:"total_week"`
	TotalAll      int64            `json:"total_all"`
	SuccessRate   float64          `json:"success_rate"`
	TopUsers      []TopUserStat    `json:"top_users"`
	TopModules    []TopModuleStat  `json:"top_modules"`
	ActionSummary []ActionSummary  `json:"action_summary"`
	HourlyCounts  []HourlyCount    `json:"hourly_counts"`
}

// TopUserStat 用户操作排行
type TopUserStat struct {
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

// TopModuleStat 模块操作排行
type TopModuleStat struct {
	Module  string `json:"module"`
	Count   int64  `json:"count"`
}

// ActionSummary 操作类型汇总
type ActionSummary struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// HourlyCount 每小时操作量
type HourlyCount struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// AuditRetentionUpdateReq 保留策略更新请求
type AuditRetentionUpdateReq struct {
	RetentionDays int  `json:"retention_days" binding:"min=0,max=3650"` // 0 = 永久
	IsPermanent   bool `json:"is_permanent"`
}

// ========== 审计模块/操作常量 ==========

const (
	// 模块常量
	AuditModuleAuth     = "auth"
	AuditModuleCluster  = "cluster"
	AuditModuleWorkload = "workload"
	AuditModuleNetwork  = "network"
	AuditModuleConfig   = "config"
	AuditModuleStorage  = "storage"
	AuditModuleCICD     = "cicd"
	AuditModuleRBAC     = "rbac"
	AuditModulePlatform = "platform"
	AuditModuleAI       = "ai"
	AuditModuleMonitor  = "monitoring"
	AuditModuleImage    = "image"

	// 操作类型常量
	AuditActionCreate  = "create"
	AuditActionUpdate  = "update"
	AuditActionDelete  = "delete"
	AuditActionLogin   = "login"
	AuditActionLogout  = "logout"
	AuditActionExec    = "exec"
	AuditActionApprove = "approve"
	AuditActionReject  = "reject"
	AuditActionDeploy  = "deploy"
	AuditActionScale   = "scale"
	AuditActionView    = "view"

	// 状态常量
	AuditStatusSuccess = "success"
	AuditStatusFailed  = "failed"
)

// ========== CRUD 方法 ==========

// Create 创建审计日志
// ⭐ Detail/Extra 为 JSON 类型列，MySQL 不接受空字符串，必须 Omit 跳过空值
func (a *AuditLog) Create(db *gorm.DB) error {
	var omits []string
	if a.Detail == "" {
		omits = append(omits, "Detail")
	}
	if a.Extra == "" {
		omits = append(omits, "Extra")
	}
	if len(omits) > 0 {
		return db.Omit(omits...).Create(a).Error
	}
	return db.Create(a).Error
}

// BatchCreate 批量创建审计日志
func BatchCreateAuditLogs(db *gorm.DB, logs []*AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return db.CreateInBatches(logs, 100).Error
}

// GetByID 根据ID获取审计日志
func (a *AuditLog) GetByID(db *gorm.DB, id int64) (*AuditLog, error) {
	var log AuditLog
	err := db.Where("id = ?", id).First(&log).Error
	return &log, err
}

// DeleteBefore 删除指定时间之前的审计日志
func (a *AuditLog) DeleteBefore(db *gorm.DB, beforeTimestamp int64) (int64, error) {
	result := db.Where("created_at < ?", beforeTimestamp).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}
