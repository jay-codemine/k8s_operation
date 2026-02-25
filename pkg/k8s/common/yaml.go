package common

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateResourceFromYAML 从 YAML 创建 Kubernetes 资源
// 返回创建的对象和错误
func CreateResourceFromYAML(ctx context.Context, clientset *kubernetes.Clientset, yamlContent string, targetObj runtime.Object) error {
	// 1. 解码 YAML 到 Unstructured
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	
	_, _, err := decUnstructured.Decode([]byte(yamlContent), nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	// 2. 转换为指定类型的对象
	objJSON, err := json.Marshal(obj.Object)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	if err := json.Unmarshal(objJSON, targetObj); err != nil {
		return fmt.Errorf("failed to unmarshal to target type: %w", err)
	}

	return nil
}

// ApplyResourceToK8s 将资源应用到 K8s 集群（使用 client-go）
func ApplyResourceToK8s(ctx context.Context, c client.Client, obj client.Object) error {
	// 尝试创建资源
	err := c.Create(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}
