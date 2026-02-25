package ingress

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"strconv"
)

// BuildIngressFromReq 转换 DTO → Kubernetes Ingress 对象
// 只设一个字段：数字 -> Number；否则 -> Name
func buildServiceBackendPort(v string) networkingv1.ServiceBackendPort {
	if n, err := strconv.Atoi(v); err == nil {
		return networkingv1.ServiceBackendPort{Number: int32(n)}
	}
	return networkingv1.ServiceBackendPort{Name: v}
}

// 可选：把字符串安全转为 PathType，空值兜底为 Prefix
func toPathType(s string) *networkingv1.PathType {
	pt := networkingv1.PathType(s)
	if pt != networkingv1.PathTypeExact &&
		pt != networkingv1.PathTypePrefix &&
		pt != networkingv1.PathTypeImplementationSpecific {
		pt = networkingv1.PathTypePrefix
	}
	return &pt
}

func BuildIngressFromReq(req *requests.KubeIngressCreateRequest) *networkingv1.Ingress {
	var rules []networkingv1.IngressRule

	for _, r := range req.Rules {
		var paths []networkingv1.HTTPIngressPath

		for _, p := range r.Paths {
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     p.Path,
				PathType: toPathType(p.PathType),
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: p.ServiceName,
						Port: buildServiceBackendPort(p.ServicePort),
					},
				},
			})
		}

		rules = append(rules, networkingv1.IngressRule{
			Host: r.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{Paths: paths},
			},
		})
	}

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: pointerStr(req.IngressClassName),
			Rules:            rules,
			TLS:              convertTLS(req.TLS),
		},
	}
}

func pointerStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func convertTLS(tls []requests.IngressTLS) []networkingv1.IngressTLS {
	var out []networkingv1.IngressTLS
	for _, t := range tls {
		out = append(out, networkingv1.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: t.SecretName,
		})
	}
	return out
}
