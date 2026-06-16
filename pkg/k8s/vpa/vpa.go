// Package vpa 通过 DynamicClient 操作 VerticalPodAutoscaler 资源
// VPA 不是 K8s 内置资源，依赖 vertical-pod-autoscaler 控制器（autoscaling.k8s.io/v1）
package vpa

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// VPA 资源 GVR
var VPAGVR = schema.GroupVersionResource{
	Group:    "autoscaling.k8s.io",
	Version:  "v1",
	Resource: "verticalpodautoscalers",
}

// UpdateMode VPA 更新模式
const (
	UpdateModeOff      = "Off"      // 仅推荐，不自动更新
	UpdateModeInitial  = "Initial"  // 仅 Pod 创建时应用
	UpdateModeRecreate = "Recreate" // 重建 Pod 应用推荐
	UpdateModeAuto     = "Auto"     // 当前等同 Recreate
)

// CreateOptions 创建 VPA 选项
type CreateOptions struct {
	Namespace      string
	Name           string
	TargetKind     string // Deployment / StatefulSet
	TargetName     string
	TargetAPIVer   string // 默认 apps/v1
	UpdateMode     string // Off/Initial/Recreate/Auto
	MinAllowedCPU  string // 例如 "100m"
	MinAllowedMem  string // 例如 "128Mi"
	MaxAllowedCPU  string // 例如 "2"
	MaxAllowedMem  string // 例如 "4Gi"
	ControlledRes  []string // ["cpu","memory"] 默认两者
	ContainerName  string   // 不填则匹配所有容器（"*"）
	Labels         map[string]string
}

// IsAvailable 通过 GET 一次资源类型存在性，判断 VPA Operator 是否安装
func IsAvailable(ctx context.Context, dyn dynamic.Interface) bool {
	_, err := dyn.Resource(VPAGVR).Namespace("").List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		return true
	}
	// IsNotFound 表示集群没有这个 CRD
	return !apierrors.IsNotFound(err) && !isNoMatchError(err)
}

func isNoMatchError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "no matches for kind") || contains(msg, "could not find the requested resource")
}

// List 获取 VPA 列表
func List(ctx context.Context, dyn dynamic.Interface, namespace, nameKeyword string) ([]unstructured.Unstructured, error) {
	res, err := dyn.Resource(VPAGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if nameKeyword == "" {
		return res.Items, nil
	}
	filtered := make([]unstructured.Unstructured, 0, len(res.Items))
	for _, item := range res.Items {
		if containsIgnoreCase(item.GetName(), nameKeyword) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// Get 获取 VPA 详情
func Get(ctx context.Context, dyn dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	obj, err := dyn.Resource(VPAGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("vpa %s/%s not found", namespace, name)
		}
		return nil, err
	}
	return obj, nil
}

// Create 创建 VPA
func Create(ctx context.Context, dyn dynamic.Interface, opts *CreateOptions) (*unstructured.Unstructured, error) {
	if err := validate(opts); err != nil {
		return nil, err
	}
	obj := buildVPA(opts)
	created, err := dyn.Resource(VPAGVR).Namespace(opts.Namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Update 更新 VPA
func Update(ctx context.Context, dyn dynamic.Interface, opts *CreateOptions) (*unstructured.Unstructured, error) {
	if err := validate(opts); err != nil {
		return nil, err
	}
	existing, err := Get(ctx, dyn, opts.Namespace, opts.Name)
	if err != nil {
		return nil, err
	}
	updated := buildVPA(opts)
	updated.SetResourceVersion(existing.GetResourceVersion())
	res, err := dyn.Resource(VPAGVR).Namespace(opts.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Delete 删除 VPA
func Delete(ctx context.Context, dyn dynamic.Interface, namespace, name string) error {
	err := dyn.Resource(VPAGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// buildVPA 构建 VPA Unstructured 对象
func buildVPA(opts *CreateOptions) *unstructured.Unstructured {
	apiVer := opts.TargetAPIVer
	if apiVer == "" {
		apiVer = "apps/v1"
	}
	updateMode := opts.UpdateMode
	if updateMode == "" {
		updateMode = UpdateModeAuto
	}

	containerName := opts.ContainerName
	if containerName == "" {
		containerName = "*"
	}

	controlled := opts.ControlledRes
	if len(controlled) == 0 {
		controlled = []string{"cpu", "memory"}
	}

	minAllowed := map[string]interface{}{}
	if opts.MinAllowedCPU != "" {
		minAllowed["cpu"] = opts.MinAllowedCPU
	}
	if opts.MinAllowedMem != "" {
		minAllowed["memory"] = opts.MinAllowedMem
	}
	maxAllowed := map[string]interface{}{}
	if opts.MaxAllowedCPU != "" {
		maxAllowed["cpu"] = opts.MaxAllowedCPU
	}
	if opts.MaxAllowedMem != "" {
		maxAllowed["memory"] = opts.MaxAllowedMem
	}

	containerPolicy := map[string]interface{}{
		"containerName":      containerName,
		"controlledResources": toIfaceList(controlled),
	}
	if len(minAllowed) > 0 {
		containerPolicy["minAllowed"] = minAllowed
	}
	if len(maxAllowed) > 0 {
		containerPolicy["maxAllowed"] = maxAllowed
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "autoscaling.k8s.io/v1",
			"kind":       "VerticalPodAutoscaler",
			"metadata": map[string]interface{}{
				"name":      opts.Name,
				"namespace": opts.Namespace,
			},
			"spec": map[string]interface{}{
				"targetRef": map[string]interface{}{
					"apiVersion": apiVer,
					"kind":       opts.TargetKind,
					"name":       opts.TargetName,
				},
				"updatePolicy": map[string]interface{}{
					"updateMode": updateMode,
				},
				"resourcePolicy": map[string]interface{}{
					"containerPolicies": []interface{}{containerPolicy},
				},
			},
		},
	}
	if len(opts.Labels) > 0 {
		obj.SetLabels(opts.Labels)
	}
	return obj
}

func toIfaceList(s []string) []interface{} {
	out := make([]interface{}, 0, len(s))
	for _, v := range s {
		out = append(out, v)
	}
	return out
}

func validate(opts *CreateOptions) error {
	if opts == nil {
		return fmt.Errorf("opts is nil")
	}
	if opts.Namespace == "" || opts.Name == "" {
		return fmt.Errorf("namespace and name are required")
	}
	if opts.TargetKind == "" || opts.TargetName == "" {
		return fmt.Errorf("target kind and name are required")
	}
	switch opts.UpdateMode {
	case "", UpdateModeOff, UpdateModeInitial, UpdateModeRecreate, UpdateModeAuto:
	default:
		return fmt.Errorf("invalid update_mode: %s", opts.UpdateMode)
	}
	return nil
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	ss := []byte(s)
	for i := range ss {
		if ss[i] >= 'A' && ss[i] <= 'Z' {
			ss[i] += 32
		}
	}
	target := []byte(sub)
	for i := range target {
		if target[i] >= 'A' && target[i] <= 'Z' {
			target[i] += 32
		}
	}
	return contains(string(ss), string(target))
}
