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
		// 读取 GitOps 配置（ArgoCD + Argo Workflows）
		// 对应 config.yaml 中的：
		// GitOps:
		if err = s.ReadSection("GitOps", &global.GitOpsSetting); err != nil {
			// GitOps 配置可选，不存在时不报错
			log.Println("[GitOps] 配置块未找到，GitOps 模式将不可用")
			global.GitOpsSetting = nil
		} else if global.GitOpsSetting != nil {
			if global.GitOpsSetting.ArgoCDURL == "" {
				log.Println("[GitOps] ArgoCD URL 未配置，GitOps 模式将不可用")
				global.GitOpsSetting = nil
			} else {
				log.Printf("[GitOps] 配置加载成功: argocd=%s, workflows=%s\n",
					global.GitOpsSetting.ArgoCDURL,
					global.GitOpsSetting.ArgoWorkflowsURL)
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
		}

		// 读取 Monitoring 配置
		// 对应 config.yaml 中的：
		// Monitoring:
		if err = s.ReadSection("Monitoring", &global.MonitoringSetting); err != nil {
			// Monitoring 可选
			log.Println("[Monitoring] 配置块未找到")
			global.MonitoringSetting = nil
		}

		// 读取 AI 助手配置
		// 对应 config.yaml 中的：
		// AIAssistant:
		if err = s.ReadSection("AIAssistant", &global.AISetting); err != nil {
			log.Println("[AIAssistant] 配置块未找到")
			global.AISetting = nil
		}
		if global.AISetting != nil && global.AISetting.Enabled {
			global.AIRegistry = openai.NewRegistry(global.AISetting)
			log.Println("[AIAssistant] AI 多模型注册中心初始化完成")
		}

		// 读取 LDAP 配置
		// 对应 config.yaml 中的：
		// LDAP:
		if err = s.ReadSection("LDAP", &global.LDAPSetting); err != nil {
			// LDAP 可选
			log.Println("[LDAP] 配置块未找到")
			global.LDAPSetting = nil
		}

	// 读取 LogControl 配置
	if err = s.ReadSection("LogControl", &global.LogControlSetting); err != nil || global.LogControlSetting == nil {
		log.Println("[LogControl] 配置块未找到，使用默认值（全部抑制）")
		global.LogControlSetting = &setting.LogControlS{
			SuppressPrometheusQueryWarn: true,
			SuppressK8sClusterInitWarn:  true,
			SuppressK8sClusterInitError: true,
		}
	}
	if global.LogControlSetting != nil {
		log.Printf("[LogControl] SuppressPrometheusQuery=%v SuppressK8sInitWarn=%v SuppressK8sInitError=%v",
			global.LogControlSetting.SuppressPrometheusQueryWarn,
			global.LogControlSetting.SuppressK8sClusterInitWarn,
			global.LogControlSetting.SuppressK8sClusterInitError)
	}

	// 读取 Canary 配置
	if err = s.ReadSection("Canary", &global.CanarySetting); err != nil || global.CanarySetting == nil {
		log.Println("[Canary] 配置块未找到，使用默认值")
		global.CanarySetting = &setting.CanarySettingS{
			Enabled: true, DefaultReplicas: 1, DefaultTrafficRatio: 10,
			DefaultDurationSec: 300, HealthCheckInterval: 30, MaxTrafficRatio: 50,
		}
	}

		// 读取 CicdWorker 配置（部署任务消费协程数）
		// 对应 config.yaml 中的：
		// CicdWorker:
		if err = s.ReadSection("CicdWorker", &global.CicdWorkerSetting); err != nil || global.CicdWorkerSetting == nil {
			log.Println("[CicdWorker] 配置块未找到，使用默认并发数 3")
			global.CicdWorkerSetting = &setting.CicdWorkerSettingS{Concurrency: 3}
		}
		if global.CicdWorkerSetting.Concurrency <= 0 {
			log.Println("[CicdWorker] Concurrency 非法，回退默认并发数 3")
			global.CicdWorkerSetting.Concurrency = 3
		}

		// 注入错误码配置到 errorcode 包
		if global.ErrorCodeSetting != nil && global.ErrorCodeSetting.AllowOverride {
			errorcode.SetAllowOverride(true)
		}

		return nil
	}
