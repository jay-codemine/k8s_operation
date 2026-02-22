package setting

import "github.com/spf13/viper"

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
	// 1. 创建一个新的 viper 实例（用于读取和管理配置）
	vp := viper.New()

	// 2. 设置配置文件名（不包含扩展名）
	//    这里是 "config" → 会去找 config.yaml（或 config.json 等，取决于 SetConfigType）
	vp.SetConfigName("config")

	// 3. 添加配置文件搜索路径
	//    这里是 configs/ 目录，也就是说程序会到 ./configs/ 下查找文件
	//    可以多次调用 AddConfigPath 添加多个搜索路径
	vp.AddConfigPath("configs")

	// 4. 设置配置文件类型为 YAML
	//    即使文件扩展名不是 .yaml，也会按 YAML 格式解析
	vp.SetConfigType("yaml")

	// 5. 尝试读取配置文件
	//    如果文件不存在、路径错误、格式不正确等，都会返回错误
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	// 6. 将 viper 实例封装进自定义 Setting 结构体并返回
	return &Setting{vp: vp}, nil
}
