package models

import dm "k8soperation/internal/domain/ai"

// ===== 类型别名 =====

type (
	AIOpsAnalysisRecord    = dm.AIOpsAnalysisRecord
	AIOpsInspectionReport  = dm.AIOpsInspectionReport
)

// ===== 常量别名 =====

const (
	AIOpsTypeAlertAnalysis = dm.AIOpsTypeAlertAnalysis
	AIOpsTypeLogDiagnosis  = dm.AIOpsTypeLogDiagnosis
	AIOpsTypeInspection    = dm.AIOpsTypeInspection
)

const (
	AIOpsStatusSuccess = dm.AIOpsStatusSuccess
	AIOpsStatusFailed  = dm.AIOpsStatusFailed
	AIOpsStatusTimeout = dm.AIOpsStatusTimeout
)

const (
	InspectionTypeScheduled = dm.InspectionTypeScheduled
	InspectionTypeManual    = dm.InspectionTypeManual
)

const (
	InspectionLevelHealthy  = dm.InspectionLevelHealthy
	InspectionLevelWarning  = dm.InspectionLevelWarning
	InspectionLevelCritical = dm.InspectionLevelCritical
)
