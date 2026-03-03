package probe

import corev1 "k8s.io/api/core/v1"

// ParseProtocol 把字符串转换成 corev1.Protocol，默认返回 TCP
func ParseProtocol(proto string) corev1.Protocol {
	switch proto {
	case "UDP", "udp":
		return corev1.ProtocolUDP
	case "SCTP", "sctp":
		return corev1.ProtocolSCTP
	case "TCP", "tcp", "":
		return corev1.ProtocolTCP
	default:
		// 防御：遇到未知协议，仍然返回 TCP
		return corev1.ProtocolTCP
	}
}
