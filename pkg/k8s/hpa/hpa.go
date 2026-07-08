// Package hpa 封装 K8s HorizontalPodAutoscaler (HPA) 资源操作
// 使用 autoscaling/v2 API（K8s 1.23+ 稳定版本，支持多指标/Behavior 配置）
package hpa

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// CreateOptions 创建 HPA 选项（基于 CPU/内存利用率）
type CreateOptions struct {
	Namespace      string
	Name           string
	TargetKind     string // Deployment / StatefulSet
	TargetName     string
	TargetAPIVer   string // 默认 apps/v1
	MinReplicas    int32
	MaxReplicas    int32
	CPUTargetUtil  *int32 // CPU 平均利用率 (百分比，例如 70 表示 70%)
	MemTargetUtil  *int32 // 内存平均利用率 (百分比)
	Labels         map[string]string
	Annotations    map[string]string
	ScaleUpStab    *int32 // 扩容稳定窗口秒数（可选）
	ScaleDownStab  *int32 // 缩容稳定窗口秒数（可选）
}

// List 获取 HPA 列表（支持命名空间过滤、名称模糊匹配）
func List(ctx context.Context, kube kubernetes.Interface, namespace, nameKeyword string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	res, err := kube.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if nameKeyword == "" {
		return res.Items, nil
	}
	filtered := make([]autoscalingv2.HorizontalPodAutoscaler, 0, len(res.Items))
	for _, item := range res.Items {
		if containsIgnoreCase(item.Name, nameKeyword) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// Get 获取 HPA 详情
func Get(ctx context.Context, kube kubernetes.Interface, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpa, err := kube.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("hpa %s/%s not found", namespace, name)
		}
		return nil, err
	}
	return hpa, nil
}

// Create 创建 HPA
func Create(ctx context.Context, kube kubernetes.Interface, opts *CreateOptions) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	hpa := buildHPA(opts)
	created, err := kube.AutoscalingV2().HorizontalPodAutoscalers(opts.Namespace).Create(ctx, hpa, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Update 更新 HPA（替换式更新，保留 resourceVersion）
func Update(ctx context.Context, kube kubernetes.Interface, opts *CreateOptions) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	existing, err := Get(ctx, kube, opts.Namespace, opts.Name)
	if err != nil {
		return nil, err
	}
	updated := buildHPA(opts)
	updated.ResourceVersion = existing.ResourceVersion
	res, err := kube.AutoscalingV2().HorizontalPodAutoscalers(opts.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Delete 删除 HPA
func Delete(ctx context.Context, kube kubernetes.Interface, namespace, name string) error {
	err := kube.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// PatchMinMax 仅修改 min/max 副本数（用于 618 等促销批量预扩容场景）
func PatchMinMax(ctx context.Context, kube kubernetes.Interface, namespace, name string, minReplicas, maxReplicas int32) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	patch := fmt.Sprintf(`{"spec":{"minReplicas":%d,"maxReplicas":%d}}`, minReplicas, maxReplicas)
	res, err := kube.AutoscalingV2().HorizontalPodAutoscalers(namespace).
		Patch(ctx, name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// buildHPA 根据选项构建 HPA 对象（基于 CPU/内存利用率指标）
func buildHPA(opts *CreateOptions) *autoscalingv2.HorizontalPodAutoscaler {
	apiVer := opts.TargetAPIVer
	if apiVer == "" {
		apiVer = "apps/v1"
	}
	metrics := make([]autoscalingv2.MetricSpec, 0, 2)
	if opts.CPUTargetUtil != nil && *opts.CPUTargetUtil > 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: "cpu",
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: opts.CPUTargetUtil,
				},
			},
		})
	}
	if opts.MemTargetUtil != nil && *opts.MemTargetUtil > 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: "memory",
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: opts.MemTargetUtil,
				},
			},
		})
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: apiVer,
				Kind:       opts.TargetKind,
				Name:       opts.TargetName,
			},
			MinReplicas: &opts.MinReplicas,
			MaxReplicas: opts.MaxReplicas,
			Metrics:     metrics,
		},
	}

	// 可选：扩缩容稳定窗口
	if opts.ScaleUpStab != nil || opts.ScaleDownStab != nil {
		behavior := &autoscalingv2.HorizontalPodAutoscalerBehavior{}
		if opts.ScaleUpStab != nil {
			behavior.ScaleUp = &autoscalingv2.HPAScalingRules{
				StabilizationWindowSeconds: opts.ScaleUpStab,
			}
		}
		if opts.ScaleDownStab != nil {
			behavior.ScaleDown = &autoscalingv2.HPAScalingRules{
				StabilizationWindowSeconds: opts.ScaleDownStab,
			}
		}
		hpa.Spec.Behavior = behavior
	}

	return hpa
}

func validateOptions(opts *CreateOptions) error {
	if opts == nil {
		return fmt.Errorf("opts is nil")
	}
	if opts.Namespace == "" || opts.Name == "" {
		return fmt.Errorf("namespace and name are required")
	}
	if opts.TargetKind == "" || opts.TargetName == "" {
		return fmt.Errorf("target kind and name are required")
	}
	if opts.MaxReplicas <= 0 {
		return fmt.Errorf("max_replicas must be > 0")
	}
	if opts.MinReplicas < 0 || opts.MinReplicas > opts.MaxReplicas {
		return fmt.Errorf("min_replicas must be in [0, max_replicas]")
	}
	if (opts.CPUTargetUtil == nil || *opts.CPUTargetUtil <= 0) && (opts.MemTargetUtil == nil || *opts.MemTargetUtil <= 0) {
		return fmt.Errorf("at least one of cpu/memory target utilization must be set")
	}
	return nil
}

func containsIgnoreCase(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(s) < len(sub) {
		return false
	}
	// 简化：直接小写比较
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

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ParseQuantity 安全解析资源量字符串（"100m"/"128Mi" 等）
func ParseQuantity(s string) (*resource.Quantity, error) {
	if s == "" {
		return nil, nil
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return nil, err
	}
	return &q, nil
}
