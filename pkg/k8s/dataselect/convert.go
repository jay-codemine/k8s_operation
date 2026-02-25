package dataselect

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"strings"
)

// ToCells 是一个泛型函数，将任意类型的切片转换为 DataCell 类型的切片
// 参数:
//
//	items - 任意类型的切片，需要被转换
//	convert - 转换函数，将类型 T 的元素转换为 DataCell 类型
//
// 返回值:
//
//	[]DataCell - 转换后的 DataCell 类型切片
func ToCells[T any](items []T, convert func(T) DataCell) []DataCell {
	// 创建一个与输入切片长度相同的 DataCell 类型切片
	cells := make([]DataCell, len(items))
	// 遍历输入切片，使用转换函数将每个元素转换为 DataCell 类型
	for i, item := range items {
		cells[i] = convert(item)
	}
	// 返回转换后的切片
	return cells
}

func FromCells[T any](cells []DataCell, convert func(DataCell) T) []T {
	// 创建一个与cells长度相同的结果切片，类型为T
	result := make([]T, len(cells))
	// 遍历cells切片中的每个元素
	for i, cell := range cells {
		// 对每个元素进行转换，并将结果存入结果切片的对应位置
		result[i] = convert(cell)
	}
	// 返回转换后的结果切片
	return result
}

func GetLabelsMap(labels []requests.Label) map[string]string {
	// 创建一个字符串到字符串的映射（map）作为结果容器
	res := map[string]string{}
	// 遍历labels切片中的每个label元素
	for _, label := range labels {
		// 将每个label的Key作为键，Value作为值存入结果map中
		res[label.Key] = label.Value
	}
	// 返回构建好的map
	return res
}

func ConvertEnvVarSpec(envs []requests.EnvironmentVariable) []corev1.EnvVar {
	// 将环境变量列表转换为Kubernetes核心API中的EnvVar数组格式
	// 这个函数接收一个环境变量列表，并将其转换为Kubernetes环境变量所需的格式
	var res []corev1.EnvVar    // 初始化一个空的Kubernetes EnvVar类型的切片
	for _, env := range envs { // 遍历输入的环境变量列表
		res = append(res, corev1.EnvVar{ // 将每个环境变量转换为Kubernetes EnvVar格式并添加到结果切片中
			Name:  env.Name,  // 设置环境变量名称
			Value: env.Value, // 设置环境变量值
		})
	}

	return res // 返回转换后的Kubernetes环境变量列表
}

func GetContainerProbe(probe requests.HealthCheckDetail) corev1.ProbeHandler {
	var ph corev1.ProbeHandler // 定义一个 ProbeHandler 类型的变量 ph，用于存储探测处理器

	// 解析端口（当前 DTO 只有数值端口）
	port := intstr.FromInt32(probe.Port) // 将输入的端口号转换为 intstr.IntOrString 类型

	// 解析协议，兼容大小写，非法值回退 HTTP
	parseScheme := func(s string) corev1.URIScheme { // 定义一个匿名函数，用于解析协议类型
		switch strings.ToUpper(strings.TrimSpace(s)) { // 将输入的协议字符串转换为大写并去除前后空格
		case "HTTPS":
			return corev1.URISchemeHTTPS // 如果是 HTTPS，则返回 HTTPS 协议
		default:
			return corev1.URISchemeHTTP // 默认返回 HTTP 协议
		}
	}

	switch strings.ToUpper(strings.TrimSpace(probe.Type)) { // 根据探测类型执行不同的处理逻辑
	case "HTTP":
		var headers []corev1.HTTPHeader // 定义 HTTP 头部切片
		if len(probe.HttpHeader) > 0 {  // 如果存在 HTTP 头部
			headers = make([]corev1.HTTPHeader, 0, len(probe.HttpHeader)) // 初始化 HTTP 头部切片
			for _, h := range probe.HttpHeader {                          // 遍历输入的 HTTP 头部
				if h.Name == "" { // 如果头部名称为空，则跳过
					continue
				}
				headers = append(headers, corev1.HTTPHeader{ // 将有效的 HTTP 头部添加到切片中
					Name:  h.Name,
					Value: h.Value,
				})
			}
		}
		ph.HTTPGet = &corev1.HTTPGetAction{ // 设置 HTTP 探测处理器
			Path:        probe.Path,                  // 设置探测路径
			Port:        port,                        // 设置探测端口
			Scheme:      parseScheme(probe.Protocol), // 解析并设置协议类型
			HTTPHeaders: headers,                     // 设置 HTTP 头部
		}

	case "TCP":
		ph.TCPSocket = &corev1.TCPSocketAction{ // 设置 TCP 探测处理器
			Port: port, // 设置探测端口
			// Host 可选：留空表示连 Pod IP
		}

	case "EXEC":
		cmd := strings.Fields(strings.TrimSpace(probe.Command)) // 将命令字符串分割为命令切片
		// 如果你允许空命令，这里可再做兜底或日志
		ph.Exec = &corev1.ExecAction{Command: cmd} // 设置命令执行探测处理器

	default:
		global.Logger.Errorf("probe type %q is not supported; expect HTTP/TCP/EXEC", probe.Type) // 记录错误日志，表示不支持的探测类型
	}

	return ph // 返回探测处理器
}
