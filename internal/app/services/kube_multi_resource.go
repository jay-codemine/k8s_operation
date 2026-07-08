package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/configmap"
	"k8soperation/pkg/k8s/deployment"
	"k8soperation/pkg/k8s/secret"
	svc "k8soperation/pkg/k8s/svc"
	"sigs.k8s.io/yaml"
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
		return secret.ApplySecretYaml(ctx, cli.Kube, resource.Content)
	case "Service":
		return svc.ApplyServiceYaml(ctx, cli.Kube, resource.Content)
	case "Deployment":
		return deployment.ApplyYaml(ctx, cli.Kube, resource.Namespace, resource.Name, resource.Content)
	case "StatefulSet":
		return s.applyStatefulSetYaml(ctx, cli, resource)
	case "DaemonSet":
		return s.applyDaemonSetYaml(ctx, cli, resource)
	case "Job":
		return s.applyJobYaml(ctx, cli, resource)
	case "CronJob":
		return s.applyCronJobYaml(ctx, cli, resource)
	case "PersistentVolumeClaim":
		return s.applyPVCYaml(ctx, cli, resource)
	default:
		return nil, errors.Errorf("暂不支持资源类型: %s", resource.Kind)
	}
}

// applyStatefulSetYaml 从 YAML 创建/更新 StatefulSet
func (s *Services) applyStatefulSetYaml(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	var sts interface{}
	if err := yaml.Unmarshal([]byte(resource.Content), &sts); err != nil {
		return nil, errors.Wrap(err, "解析 StatefulSet YAML 失败")
	}

	// 直接使用 dynamic 风格：先 Get，存在则 Update，不存在则 Create
	existing, err := cli.Kube.AppsV1().StatefulSets(resource.Namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	if err != nil {
		// 不存在，创建
		var obj appv1.StatefulSet
		if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
			return nil, errors.Wrap(err, "解析 StatefulSet YAML 失败")
		}
		if obj.Namespace == "" {
			obj.Namespace = resource.Namespace
		}
		created, createErr := cli.Kube.AppsV1().StatefulSets(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
		if createErr != nil {
			return nil, errors.Wrap(createErr, "创建 StatefulSet 失败")
		}
		global.Logger.Infof("[MultiResource] StatefulSet %s/%s 创建成功", created.Namespace, created.Name)
		return created, nil
	}

	// 存在，更新
	var obj appv1.StatefulSet
	if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
		return nil, errors.Wrap(err, "解析 StatefulSet YAML 失败")
	}
	obj.ResourceVersion = existing.ResourceVersion
	obj.Spec.Selector = existing.Spec.Selector // selector 是 immutable
	updated, err := cli.Kube.AppsV1().StatefulSets(resource.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "更新 StatefulSet 失败")
	}
	global.Logger.Infof("[MultiResource] StatefulSet %s/%s 更新成功", updated.Namespace, updated.Name)
	return updated, nil
}

// applyDaemonSetYaml 从 YAML 创建/更新 DaemonSet
func (s *Services) applyDaemonSetYaml(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	existing, err := cli.Kube.AppsV1().DaemonSets(resource.Namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	if err != nil {
		var obj appv1.DaemonSet
		if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
			return nil, errors.Wrap(err, "解析 DaemonSet YAML 失败")
		}
		if obj.Namespace == "" {
			obj.Namespace = resource.Namespace
		}
		created, createErr := cli.Kube.AppsV1().DaemonSets(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
		if createErr != nil {
			return nil, errors.Wrap(createErr, "创建 DaemonSet 失败")
		}
		global.Logger.Infof("[MultiResource] DaemonSet %s/%s 创建成功", created.Namespace, created.Name)
		return created, nil
	}

	var obj appv1.DaemonSet
	if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
		return nil, errors.Wrap(err, "解析 DaemonSet YAML 失败")
	}
	obj.ResourceVersion = existing.ResourceVersion
	obj.Spec.Selector = existing.Spec.Selector
	updated, err := cli.Kube.AppsV1().DaemonSets(resource.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "更新 DaemonSet 失败")
	}
	global.Logger.Infof("[MultiResource] DaemonSet %s/%s 更新成功", updated.Namespace, updated.Name)
	return updated, nil
}

// applyJobYaml 从 YAML 创建 Job
func (s *Services) applyJobYaml(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	var obj batchv1.Job
	if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
		return nil, errors.Wrap(err, "解析 Job YAML 失败")
	}
	if obj.Namespace == "" {
		obj.Namespace = resource.Namespace
	}
	created, err := cli.Kube.BatchV1().Jobs(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "创建 Job 失败")
	}
	global.Logger.Infof("[MultiResource] Job %s/%s 创建成功", created.Namespace, created.Name)
	return created, nil
}

// applyCronJobYaml 从 YAML 创建 CronJob
func (s *Services) applyCronJobYaml(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	var obj batchv1.CronJob
	if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
		return nil, errors.Wrap(err, "解析 CronJob YAML 失败")
	}
	if obj.Namespace == "" {
		obj.Namespace = resource.Namespace
	}
	created, err := cli.Kube.BatchV1().CronJobs(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "创建 CronJob 失败")
	}
	global.Logger.Infof("[MultiResource] CronJob %s/%s 创建成功", created.Namespace, created.Name)
	return created, nil
}

// applyPVCYaml 从 YAML 创建 PVC
func (s *Services) applyPVCYaml(ctx context.Context, cli *K8sClients, resource requests.ParsedResource) (interface{}, error) {
	var obj corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(resource.Content), &obj); err != nil {
		return nil, errors.Wrap(err, "解析 PVC YAML 失败")
	}
	if obj.Namespace == "" {
		obj.Namespace = resource.Namespace
	}
	created, err := cli.Kube.CoreV1().PersistentVolumeClaims(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "创建 PVC 失败")
	}
	global.Logger.Infof("[MultiResource] PVC %s/%s 创建成功", created.Namespace, created.Name)
	return created, nil
}