package services

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8soperation/global"
	"k8soperation/pkg/k8s/dynamicresource"
)

// ==================== CRD 管理 ====================

// KubeCRDList 列出所有 CRD
func (s *Services) KubeCRDList(ctx context.Context, cli *K8sClients, keyword, group string) ([]dynamicresource.CRDItem, int64, error) {
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)
	items, total, err := crud.ListCRDs(ctx, keyword, group)
	if err != nil {
		global.Logger.Errorf("KubeCRDList error: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

// KubeCRDGet 获取单个 CRD 详情
func (s *Services) KubeCRDGet(ctx context.Context, cli *K8sClients, name string) (*unstructured.Unstructured, error) {
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)
	obj, err := crud.GetCRD(ctx, name)
	if err != nil {
		global.Logger.Errorf("KubeCRDGet error: %v", err)
		return nil, err
	}
	return obj, nil
}

// KubeCRDDelete 删除 CRD（含保护检查）
func (s *Services) KubeCRDDelete(ctx context.Context, cli *K8sClients, name string) error {
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	// 检查删除保护
	obj, err := crud.GetCRD(ctx, name)
	if err != nil {
		return err
	}
	if allowed, reason := dynamicresource.CheckDeleteProtection(obj); !allowed {
		return fmt.Errorf("%s", reason)
	}

	if err := crud.DeleteCRD(ctx, name); err != nil {
		global.Logger.Errorf("KubeCRDDelete error: %v", err)
		return err
	}
	global.Logger.Infof("CRD %s deleted successfully", name)
	return nil
}

// ==================== CR 实例管理 ====================

// KubeCRList 列出 CR 实例
func (s *Services) KubeCRList(ctx context.Context, cli *K8sClients, group, version, resource, namespace string, labelSelector string) (*unstructured.UnstructuredList, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}

	list, err := crud.ListCRs(ctx, gvr, namespace, opts)
	if err != nil {
		global.Logger.Errorf("KubeCRList error: %v", err)
		return nil, err
	}
	return list, nil
}

// KubeCRGet 获取单个 CR 实例
func (s *Services) KubeCRGet(ctx context.Context, cli *K8sClients, group, version, resource, namespace, name string) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	obj, err := crud.GetCR(ctx, gvr, namespace, name)
	if err != nil {
		global.Logger.Errorf("KubeCRGet error: %v", err)
		return nil, err
	}
	return obj, nil
}

// KubeCRCreate 创建 CR 实例
func (s *Services) KubeCRCreate(ctx context.Context, cli *K8sClients, group, version, resource, namespace, yamlContent string, dryRun bool) (*unstructured.Unstructured, *dynamicresource.DryRunResult, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	obj, err := dynamicresource.ParseYAMLToUnstructured(yamlContent)
	if err != nil {
		return nil, nil, err
	}

	// 确保 namespace
	if namespace != "" && obj.GetNamespace() == "" {
		obj.SetNamespace(namespace)
	}

	// DryRun 模式
	if dryRun {
		result := crud.DryRunCreate(ctx, gvr, namespace, obj)
		return nil, result, nil
	}

	created, err := crud.CreateCR(ctx, gvr, namespace, obj, false)
	if err != nil {
		global.Logger.Errorf("KubeCRCreate error: %v", err)
		return nil, nil, err
	}
	global.Logger.Infof("CR %s/%s created successfully", gvr.Resource, created.GetName())
	return created, nil, nil
}

// KubeCRUpdate 更新 CR 实例
func (s *Services) KubeCRUpdate(ctx context.Context, cli *K8sClients, group, version, resource, namespace, name, yamlContent string, dryRun bool) (*unstructured.Unstructured, *dynamicresource.DryRunResult, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	obj, err := dynamicresource.ParseYAMLToUnstructured(yamlContent)
	if err != nil {
		return nil, nil, err
	}

	// 获取现有资源的 resourceVersion（乐观锁）
	existing, err := crud.GetCR(ctx, gvr, namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("get existing resource failed: %w", err)
	}
	obj.SetResourceVersion(existing.GetResourceVersion())

	if namespace != "" && obj.GetNamespace() == "" {
		obj.SetNamespace(namespace)
	}

	// DryRun 模式
	if dryRun {
		result := crud.DryRunUpdate(ctx, gvr, namespace, obj)
		return nil, result, nil
	}

	updated, err := crud.UpdateCR(ctx, gvr, namespace, obj, false)
	if err != nil {
		global.Logger.Errorf("KubeCRUpdate error: %v", err)
		return nil, nil, err
	}
	global.Logger.Infof("CR %s/%s updated successfully", gvr.Resource, updated.GetName())
	return updated, nil, nil
}

// KubeCRDelete 删除 CR 实例（含保护检查）
func (s *Services) KubeCRDelete(ctx context.Context, cli *K8sClients, group, version, resource, namespace, name string) error {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	// 检查删除保护
	obj, err := crud.GetCR(ctx, gvr, namespace, name)
	if err != nil {
		return err
	}
	if allowed, reason := dynamicresource.CheckDeleteProtection(obj); !allowed {
		return fmt.Errorf("%s", reason)
	}

	if err := crud.DeleteCR(ctx, gvr, namespace, name); err != nil {
		global.Logger.Errorf("KubeCRDelete error: %v", err)
		return err
	}
	global.Logger.Infof("CR %s/%s/%s deleted successfully", gvr.Resource, namespace, name)
	return nil
}

// KubeCRGetYaml 获取 CR 的 YAML 表示
func (s *Services) KubeCRGetYaml(ctx context.Context, cli *K8sClients, group, version, resource, namespace, name string) (string, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	crud := dynamicresource.NewDynamicCRUD(cli.Dynamic, cli.Kube)

	obj, err := crud.GetCR(ctx, gvr, namespace, name)
	if err != nil {
		return "", err
	}

	yamlStr, err := dynamicresource.UnstructuredToYAML(obj)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}
