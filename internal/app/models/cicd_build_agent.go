package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type CicdBuildAgent = dm.CicdBuildAgent

// ===== 常量别名 =====
const (
	AgentCategoryObservability = dm.AgentCategoryObservability
	AgentCategoryDiagnostics   = dm.AgentCategoryDiagnostics
	AgentCategorySecurity      = dm.AgentCategorySecurity
	AgentCategoryCustom        = dm.AgentCategoryCustom

	AgentScopeJava   = dm.AgentScopeJava
	AgentScopeGo     = dm.AgentScopeGo
	AgentScopePython = dm.AgentScopePython
	AgentScopeAll    = dm.AgentScopeAll

	AgentStatusActive   = dm.AgentStatusActive
	AgentStatusInactive = dm.AgentStatusInactive
)

// ===== 变量别名 =====
var (
	ValidAgentCategories = dm.ValidAgentCategories
	ValidAgentScopes     = dm.ValidAgentScopes
)
