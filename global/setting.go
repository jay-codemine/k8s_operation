package global

import (
	"k8soperation/pkg/setting"
)

// 分页默认配置
const (
	DefaultPageSize = 10   // 默认每页条数
	MaxPageSize     = 1000 // 最大每页条数
)

var (
	ServerSetting   *setting.ServerSettingS
	DatabaseSetting *setting.DatabaseSettingS
	AppSetting      *setting.AppSettingS
	CacheSetting    *setting.CacheSettingS
	ClusterTTL      *setting.ClusterClientConfig
	JenkinsSetting  *setting.JenkinsSettingS // Jenkins CI/CD 配置
)
