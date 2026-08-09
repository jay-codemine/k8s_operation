package cicd

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// 部署模式常量
const (
	DeployModeJenkins = "jenkins"
	DeployModeGitOps  = "gitops"
)

// ValidDeployModes 合法的部署模式列表
var ValidDeployModes = []string{DeployModeJenkins, DeployModeGitOps}

// ArgoCD 同步状态常量
const (
	SyncStatusSynced    = "Synced"
	SyncStatusOutOfSync = "OutOfSync"
	SyncStatusUnknown   = "Unknown"
	SyncStatusSyncing   = "Progressing"
)

// Argo Workflow 状态常量
const (
	WorkflowStatusPending   = "Pending"
	WorkflowStatusRunning   = "Running"
	WorkflowStatusSucceeded = "Succeeded"
	WorkflowStatusFailed    = "Failed"
	WorkflowStatusError     = "Error"
)

// DefaultArgoWorkflowTemplateMap 语言类型 -> Argo WorkflowTemplate 名称
var DefaultArgoWorkflowTemplateMap = map[string]string{
	LanguageTypeGo:       "go-build-workflow",
	LanguageTypeJava:     "java-spring-build-workflow",
	LanguageTypeFrontend: "frontend-build-workflow",
	LanguageTypePython:   "python-build-workflow",
}

// GitOpsConfig ArgoCD + Argo Workflows 流水线配置
type GitOpsConfig struct {
	ArgoCDAppName    string `json:"argo_app_name"`
	GitManifestRepo  string `json:"git_manifest_repo"`
	ManifestPath     string `json:"manifest_path"`
	TargetRevision   string `json:"target_revision"`
	ArgoCDProject    string `json:"argo_project"`

	WorkflowTemplate  string `json:"workflow_template"`
	WorkflowNamespace string `json:"workflow_namespace"`

	ImageRegistry  string `json:"image_registry"`
	ImageRepo      string `json:"image_repo"`
	DockerfilePath string `json:"dockerfile_path"`
	BuildContext   string `json:"build_context"`

	AutoSync      bool `json:"auto_sync"`
	PruneResource bool `json:"prune_resource"`
}

func (g GitOpsConfig) Value() (driver.Value, error) {
	return json.Marshal(g)
}

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

// ToJSONMap 将 GitOpsConfig 转换为 JSONMap
func (g *GitOpsConfig) ToJSONMap() JSONMap {
	if g == nil {
		return nil
	}
	data, _ := json.Marshal(g)
	var m JSONMap
	_ = json.Unmarshal(data, &m)
	return m
}
