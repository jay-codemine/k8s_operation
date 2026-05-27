package dynamicresource

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// DynamicCRUD 通用 CRD/CR 动态资源 CRUD 引擎
type DynamicCRUD struct {
	client    dynamic.Interface
	kube      kubernetes.Interface
}

// NewDynamicCRUD 创建 CRUD 引擎
func NewDynamicCRUD(dynClient dynamic.Interface, kube kubernetes.Interface) *DynamicCRUD {
	return &DynamicCRUD{client: dynClient, kube: kube}
}

// ==================== CRD 管理 ====================

// CRDItem CRD 列表项
type CRDItem struct {
	Name        string   `json:"name"`
	Group       string   `json:"group"`
	Version     string   `json:"version"`
	Kind        string   `json:"kind"`
	Resource    string   `json:"resource"` // 复数名（如 prometheusrules）
	Scope       string   `json:"scope"`
	Versions    []string `json:"versions"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"created_at"`
}

// ListCRDs 列出所有 CRD
func (d *DynamicCRUD) ListCRDs(ctx context.Context, keyword, group string) ([]CRDItem, int64, error) {
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	list, err := d.client.Resource(crdGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("list CRDs failed: %w", err)
	}

	var items []CRDItem
	for _, item := range list.Items {
		crdItem := parseCRDItem(&item)
		// 关键词过滤
		if keyword != "" && !strings.Contains(strings.ToLower(crdItem.Name), strings.ToLower(keyword)) &&
			!strings.Contains(strings.ToLower(crdItem.Kind), strings.ToLower(keyword)) {
			continue
		}
		// Group 过滤
		if group != "" && crdItem.Group != group {
			continue
		}
		items = append(items, crdItem)
	}

	return items, int64(len(items)), nil
}

// GetCRD 获取单个 CRD 详情
func (d *DynamicCRUD) GetCRD(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	return d.client.Resource(crdGVR).Get(ctx, name, metav1.GetOptions{})
}

// DeleteCRD 删除 CRD
func (d *DynamicCRUD) DeleteCRD(ctx context.Context, name string) error {
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	return d.client.Resource(crdGVR).Delete(ctx, name, metav1.DeleteOptions{})
}

// ==================== CR 实例管理 ====================

// ListCRs 列出自定义资源实例
func (d *DynamicCRUD) ListCRs(ctx context.Context, gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	if namespace == "" {
		return d.client.Resource(gvr).List(ctx, opts)
	}
	return d.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

// GetCR 获取单个 CR
func (d *DynamicCRUD) GetCR(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return d.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	return d.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateCR 创建 CR（支持 DryRun）
func (d *DynamicCRUD) CreateCR(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, dryRun bool) (*unstructured.Unstructured, error) {
	opts := metav1.CreateOptions{}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}
	if namespace == "" {
		return d.client.Resource(gvr).Create(ctx, obj, opts)
	}
	return d.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, opts)
}

// UpdateCR 更新 CR（支持 DryRun）
func (d *DynamicCRUD) UpdateCR(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, dryRun bool) (*unstructured.Unstructured, error) {
	opts := metav1.UpdateOptions{}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}
	if namespace == "" {
		return d.client.Resource(gvr).Update(ctx, obj, opts)
	}
	return d.client.Resource(gvr).Namespace(namespace).Update(ctx, obj, opts)
}

// DeleteCR 删除 CR
func (d *DynamicCRUD) DeleteCR(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	if namespace == "" {
		return d.client.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return d.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// PatchCR Patch CR (用于局部更新)
func (d *DynamicCRUD) PatchCR(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, patchData []byte) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return d.client.Resource(gvr).Patch(ctx, name, types.MergePatchType, patchData, metav1.PatchOptions{})
	}
	return d.client.Resource(gvr).Namespace(namespace).Patch(ctx, name, types.MergePatchType, patchData, metav1.PatchOptions{})
}

// ==================== YAML 相关 ====================

// ParseYAMLToUnstructured 将 YAML 解析为 Unstructured 对象
func ParseYAMLToUnstructured(yamlContent string) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlContent), &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if obj.GetKind() == "" {
		return nil, fmt.Errorf("YAML missing required field: kind")
	}
	if obj.GetName() == "" {
		return nil, fmt.Errorf("YAML missing required field: metadata.name")
	}
	return obj, nil
}

