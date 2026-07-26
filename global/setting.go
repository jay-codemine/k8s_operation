package global

import (
	"k8soperation/pkg/openai"
	"k8soperation/pkg/setting"
)

// 分页默认配置
const (
	DefaultPageSize = 10   // 默认每页条数
	MaxPageSize     = 1000 // 最大每页条数
)

var (
	ServerSetting    *setting.ServerSettingS
	DatabaseSetting  *setting.DatabaseSettingS
	AppSetting       *setting.AppSettingS
	CacheSetting     *setting.CacheSettingS
	ClusterTTL       *setting.ClusterClientConfig
	JenkinsSetting   *setting.JenkinsSettingS   // Jenkins CI/CD 配置
	GitOpsSetting    *setting.GitOpsSettingS    // GitOps (ArgoCD + Argo Workflows) 配置
	SecuritySetting  *setting.SecuritySettingS  // 安全配置（加密密钥等）
	PlatformSetting  *setting.PlatformSettingsS // 平台系统设置（默认值，优先级低于数据库）
	MonitoringSetting *setting.MonitoringSettingS  // 监控配置
	AISetting         *setting.AIAssistantSettingS // AI 助手配置
	AIRegistry        *openai.Registry             // AI 多模型注册中心
	LDAPSetting       *setting.LDAPSettingS        // LDAP 认证配置
	LogControlSetting *setting.LogControlS          // 日志控制配置
	CanarySetting     *setting.CanarySettingS       // 金丝雀部署配置        // LDAP 认证配置
	CicdSetting       *setting.CicdSettingS         // CICD Worker 配置
)

// DefaultBranch 返回配置的默认 Git 分支，未配置时返回 "master"
func DefaultBranch() string {
	if JenkinsSetting != nil && JenkinsSetting.DefaultBranch != "" {
		return JenkinsSetting.DefaultBranch
	}
	return "master"
}
