package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8soperation/global"
)

// SecretYamlView 用于 YAML 序列化的结构体
type SecretYamlView struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Type       string            `json:"type,omitempty"`
	Data       map[string]string `json:"data,omitempty"`
	StringData map[string]string `json:"stringData,omitempty"`
}

// GetSecretYaml 获取 Secret 的 YAML 表示
func GetSecretYaml(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (string, error) {
	secret, err := Kube.CoreV1().
		Secrets(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 Secret 失败: %w", err)
	}

	// 构建干净的 YAML 结构
	view := SecretYamlView{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata: map[string]interface{}{
			"name":      secret.Name,
			"namespace": secret.Namespace,
		},
		Type: string(secret.Type),
	}

	// 添加标签
	if len(secret.Labels) > 0 {
		view.Metadata["labels"] = secret.Labels
	}

	// 添加注解
	if len(secret.Annotations) > 0 {
		view.Metadata["annotations"] = secret.Annotations
	}

	// 转换 data 字段（[]byte -> base64 string）
	if len(secret.Data) > 0 {
		view.Data = make(map[string]string)
		for k, v := range secret.Data {
			view.Data[k] = base64.StdEncoding.EncodeToString(v)
		}
	}

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("转换 YAML 失败: %w", err)
	}

	return string(yamlBytes), nil
}

// ApplySecretYaml 从 YAML 字符串创建或更新 Secret
func ApplySecretYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*corev1.Secret, error) {
	// 解析 YAML
	var secret corev1.Secret
	if err := yaml.Unmarshal([]byte(yamlContent), &secret); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 检查必要字段
	if secret.Name == "" {
		return nil, fmt.Errorf("YAML 中缺少 metadata.name")
	}
	if secret.Namespace == "" {
		secret.Namespace = "default"
	}

	// 尝试获取现有 Secret
	existing, err := Kube.CoreV1().
		Secrets(secret.Namespace).
		Get(ctx, secret.Name, metav1.GetOptions{})

	if err != nil {
		// Secret 不存在，创建新的
		created, createErr := Kube.CoreV1().
			Secrets(secret.Namespace).
			Create(ctx, &secret, metav1.CreateOptions{})
		if createErr != nil {
			return nil, fmt.Errorf("创建 Secret 失败: %w", createErr)
		}

		global.Logger.Infof("Secret [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		return created, nil
	}

	// Secret 已存在，更新它
	// 继承 ResourceVersion
	secret.ResourceVersion = existing.ResourceVersion

	// 更新 Secret
	updated, err := Kube.CoreV1().
		Secrets(secret.Namespace).
		Update(ctx, &secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Secret 失败: %w", err)
	}

	global.Logger.Infof("Secret [%s] 在命名空间 [%s] YAML 更新成功",
		updated.Name, updated.Namespace)

	return updated, nil
}

// SecretRaw 用于解析含 stringData 的 YAML
type SecretRaw struct {
	corev1.Secret `json:",inline"`
	StringData    map[string]string `json:"stringData,omitempty"`
}

// CreateSecretFromYaml 从 YAML 创建 Secret（支持多资源）
func CreateSecretFromYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*corev1.Secret, error) {
	// 分割多文档 YAML
	docs := strings.Split(yamlContent, "---")
	
	var lastSecret *corev1.Secret
	var createdCount int
	
	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" || strings.HasPrefix(doc, "#") {
			continue
		}
		
		// 检查是否是 Secret 类型
		if !strings.Contains(doc, "kind: Secret") {
			global.Logger.Warnf("跳过非 Secret 类型资源 (文档 %d)", i+1)
			continue
		}
		
		// 使用自定义结构体解析，支持 stringData
		var raw SecretRaw
		if err := yaml.Unmarshal([]byte(doc), &raw); err != nil {
			// 提供更友好的错误信息
			if strings.Contains(err.Error(), "illegal base64") {
				return nil, fmt.Errorf("解析文档 %d 失败: data 字段包含无效的 Base64 编码，请检查值是否完整", i+1)
			}
			return nil, fmt.Errorf("解析文档 %d 失败: %w", i+1, err)
		}
		
		secret := &raw.Secret
		
		// 处理 stringData: 转换为 base64 并合并到 data
		if len(raw.StringData) > 0 {
			if secret.Data == nil {
				secret.Data = make(map[string][]byte)
			}
			for k, v := range raw.StringData {
				secret.Data[k] = []byte(v)
			}
		}
		
		// 检查必要字段
		if secret.Name == "" {
			return nil, fmt.Errorf("文档 %d 中缺少 metadata.name", i+1)
		}
		if secret.Namespace == "" {
			secret.Namespace = "default"
		}
		
		// 创建 Secret
		created, err := Kube.CoreV1().
			Secrets(secret.Namespace).
			Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建 Secret [%s] 失败: %w", secret.Name, err)
		}
		
		global.Logger.Infof("Secret [%s] 在命名空间 [%s] 创建成功",
			created.Name, created.Namespace)
		lastSecret = created
		createdCount++
	}
	
	if lastSecret == nil {
		return nil, fmt.Errorf("YAML 中没有找到有效的 Secret 资源")
	}
	
	global.Logger.Infof("共创建 %d 个 Secret", createdCount)
	return lastSecret, nil
}
