package probe

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8soperation/internal/app/requests"
)

// 将请求里的端口映射转换为容器端口声明（ContainerPort）
// 说明：ContainerPort 只需要容器内暴露的端口和协议，TargetPort 概念属于 Service 层，这里不会用到。
func ConvertContainerPorts(ports []requests.PortMapping) []corev1.ContainerPort {
	out := make([]corev1.ContainerPort, 0, len(ports))
	for _, p := range ports {
		if p.Port <= 0 {
			continue
		}
		out = append(out, corev1.ContainerPort{
			ContainerPort: p.Port,
			Protocol:      parseProtocol(p.Protocol),
			// 可选：Name/HostPort/HostIP 等按需补充
		})
	}
	return out
}

// 如果你要同时创建 Service，可用这个把同一份端口映射转换为 ServicePort
func convertServicePorts(ports []requests.PortMapping) []corev1.ServicePort {
	out := make([]corev1.ServicePort, 0, len(ports))
	for _, p := range ports {
		if p.Port <= 0 {
			continue
		}
		out = append(out, corev1.ServicePort{
			Port:       p.Port,                         // Service 对外端口
			TargetPort: intstr.FromInt32(p.TargetPort), // 指向 Pod 里的端口
			Protocol:   parseProtocol(p.Protocol),
			// 可选：Name/NodePort 等
		})
	}
	return out
}

// 协议解析：默认 TCP，支持 UDP/SCTP（大小写不敏感）
func parseProtocol(s string) corev1.Protocol {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UDP":
		return corev1.ProtocolUDP
	case "SCTP":
		return corev1.ProtocolSCTP
	default:
		return corev1.ProtocolTCP
	}
}