// UnstructuredToYAML 将 Unstructured 对象转为 YAML（剥离系统字段）
func UnstructuredToYAML(obj *unstructured.Unstructured) (string, error) {
	// 深拷贝，避免修改原对象
	clean := obj.DeepCopy()

	// 剥离系统管理字段
	metadata, ok := clean.Object["metadata"].(map[string]interface{})
	if ok {
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")
		delete(metadata, "generation")
		delete(metadata, "creationTimestamp")
		// 清理空 annotations
		if ann, ok := metadata["annotations"].(map[string]interface{}); ok {
			delete(ann, "kubectl.kubernetes.io/last-applied-configuration")
			if len(ann) == 0 {
				delete(metadata, "annotations")
			}
		}
	}
	delete(clean.Object, "status")

	data, err := yaml.Marshal(clean.Object)
	if err != nil {
		return "", fmt.Errorf("marshal to YAML failed: %w", err)
	}
	return string(data), nil
}

// ==================== DryRun ====================

// DryRunResult 校验结果
type DryRunResult struct {
	Valid    bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

// DryRunCreate 创建资源预校验
func (d *DynamicCRUD) DryRunCreate(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) *DryRunResult {
	_, err := d.CreateCR(ctx, gvr, namespace, obj, true)
	if err != nil {
		return &DryRunResult{
			Valid:   false,
			Errors:  []string{err.Error()},
			Message: "DryRun 校验失败",
		}
	}
	return &DryRunResult{Valid: true, Message: "DryRun 校验通过，可安全提交"}
}

// DryRunUpdate 更新资源预校验
func (d *DynamicCRUD) DryRunUpdate(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) *DryRunResult {
	_, err := d.UpdateCR(ctx, gvr, namespace, obj, true)
	if err != nil {
		return &DryRunResult{
			Valid:   false,
			Errors:  []string{err.Error()},
			Message: "DryRun 校验失败",
		}
	}
	return &DryRunResult{Valid: true, Message: "DryRun 校验通过，可安全提交"}
}

// ==================== 删除保护 ====================

const (
	ProtectionAnnotationKey   = "k8soperation.io/delete-protection"
	ProtectionAnnotationValue = "enabled"
)

// CheckDeleteProtection 检查删除保护
func CheckDeleteProtection(obj *unstructured.Unstructured) (bool, string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return true, "" // 允许删除
	}
	if annotations[ProtectionAnnotationKey] == ProtectionAnnotationValue {
		return false, fmt.Sprintf("资源受删除保护（annotation: %s=%s），请先移除保护标记", ProtectionAnnotationKey, ProtectionAnnotationValue)
	}
	return true, ""
}

// ==================== 辅助函数 ====================

// ParseGVR 解析 group/version/resource 字符串
func ParseGVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}

// parseCRDItem 从 Unstructured 解析 CRD 信息
func parseCRDItem(item *unstructured.Unstructured) CRDItem {
	spec, _ := item.Object["spec"].(map[string]interface{})
	status, _ := item.Object["status"].(map[string]interface{})

	crdItem := CRDItem{
		Name:      item.GetName(),
		CreatedAt: item.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
	}

	if spec != nil {
		crdItem.Group, _ = spec["group"].(string)
		if scope, ok := spec["scope"].(string); ok {
			crdItem.Scope = scope
		}
		// names
		if names, ok := spec["names"].(map[string]interface{}); ok {
			crdItem.Kind, _ = names["kind"].(string)
			crdItem.Resource, _ = names["plural"].(string)
		}
		// versions
		if versions, ok := spec["versions"].([]interface{}); ok {
			for _, v := range versions {
				if vMap, ok := v.(map[string]interface{}); ok {
					if name, ok := vMap["name"].(string); ok {
						crdItem.Versions = append(crdItem.Versions, name)
						if crdItem.Version == "" {
							// 取 served 的第一个版本
							if served, _ := vMap["served"].(bool); served {
								crdItem.Version = name
							}
						}
					}
				}
			}
			if crdItem.Version == "" && len(crdItem.Versions) > 0 {
				crdItem.Version = crdItem.Versions[0]
			}
		}
	}

	// status conditions
	if status != nil {
		if conditions, ok := status["conditions"].([]interface{}); ok {
			for _, c := range conditions {
				if cMap, ok := c.(map[string]interface{}); ok {
					if cType, _ := cMap["type"].(string); cType == "Established" {
						if cStatus, _ := cMap["status"].(string); cStatus == "True" {
							crdItem.Status = "Established"
						}
					}
				}
			}
		}
		if crdItem.Status == "" {
			crdItem.Status = "Unknown"
		}
	}

	return crdItem
}
