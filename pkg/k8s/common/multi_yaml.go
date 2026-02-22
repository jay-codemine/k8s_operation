package common

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// ResourceType 资源类型枚举
type ResourceType string

const (
	ResourceTypePVC        ResourceType = "PersistentVolumeClaim"
	ResourceTypeConfigMap  ResourceType = "ConfigMap"
	ResourceTypeSecret     ResourceType = "Secret"
	ResourceTypeService    ResourceType = "Service"
	ResourceTypeDeployment ResourceType = "Deployment"
	ResourceTypeStatefulSet ResourceType = "StatefulSet"
	ResourceTypeDaemonSet  ResourceType = "DaemonSet"
	ResourceTypePod        ResourceType = "Pod"
	ResourceTypeJob        ResourceType = "Job"
	ResourceTypeCronJob    ResourceType = "CronJob"
)

// ParsedResource 解析后的资源对象
type ParsedResource struct {
	Type      ResourceType           // 资源类型
	Name      string                 // 资源名称
	Namespace string                 // 命名空间
	Raw       *unstructured.Unstructured // 原始对象
}

// MultiYAMLParser 多资源 YAML 解析器
type MultiYAMLParser struct {
	yamlContent string
	resources   []*ParsedResource
}

// NewMultiYAMLParser 创建解析器
func NewMultiYAMLParser(yamlContent string) *MultiYAMLParser {
	return &MultiYAMLParser{
		yamlContent: yamlContent,
		resources:   make([]*ParsedResource, 0),
	}
}

// Parse 解析 YAML 内容，支持 --- 分隔的多资源
func (p *MultiYAMLParser) Parse() error {
	// 按 --- 分隔 YAML 文档
	documents := strings.Split(p.yamlContent, "\n---")
	
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "---" {
			continue
		}
		
		// 解析为 Unstructured 对象
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), obj); err != nil {
			return fmt.Errorf("failed to parse YAML document: %w", err)
		}
		
		// 提取资源元数据
		kind := obj.GetKind()
		name := obj.GetName()
		namespace := obj.GetNamespace()
		
		if kind == "" {
			return fmt.Errorf("resource kind is required")
		}
		if name == "" {
			return fmt.Errorf("resource name is required for kind: %s", kind)
		}
		
		// 默认命名空间
		if namespace == "" {
			namespace = "default"
		}
		
		resource := &ParsedResource{
			Type:      ResourceType(kind),
			Name:      name,
			Namespace: namespace,
			Raw:       obj,
		}
		
		p.resources = append(p.resources, resource)
	}
	
	if len(p.resources) == 0 {
		return fmt.Errorf("no valid resources found in YAML")
	}
	
	return nil
}

// GetResources 获取解析后的资源列表
func (p *MultiYAMLParser) GetResources() []*ParsedResource {
	return p.resources
}

// GetResourcesByType 按类型筛选资源
func (p *MultiYAMLParser) GetResourcesByType(resType ResourceType) []*ParsedResource {
	result := make([]*ParsedResource, 0)
	for _, res := range p.resources {
		if res.Type == resType {
			result = append(result, res)
		}
	}
	return result
}

// HasResourceType 检查是否包含指定类型的资源
func (p *MultiYAMLParser) HasResourceType(resType ResourceType) bool {
	for _, res := range p.resources {
		if res.Type == resType {
			return true
		}
	}
	return false
}

// ValidateMainResource 验证主资源（Deployment/StatefulSet/DaemonSet/Pod）是否存在且唯一
func (p *MultiYAMLParser) ValidateMainResource(expectedType ResourceType) (*ParsedResource, error) {
	mainResources := p.GetResourcesByType(expectedType)
	
	if len(mainResources) == 0 {
		return nil, fmt.Errorf("no %s found in YAML", expectedType)
	}
	
	if len(mainResources) > 1 {
		return nil, fmt.Errorf("multiple %s resources found, only one is allowed", expectedType)
	}
	
	return mainResources[0], nil
}

// CreateResourcesInOrder 按依赖顺序创建资源
// 顺序：PVC/ConfigMap/Secret -> Service
// 注意：工作负载（Deployment/StatefulSet/DaemonSet/Pod）由服务层单独创建
func CreateResourcesInOrder(ctx context.Context, cli *kubernetes.Clientset, parser *MultiYAMLParser) (map[ResourceType][]string, error) {
	created := make(map[ResourceType][]string)
	
	// 阶段 1: 创建依赖资源（PVC, ConfigMap, Secret）
	dependencyTypes := []ResourceType{ResourceTypePVC, ResourceTypeConfigMap, ResourceTypeSecret}
	for _, resType := range dependencyTypes {
		resources := parser.GetResourcesByType(resType)
		for _, res := range resources {
			if err := createResource(ctx, cli, res); err != nil {
				return created, fmt.Errorf("failed to create %s %s/%s: %w", resType, res.Namespace, res.Name, err)
			}
			created[resType] = append(created[resType], res.Name)
		}
	}
	
	// 阶段 2: 创建 Service（如果有）
	services := parser.GetResourcesByType(ResourceTypeService)
	for _, svc := range services {
		if err := createResource(ctx, cli, svc); err != nil {
			return created, fmt.Errorf("failed to create Service %s/%s: %w", svc.Namespace, svc.Name, err)
		}
		created[ResourceTypeService] = append(created[ResourceTypeService], svc.Name)
	}
	
	// 注意：工作负载（Deployment/StatefulSet/DaemonSet/Pod/Job/CronJob）由服务层单独创建，不在这里处理
	
	return created, nil
}

// createResource 创建单个资源（统一处理）
func createResource(ctx context.Context, cli *kubernetes.Clientset, res *ParsedResource) error {
	switch res.Type {
	case ResourceTypePVC:
		return createPVC(ctx, cli, res)
	case ResourceTypeConfigMap:
		return createConfigMap(ctx, cli, res)
	case ResourceTypeSecret:
		return createSecret(ctx, cli, res)
	case ResourceTypeService:
		return createService(ctx, cli, res)
	default:
		return fmt.Errorf("resource type %s is not supported for pre-creation", res.Type)
	}
}

// createPVC 创建 PVC
func createPVC(ctx context.Context, cli *kubernetes.Clientset, res *ParsedResource) error {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Raw.Object, pvc); err != nil {
		return fmt.Errorf("failed to convert to PVC: %w", err)
	}
	
	_, err := cli.CoreV1().PersistentVolumeClaims(res.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// createConfigMap 创建 ConfigMap
func createConfigMap(ctx context.Context, cli *kubernetes.Clientset, res *ParsedResource) error {
	cm := &corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Raw.Object, cm); err != nil {
		return fmt.Errorf("failed to convert to ConfigMap: %w", err)
	}
	
	_, err := cli.CoreV1().ConfigMaps(res.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

// createSecret 创建 Secret
func createSecret(ctx context.Context, cli *kubernetes.Clientset, res *ParsedResource) error {
	secret := &corev1.Secret{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Raw.Object, secret); err != nil {
		return fmt.Errorf("failed to convert to Secret: %w", err)
	}
	
	_, err := cli.CoreV1().Secrets(res.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// createService 创建 Service
func createService(ctx context.Context, cli *kubernetes.Clientset, res *ParsedResource) error {
	svc := &corev1.Service{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Raw.Object, svc); err != nil {
		return fmt.Errorf("failed to convert to Service: %w", err)
	}
	
	_, err := cli.CoreV1().Services(res.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}
