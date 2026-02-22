package common

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8soperation/internal/app/requests"
)

// ParseMultiYaml 将多文档 YAML 分割并解析为资源列表
func ParseMultiYaml(yamlContent string) ([]requests.ParsedResource, error) {
	if strings.TrimSpace(yamlContent) == "" {
		return nil, errors.New("YAML 内容不能为空")
	}

	// 按 --- 分割多个文档
	documents := splitYamlDocuments(yamlContent)
	var resources []requests.ParsedResource

	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "---" {
			continue
		}

		// 解析单个文档
		resource, err := parseSingleResource(doc, i)
		if err != nil {
			return nil, errors.Wrapf(err, "解析第 %d 个文档失败", i+1)
		}

		resources = append(resources, *resource)
	}

	return resources, nil
}

// splitYamlDocuments 按 --- 分割 YAML 文档
func splitYamlDocuments(content string) []string {
	lines := strings.Split(content, "\n")
	var documents []string
	var currentDoc []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if len(currentDoc) > 0 {
				documents = append(documents, strings.Join(currentDoc, "\n"))
				currentDoc = []string{}
			}
		} else {
			currentDoc = append(currentDoc, line)
		}
	}

	// 添加最后一个文档
	if len(currentDoc) > 0 {
		documents = append(documents, strings.Join(currentDoc, "\n"))
	}

	return documents
}

// parseSingleResource 解析单个资源文档
func parseSingleResource(yamlContent string, index int) (*requests.ParsedResource, error) {
	var obj map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return nil, errors.Wrap(err, "YAML 格式错误")
	}

	// 提取基础信息
	kind, ok := obj["kind"].(string)
	if !ok {
		return nil, errors.New("缺少 kind 字段")
	}

	apiVersion, ok := obj["apiVersion"].(string)
	if !ok {
		return nil, errors.New("缺少 apiVersion 字段")
	}

	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return nil, errors.New("缺少 metadata 字段")
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return nil, errors.New("metadata 中缺少 name 字段")
	}

	namespace, _ := metadata["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	// 计算创建顺序
	order := calculateCreateOrder(kind)

	// 分析依赖关系
	dependsOn := analyzeDependencies(obj, namespace)

	resource := &requests.ParsedResource{
		Index:      index,
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
		Content:    yamlContent,
		Order:      order,
		DependsOn:  dependsOn,
	}

	// 添加警告信息
	warnings := checkResourceWarnings(resource, obj)
	if len(warnings) > 0 {
		resource.Warnings = warnings
	}

	return resource, nil
}

// calculateCreateOrder 计算资源创建顺序
func calculateCreateOrder(kind string) int {
	// 创建顺序优先级（数值越小优先级越高）
	orderMap := map[string]int{
		"Namespace":        1,
		"StorageClass":     2,
		"PersistentVolume": 3,
		"ConfigMap":        4,
		"Secret":           4,
		"ServiceAccount":   5,
		"Role":             6,
		"RoleBinding":      6,
		"Service":          7,
		"Ingress":          8,
		"Deployment":       9,
		"StatefulSet":      9,
		"DaemonSet":        9,
		"Job":              10,
		"CronJob":          10,
	}

	if order, exists := orderMap[kind]; exists {
		return order
	}
	return 100 // 默认顺序
}

// analyzeDependencies 分析资源依赖关系
func analyzeDependencies(obj map[string]interface{}, defaultNamespace string) []requests.ResourceRef {
	var deps []requests.ResourceRef

	// 分析 spec 字段中的依赖
	if spec, ok := obj["spec"].(map[string]interface{}); ok {
		deps = append(deps, analyzeSpecDependencies(spec, defaultNamespace)...)
	}

	return deps
}

