package monitor

import "time"

// QueryTimeout 根据配置值返回 Prometheus/Loki 查询超时（秒）
func QueryTimeout(timeoutSeconds int) time.Duration {
	if timeoutSeconds > 0 {
		return time.Duration(timeoutSeconds) * time.Second
	}
	return 30 * time.Second
}
