// Package vpa - YAML 创建支持
package vpa

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// CreateFromYaml 通过原始 YAML 创建 VPA 资源（autoscaling.k8s.io/v1 VerticalPodAutoscaler）
func CreateFromYaml(ctx context.Context, dyn dynamic.Interface, yamlContent string) (*unstructured.Unstructured, error) {
	yamlContent = strings.TrimSpace(yamlContent)
	if yamlContent == "" {
		return nil, fmt.Errorf("yaml 内容不能为空")
	}

	// 1. 解析 YAML 到 map[string]interface{}
	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("yaml 内容为空")
	}
	obj := &unstructured.Unstructured{Object: raw}

	// 2. 校验 Kind / APIVersion
	kind := obj.GetKind()
	apiVer := obj.GetAPIVersion()
	if kind != "" && kind != "VerticalPodAutoscaler" {
		return nil, fmt.Errorf("不支持的资源类型 %q，仅支持 VerticalPodAutoscaler", kind)
	}
	if apiVer != "" && !strings.HasPrefix(apiVer, "autoscaling.k8s.io/") {
		return nil, fmt.Errorf("不支持的 apiVersion %q，请使用 autoscaling.k8s.io/v1", apiVer)
	}
	// 补全默认 GVK，避免未填写时被拒绝
	if kind == "" {
		obj.SetKind("VerticalPodAutoscaler")
	}
	if apiVer == "" {
		obj.SetAPIVersion("autoscaling.k8s.io/v1")
	}

	// 3. 校验必填字段
	if obj.GetName() == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	if obj.GetNamespace() == "" {
		obj.SetNamespace("default")
	}
	targetRef, found, err := unstructured.NestedMap(obj.Object, "spec", "targetRef")
	if err != nil || !found {
		return nil, fmt.Errorf("spec.targetRef 不能为空")
	}
	if s, _ := targetRef["kind"].(string); s == "" {
		return nil, fmt.Errorf("spec.targetRef.kind 不能为空")
	}
	if s, _ := targetRef["name"].(string); s == "" {
		return nil, fmt.Errorf("spec.targetRef.name 不能为空")
	}

	// 4. 清理只读字段
	obj.SetResourceVersion("")
	obj.SetUID("")
	obj.SetSelfLink("")
	obj.SetCreationTimestamp(metav1.Time{})
	obj.SetGeneration(0)
	obj.SetManagedFields(nil)
	unstructured.RemoveNestedField(obj.Object, "status")

	// 5. 通过 dynamic client 创建
	created, err := dyn.Resource(VPAGVR).Namespace(obj.GetNamespace()).
		Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("VPA %q 在命名空间 %q 中已存在", obj.GetName(), obj.GetNamespace())
		}
		return nil, err
	}
	return created, nil
}
