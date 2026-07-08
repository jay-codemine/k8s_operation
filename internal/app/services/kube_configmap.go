package services

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/configmap"
)

func (s *Services) KubeCreateConfigMap(ctx context.Context, cli *K8sClients, req *requests.KubeConfigMapCreateRequest,
) (*corev1.ConfigMap, error) {
	return configmap.CreateConfigMap(ctx, cli.Kube, req)
}

// KubeConfigMapList 获取 ConfigMap 列表（支持名称过滤 + 分页）
func (s *Services) KubeConfigMapList(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapListRequest) ([]corev1.ConfigMap, int, error) {
	return configmap.GetConfigMapList(ctx, cli.Kube, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeConfigMapDetail(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapDetailRequest) (*corev1.ConfigMap, error) {
	return configmap.GetConfigMapDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

func (s *Services) KubeConfigMapDelete(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapDeleteRequest) error {
	return configmap.DeleteConfigMap(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeConfigMapPatch Strategic Merge Patch 更新 ConfigMap
func (s *Services) KubeConfigMapPatch(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapUpdateRequest) (*corev1.ConfigMap, error) {
	return configmap.PatchConfigMap(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// KubeConfigMapUpdate JSON 全量更新 ConfigMap
func (s *Services) KubeConfigMapUpdate(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapUpdateRequest) (*corev1.ConfigMap, error) {
	return configmap.UpdateConfigMapJson(ctx, cli.Kube, param.Namespace, param.Content)
}

// KubeConfigMapYaml 获取 ConfigMap YAML
func (s *Services) KubeConfigMapYaml(ctx context.Context, cli *K8sClients, namespace, name string) (string, error) {
	return configmap.GetConfigMapYaml(ctx, cli.Kube, namespace, name)
}

// KubeConfigMapUpdateData 更新 ConfigMap data 字段
func (s *Services) KubeConfigMapUpdateData(ctx context.Context, cli *K8sClients, param *requests.KubeConfigMapUpdateDataRequest) (*corev1.ConfigMap, error) {
	return configmap.UpdateConfigMapData(ctx, cli.Kube, param.Namespace, param.Name, param.Data)
}

// KubeConfigMapApplyYaml 从 YAML 字符串更新 ConfigMap
func (s *Services) KubeConfigMapApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.ConfigMap, error) {
	return configmap.ApplyConfigMapYaml(ctx, cli.Kube, yamlContent)
}
