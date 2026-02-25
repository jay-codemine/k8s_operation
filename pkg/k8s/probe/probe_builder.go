package probe

import (
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/dataselect"
)

// BuildProbe 根据请求 DTO 构造 Kubernetes Probe
func BuildProbe(detail requests.HealthCheckDetail) *corev1.Probe {
	probeHandler := dataselect.GetContainerProbe(detail)

	return &corev1.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: detail.InitialDelaySeconds,
		TimeoutSeconds:      detail.TimeoutSeconds,
		PeriodSeconds:       detail.PeriodSeconds,
		SuccessThreshold:    detail.SuccessThreshold,
		FailureThreshold:    detail.FailureThreshold,
	}
}
