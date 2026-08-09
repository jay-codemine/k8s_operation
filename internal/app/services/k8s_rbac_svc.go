package services

import "k8soperation/internal/app/services/k8srbac"

// K8sRBACSvc 返回 K8s RBAC 服务（操作 K8s API，无 DB 依赖）
func (s *Services) K8sRBACSvc() *k8srbac.K8sRBACService {
	return k8srbac.NewK8sRBACService(s.logger)
}
