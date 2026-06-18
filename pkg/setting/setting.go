package setting

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Setting 结构体定义了一个配置结构体，用于管理应用程序的配置信息
type Setting struct {
	// vp 是一个指向 viper.Viper 类型的指针，用于读取和解析配置文件
	vp *viper.Viper
}

// NewSetting 初始化配置读取
// 返回值：
//
//	*Setting - 包含 viper 配置实例的结构体
//	error    - 如果读取配置失败则返回错误
func NewSetting() (*Setting, error) {
	// 如果环境变量 APP_CONFIG 指定了配置文件路径，则直接使用该路径
	if configPath := os.Getenv("APP_CONFIG"); configPath != "" {
		vp := viper.New()
		vp.SetConfigFile(configPath)
		if err := vp.ReadInConfig(); err != nil {
			return nil, err
		}
		expandEnvInViper(vp)
		return &Setting{vp: vp}, nil
	}

	// 1. 创建一个新的 viper 实例（用于读取和管理配置）
	vp := viper.New()

	// 2. 设置配置文件名（不包含扩展名）
	//    这里是 "config" → 会去找 config.yaml（或 config.json 等，取决于 SetConfigType）
	vp.SetConfigName("config")

	// 3. 添加配置文件搜索路径（按优先级从高到低）
	//    支持本地开发（相对路径）和容器部署（绝对路径）
	vp.AddConfigPath("configs")
	vp.AddConfigPath("/app/configs")

	// 获取可执行文件所在目录，支持从任意工作目录启动
	if exe, err := os.Executable(); err == nil {
		vp.AddConfigPath(filepath.Join(filepath.Dir(exe), "configs"))
	}

	// 4. 设置配置文件类型为 YAML
	//    即使文件扩展名不是 .yaml，也会按 YAML 格式解析
	vp.SetConfigType("yaml")

	// 5. 尝试读取配置文件
	//    如果文件不存在、路径错误、格式不正确等，都会返回错误
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	// 6. 展开配置值中的 ${ENV_VAR} 环境变量占位符
	expandEnvInViper(vp)

	// 7. 将 viper 实例封装进自定义 Setting 结构体并返回
	return &Setting{vp: vp}, nil
}

// expandEnvInViper 遍历 viper 中所有配置项，将字符串值中的 ${ENV_VAR} 替换为实际环境变量
// 支持 ConfigMap 中使用 ${DB_PASSWORD} 等占位符，由 Secret 注入的环境变量在运行时替换
func expandEnvInViper(vp *viper.Viper) {
	for _, key := range vp.AllKeys() {
		val := vp.Get(key)
		if strVal, ok := val.(string); ok {
			if strings.Contains(strVal, "${") {
				vp.Set(key, os.ExpandEnv(strVal))
			}
		}
	}
}
