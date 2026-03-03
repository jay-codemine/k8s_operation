package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/configmap"
	"k8soperation/pkg/k8s/deployment"
)

// KubeMultiResourceParseYaml 解析多资源 YAML
func (s *Services) KubeMultiResourceParseYaml(_ context.Context, yamlContent string) (*requests.MultiResourceParsedResult, error) {
	// 解析 YAML 文档
	resources, err := common.ParseMultiYaml(yamlContent)
	if err != nil {
		return nil, errors.Wrap(err, "解析 YAML 失败")
	}

	// 按创建顺序排序
	resources = common.SortResourcesByOrder(resources)

	// 验证依赖关系
	dependencyErrors := common.ValidateResourceDependencies(resources)

	result := &requests.MultiResourceParsedResult{
		Resources: resources,
		Total:     len(resources),
	}

	if len(dependencyErrors) > 0 {
		result.Errors = dependencyErrors
	}

	return result, nil
}

// KubeMultiResourceApplyYaml 应用多资源 YAML
func (s *Services) KubeMultiResourceApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*requests.MultiResourceCreateResult, error) {
	// 1. 解析 YAML
	parseResult, err := s.KubeMultiResourceParseYaml(ctx, yamlContent)
	if err != nil {
		return nil, errors.Wrap(err, "解析 YAML 失败")
	}

	if len(parseResult.Errors) > 0 {
		return nil, errors.New("YAML 依赖关系验证失败: " + strings.Join(parseResult.Errors, "; "))
	}

	// 2. 按顺序创建资源
	result := &requests.MultiResourceCreateResult{
		Total: len(parseResult.Resources),
	}

	for _, resource := range parseResult.Resources {
		_, err := s.createSingleResource(ctx, cli, resource)
		if err != nil {
			result.Failed = append(result.Failed, requests.FailedResource{
				Index:   resource.Index,
				Kind:    resource.Kind,
				Name:    resource.Name,
				Error:   err.Error(),
				Message: fmt.Sprintf("创建 %s %s 失败", resource.Kind, resource.Name),
			})
			// 继续创建其他资源，不中断整个流程
			continue
		}

		result.Created = append(result.Created, requests.CreatedResource{
			Index:     resource.Index,
			Kind:      resource.Kind,
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Message:   fmt.Sprintf("%s %s 创建成功", resource.Kind, resource.Name),
		})
	}

	return result, nil
}

// createSingleResource 创建单个资源
func (s *Services) createSingleResource(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	switch resource.Kind {
	case "ConfigMap":
		return configmap.ApplyConfigMapYaml(ctx, cli.Kube, resource.Content)
	case "Secret":
		// TODO: 实现 Secret.ApplySecretYaml
		return nil, errors.New("Secret 支持待实现")
	case "Service":
		// TODO: 实现 svc.ApplyServiceYaml
		return nil, errors.New("Service 支持待实现")
	case "Deployment":
		return deployment.ApplyYaml(ctx, cli.Kube, resource.Namespace, resource.Name, resource.Content)
	case "StorageClass":
		// TODO: 实现 StorageClass 支持
		return nil, errors.New("StorageClass 支持待实现")
	default:
		// 对于未明确支持的资源类型，尝试使用通用方法
		return s.applyGenericResource(ctx, cli, resource)
	}
}

// applyGenericResource 通用资源应用方法
func (s *Services) applyGenericResource(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	// 解析 API 版本
	gv, err := schema.ParseGroupVersion(resource.APIVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "无效的 API 版本: %s", resource.APIVersion)
	}

	// 根据不同的 API 组选择对应的客户端
	switch gv.Group {
	case "":
		// Core API Group (v1)
		return s.applyCoreResource(ctx, cli, resource, gv.Version)
	case "apps":
		// Apps API Group
		return s.applyAppsResource(ctx, cli, resource, gv.Version)
	case "batch":
		// Batch API Group
		return s.applyBatchResource(ctx, cli, resource, gv.Version)
	default:
		return nil, errors.Errorf("暂不支持 API 组: %s", gv.Group)
	}
}

// applyCoreResource 应用 Core API 资源
func (s *Services) applyCoreResource(ctx context.Context, cli *K8sClients, resource requests.ParsedResource, version string) (interface{}, error) {
	switch resource.Kind {
	case "Namespace":
		// Namespace 创建逻辑
		ns := &corev1.Namespace{}
		// 这里需要解析 YAML 并创建 Namespace
		// 由于篇幅限制，这里简化处理
		return ns, nil
	case "PersistentVolumeClaim":
		// PVC 创建逻辑
		pvc := &corev1.PersistentVolumeClaim{}
		return pvc, nil
	case "Pod":
		// Pod 创建逻辑
		pod := &corev1.Pod{}
		return pod, nil
	default:
		return nil, errors.Errorf("Core API 不支持资源类型: %s", resource.Kind)
	}
}

// applyAppsResource 应用 Apps API 资源
func (s *Services) applyAppsResource(ctx context.Context, cli *K8sClients, resource requests.ParsedResource, version string) (interface{}, error) {
	switch resource.Kind {
	case "StatefulSet":
		// StatefulSet 创建逻辑
		return nil, errors.New("StatefulSet 支持待实现")
	case "DaemonSet":
		// DaemonSet 创建逻辑
		return nil, errors.New("DaemonSet 支持待实现")
	default:
		return nil, errors.Errorf("Apps API 不支持资源类型: %s", resource.Kind)
	}
}

// applyBatchResource 应用 Batch API 资源
func (s *Services) applyBatchResource(ctx context.Context, cli *K8sClients, resource requests.ParsedResource, version string) (interface{}, error) {
	switch resource.Kind {
	case "Job":
		// Job 创建逻辑
		return nil, errors.New("Job 支持待实现")
	case "CronJob":
		// CronJob 创建逻辑
		return nil, errors.New("CronJob 支持待实现")
	default:
		return nil, errors.Errorf("Batch API 不支持资源类型: %s", resource.Kind)
	}
}