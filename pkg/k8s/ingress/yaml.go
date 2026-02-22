package ingress

import (
	"context"
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8soperation/global"
)

// IngressYamlView 用于 YAML 序列化的结构体
type IngressYamlView struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

// GetIngressYaml 获取 Ingress 的 YAML 表示
func GetIngressYaml(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (string, error) {
	ingress, err := Kube.NetworkingV1().
		Ingresses(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 Ingress 失败: %w", err)
	}

	// 构建干净的 YAML 结构
	view := IngressYamlView{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "Ingress",
		Metadata: map[string]interface{}{
			"name":      ingress.Name,
			"namespace": ingress.Namespace,
		},
		Spec: make(map[string]interface{}),
	}

	// 添加标签
	if len(ingress.Labels) > 0 {
		view.Metadata["labels"] = ingress.Labels
	}

	// 添加注解
	if len(ingress.Annotations) > 0 {
		view.Metadata["annotations"] = ingress.Annotations
	}

	// IngressClassName
	if ingress.Spec.IngressClassName != nil {
		view.Spec["ingressClassName"] = *ingress.Spec.IngressClassName
	}

	// 转换 Rules
	if len(ingress.Spec.Rules) > 0 {
		rules := make([]map[string]interface{}, 0, len(ingress.Spec.Rules))
		for _, rule := range ingress.Spec.Rules {
			r := make(map[string]interface{})
			if rule.Host != "" {
				r["host"] = rule.Host
			}
			if rule.HTTP != nil && len(rule.HTTP.Paths) > 0 {
				paths := make([]map[string]interface{}, 0, len(rule.HTTP.Paths))
				for _, p := range rule.HTTP.Paths {
					path := map[string]interface{}{
						"path":     p.Path,
						"pathType": string(*p.PathType),
						"backend": map[string]interface{}{
							"service": map[string]interface{}{
								"name": p.Backend.Service.Name,
								"port": map[string]interface{}{},
							},
						},
					}
					// 端口
					if p.Backend.Service.Port.Number > 0 {
						path["backend"].(map[string]interface{})["service"].(map[string]interface{})["port"].(map[string]interface{})["number"] = p.Backend.Service.Port.Number
					} else if p.Backend.Service.Port.Name != "" {
						path["backend"].(map[string]interface{})["service"].(map[string]interface{})["port"].(map[string]interface{})["name"] = p.Backend.Service.Port.Name
					}
					paths = append(paths, path)
				}
				r["http"] = map[string]interface{}{"paths": paths}
			}
			rules = append(rules, r)
		}
		view.Spec["rules"] = rules
	}

	// TLS
	if len(ingress.Spec.TLS) > 0 {
		tls := make([]map[string]interface{}, 0, len(ingress.Spec.TLS))
		for _, t := range ingress.Spec.TLS {
			tlsItem := map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			}
			tls = append(tls, tlsItem)
		}
		view.Spec["tls"] = tls
	}

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("转换 YAML 失败: %w", err)
	}

	return string(yamlBytes), nil
}

// ApplyIngressYaml 从 YAML 字符串创建或更新 Ingress
func ApplyIngressYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*networkingv1.Ingress, error) {
	// 解析 YAML
	var ingress networkingv1.Ingress
	if err := yaml.Unmarshal([]byte(yamlContent), &ingress); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 检查必要字段
	if ingress.Name == "" {
		return nil, fmt.Errorf("YAML 中缺少 metadata.name")
	}
	if ingress.Namespace == "" {
		ingress.Namespace = "default"
	}

	// 尝试获取现有 Ingress
	existing, err := Kube.NetworkingV1().
		Ingresses(ingress.Namespace).
		Get(ctx, ingress.Name, metav1.GetOptions{})

	if err != nil {
		// Ingress 不存在，创建新的
		created, createErr := Kube.NetworkingV1().
			Ingresses(ingress.Namespace).
			Create(ctx, &ingress, metav1.CreateOptions{})
		if createErr != nil {
			return nil, fmt.Errorf("创建 Ingress 失败: %w", createErr)
		}

		global.Logger.Infof("Ingress [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		return created, nil
	}

	// Ingress 已存在，更新它
	ingress.ResourceVersion = existing.ResourceVersion

	// 更新 Ingress
	updated, err := Kube.NetworkingV1().
		Ingresses(ingress.Namespace).
		Update(ctx, &ingress, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	global.Logger.Infof("Ingress [%s] 在命名空间 [%s] YAML 更新成功",
		updated.Name, updated.Namespace)

	return updated, nil
}

// CreateIngressFromYaml 从 YAML 创建 Ingress（支持多资源）
func CreateIngressFromYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*networkingv1.Ingress, error) {
	// 分割多文档 YAML
	docs := strings.Split(yamlContent, "---")

	var lastIngress *networkingv1.Ingress
	var createdCount int

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 移除开头的注释行（以 # 开头）
		lines := strings.Split(doc, "\n")
		var cleanedLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// 跳过空行和纯注释行
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			cleanedLines = append(cleanedLines, line)
		}

		if len(cleanedLines) == 0 {
			continue
		}

		cleanedDoc := strings.Join(cleanedLines, "\n")

		// 检查是否是 Ingress 类型
		if !strings.Contains(cleanedDoc, "kind:") || !strings.Contains(cleanedDoc, "Ingress") {
			global.Logger.Warnf("跳过非 Ingress 类型资源 (文档 %d)", i+1)
			continue
		}

		var ingress networkingv1.Ingress
		if err := yaml.Unmarshal([]byte(cleanedDoc), &ingress); err != nil {
			return nil, fmt.Errorf("解析文档 %d 失败: %w", i+1, err)
		}

		// 检查必要字段
		if ingress.Name == "" {
			return nil, fmt.Errorf("文档 %d 中缺少 metadata.name", i+1)
		}
		if ingress.Namespace == "" {
			ingress.Namespace = "default"
		}

		// 创建 Ingress
		created, err := Kube.NetworkingV1().
			Ingresses(ingress.Namespace).
			Create(ctx, &ingress, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建 Ingress [%s] 失败: %w", ingress.Name, err)
		}

		global.Logger.Infof("Ingress [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		lastIngress = created
		createdCount++
	}

	if lastIngress == nil {
		return nil, fmt.Errorf("YAML 中没有找到有效的 Ingress 资源")
	}

	global.Logger.Infof("共创建 %d 个 Ingress", createdCount)
	return lastIngress, nil
}
