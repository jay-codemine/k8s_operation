package svc

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8soperation/internal/app/requests"
	"strings"
)

// 将 DTO 端口转换为 k8s ServicePort
func buildServicePorts(in []requests.ServicePortRule) []corev1.ServicePort {
	if len(in) == 0 {
		return nil
	}
	ports := make([]corev1.ServicePort, 0, len(in))
	for _, p := range in {
		sp := corev1.ServicePort{
			Name: p.Name,
			Port: p.Port,
		}
		// TargetPort 既支持数字也支持“命名端口”
		if p.TargetPort != "" {
			sp.TargetPort = intstr.Parse(p.TargetPort)
		}
		if p.Protocol != "" {
			sp.Protocol = corev1.Protocol(p.Protocol) // TCP/UDP/SCTP
		}
		if p.NodePort != nil && *p.NodePort != 0 {
			sp.NodePort = *p.NodePort // 仅 NodePort/LB 类型可用
		}
		ports = append(ports, sp)
	}
	return ports
}

// 把 selector 两种来源合并
func mergeSelector(req *requests.KubeServiceCreateRequest) map[string]string {
	sel := map[string]string{}
	for _, kv := range req.SelectorLabels {
		if kv.Key != "" && kv.Value != "" {
			sel[kv.Key] = kv.Value
		}
	}
	for k, v := range req.Selector {
		if k != "" && v != "" {
			sel[k] = v
		}
	}
	return sel
}

func BuildServiceFromSvcReq(req *requests.KubeServiceCreateRequest) *corev1.Service {
	ports := buildServicePorts(req.Ports)
	spec := corev1.ServiceSpec{
		Type:     corev1.ServiceType(req.Type), // ClusterIP/NodePort/LoadBalancer/Headless(见下)
		Selector: mergeSelector(req),
		Ports:    ports,
	}

	// Headless 约定：Type 传 "Headless" 时转成 ClusterIP + ClusterIP=None
	if strings.EqualFold(req.Type, "Headless") {
		spec.Type = corev1.ServiceTypeClusterIP
		spec.ClusterIP = "None"
	}

	// SessionAffinity / ExternalTrafficPolicy（可选）
	if req.SessionAffinity != nil && *req.SessionAffinity != "" {
		spec.SessionAffinity = corev1.ServiceAffinity(*req.SessionAffinity) // None / ClientIP
	}
	if req.ExternalTrafficPol != nil && *req.ExternalTrafficPol != "" {
		spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyType(*req.ExternalTrafficPol) // Local / Cluster
	}

	// clusterIP（可选直传）
	if req.ClusterIP != nil && *req.ClusterIP != "" {
		spec.ClusterIP = *req.ClusterIP
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Annotations: req.Annotations,
		},
		Spec: spec,
	}
}

// BuildSimpleEndpointList 把 corev1.Endpoints 转成更直观的结构
func BuildSimpleEndpointList(ep *corev1.Endpoints) []map[string]interface{} {
	if ep == nil {
		return nil
	}

	var result []map[string]interface{}

	for _, subset := range ep.Subsets {
		var ports []int32
		for _, p := range subset.Ports {
			ports = append(ports, p.Port)
		}

		for _, addr := range subset.Addresses {
			result = append(result, map[string]interface{}{
				"ip":       addr.IP,
				"nodeName": addr.NodeName,
				"hostname": addr.Hostname,
				"ports":    ports,
				"ready":    true,
			})
		}

		for _, addr := range subset.NotReadyAddresses {
			result = append(result, map[string]interface{}{
				"ip":       addr.IP,
				"nodeName": addr.NodeName,
				"hostname": addr.Hostname,
				"ports":    ports,
				"ready":    false,
			})
		}
	}

	return result
}
