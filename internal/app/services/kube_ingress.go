package services

import (
	"context"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/ingress"
)

// KubeIngressCreate 创建 Ingress
func (s *Services) KubeIngressCreate(ctx context.Context, cli *K8sClients, req *requests.KubeIngressCreateRequest) (*networkingv1.Ingress, error) {
	ing, err := ingress.CreateIngress(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("ingress %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("ingress %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create ingress failed: %w", err)
	}

	global.Logger.Infof("ingress %s/%s created successfully", ing.Namespace, ing.Name)
	return ing, nil
}

func (s *Services) KubeIngressList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubeIngressListRequest,
) ([]networkingv1.Ingress, int64, error) {

	ingresses, total, err := ingress.GetIngressList(
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

	return ingresses, total, nil
}

func (s *Services) KubeIngressDetail(ctx context.Context, cli *K8sClients, param *requests.KubeIngressDetailRequest) (*networkingv1.Ingress, error) {
	return ingress.GetIngressDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeIngressPatch(ctx context.Context, cli *K8sClients, param *requests.KubeIngressUpdateRequest) (*networkingv1.Ingress, error) {
	return ingress.PatchIngress(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// 覆盖更新
func (s *Services) KubeIngressPatchJSON(ctx context.Context, cli *K8sClients, req *requests.KubeIngressUpdateRequest,
) (*networkingv1.Ingress, error) {
	return ingress.UpdateIngressFromJSON(ctx, cli.Kube, req.Namespace, req.Content)
}

// 删除 Ingress
func (s *Services) KubeIngressDelete(ctx context.Context, cli *K8sClients, param *requests.KubeIngressDeleteRequest) error {
	return ingress.DeleteIngress(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeIngressYaml 获取 Ingress 的 YAML
func (s *Services) KubeIngressYaml(ctx context.Context, cli *K8sClients, namespace, name string) (string, error) {
	return ingress.GetIngressYaml(ctx, cli.Kube, namespace, name)
}

// KubeIngressApplyYaml 从 YAML 创建或更新 Ingress
func (s *Services) KubeIngressApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*networkingv1.Ingress, error) {
	return ingress.ApplyIngressYaml(ctx, cli.Kube, yamlContent)
}

// KubeIngressCreateFromYaml 从 YAML 创建 Ingress（支持多文档）
func (s *Services) KubeIngressCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*networkingv1.Ingress, error) {
	return ingress.CreateIngressFromYaml(ctx, cli.Kube, yamlContent)
}
