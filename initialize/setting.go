package initialize

import (
	"log"

	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/openai"
	"k8soperation/pkg/setting"
	"k8soperation/pkg/utils"
)

// SetupSetting 初始化全局配置
// 1. 创建 viper 实例（读取配置文件）
// 2. 将配置文件中 "Server" 部分映射到 global.Setting
// SetupSetting 初始化全局配置
//
// 作用说明：
// 1. 创建配置读取器（viper 封装）
// 2. 按 YAML 顶层 key 分段读取配置
// 3. 将配置反序列化到 global 包中的全局只读配置
// 4. 注入部分配置到子模块（如 errorcode）
func SetupSetting() error {
	// 创建 Setting 实例
	// - 内部一般封装 viper
	// - 负责读取 config.yaml / env / 默认值
	s, err := setting.NewSetting()
	if err != nil {
		return err
	}

	// 读取 Server 配置
	// 对应 config.yaml 中的：
	// Server:
	if err = s.ReadSection("Server", &global.ServerSetting); err != nil {
		return err
	}

	//  读取 App 配置
	// 对应 config.yaml 中的：
	// App:
	if err = s.ReadSection("App", &global.AppSetting); err != nil {
		return err
	}

	// 读取 Database 配置
	// 对应 config.yaml 中的：
	// Database:
	if err = s.ReadSection("Database", &global.DatabaseSetting); err != nil {
		return err
	}

	// 读取 Cache（Redis）配置
	// 对应 config.yaml 中的：
	// Cache:
	if err = s.ReadSection("Cache", &global.CacheSetting); err != nil {
		return err
	}

	// 读取 Pod 日志配置
	// 注意：这里的 key 必须是 PodLog
	// 对应 config.yaml 中的：
	// PodLog:
	if err = s.ReadSection("PodLog", &global.PodLogSetting); err != nil {
		return err
	}

	if err = s.ReadSection("Pod", &global.PodSetting); err != nil {
		return err
	}

	// 读取 Node 配置
	// 前提：config.yaml 中必须存在 Node 段
	// 如果暂时不需要 Node，可以：
	// - 在 YAML 中补 Node
	// - 或改 ReadSection 为“允许缺省”
	if err = s.ReadSection("Node", &global.NodeSetting); err != nil {
		return err
	}

	// 读取错误码配置
	// 对应 config.yaml 中的：
	// ErrorCode:
	if err = s.ReadSection("ErrorCode", &global.ErrorCodeSetting); err != nil {
		return err
	}

	// 读取 K8s Cluster Client 配置（TTL / Jitter）
	// 对应 config.yaml 中的：
	// ClusterClient:
	if err = s.ReadSection("ClusterClient", &global.ClusterTTL); err != nil {
		return err
	}

	// 读取 Jenkins 配置
	// 对应 config.yaml 中的：
	// Jenkins:
	if err = s.ReadSection("Jenkins", &global.JenkinsSetting); err != nil {
		// Jenkins 配置可选，不存在时不报错
		log.Println("[Jenkins] 配置块未找到，CI/CD 功能将不可用")
		global.JenkinsSetting = nil
	} else if global.JenkinsSetting != nil {
		// 校验关键字段：如果 URL 为空则表示未启用，置为 nil
		if global.JenkinsSetting.URL == "" {
			log.Println("[Jenkins] URL 未配置，CI/CD 功能将不可用")
			global.JenkinsSetting = nil
		} else {
			log.Printf("[Jenkins] 配置加载成功: url=%s, username=%s, has_token=%v\n",
				global.JenkinsSetting.URL,
				global.JenkinsSetting.Username,
				global.JenkinsSetting.APIToken != "",
			)
			if global.JenkinsSetting.Username == "" || global.JenkinsSetting.APIToken == "" {
				log.Println("[Jenkins] 凭据不完整，请配置 Username 和 APIToken")
			}
		}
	}

	// 读取 Security 配置
	// 对应 config.yaml 中的：
	// Security:
	if err = s.ReadSection("Security", &global.SecuritySetting); err != nil {
		// 安全配置可选，使用默认值
		log.Println("[Security] 配置块未找到，使用默认安全配置")
		global.SecuritySetting = &setting.SecuritySettingS{
			KubeConfigEncryptKey:  "k8s-operation-default-secret-key",
			PasswordBcryptCost:    10,
			AutoEncryptLegacyData: false,
		}
	}

	// 初始化全局加密服务
	if global.SecuritySetting != nil && global.SecuritySetting.KubeConfigEncryptKey != "" {
		utils.InitGlobalCrypto(global.SecuritySetting.KubeConfigEncryptKey)
	} else {
		log.Println("[Security] 警告: 加密密钥未配置，数据将不加密存储")
	}

	// 读取 PlatformSettings 配置（平台系统设置默认值）
	// 对应 config.yaml 中的：
	// PlatformSettings:
	if err = s.ReadSection("PlatformSettings", &global.PlatformSetting); err != nil {
		// PlatformSettings 可选，使用程序内默认值
		log.Println("[PlatformSettings] 配置块未找到，使用内置默认值")
		global.PlatformSetting = &setting.PlatformSettingsS{
			Basic: setting.PlatformBasicSettings{
				DefaultPage:    "/clusters",
				DefaultCluster: "auto",
				Language:       "zh-CN",
				Timezone:       "Asia/Shanghai",
			},
			Security: setting.PlatformSecuritySettings{
				SessionTimeout: 120,
				Enable2FA:      false,
				PasswordPolicy: "medium",
				AuditRetention: 30,
			},
			Alert: setting.PlatformAlertSettings{
				CPUThreshold:  80,
				MemThreshold:  80,
				DiskThreshold: 85,
				AlertSilence:  15,
			},
			Notification: setting.PlatformNotificationSettings{
				EnableEmail:    false,
				EnableDingTalk: false,
				EnableWebhook:  false,
			},
			About: setting.PlatformAboutSettings{
				Version:    "2.0.0",
				BuildDate:  "2026-03-04",
				GoVersion:  "1.21",
				VueVersion: "3.5.13",
				DBType:     "MySQL 8.0",
				K8sSupport: "v1.25+",
			},
		}
	} else {
		// PlatformSettings 加载成功，静默不打印
	}

	// 读取 AI 助手配置
	// 对应 config.yaml 中的：
	// Monitoring:
	if err = s.ReadSection("Monitoring", &global.MonitoringSetting); err != nil {
		log.Println("[Monitoring] 配置块未找到，监控功能将不可用")
		global.MonitoringSetting = &setting.MonitoringSettingS{Enabled: false}
	} else if global.MonitoringSetting != nil {
		if global.MonitoringSetting.PrometheusURL == "" {
			global.MonitoringSetting.Enabled = false
			log.Println("[Monitoring] PrometheusURL 未配置，监控功能将不可用")
		} else {
			log.Printf("[Monitoring] 配置加载成功: enabled=%v, url=%s, timeout=%ds\n",
				global.MonitoringSetting.Enabled,
				global.MonitoringSetting.PrometheusURL,
				global.MonitoringSetting.QueryTimeout)
		}
	}

	// 读取 AI 助手配置
	// 对应 config.yaml 中的：
	// AIAssistant:
	if err = s.ReadSection("AIAssistant", &global.AISetting); err != nil {
		log.Println("[AIAssistant] 配置块未找到，AI 助手功能将不可用")
		global.AISetting = &setting.AIAssistantSettingS{Enabled: false}
	} else if global.AISetting != nil && global.AISetting.Enabled {
		// 初始化多模型Provider注册中心
		global.AIRegistry = openai.NewRegistry(global.AISetting)
		providers := global.AIRegistry.ListProviders()
		if len(providers) == 0 {
			log.Println("[AIAssistant] 警告: 无有效的 AI 提供商配置（检查 APIKey），AI 助手功能将不可用")
			global.AISetting.Enabled = false
		} else {
			var modelCount int
			for _, p := range providers {
				modelCount += len(p.Models)
			}
			log.Printf("[AIAssistant] AI 助手已启用: %d 个提供商, %d 个模型, 默认: %s/%s\n",
				len(providers), modelCount,
				global.AIRegistry.GetDefaultProviderID(),
				global.AIRegistry.GetDefaultModelID())
		}
	}
	// 将 ErrorCode 配置注入 errorcode 包
	// - AllowOverride=true：开发环境，允许错误码覆盖
	// - AllowOverride=false：生产环境，发现重复直接 panic
	if global.ErrorCodeSetting != nil {
		errorcode.SetAllowOverride(global.ErrorCodeSetting.AllowOverride)
	} else {
		log.Println("[ErrorCode] 配置未加载，使用默认值 (AllowOverride=true)")
		errorcode.SetAllowOverride(true)
	}

	// 注册所有错误码
	// 一般在这里做启动期校验
	errorcode.Register()

	// 初始化成功
	return nil
}