// analyzeSpecDependencies 分析 spec 中的依赖
func analyzeSpecDependencies(spec map[string]interface{}, defaultNamespace string) []requests.ResourceRef {
	var deps []requests.ResourceRef

	// 检查 volumes 中的 PVC 引用
	if volumes, ok := spec["volumes"].([]interface{}); ok {
		for _, vol := range volumes {
			if volMap, ok := vol.(map[string]interface{}); ok {
				if pvc := volMap["persistentVolumeClaim"]; pvc != nil {
					if pvcMap, ok := pvc.(map[string]interface{}); ok {
						if claimName, ok := pvcMap["claimName"].(string); ok {
							deps = append(deps, requests.ResourceRef{
								Kind:      "PersistentVolumeClaim",
								Name:      claimName,
								Namespace: defaultNamespace,
								FieldPath: "spec.volumes[].persistentVolumeClaim.claimName",
							})
						}
					}
				}
			}
		}
	}

	// 检查 envFrom 中的 ConfigMap/Secret 引用
	if containers, ok := spec["containers"].([]interface{}); ok {
		for _, container := range containers {
			if containerMap, ok := container.(map[string]interface{}); ok {
				if envFrom, ok := containerMap["envFrom"].([]interface{}); ok {
					for _, ef := range envFrom {
						if efMap, ok := ef.(map[string]interface{}); ok {
							// ConfigMap 引用
							if cmRef := efMap["configMapRef"]; cmRef != nil {
								if cmMap, ok := cmRef.(map[string]interface{}); ok {
									if name, ok := cmMap["name"].(string); ok {
										deps = append(deps, requests.ResourceRef{
											Kind:      "ConfigMap",
											Name:      name,
											Namespace: defaultNamespace,
											FieldPath: "spec.containers[].envFrom.configMapRef.name",
										})
									}
								}
							}
							// Secret 引用
							if secRef := efMap["secretRef"]; secRef != nil {
								if secMap, ok := secRef.(map[string]interface{}); ok {
									if name, ok := secMap["name"].(string); ok {
										deps = append(deps, requests.ResourceRef{
											Kind:      "Secret",
											Name:      name,
											Namespace: defaultNamespace,
											FieldPath: "spec.containers[].envFrom.secretRef.name",
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return deps
}

// checkResourceWarnings 检查资源配置警告
func checkResourceWarnings(resource *requests.ParsedResource, obj map[string]interface{}) []string {
	var warnings []string

	// 检查是否缺少重要字段
	switch resource.Kind {
	case "Deployment", "StatefulSet", "DaemonSet":
		if spec, ok := obj["spec"].(map[string]interface{}); ok {
			if _, hasSelector := spec["selector"]; !hasSelector {
				warnings = append(warnings, "缺少 selector 字段，可能导致无法正确管理 Pod")
			}
			if _, hasTemplate := spec["template"]; !hasTemplate {
				warnings = append(warnings, "缺少 template 字段，无法创建 Pod")
			}
		}
	case "Service":
		if spec, ok := obj["spec"].(map[string]interface{}); ok {
			if _, hasPorts := spec["ports"]; !hasPorts {
				warnings = append(warnings, "缺少 ports 字段，Service 将无法暴露端口")
			}
			if _, hasSelector := spec["selector"]; !hasSelector {
				warnings = append(warnings, "缺少 selector 字段，Service 将无法关联到任何 Pod")
			}
		}
	}

	// 检查命名空间是否为 default
	if resource.Namespace == "default" {
		warnings = append(warnings, "使用 default 命名空间，建议明确指定命名空间")
	}

	return warnings
}

// SortResourcesByOrder 按创建顺序排序资源
func SortResourcesByOrder(resources []requests.ParsedResource) []requests.ParsedResource {
	// 简单的冒泡排序
	for i := 0; i < len(resources)-1; i++ {
		for j := 0; j < len(resources)-1-i; j++ {
			if resources[j].Order > resources[j+1].Order {
				resources[j], resources[j+1] = resources[j+1], resources[j]
			}
		}
	}
	return resources
}

// ValidateResourceDependencies 验证资源依赖关系
func ValidateResourceDependencies(resources []requests.ParsedResource) []string {
	var errors []string
	resourceMap := make(map[string]*requests.ParsedResource)

	// 构建资源映射
	for i := range resources {
		key := fmt.Sprintf("%s/%s/%s", resources[i].Kind, resources[i].Namespace, resources[i].Name)
		resourceMap[key] = &resources[i]
	}

	// 检查每个资源的依赖
	for _, resource := range resources {
		for _, dep := range resource.DependsOn {
			key := fmt.Sprintf("%s/%s/%s", dep.Kind, dep.Namespace, dep.Name)
			if _, exists := resourceMap[key]; !exists {
				errors = append(errors, fmt.Sprintf(
					"资源 %s/%s/%s 依赖于不存在的资源 %s/%s/%s (%s)",
					resource.Kind, resource.Namespace, resource.Name,
					dep.Kind, dep.Namespace, dep.Name,
					dep.FieldPath,
				))
			}
		}
	}

	return errors
}