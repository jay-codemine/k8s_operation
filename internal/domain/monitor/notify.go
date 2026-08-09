package monitor

// AlertNotification 告警通知数据（跨域 DTO，供 Worker / AIOps 使用）
type AlertNotification struct {
	RuleName    string            `json:"rule_name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"` // firing/resolved
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Value       string            `json:"value"`
	FiredAt     int64             `json:"fired_at"`
	ResolvedAt  int64             `json:"resolved_at"`
	Labels      map[string]string `json:"labels"`
}
