package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ==================== 部署模式常量 ====================

const (
	DeployModeJenkins = "jenkins" // Push-based CI/CD via Jenkins
	DeployModeGitOps  = "gitops"  // Pull-based CD via ArgoCD + Argo Workflows
)

// ValidDeployModes 合法的部署模式列表
var ValidDeployModes = []string{DeployModeJenkins, DeployModeGitOps}

// ==================== ArgoCD 同步状态常量 ====================

const (
	SyncStatusSynced    = "Synced"
	SyncStatusOutOfSync = "OutOfSync"
	SyncStatusUnknown   = "Unknown"
	SyncStatusSyncing   = "Progressing"
)

// ==================== Argo Workflow 状态常量 ====================

const (
	WorkflowStatusPending   = "Pending"
	WorkflowStatusRunning   = "Running"
	WorkflowStatusSucceeded = "Succeeded"
	WorkflowStatusFailed    = "Failed"
	WorkflowStatusError     = "Error"
)

// ==================== Argo Workflow 模板映射 ====================

// DefaultArgoWorkflowTemplateMap 语言类型 -> Argo WorkflowTemplate 名称
var DefaultArgoWorkflowTemplateMap = map[string]string{
	LanguageTypeGo:       "go-build-workflow",
	LanguageTypeJava:     "java-spring-build-workflow",
	LanguageTypeFrontend: "frontend-build-workflow",
	LanguageTypePython:   "python-build-workflow",
}

// ==================== GitOpsConfig 结构体 ====================

// GitOpsConfig ArgoCD + Argo Workflows 流水线配置（存储在 cicd_pipeline.gitops_config JSON 列）
type GitOpsConfig struct {
	// ArgoCD 配置
	ArgoCDAppName   string `json:"argo_app_name"`   // ArgoCD Application 名称
	GitManifestRepo string `json:"git_manifest_repo"` // Git 仓库（存放 K8s manifests）
	ManifestPath    string `json:"manifest_path"`    // manifests 路径 (如 "manifests/overlays/prod")
	TargetRevision  string `json:"target_revision"`   // 目标分支/标签
	ArgoCDProject   string `json:"argo_project"`      // ArgoCD Project (默认 "default")

	// Argo Workflows 配置
	WorkflowTemplate  string `json:"workflow_template"`  // Argo WorkflowTemplate 名称
	WorkflowNamespace string `json:"workflow_namespace"` // WorkflowTemplate 命名空间

	// 镜像构建配置
	ImageRegistry  string `json:"image_registry"`  // e.g., registry.example.com/myproject
	ImageRepo      string `json:"image_repo"`      // 镜像仓库名 (e.g., myapp)
	DockerfilePath string `json:"dockerfile_path"` // Dockerfile 路径 (默认 "Dockerfile")
	BuildContext   string `json:"build_context"`   // 构建上下文 (默认 ".")

	// 同步策略
	AutoSync      bool `json:"auto_sync"`       // 启用 ArgoCD 自动同步
	PruneResource bool `json:"prune_resource"` // 启用资源清理 (不在 Git 中的资源自动删除)
}

// Value 实现 driver.Valuer 接口（用于 GORM JSON 列写入）
func (g GitOpsConfig) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan 实现 sql.Scanner 接口（用于 GORM JSON 列读取）
func (g *GitOpsConfig) Scan(value interface{}) error {
	if value == nil {
		*g = GitOpsConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for GitOpsConfig")
	}
	return json.Unmarshal(bytes, g)
}

// GitOpsConfigFromJSONMap 从 JSONMap 解析 GitOpsConfig
func GitOpsConfigFromJSONMap(m JSONMap) (*GitOpsConfig, error) {
	if m == nil {
		return nil, nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var config GitOpsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ToJSONMap 将 GitOpsConfig 转换为 JSONMap（用于存储到 CicdPipeline.GitOpsConfig）
func (g *GitOpsConfig) ToJSONMap() JSONMap {
	if g == nil {
		return nil
	}
	data, _ := json.Marshal(g)
	var m JSONMap
	_ = json.Unmarshal(data, &m)
	return m
}
