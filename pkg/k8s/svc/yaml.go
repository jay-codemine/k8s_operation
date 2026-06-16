package svc

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8soperation/global"
)

// ServiceYamlView 用于 YAML 序列化的结构体
type ServiceYamlView struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

// GetServiceYaml 获取 Service 的 YAML 表示
func GetServiceYaml(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (string, error) {
	service, err := Kube.CoreV1().
		Services(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 Service 失败: %w", err)
	}

	// 构建干净的 YAML 结构
	view := ServiceYamlView{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: map[string]interface{}{
			"name":      service.Name,
			"namespace": service.Namespace,
		},
		Spec: map[string]interface{}{
			"type":     string(service.Spec.Type),
			"selector": service.Spec.Selector,
		},
	}

	// 添加标签
	if len(service.Labels) > 0 {
		view.Metadata["labels"] = service.Labels
	}

	// 添加注解
	if len(service.Annotations) > 0 {
		view.Metadata["annotations"] = service.Annotations
	}

	// 添加 ClusterIP
	if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
		view.Spec["clusterIP"] = service.Spec.ClusterIP
	}

	// 转换端口
	if len(service.Spec.Ports) > 0 {
		ports := make([]map[string]interface{}, 0, len(service.Spec.Ports))
		for _, p := range service.Spec.Ports {
			port := map[string]interface{}{
				"port":       p.Port,
				"targetPort": p.TargetPort.IntValue(),
				"protocol":   string(p.Protocol),
			}
			if p.Name != "" {
				port["name"] = p.Name
			}
			if p.NodePort > 0 {
				port["nodePort"] = p.NodePort
			}
			ports = append(ports, port)
		}
		view.Spec["ports"] = ports
	}

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("转换 YAML 失败: %w", err)
	}

	return string(yamlBytes), nil
}

// ApplyServiceYaml 从 YAML 字符串创建或更新 Service
func ApplyServiceYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*corev1.Service, error) {
	// 解析 YAML
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(yamlContent), &service); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 检查必要字段
	if service.Name == "" {
		return nil, fmt.Errorf("YAML 中缺少 metadata.name")
	}
	if service.Namespace == "" {
		service.Namespace = "default"
	}

	// 尝试获取现有 Service
	existing, err := Kube.CoreV1().
		Services(service.Namespace).
		Get(ctx, service.Name, metav1.GetOptions{})

	if err != nil {
		// Service 不存在，创建新的
		created, createErr := Kube.CoreV1().
			Services(service.Namespace).
			Create(ctx, &service, metav1.CreateOptions{})
		if createErr != nil {
			return nil, fmt.Errorf("创建 Service 失败: %w", createErr)
		}

		global.Logger.Infof("Service [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		return created, nil
	}

	// Service 已存在，更新它
	// 继承 ResourceVersion 和 ClusterIP
	service.ResourceVersion = existing.ResourceVersion
	if service.Spec.ClusterIP == "" {
		service.Spec.ClusterIP = existing.Spec.ClusterIP
	}

	// 更新 Service
	updated, err := Kube.CoreV1().
		Services(service.Namespace).
		Update(ctx, &service, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Service 失败: %w", err)
	}

	global.Logger.Infof("Service [%s] 在命名空间 [%s] YAML 更新成功",
		updated.Name, updated.Namespace)

	return updated, nil
}

// CreateServiceFromYaml 从 YAML 创建 Service（支持多资源）
func CreateServiceFromYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*corev1.Service, error) {
	// 分割多文档 YAML
	docs := strings.Split(yamlContent, "---")

	var lastService *corev1.Service
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

		// 检查是否是 Service 类型
		if !strings.Contains(cleanedDoc, "kind:") || !strings.Contains(cleanedDoc, "Service") {
			global.Logger.Warnf("跳过非 Service 类型资源 (文档 %d)", i+1)
			continue
		}

		var service corev1.Service
		if err := yaml.Unmarshal([]byte(cleanedDoc), &service); err != nil {
			return nil, fmt.Errorf("解析文档 %d 失败: %w", i+1, err)
		}

		// 检查必要字段
		if service.Name == "" {
			return nil, fmt.Errorf("文档 %d 中缺少 metadata.name", i+1)
		}
		if service.Namespace == "" {
			service.Namespace = "default"
		}

		// 创建 Service
		created, err := Kube.CoreV1().
			Services(service.Namespace).
			Create(ctx, &service, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建 Service [%s] 失败: %w", service.Name, err)
		}

		global.Logger.Infof("Service [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		lastService = created
		createdCount++
	}

	if lastService == nil {
		return nil, fmt.Errorf("YAML 中没有找到有效的 Service 资源")
	}

	global.Logger.Infof("共创建 %d 个 Service", createdCount)
	return lastService, nil
}
