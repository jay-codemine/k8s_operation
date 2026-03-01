package initialize

import (
	"log"

	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/setting"
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

	// 将 ErrorCode 配置注入 errorcode 包
	// - AllowOverride=true：开发环境，允许错误码覆盖
	// - AllowOverride=false：生产环境，发现重复直接 panic
	errorcode.SetAllowOverride(global.ErrorCodeSetting.AllowOverride)

	// 注册所有错误码
	// 一般在这里做启动期校验
	errorcode.Register()

	// 初始化成功
	return nil
}
