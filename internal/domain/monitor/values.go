package monitor

import (
	"fmt"
	"strings"
)

// AlertSeverity 告警严重级别值对象
type AlertSeverity struct{ val string }

var severities = map[string]bool{"critical": true, "warning": true, "info": true}

func NewAlertSeverity(s string) (AlertSeverity, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return AlertSeverity{}, fmt.Errorf("告警级别不能为空")
	}
	if !severities[s] {
		return AlertSeverity{}, fmt.Errorf("无效的告警级别: %s (允许 critical/warning/info)", s)
	}
	return AlertSeverity{val: s}, nil
}

func (s AlertSeverity) String() string  { return s.val }
func (s AlertSeverity) IsCritical() bool { return s.val == "critical" }

// DatasourceType 数据源类型
type DatasourceType struct{ val string }

var dsTypes = map[string]bool{"prometheus": true, "victoriametrics": true, "thanos": true, "loki": true}

func NewDatasourceType(s string) (DatasourceType, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return DatasourceType{}, fmt.Errorf("数据源类型不能为空")
	}
	return DatasourceType{val: s}, nil // 保留校验，但不强制白名单（扩展性）
}

func (t DatasourceType) String() string  { return t.val }
func (t DatasourceType) IsPrometheus() bool { return dsTypes[t.val] }
