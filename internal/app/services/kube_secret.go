package services

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/secret"
)

func (s *Services) KubeCreateSecret(ctx context.Context, cli *K8sClients,
	req *requests.KubeSecretCreateRequest) (*corev1.Secret, error) {
	return secret.CreateSecret(ctx, cli.Kube, req)
}

// KubeSecretList 获取 Secret 列表（支持名称过滤 + 分页）
func (s *Services) KubeSecretList(ctx context.Context, cli *K8sClients, param *requests.KubeSecretListRequest) ([]corev1.Secret, int, error) {
	return secret.GetSecretList(ctx, cli.Kube, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeSecretDetail(ctx context.Context, cli *K8sClients, param *requests.KubeSecretDetailRequest) (*corev1.Secret, error) {
	return secret.GetSecretDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// 删除 Secret
func (s *Services) KubeSecretDelete(ctx context.Context, cli *K8sClients, param *requests.KubeSecretDeleteRequest) error {
	return secret.DeleteSecret(ctx, cli.Kube, param.Name, param.Namespace)
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeSecretPatch(ctx context.Context, cli *K8sClients, param *requests.KubeSecretUpdateRequest) (*corev1.Secret, error) {
	return secret.PatchSecret(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeSecretUpdate(ctx context.Context, cli *K8sClients, req *requests.KubeSecretUpdateRequest,
) (*corev1.Secret, error) {
	return secret.UpdateSecretFromJSON(ctx, cli.Kube, req.Namespace, req.Content)
}

func (s *Services) KubeSecretDecode(ctx context.Context, cli *K8sClients, param *requests.KubeSecretDecodeRequest) (map[string]string, error) {
	return secret.DecodeSecretData(ctx, cli.Kube, param)
}

// KubeSecretYaml 获取 Secret 的 YAML 表示
func (s *Services) KubeSecretYaml(ctx context.Context, cli *K8sClients, namespace, name string) (string, error) {
	return secret.GetSecretYaml(ctx, cli.Kube, namespace, name)
}

// KubeSecretApplyYaml 从 YAML 创建或更新 Secret
func (s *Services) KubeSecretApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.Secret, error) {
	return secret.ApplySecretYaml(ctx, cli.Kube, yamlContent)
}

// KubeSecretCreateFromYaml 从 YAML 创建 Secret
func (s *Services) KubeSecretCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.Secret, error) {
	return secret.CreateSecretFromYaml(ctx, cli.Kube, yamlContent)
}
