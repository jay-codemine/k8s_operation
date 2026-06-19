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
		vp.SetConfigType("yaml")
		// 读取 YAML 原文，在文本层面展开 ${ENV_VAR}，然后再交给 Viper 解析
		// 这样 bool/int/duration 等非字符串字段也能正确解析
		raw, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		expanded := os.ExpandEnv(string(raw))
		if err := vp.ReadConfig(strings.NewReader(expanded)); err != nil {
			return nil, err
		}
		return &Setting{vp: vp}, nil
	}

	// 1. 创建一个新的 viper 实例（用于读取和管理配置）
	vp := viper.New()

	// 2. 设置配置文件名（不包含扩展名）
	vp.SetConfigName("config")

	// 3. 添加配置文件搜索路径
	vp.AddConfigPath("configs")
	vp.AddConfigPath("/app/configs")

	// 获取可执行文件所在目录，支持从任意工作目录启动
	if exe, err := os.Executable(); err == nil {
		vp.AddConfigPath(filepath.Join(filepath.Dir(exe), "configs"))
	}

	// 4. 设置配置文件类型为 YAML
	vp.SetConfigType("yaml")

	// 5. 尝试读取配置文件
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	// 对通过文件搜索路径加载的配置也做环境变量展开
	cfgFile := vp.ConfigFileUsed()
	if cfgFile != "" {
		raw, err := os.ReadFile(cfgFile)
		if err == nil {
			expanded := os.ExpandEnv(string(raw))
			// 重新创建 viper 实例，用展开后的内容
			vp2 := viper.New()
			vp2.SetConfigType("yaml")
			if err := vp2.ReadConfig(strings.NewReader(expanded)); err == nil {
				vp = vp2
			}
		}
	}

	return &Setting{vp: vp}, nil
}
