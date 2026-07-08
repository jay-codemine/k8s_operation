package services

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/svc"
)

func (s *Services) KubeCreateService(ctx context.Context, cli *K8sClients, req *requests.KubeServiceCreateRequest) (*corev1.Service, error) {
	return svc.CreateService(ctx, cli.Kube, req)
}

func (s *Services) KubeServiceList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubeServiceListRequest,
) ([]corev1.Service, int64, error) {

	list, total, err := svc.GetServiceList(
		ctx,
		cli.Kube,
		param.Name,
		param.Namespace,
		param.Page,
		param.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Services) KubeServiceDetail(ctx context.Context, cli *K8sClients, param *requests.KubeServiceDetailRequest) (*corev1.Service, error) {
	return svc.GetServiceDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

func (s *Services) KubeServiceDelete(ctx context.Context, cli *K8sClients, param *requests.KubeServiceDeleteRequest) error {
	return svc.DeleteService(ctx, cli.Kube, param.Name, param.Namespace)
}

func (s *Services) KubeUpdateServiceTemplate(ctx context.Context, cli *K8sClients, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	return svc.PatchService(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeServicePatch(ctx context.Context, cli *K8sClients, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	return svc.PatchService(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeServicePatchJSON(ctx context.Context, cli *K8sClients, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	// 建议在这里做 JSON 合法性校验，返回更友好的错误
	if !json.Valid([]byte(param.Content)) {
		return nil, fmt.Errorf("invalid json")
	}
	return svc.PatchJsonService(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// 获取 Endpoints 列表
func (s *Services) KubeServiceEndpoints(ctx context.Context, cli *K8sClients, param *requests.KubeServiceEndpointsRequest) (*corev1.Endpoints, error) {
	return svc.GetServiceEndpoints(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeServiceYaml 获取 Service 的 YAML 表示
func (s *Services) KubeServiceYaml(ctx context.Context, cli *K8sClients, namespace, name string) (string, error) {
	return svc.GetServiceYaml(ctx, cli.Kube, namespace, name)
}

// KubeServiceApplyYaml 从 YAML 创建或更新 Service
func (s *Services) KubeServiceApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.Service, error) {
	return svc.ApplyServiceYaml(ctx, cli.Kube, yamlContent)
}

// KubeServiceCreateFromYaml 从 YAML 创建 Service
func (s *Services) KubeServiceCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.Service, error) {
	return svc.CreateServiceFromYaml(ctx, cli.Kube, yamlContent)
}
