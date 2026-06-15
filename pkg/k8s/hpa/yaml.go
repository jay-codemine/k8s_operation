// Package hpa - YAML 创建支持
package hpa

import (
	"context"
	"fmt"
	"strings"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// CreateFromYaml 通过原始 YAML 创建 HPA 资源
// 仅支持 autoscaling/v2 HorizontalPodAutoscaler 单资源
func CreateFromYaml(ctx context.Context, kube kubernetes.Interface, yamlContent string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	yamlContent = strings.TrimSpace(yamlContent)
	if yamlContent == "" {
		return nil, fmt.Errorf("yaml 内容不能为空")
	}

	// 1. 解析 YAML 为 HorizontalPodAutoscaler
	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	if err := yaml.Unmarshal([]byte(yamlContent), hpa); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 2. 校验 Kind / APIVersion
	if hpa.Kind != "" && hpa.Kind != "HorizontalPodAutoscaler" {
		return nil, fmt.Errorf("不支持的资源类型 %q，仅支持 HorizontalPodAutoscaler", hpa.Kind)
	}
	if hpa.APIVersion != "" && !strings.HasPrefix(hpa.APIVersion, "autoscaling/") {
		return nil, fmt.Errorf("不支持的 apiVersion %q，请使用 autoscaling/v2", hpa.APIVersion)
	}

	// 3. 校验必填字段
	if hpa.Name == "" {
		return nil, fmt.Errorf("metadata.name 不能为空")
	}
	if hpa.Namespace == "" {
		hpa.Namespace = "default"
	}
	if hpa.Spec.ScaleTargetRef.Kind == "" || hpa.Spec.ScaleTargetRef.Name == "" {
		return nil, fmt.Errorf("spec.scaleTargetRef.kind 和 spec.scaleTargetRef.name 必填")
	}
	if hpa.Spec.MaxReplicas <= 0 {
		return nil, fmt.Errorf("spec.maxReplicas 必须大于 0")
	}

	// 4. 清理只读字段，避免冲突
	hpa.ResourceVersion = ""
	hpa.UID = ""
	hpa.SelfLink = ""
	hpa.CreationTimestamp = metav1.Time{}
	hpa.Generation = 0
	hpa.ManagedFields = nil
	hpa.Status = autoscalingv2.HorizontalPodAutoscalerStatus{}

	// 5. 调用 typed client 创建
	created, err := kube.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Create(ctx, hpa, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("HPA %q 在命名空间 %q 中已存在", hpa.Name, hpa.Namespace)
		}
		return nil, err
	}
	return created, nil
}
